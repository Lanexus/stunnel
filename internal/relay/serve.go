package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type ServeClient struct {
	relayAddr string
	secret    string
	localAddr string
}

func NewServeClient(relayAddr, secret, localAddr string) *ServeClient {
	return &ServeClient{
		relayAddr: relayAddr,
		secret:    secret,
		localAddr: localAddr,
	}
}

func (s *ServeClient) Connect() error {
	conn, err := net.DialTimeout("tcp", s.relayAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to relay: %w", err)
	}

	// Register with relay
	msg := Message{
		Type:   MsgRegister,
		Secret: s.secret,
	}
	data, _ := json.Marshal(msg)
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return fmt.Errorf("send register: %w", err)
	}

	// Read response (should be MATCHED with tunnel ID)
	dec := json.NewDecoder(conn)
	var resp Message
	if err := dec.Decode(&resp); err != nil {
		conn.Close()
		return fmt.Errorf("read response: %w", err)
	}

	if resp.Type != MsgMatched {
		conn.Close()
		return fmt.Errorf("unexpected response: %v", resp.Type)
	}

	log.Printf("registered with relay, tunnel: %s", resp.TunnelID)
	log.Printf("waiting for connections...")

	// Wait for READY message (when a connect client arrives)
	var ready Message
	if err := dec.Decode(&ready); err != nil {
		conn.Close()
		return fmt.Errorf("read ready: %w", err)
	}

	if ready.Type != MsgReady {
		conn.Close()
		return fmt.Errorf("unexpected message: %v", ready.Type)
	}

	log.Printf("client connected, bridging to %s", s.localAddr)

	// Connect to local service
	localConn, err := net.DialTimeout("tcp", s.localAddr, 5*time.Second)
	if err != nil {
		conn.Close()
		return fmt.Errorf("connect local: %w", err)
	}

	// Bridge connections
	done := make(chan struct{})
	go func() {
		io.Copy(localConn, conn)
		close(done)
	}()
	io.Copy(conn, localConn)
	conn.Close()
	localConn.Close()
	<-done

	return nil
}
