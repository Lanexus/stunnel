package relay

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestRelayStartStop(t *testing.T) {
	ln, _ := net.Listen("tcp", ":0")
	addr := ln.Addr().String()
	ln.Close()

	r := New(addr)
	go r.Start()
	time.Sleep(100 * time.Millisecond)
	r.Stop()
}

func TestRelayRegisterAndRequest(t *testing.T) {
	ln, _ := net.Listen("tcp", ":0")
	addr := ln.Addr().String()
	ln.Close()

	r := New(addr)
	go r.Start()
	time.Sleep(100 * time.Millisecond)
	defer r.Stop()

	// Simulate serve client
	serveConn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer serveConn.Close()

	// Register
	regMsg := Message{Type: MsgRegister, Secret: "test123"}
	data, _ := json.Marshal(regMsg)
	data = append(data, '\n')
	serveConn.Write(data)

	// Read tunnel ID
	dec := json.NewDecoder(serveConn)
	var resp Message
	if err := dec.Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Type != MsgMatched {
		t.Errorf("expected MATCHED, got %v", resp.Type)
	}
	if resp.TunnelID == "" {
		t.Error("expected non-empty tunnel ID")
	}
}
