package eslgo

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Connection struct {
	config *Config
	conn   net.Conn
	reader *bufio.Reader

	closeChan chan error
	closeOnce sync.Once

	cmdMux    sync.Mutex
	cmdChan   chan *Response
	eventChan chan *Response
}

func newConnection(conn net.Conn, config *Config) *Connection {
	c := &Connection{
		config: config,
		conn:   conn,
	}

	c.reader = bufio.NewReader(c.conn)
	c.closeChan = make(chan error, 1)

	return c
}

func (c *Connection) startRecvLoop() {
	c.cmdChan = make(chan *Response)
	c.eventChan = make(chan *Response, int(c.config.eventChanCap))
	go c.recvLoop()
}

func (c *Connection) Close() {
	c.close(nil)
}

func (c *Connection) close(err error) {
	c.closeOnce.Do(
		func() {
			c.conn.Close()
			c.closeChan <- err
			close(c.closeChan)
		})
}

func (c *Connection) CloseNotify() <-chan error {
	return c.closeChan
}

func (c *Connection) recvLoop() {
	defer close(c.cmdChan)
	defer close(c.eventChan)

	for {
		resp, err := c.recvOne()
		if err != nil {
			c.close(err)
			return
		}

		switch resp.Mime["Content-Type"] {
		case CONTENT_TYPE_COMMAND_REPLY, CONTENT_TYPE_API_RESPONSE:
			c.cmdChan <- resp

		case CONTENT_TYPE_TEXT_EVENT_PLAIN, CONTENT_TYPE_TEXT_EVENT_XML, CONTENT_TYPE_TEXT_EVENT_JSON:
			c.eventChan <- resp

		case CONTENT_TYPE_TEXT_DISCONNECT_NOTICE:
			c.close(ErrDisconnectNotice)
			return
		}
	}
}

func (c *Connection) readMIME() (map[string]string, error) {
	mime := make(map[string]string)

	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" || line == "\n" {
			break
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("readMIME error, invalid MIME header line: %q", line)
		}

		mime[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return mime, nil
}

func (c *Connection) recvOne() (*Response, error) {
	var err error
	resp := &Response{}
	resp.Mime, err = c.readMIME()
	if err != nil {
		return nil, err
	}

	if v := resp.Mime["Content-Length"]; v != "" {
		length, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		resp.Content = make([]byte, length)
		if _, err := io.ReadFull(c.reader, resp.Content); err != nil {
			return nil, err
		}
	} else {
		resp.Content = make([]byte, 0)
	}

	return resp, nil
}

func (c *Connection) GetEventChan() <-chan *Response {
	return c.eventChan
}

func (c *Connection) SendCommand(cmd string) (*Response, error) {
	c.cmdMux.Lock()
	defer c.cmdMux.Unlock()

	_, err := fmt.Fprintf(c.conn, "%s%s", cmd, MSG_END)
	if err != nil {
		return nil, err
	}

	timer := time.NewTimer(c.config.commandTimeOut)
	defer timer.Stop()

	select {
	case resp, ok := <-c.cmdChan:
		if !ok {
			return nil, ErrConnectClosed
		}
		return resp, nil
	case <-timer.C:
		return nil, ErrCommandTimeout
	}
}

func (c *Connection) SendAuthCommand(password string) error {
	resp, err := c.SendCommand("auth " + password)
	if err != nil {
		return err
	}
	err = resp.HasError()
	if err != nil {
		return err
	}

	return nil
}

func (c *Connection) SendConnectCommand() (*Response, error) {
	resp, err := c.SendCommand("connect")
	if err != nil {
		return nil, err
	}
	err = resp.HasError()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 订阅event消息
func (c *Connection) SendEventCommand(cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "event") {
		cmd = "event " + cmd
	}

	resp, err := c.SendCommand(cmd)
	if err != nil {
		return err
	}
	err = resp.HasError()
	if err != nil {
		return err
	}

	return nil
}

// 发送同步api指令
func (c *Connection) SendApiCommand(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "api") {
		cmd = "api " + cmd
	}

	resp, err := c.SendCommand(cmd)
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
func (c *Connection) SendBgApiCommand(cmd string) (JobUUID, error) {
	cmd = strings.TrimSpace(cmd)
	if !strings.HasPrefix(cmd, "bgapi") {
		cmd = "bgapi " + cmd
	}

	resp, err := c.SendCommand(cmd)
	if err != nil {
		return "", err
	}
	err = resp.HasError()
	if err != nil {
		return "", err
	}

	return JobUUID(resp.Mime["Job-UUID"]), nil
}
