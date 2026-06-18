package client

import (
	"fmt"
	"log"
	"net"
	"stunnel/internal/protocol"
	"time"
)

type Client struct {
	serverAddr string
	secret     string
	localAddr  string
	tunnelID   string
}

func New(serverAddr, secret, localAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		secret:     secret,
		localAddr:  localAddr,
	}
}

func (c *Client) TunnelID() string {
	return c.tunnelID
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer conn.Close()

	err = protocol.Encode(conn, protocol.Message{
		Type: protocol.MsgAuth,
		Data: protocol.AuthData{
			Secret:   c.secret,
			TunnelID: c.tunnelID,
		},
	})
	if err != nil {
		return fmt.Errorf("send auth: %w", err)
	}

	msg, err := protocol.Decode(conn)
	if err != nil {
		return fmt.Errorf("read auth response: %w", err)
	}

	if msg.Type != protocol.MsgAuthOK {
		return fmt.Errorf("auth failed: %v", msg.Type)
	}

	data := msg.Data.(protocol.AuthOKData)
	c.tunnelID = data.TunnelID
	log.Printf("connected, tunnel: %s, public port: %d", data.TunnelID, data.PublicPort)

	return c.handleMessages(conn)
}

func (c *Client) handleMessages(conn net.Conn) error {
	for {
		msg, err := protocol.Decode(conn)
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		switch msg.Type {
		case protocol.MsgNewTunnel:
			data := msg.Data.(protocol.NewTunnelData)
			go c.handleTunnel(data.TunnelID, data.ConnID)
		}
	}
}

func (c *Client) handleTunnel(tunnelID, connID string) {
	log.Printf("tunnel %s conn %s -> local %s", tunnelID, connID, c.localAddr)
}
