package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

const (
	defaultHost           = "localhost"
	defaultPort           = 6379
	defaultMaxConnections = 256
)

var ErrStoppedAbnormally = errors.New("server stopped abnormally")

type Handler interface {
	Handle(ctx context.Context, req io.Reader, resp io.Writer) error
}

type HandlerFunc func(ctx context.Context, req io.Reader, resp io.Writer) error

func (f HandlerFunc) Handle(ctx context.Context, req io.Reader, resp io.Writer) error {
	return f(ctx, req, resp)
}

type Server struct {
	Host           string
	Port           int
	Logger         *slog.Logger
	MaxConnections int
	Handler        Handler
	running        bool
	connections    chan net.Conn
	workers        sync.WaitGroup
	softDone       chan struct{}
	hardDone       chan struct{}
	listener       net.Listener
}

func Default(handler Handler) *Server {
	return &Server{
		Host:           defaultHost,
		Port:           defaultPort,
		Logger:         slog.Default(),
		MaxConnections: defaultMaxConnections,
		Handler:        handler,
		running:        false,
		connections:    nil,
		workers:        sync.WaitGroup{},
		softDone:       make(chan struct{}),
		hardDone:       make(chan struct{}),
		listener:       nil,
	}
}

func (s *Server) setRunning(val bool) {
	s.running = val
}

func (s *Server) Run() error {
	if s.running {
		panic("server is already running")
	}

	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	s.setRunning(true)
	defer s.setRunning(false)

	s.connections = make(chan net.Conn, s.MaxConnections)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = lis

	s.createWorkers()

	s.Logger.Info("Server has started", "addr", addr)
	for {
		select {
		case <-s.softDone:
			return nil
		default:
			if err := s.accept(); err != nil && !errors.Is(err, net.ErrClosed) {
				s.Logger.Info("failed to accept connection", "error", err)
			}
		}
	}
}

func (s *Server) Stop(timeout time.Duration) error {
	if !s.running {
		return nil
	}
	close(s.softDone)

	done := make(chan struct{})
	go func() {
		s.workers.Wait()
		close(done)
	}()

	var resultErr error
	select {
	case <-done:
		resultErr = nil
	case <-time.After(timeout):
		close(s.hardDone)
		resultErr = ErrStoppedAbnormally
	}

	if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		s.Logger.Error("failed to close listener", "error", err)
	}
	s.workers.Wait()
	return resultErr
}

func (s *Server) accept() error {
	for {
		select {
		case <-s.softDone:
			close(s.connections)
			return nil
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				return err
			}
			s.Logger.Debug("accepted a new connection ", "addr", conn.RemoteAddr())
			s.connections <- conn
			return nil
		}
	}
}

func (s *Server) createWorkers() {
	s.workers.Add(s.MaxConnections)
	for i := 0; i < s.MaxConnections; i++ {
		go s.worker()
	}
}

func (s *Server) worker() {
	defer s.workers.Done()
	for {
		select {
		case conn, ok := <-s.connections:
			if !ok {
				return
			}

			s.processConnection(conn)

			if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				s.Logger.Warn("failed properly to close a connection", "error", err)
			}

		case <-s.softDone:
			return
		}
	}
}

func (s *Server) processConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("recovered from panic", "error", r)
			s.Logger.Debug(string(debug.Stack()))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)
		var err error
		if err = s.Handler.Handle(ctx, conn, conn); err != nil {
			s.Logger.Error("failed to handle connection", "error", err)
		}
	}()

	select {
	case <-s.hardDone:
		cancel()
	case <-done:
		return
	}
}
