package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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

type pendingConn struct {
	conn    net.Conn
	connID  string
	ready   chan struct{}
}

type Server struct {
	addr       string
	publicAddr string
	secret     string
	listener   net.Listener
	pubLn      net.Listener
	clients    map[string]*Client
	mu         sync.RWMutex
	tunnelIDs  map[string]string
	pending    map[string]*pendingConn
	done       chan struct{}
}

func New(addr, publicAddr, secret string) *Server {
	return &Server{
		addr:       addr,
		publicAddr: publicAddr,
		secret:     secret,
		clients:    make(map[string]*Client),
		tunnelIDs:  make(map[string]string),
		pending:    make(map[string]*pendingConn),
		done:       make(chan struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = ln
	log.Printf("server listening on %s", ln.Addr())

	if s.publicAddr != "" {
		pubLn, err := net.Listen("tcp", s.publicAddr)
		if err != nil {
			ln.Close()
			return fmt.Errorf("public listen: %w", err)
		}
		s.pubLn = pubLn
		log.Printf("public listener on %s", pubLn.Addr())
		go s.acceptPublic()
	}

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
	if s.pubLn != nil {
		s.pubLn.Close()
	}
}

func (s *Server) PublicAddr() net.Addr {
	if s.pubLn == nil {
		return nil
	}
	return s.pubLn.Addr()
}

func (s *Server) Addr() net.Addr {
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) acceptPublic() {
	for {
		conn, err := s.pubLn.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				log.Printf("public accept error: %v", err)
				return
			}
		}
		go s.handleUserConnection(conn)
	}
}

func (s *Server) handleUserConnection(conn net.Conn) {
	log.Printf("user connection from %s", conn.RemoteAddr())

	s.mu.RLock()
	var targetTunnelID string
	for tunnelID, clientID := range s.tunnelIDs {
		if _, ok := s.clients[clientID]; ok {
			targetTunnelID = tunnelID
			break
		}
	}
	s.mu.RUnlock()

	if targetTunnelID == "" {
		log.Printf("no available tunnel")
		conn.Close()
		return
	}

	s.mu.RLock()
	clientID := s.tunnelIDs[targetTunnelID]
	client := s.clients[clientID]
	s.mu.RUnlock()

	if client == nil {
		log.Printf("client not found for tunnel %s", targetTunnelID)
		conn.Close()
		return
	}

	connID := generateID()
	pc := &pendingConn{conn: conn, connID: connID, ready: make(chan struct{})}

	s.mu.Lock()
	key := targetTunnelID + ":" + connID
	s.pending[key] = pc
	s.mu.Unlock()

	protocol.Encode(client.Conn, protocol.Message{
		Type: protocol.MsgNewTunnel,
		Data: protocol.NewTunnelData{TunnelID: targetTunnelID, ConnID: connID},
	})

	select {
	case <-pc.ready:
	case <-time.After(30 * time.Second):
		log.Printf("timeout waiting for data connection for tunnel %s conn %s", targetTunnelID, connID)
		s.mu.Lock()
		delete(s.pending, key)
		s.mu.Unlock()
		conn.Close()
		return
	}

	go io.Copy(conn, pc.conn)
	io.Copy(pc.conn, conn)
	conn.Close()
	pc.conn.Close()
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
	case protocol.MsgDataOpen:
		s.handleDataOpen(conn, msg.Data.(protocol.DataOpenData))
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

	publicPort := 0
	if s.pubLn != nil {
		publicPort = s.pubLn.Addr().(*net.TCPAddr).Port
	}

	protocol.Encode(conn, protocol.Message{
		Type: protocol.MsgAuthOK,
		Data: protocol.AuthOKData{
			TunnelID:   tunnelID,
			PublicPort: publicPort,
		},
	})
}

func (s *Server) handleDataOpen(conn net.Conn, data protocol.DataOpenData) {
	conn.SetDeadline(time.Time{})

	key := data.TunnelID + ":" + data.ConnID
	s.mu.Lock()
	pc, ok := s.pending[key]
	if ok {
		delete(s.pending, key)
	}
	s.mu.Unlock()

	if !ok {
		log.Printf("no pending connection for tunnel %s conn %s", data.TunnelID, data.ConnID)
		conn.Close()
		return
	}

	pc.conn = conn
	close(pc.ready)
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
