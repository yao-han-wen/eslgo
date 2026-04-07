package eslgo

import (
	"net"
)

func NewInboundSocket(addr string, options ...Option) (*Connection, error) {
	var err error

	config := DefaultConfig()
	// 应用选项
	for _, opt := range options {
		opt(config)
	}

	conn, err := net.DialTimeout("tcp", addr, config.connectTimeOut)
	if err != nil {
		return nil, err
	}

	c := newConnection(conn, config)

	// 获取密码请求
	resp, err := c.recvOne()
	if err != nil {
		c.close(err)
		return nil, err
	}
	if resp.Mime["Content-Type"] != CONTENT_TYPE_AUTH_REQUEST {
		c.close(ErrNoResponseToAuthRequest)
		return nil, ErrNoResponseToAuthRequest
	}

	c.startRecvLoop()

	// 检验密码
	err = c.SendAuthCommand(c.config.connectPassword)
	if err != nil {
		c.close(err)
		return nil, err
	}

	return c, nil
}
