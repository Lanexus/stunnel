package protocol

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeAuth(t *testing.T) {
	msg := Message{
		Type: MsgAuth,
		Data: AuthData{Secret: "test123"},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != MsgAuth {
		t.Errorf("expected MsgAuth, got %v", decoded.Type)
	}
}

func TestEncodeDecodeAuthOK(t *testing.T) {
	msg := Message{
		Type: MsgAuthOK,
		Data: AuthOKData{TunnelID: "abc123", PublicPort: 8080},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != MsgAuthOK {
		t.Errorf("expected MsgAuthOK, got %v", decoded.Type)
	}
}

func TestEncodeDecodeNewTunnel(t *testing.T) {
	msg := Message{
		Type: MsgNewTunnel,
		Data: NewTunnelData{TunnelID: "abc123", ConnID: "conn1"},
	}
	var buf bytes.Buffer
	if err := Encode(&buf, msg); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Type != MsgNewTunnel {
		t.Errorf("expected MsgNewTunnel, got %v", decoded.Type)
	}
}
