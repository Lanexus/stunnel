package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type PeerInfo struct {
	Conn      net.Conn
	Timestamp int64
}

type Server struct {
	peers map[string]*PeerInfo
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		peers: make(map[string]*PeerInfo),
	}
}

func (s *Server) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("Relay listening on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Read: REGISTER <secret> or CONNECT <secret>
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	msg := string(buf[:n])
	if len(msg) < 3 {
		conn.Write([]byte("ERROR Invalid\n"))
		return
	}

	// Parse command
	cmd := ""
	secret := ""
	if len(msg) > 9 && msg[:9] == "REGISTER " {
		cmd = "REGISTER"
		secret = msg[9:]
	} else if len(msg) > 8 && msg[:8] == "CONNECT " {
		cmd = "CONNECT"
		secret = msg[8:]
	} else {
		conn.Write([]byte("ERROR Unknown command\n"))
		return
	}

	// Trim newline
	if len(secret) > 0 && secret[len(secret)-1] == '\n' {
		secret = secret[:len(secret)-1]
	}

	switch cmd {
	case "REGISTER":
		s.handleRegister(conn, secret)
	case "CONNECT":
		s.handleConnect(conn, secret)
	}
}

func (s *Server) handleRegister(conn net.Conn, secret string) {
	// Store the server connection
	s.mu.Lock()
	s.peers[secret] = &PeerInfo{Conn: conn, Timestamp: time.Now().Unix()}
	s.mu.Unlock()

	log.Printf("Server registered: %s", secret[:8]+"...")
	conn.Write([]byte("OK Registered\n"))

	// Keep connection alive
	conn.SetDeadline(time.Time{})
	buf := make([]byte, 1)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("Server disconnected: %s", secret[:8]+"...")
			s.mu.Lock()
			delete(s.peers, secret)
			s.mu.Unlock()
			return
		}
	}
}

func (s *Server) handleConnect(conn net.Conn, secret string) {
	s.mu.RLock()
	peer, exists := s.peers[secret]
	s.mu.RUnlock()

	if !exists {
		log.Printf("No server found: %s", secret[:8]+"...")
		conn.Write([]byte("ERROR No server found\n"))
		return
	}

	log.Printf("Client connecting: %s", secret[:8]+"...")
	conn.Write([]byte("OK Connected\n"))

	// Notify server
	peer.Conn.Write([]byte("CLIENT_CONNECTED\n"))

	// Bridge connections
	bridge(conn, peer.Conn)
}

func bridge(a, b net.Conn) {
	done := make(chan struct{})
	go func() {
		io.Copy(b, a)
		close(done)
	}()
	io.Copy(a, b)
	<-done
}

func main() {
	addr := ":7000"
	if len(os.Args) > 1 {
		addr = ":" + os.Args[1]
	}

	server := NewServer()
	if err := server.Start(addr); err != nil {
		log.Fatal(err)
	}
}
