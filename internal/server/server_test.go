package server

import (
	"net"
	"testing"
	"time"
)

func TestServerStartStop(t *testing.T) {
	srv := New(":0", "testsecret")
	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	srv.Stop()
}

func TestServerAssignsTunnelID(t *testing.T) {
	srv := New(":0", "testsecret")
	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Stop()

	addr := srv.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	tunnelID := srv.AssignTunnelID(conn)
	if tunnelID == "" {
		t.Error("expected non-empty tunnel ID")
	}
}
