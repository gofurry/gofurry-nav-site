package health

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const probeTimeout = 3 * time.Second

type Config struct {
	Enabled    bool   `yaml:"enabled"`
	ListenAddr string `yaml:"listen_addr"`
}

type ReadyCheck func(context.Context) error

type Server struct {
	config     Config
	readyCheck ReadyCheck
	server     *http.Server
	errors     chan error
	ready      atomic.Bool
	started    atomic.Bool
	stopOnce   sync.Once
}

func ValidateConfig(config Config) error {
	if !config.Enabled {
		return nil
	}
	address := strings.TrimSpace(config.ListenAddr)
	if address == "" {
		return errors.New("health.listen_addr is required when health is enabled")
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("health.listen_addr must be a valid host:port: %w", err)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return errors.New("health.listen_addr must contain a valid non-zero port")
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || (!ip.IsLoopback() && !ip.IsPrivate()) {
		return errors.New("health.listen_addr must bind a loopback or private IP address")
	}
	return nil
}

func New(config Config, readyCheck ReadyCheck) (*Server, error) {
	config.ListenAddr = strings.TrimSpace(config.ListenAddr)
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	server := &Server{
		config:     config,
		readyCheck: readyCheck,
		errors:     make(chan error, 1),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", server.handleLive)
	mux.HandleFunc("GET /readyz", server.handleReady)
	server.server = &http.Server{
		Addr:              config.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return server, nil
}

func (s *Server) Start() error {
	if !s.config.Enabled {
		return nil
	}
	listener, err := net.Listen("tcp", s.config.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen on collector health address %s: %w", s.config.ListenAddr, err)
	}
	s.started.Store(true)
	go func() {
		if serveErr := s.server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			select {
			case s.errors <- fmt.Errorf("serve collector health endpoint: %w", serveErr):
			default:
			}
		}
	}()
	return nil
}

func (s *Server) Errors() <-chan error { return s.errors }

func (s *Server) MarkReady()    { s.ready.Store(true) }
func (s *Server) MarkNotReady() { s.ready.Store(false) }

func (s *Server) Shutdown(ctx context.Context) error {
	s.MarkNotReady()
	if !s.started.Load() {
		return nil
	}
	var err error
	s.stopOnce.Do(func() {
		err = s.server.Shutdown(ctx)
	})
	return err
}

func (s *Server) handleLive(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok\n"))
}

func (s *Server) handleReady(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if !s.ready.Load() {
		http.Error(writer, "not ready", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(request.Context(), probeTimeout)
	defer cancel()
	if s.readyCheck == nil || s.readyCheck(ctx) != nil {
		http.Error(writer, "not ready", http.StatusServiceUnavailable)
		return
	}
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ready\n"))
}
