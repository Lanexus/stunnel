package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type PeerInfo struct {
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Timestamp int64  `json:"ts"`
}

type SignalingServer struct {
	peers map[string]*PeerInfo
	mu    sync.RWMutex
}

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		peers: make(map[string]*PeerInfo),
	}
}

func (s *SignalingServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", s.handleRegister)
	mux.HandleFunc("/lookup/", s.handleLookup)
	mux.HandleFunc("/health", s.handleHealth)

	log.Printf("Signaling server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *SignalingServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var info PeerInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get secret from query params
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		http.Error(w, "Secret required", http.StatusBadRequest)
		return
	}

	// Store peer info
	s.mu.Lock()
	s.peers[secret] = &info
	s.mu.Unlock()

	// Clean old entries
	go s.cleanOldEntries()

	log.Printf("Registered peer: %s:%s (secret: %s)", info.IP, info.Port, secret)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *SignalingServer) handleLookup(w http.ResponseWriter, r *http.Request) {
	// Extract secret from path
	secret := r.URL.Path[len("/lookup/"):]
	if secret == "" {
		http.Error(w, "Secret required", http.StatusBadRequest)
		return
	}

	log.Printf("Looking up secret: %s", secret)
	log.Printf("Current peers: %v", s.peers)

	s.mu.RLock()
	info, exists := s.peers[secret]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Peer not found", http.StatusNotFound)
		return
	}

	// Check if peer is still alive (within 5 minutes)
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

func (s *SignalingServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"peers":  len(s.peers),
	})
}

func (s *SignalingServer) cleanOldEntries() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().Unix()
	for secret, info := range s.peers {
		if now-info.Timestamp > 300 {
			delete(s.peers, secret)
		}
	}
}

func main() {
	addr := ":8080"

	// Check for custom port
	if len(os.Args) > 1 {
		addr = ":" + os.Args[1]
	}

	server := NewSignalingServer()
	if err := server.Start(addr); err != nil {
		log.Fatal(err)
	}
}
