package relay

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type ConnectClient struct {
	relayAddr string
	secret    string
}

func NewConnectClient(relayAddr, secret string) *ConnectClient {
	return &ConnectClient{
		relayAddr: relayAddr,
		secret:    secret,
	}
}

func (c *ConnectClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.relayAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to relay: %w", err)
	}

	// Request connection
	msg := Message{
		Type:   MsgRequest,
		Secret: c.secret,
	}
	data, _ := json.Marshal(msg)
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		conn.Close()
		return fmt.Errorf("send request: %w", err)
	}

	// Read response (should be MATCHED)
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

	log.Printf("matched to tunnel %s", resp.TunnelID)
	log.Printf("connected! pipe stdin/stdout...")

	// Pipe stdin to connection and connection to stdout
	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	io.Copy(conn, os.Stdin)
	conn.Close()
	<-done

	return nil
}
