package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	for _, test := range []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{name: "omitted", config: Config{}},
		{name: "loopback", config: Config{Enabled: true, ListenAddr: "127.0.0.1:19092"}},
		{name: "private", config: Config{Enabled: true, ListenAddr: "10.0.0.8:19092"}},
		{name: "missing address", config: Config{Enabled: true}, wantErr: true},
		{name: "public", config: Config{Enabled: true, ListenAddr: "8.8.8.8:19092"}, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateConfig(test.config); (err != nil) != test.wantErr {
				t.Fatalf("ValidateConfig() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestHealthHandlers(t *testing.T) {
	dependencyReady := true
	server, err := New(Config{}, func(context.Context) error {
		if !dependencyReady {
			return errors.New("dependency unavailable")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	assertStatus(t, server.server.Handler, "/livez", http.StatusOK)
	assertStatus(t, server.server.Handler, "/readyz", http.StatusServiceUnavailable)
	server.MarkReady()
	assertStatus(t, server.server.Handler, "/readyz", http.StatusOK)
	dependencyReady = false
	assertStatus(t, server.server.Handler, "/readyz", http.StatusServiceUnavailable)
	dependencyReady = true
	assertStatus(t, server.server.Handler, "/readyz", http.StatusOK)
	server.MarkNotReady()
	assertStatus(t, server.server.Handler, "/readyz", http.StatusServiceUnavailable)
}

func TestDisabledServerDoesNotListen(t *testing.T) {
	server, err := New(Config{}, func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if err = server.Start(); err != nil {
		t.Fatal(err)
	}
	if server.started.Load() {
		t.Fatal("disabled health server started a listener")
	}
}

func TestEnabledServerShutsDown(t *testing.T) {
	server, err := New(Config{Enabled: true, ListenAddr: "127.0.0.1:19092"}, func(context.Context) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	server.config.ListenAddr = "127.0.0.1:0"
	if err = server.Start(); err != nil {
		t.Fatal(err)
	}
	server.MarkReady()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if server.ready.Load() {
		t.Fatal("server remained ready during shutdown")
	}
}

func assertStatus(t *testing.T, handler http.Handler, path string, want int) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("GET %s status = %d, want %d; body=%q", path, response.Code, want, response.Body.String())
	}
}
