package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"stunnel/internal/protocol"
)

type Client struct {
	ID   string
	Conn net.Conn
}

type Server struct {
	addr      string
	secret    string
	listener  net.Listener
	clients   map[string]*Client
	mu        sync.RWMutex
	tunnelIDs map[string]string
	done      chan struct{}
}

func New(addr, secret string) *Server {
	return &Server{
		addr:      addr,
		secret:    secret,
		clients:   make(map[string]*Client),
		tunnelIDs: make(map[string]string),
		done:      make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = ln
	log.Printf("server listening on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				log.Printf("accept error: %v", err)
				return fmt.Errorf("accept: %w", err)
			}
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() {
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) AssignTunnelID(conn net.Conn) string {
	id := generateID()
	clientID := generateID()
	s.mu.Lock()
	s.clients[clientID] = &Client{ID: clientID, Conn: conn}
	s.tunnelIDs[id] = clientID
	s.mu.Unlock()
	return id
}

func (s *Server) handleConnection(conn net.Conn) {
	log.Printf("new connection from %s", conn.RemoteAddr())

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	msg, err := protocol.Decode(conn)
	if err != nil {
		log.Printf("decode error: %v", err)
		conn.Close()
		return
	}

	switch msg.Type {
	case protocol.MsgAuth:
		s.handleAuth(conn, msg.Data.(protocol.AuthData))
	default:
		log.Printf("unexpected message type: %v", msg.Type)
		conn.Close()
	}
}

func (s *Server) handleAuth(conn net.Conn, data protocol.AuthData) {
	if data.Secret != s.secret {
		log.Printf("auth failed: wrong secret")
		protocol.Encode(conn, protocol.Message{Type: "AUTH_FAIL"})
		conn.Close()
		return
	}

	conn.SetDeadline(time.Time{})

	var tunnelID string
	if data.TunnelID != "" {
		tunnelID = data.TunnelID
	} else {
		tunnelID = s.AssignTunnelID(conn)
	}

	clientID := generateID()
	s.mu.Lock()
	s.clients[clientID] = &Client{ID: clientID, Conn: conn}
	s.tunnelIDs[tunnelID] = clientID
	s.mu.Unlock()

	log.Printf("client authenticated, tunnel: %s", tunnelID)

	protocol.Encode(conn, protocol.Message{
		Type: protocol.MsgAuthOK,
		Data: protocol.AuthOKData{
			TunnelID:   tunnelID,
			PublicPort: 8080,
		},
	})
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
