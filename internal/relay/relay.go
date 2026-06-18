package relay

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type MsgType string

const (
	MsgRegister  MsgType = "REGISTER"  // serve client registers
	MsgRequest   MsgType = "REQUEST"   // connect client requests
	MsgMatched   MsgType = "MATCHED"   // both sides notified
	MsgReady     MsgType = "READY"     // serve client ready for data
)

type Message struct {
	Type    MsgType     `json:"type"`
	Secret  string      `json:"secret,omitempty"`
	TunnelID string     `json:"tunnel_id,omitempty"`
}

type tunnel struct {
	secret   string
	serveCh  chan net.Conn // channel for serve client connection
	mu       sync.Mutex
	closed   bool
}

type Relay struct {
	addr    string
	tunnels map[string]*tunnel
	mu      sync.RWMutex
	done    chan struct{}
}

func New(addr string) *Relay {
	return &Relay{
		addr:    addr,
		tunnels: make(map[string]*tunnel),
		done:    make(chan struct{}),
	}
}

func (r *Relay) Start() error {
	ln, err := net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("relay listening on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-r.done:
				return nil
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}
		go r.handleConnection(conn)
	}
}

func (r *Relay) Stop() {
	close(r.done)
}

func (r *Relay) handleConnection(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	var msg Message
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&msg); err != nil {
		log.Printf("decode error: %v", err)
		conn.Close()
		return
	}

	conn.SetDeadline(time.Time{})

	switch msg.Type {
	case MsgRegister:
		r.handleRegister(conn, msg)
	case MsgRequest:
		r.handleRequest(conn, msg)
	default:
		log.Printf("unknown message type: %v", msg.Type)
		conn.Close()
	}
}

func (r *Relay) handleRegister(conn net.Conn, msg Message) {
	tunnelID := generateID()
	
	t := &tunnel{
		secret:  msg.Secret,
		serveCh: make(chan net.Conn, 1),
	}

	r.mu.Lock()
	r.tunnels[tunnelID] = t
	r.mu.Unlock()

	// Send tunnel ID to serve client
	resp := Message{
		Type:     MsgMatched,
		TunnelID: tunnelID,
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)

	log.Printf("tunnel registered: %s", tunnelID)

	// Wait for a connect client to match
	select {
	case connectConn := <-t.serveCh:
		// Matched! Send READY to serve client
		ready := Message{Type: MsgReady}
		readyData, _ := json.Marshal(ready)
		readyData = append(readyData, '\n')
		conn.Write(readyData)

		// Bridge the connections
		r.bridge(conn, connectConn)
		
		// Cleanup
		r.mu.Lock()
		delete(r.tunnels, tunnelID)
		r.mu.Unlock()
		
	case <-r.done:
		r.mu.Lock()
		delete(r.tunnels, tunnelID)
		r.mu.Unlock()
		conn.Close()
	}
}

func (r *Relay) handleRequest(conn net.Conn, msg Message) {
	r.mu.RLock()
	var matchedTunnel *tunnel
	var matchedID string
	for id, t := range r.tunnels {
		if t.secret == msg.Secret {
			matchedTunnel = t
			matchedID = id
			break
		}
	}
	r.mu.RUnlock()

	if matchedTunnel == nil {
		log.Printf("no tunnel found for secret")
		conn.Close()
		return
	}

	// Send tunnel ID to connect client
	resp := Message{
		Type:     MsgMatched,
		TunnelID: matchedID,
	}
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)

	// Send connect client to serve client
	matchedTunnel.serveCh <- conn

	log.Printf("matched connect client to tunnel %s", matchedID)
}

func (r *Relay) bridge(a, b net.Conn) {
	done := make(chan struct{})
	go func() {
		io.Copy(b, a)
		close(done)
	}()
	io.Copy(a, b)
	a.Close()
	b.Close()
	<-done
}

func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
