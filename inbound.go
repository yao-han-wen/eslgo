package eslgo

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InboundSocket struct {
	config     *Config
	conn       net.Conn
	reader     *bufio.Reader
	mimeReader *textproto.Reader

	cmdMux      sync.Mutex
	cmdChan     chan *Response
	eventChan   chan *Event
	closeNotify chan error
	closeOnce   sync.Once
}

func NewInboundSocket(addr, passwd string, options ...Option) (*InboundSocket, error) {
	config := &Config{
		connectTimeOut: OPT_CONNECT_TIMEOUT * time.Second,
		commandTimeOut: OPT_COMMAND_TIMEOUT * time.Second,
		eventChanCap:   OPT_EVENT_CHANNEL_CAPACITY,
	}
	// 应用选项
	for _, opt := range options {
		opt(config)
	}

	conn, err := net.DialTimeout("tcp", addr, config.connectTimeOut)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	mimeReader := textproto.NewReader(reader)

	// 获取密码请求
	mime, err := mimeReader.ReadMIMEHeader()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if mime.Get("Content-Type") != CONTENT_TYPE_AUTH_REQUEST {
		conn.Close()
		return nil, ErrNoResponseToAuthRequest
	}

	// 检验密码
	fmt.Fprintf(conn, "auth %s\r\n\r\n", passwd)
	mime, err = mimeReader.ReadMIMEHeader()
	if err != nil {
		conn.Close()
		return nil, err
	}
	if mime.Get("Reply-Text") != "+OK accepted" {
		conn.Close()
		return nil, ErrInvalidPassword
	}

	is := &InboundSocket{
		config:     config,
		conn:       conn,
		reader:     reader,
		mimeReader: mimeReader,
	}
	is.cmdChan = make(chan *Response)
	is.eventChan = make(chan *Event, is.config.eventChanCap)
	is.closeNotify = make(chan error, 1)

	go is.recvLoop()

	return is, nil
}

func (is *InboundSocket) Close() {
	is.close(nil)
}

func (is *InboundSocket) close(err error) {
	is.closeOnce.Do(
		func() {
			is.conn.Close()
			is.closeNotify <- err
		})
}

func (is *InboundSocket) CloseNotify() <-chan error {
	return is.closeNotify
}

func (is *InboundSocket) recvLoop() {
	var err error
	for {
		err = is.recv()
		if err != nil {
			break
		}
	}

	//出错直接关闭连接
	is.close(err)

	close(is.cmdChan)
	close(is.eventChan)
	close(is.closeNotify)
}

func (is *InboundSocket) recv() error {
	var err error
	resp := &Response{}
	resp.Mime, err = is.mimeReader.ReadMIMEHeader()
	if err != nil {
		return err
	}
	// log.Println("mime:", resp.Mime)

	if v := resp.Mime.Get("Content-Length"); v != "" {
		length, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		resp.Content = make([]byte, length)
		if _, err := io.ReadFull(is.reader, resp.Content); err != nil {
			return err
		}
	} else {
		resp.Content = make([]byte, 0)
	}
	// log.Println("content:", string(resp.Content))

	switch resp.Mime.Get("Content-Type") {
	case CONTENT_TYPE_COMMAND_REPLY:
		is.cmdChan <- resp

	case CONTENT_TYPE_API_RESPONSE:
		is.cmdChan <- resp

	case CONTENT_TYPE_TEXT_EVENT_PLAIN, CONTENT_TYPE_TEXT_EVENT_XML, CONTENT_TYPE_TEXT_EVENT_JSON:
		e, err := resp.ToEvent()
		if err != nil {
			return err
		}
		is.eventChan <- e
	}

	return nil
}

func (is *InboundSocket) SendCommand(cmd string) (*Response, error) {
	is.cmdMux.Lock()
	defer is.cmdMux.Unlock()

	_, err := fmt.Fprintf(is.conn, "%s\r\n\r\n", cmd)
	if err != nil {
		return nil, err
	}

	timer := time.NewTimer(is.config.commandTimeOut)
	defer timer.Stop()

	select {
	case resp, ok := <-is.cmdChan:
		if !ok {
			return nil, ErrConnectClosed
		}
		return resp, nil
	case <-timer.C:
		//命令异常会导致后续指令获取错误，这里直接关闭连接
		is.close(ErrCommandTimeout)
		return nil, ErrCommandTimeout
	}
}

// 订阅event消息
func (is *InboundSocket) SendEventCommand(cmd string) (<-chan *Event, error) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "event") {
		cmd = "event " + cmd
	}

	resp, err := is.SendCommand(cmd)
	if err != nil {
		return nil, err
	}
	err = resp.HasError()
	if err != nil {
		return nil, err
	}

	return is.eventChan, nil
}

// 发送同步api指令
func (is *InboundSocket) SendApiCommand(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "api") {
		cmd = "api " + cmd
	}

	resp, err := is.SendCommand(cmd)
	if err != nil {
		return "", err
	}
	err = resp.HasError()
	if err != nil {
		return "", err
	}

	return string(resp.Content), nil
}

// 发送异步bgapi指令
func (is *InboundSocket) SendBgApiCommand(cmd string) (JobUUID, error) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "bgapi") {
		cmd = "bgapi " + cmd
	}

	resp, err := is.SendCommand(cmd)
	if err != nil {
		return "", err
	}
	err = resp.HasError()
	if err != nil {
		return "", err
	}

	return JobUUID(resp.Mime.Get("Job-Uuid")), nil
}
