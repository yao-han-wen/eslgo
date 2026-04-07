package eslgo

import (
	"context"
	"log"
	"net"
	"sync"
)

type OutboundHandler func(ctx context.Context, conn *Connection)

type OutboundServer struct {
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc

	handler OutboundHandler
	options []Option

	mu    sync.Mutex
	conns map[*Connection]struct{}
	wg    sync.WaitGroup
}

func NewOutboundServer(addr string, handler OutboundHandler, options ...Option) (*OutboundServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &OutboundServer{
		listener: ln,
		ctx:      ctx,
		cancel:   cancel,
		handler:  handler,
		options:  options,
		conns:    make(map[*Connection]struct{}),
	}, nil
}

func (s *OutboundServer) Serve() error {
	for {
		rawConn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				return err
			}
		}

		conn, err := NewOutboundSocket(rawConn, s.options...)
		if err != nil {
			log.Println("NewOutboundSocket err:", err)
			rawConn.Close()
			continue
		}

		s.mu.Lock()
		s.conns[conn] = struct{}{}
		s.mu.Unlock()

		s.wg.Go(func() {
			defer func() {
				s.mu.Lock()
				delete(s.conns, conn)
				s.mu.Unlock()
			}()

			s.handler(s.ctx, conn)
		})
	}
}

func (s *OutboundServer) Shutdown(ctx context.Context) error {
	s.cancel()
	_ = s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
	}

	s.mu.Lock()
	for conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func NewOutboundSocket(conn net.Conn, options ...Option) (*Connection, error) {
	config := DefaultConfig()
	// 应用选项
	for _, opt := range options {
		opt(config)
	}

	c := newConnection(conn, config)

	c.startRecvLoop()

	return c, nil
}
