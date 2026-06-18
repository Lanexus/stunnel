package main

import (
	"net"
	"testing"
	"time"

	"stunnel/internal/client"
	"stunnel/internal/server"
)

func TestEndToEnd(t *testing.T) {
	localLn, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer localLn.Close()

	go func() {
		conn, err := localLn.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("hello from local"))
	}()

	srv := server.New(":0", ":0", "testsecret")
	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Stop()

	c := client.New(srv.Addr().String(), "testsecret", localLn.Addr().String())
	go c.Connect()

	deadline := time.After(2 * time.Second)
	for {
		if c.TunnelID() != "" {
			break
		}
		select {
		case <-deadline:
			t.Error("expected tunnel ID to be set")
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	pubAddr := srv.PublicAddr()
	if pubAddr == nil {
		t.Fatal("expected public address to be set")
	}

	userConn, err := net.DialTimeout("tcp", pubAddr.String(), 2*time.Second)
	if err != nil {
		t.Fatalf("connect to public port: %v", err)
	}
	defer userConn.Close()

	buf := make([]byte, 1024)
	n, err := userConn.Read(buf)
	if err != nil {
		t.Fatalf("read from public port: %v", err)
	}
	if string(buf[:n]) != "hello from local" {
		t.Errorf("expected 'hello from local', got %q", string(buf[:n]))
	}
}
