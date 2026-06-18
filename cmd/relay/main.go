package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct {
	ID     string
	Secret string
	Conn   net.Conn
	IsServer bool
}

type Relay struct {
	addr    string
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewRelay(addr string) *Relay {
	return &Relay{
		addr:    addr,
		clients: make(map[string]*Client),
	}
}

func (r *Relay) Start() error {
	ln, err := net.Listen("tcp", r.addr)
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
		go r.handleConnection(conn)
	}
}

func (r *Relay) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		log.Printf("Invalid command: %s", line)
		conn.Write([]byte("ERROR Invalid command\n"))
		return
	}

	command := parts[0]
	secret := parts[1]

	switch command {
	case "REGISTER":
		r.handleRegister(conn, secret)
	case "CONNECT":
		r.handleConnect(conn, secret)
	default:
		log.Printf("Unknown command: %s", command)
		conn.Write([]byte("ERROR Unknown command\n"))
	}
}

func (r *Relay) handleRegister(conn net.Conn, secret string) {
	client := &Client{
		Secret:   secret,
		Conn:     conn,
		IsServer: true,
	}

	r.mu.Lock()
	r.clients[secret] = client
	r.mu.Unlock()

	log.Printf("Server registered with secret: %s", secret[:8]+"...")
	conn.Write([]byte("OK Registered\n"))

	// Keep connection alive and wait for client
	buf := make([]byte, 1)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("Server disconnected: %s", secret[:8]+"...")
			r.mu.Lock()
			delete(r.clients, secret)
			r.mu.Unlock()
			return
		}
	}
}

func (r *Relay) handleConnect(conn net.Conn, secret string) {
	r.mu.RLock()
	server, exists := r.clients[secret]
	r.mu.RUnlock()

	if !exists {
		log.Printf("No server found for secret: %s", secret[:8]+"...")
		conn.Write([]byte("ERROR No server found\n"))
		return
	}

	log.Printf("Client connecting to server: %s", secret[:8]+"...")
	
	// Notify server
	server.Conn.Write([]byte("CLIENT_CONNECTED\n"))
	
	// Notify client
	conn.Write([]byte("OK Connected\n"))

	// Bridge connections
	bridge(conn, server.Conn)
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
		addr = os.Args[1]
	}
	
	relay := NewRelay(addr)
	if err := relay.Start(); err != nil {
		log.Fatal(err)
	}
}
