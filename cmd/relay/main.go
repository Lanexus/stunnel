package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type PeerInfo struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Timestamp int64  `json:"ts"`
}

type Client struct {
	ID       string
	Secret   string
	Conn     net.Conn
	IsServer bool
}

type Server struct {
	// HTTP signaling
	peers map[string]*PeerInfo
	mu    sync.RWMutex
	
	// TCP relay
	clients map[string]*Client
	clientsMu sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		peers:   make(map[string]*PeerInfo),
		clients: make(map[string]*Client),
	}
}

func (s *Server) StartHTTP(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", s.handleRegister)
	mux.HandleFunc("/lookup/", s.handleLookup)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/", s.handleRoot)

	log.Printf("HTTP signaling listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) StartTCP(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("TCP relay listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleTCPConnection(conn)
	}
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    "stunnel",
		"version": "0.5.0",
		"status":  "ok",
		"peers":   len(s.peers),
		"clients": len(s.clients),
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info PeerInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	secret := r.URL.Query().Get("secret")
	if secret == "" {
		http.Error(w, "Secret required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.peers[secret] = &info
	s.mu.Unlock()

	go s.cleanOldEntries()

	log.Printf("Registered peer: %s:%s (secret: %s)", info.IP, info.Port, secret)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleLookup(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Path[len("/lookup/"):]
	if secret == "" {
		http.Error(w, "Secret required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	info, exists := s.peers[secret]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Peer not found", http.StatusNotFound)
		return
	}

	if time.Now().Unix()-info.Timestamp > 300 {
		s.mu.Lock()
		delete(s.peers, secret)
		s.mu.Unlock()
		http.Error(w, "Peer expired", http.StatusNotFound)
		return
	}

	log.Printf("Lookup peer: %s -> %s:%s", secret, info.IP, info.Port)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"peers":  len(s.peers),
	})
}

func (s *Server) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	// Read first message to determine type
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	msg := strings.TrimSpace(string(buf[:n]))
	parts := strings.SplitN(msg, " ", 2)
	if len(parts) != 2 {
		conn.Write([]byte("ERROR Invalid command\n"))
		return
	}

	command := parts[0]
	secret := parts[1]

	switch command {
	case "REGISTER":
		s.handleTCPRegister(conn, secret)
	case "CONNECT":
		s.handleTCPConnect(conn, secret)
	default:
		conn.Write([]byte("ERROR Unknown command\n"))
	}
}

func (s *Server) handleTCPRegister(conn net.Conn, secret string) {
	client := &Client{
		Secret:   secret,
		Conn:     conn,
		IsServer: true,
	}

	s.clientsMu.Lock()
	s.clients[secret] = client
	s.clientsMu.Unlock()

	log.Printf("TCP server registered: %s", secret)
	conn.Write([]byte("OK Registered\n"))

	// Keep connection alive
	buf := make([]byte, 1)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("TCP server disconnected: %s", secret)
			s.clientsMu.Lock()
			delete(s.clients, secret)
			s.clientsMu.Unlock()
			return
		}
	}
}

func (s *Server) handleTCPConnect(conn net.Conn, secret string) {
	s.clientsMu.RLock()
	server, exists := s.clients[secret]
	s.clientsMu.RUnlock()

	if !exists {
		log.Printf("No TCP server found: %s", secret)
		conn.Write([]byte("ERROR No server found\n"))
		return
	}

	log.Printf("TCP client connecting: %s", secret)
	server.Conn.Write([]byte("CLIENT_CONNECTED\n"))
	conn.Write([]byte("OK Connected\n"))

	// Bridge connections
	bridge(conn, server.Conn)
}

func (s *Server) cleanOldEntries() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for secret, info := range s.peers {
		if now-info.Timestamp > 300 {
			delete(s.peers, secret)
		}
	}
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
	httpAddr := ":8080"
	tcpAddr := ":7000"

	if len(os.Args) > 1 {
		httpAddr = ":" + os.Args[1]
		tcpAddr = ":" + os.Args[1]
	}

	server := NewServer()

	go server.StartTCP(tcpAddr)
	
	if err := server.StartHTTP(httpAddr); err != nil {
		log.Fatal(err)
	}
}
