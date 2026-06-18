package client

import (
	"testing"
	"time"

	"stunnel/internal/server"
)

func TestClientConnect(t *testing.T) {
	client := New("localhost:7000", "testsecret", "localhost:3000")
	if client.serverAddr != "localhost:7000" {
		t.Error("wrong server addr")
	}
	if client.localAddr != "localhost:3000" {
		t.Error("wrong local addr")
	}
}

func TestClientAuthHandshake(t *testing.T) {
	srv := server.New("127.0.0.1:0", "testsecret")
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr().String()
	client := New(addr, "testsecret", "localhost:3000")

	go client.Connect()

	deadline := time.After(2 * time.Second)
	for {
		if client.TunnelID() != "" {
			break
		}
		select {
		case <-deadline:
			t.Fatal("auth did not complete in time")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestClientAuthWrongSecret(t *testing.T) {
	srv := server.New("127.0.0.1:0", "testsecret")
	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr().String()
	client := New(addr, "wrongsecret", "localhost:3000")

	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Connect()
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error for wrong secret")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Connect timed out")
	}
}
