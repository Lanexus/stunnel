package nat

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type HolePuncher struct {
	localPort  string
	remoteAddr string
	timeout    time.Duration
}

func NewHolePuncher(localPort, remoteAddr string) *HolePuncher {
	return &HolePuncher{
		localPort:  localPort,
		remoteAddr: remoteAddr,
		timeout:    10 * time.Second,
	}
}

func (hp *HolePuncher) Punch() (net.Conn, error) {
	// Get local address
	localAddr, err := net.ResolveUDPAddr("udp", ":"+hp.localPort)
	if err != nil {
		return nil, fmt.Errorf("resolve local: %w", err)
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", hp.remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve remote: %w", err)
	}

	// Create UDP connection
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}

	// Send packets to punch hole
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(2)

	// Sender goroutine
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_, err := conn.WriteToUDP([]byte("PUNCH"), remoteAddr)
				if err != nil {
					log.Printf("send error: %v", err)
				}
			}
		}
	}()

	// Receiver goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)

		for {
			select {
			case <-done:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					continue
				}

				if addr.String() == remoteAddr.String() {
					log.Printf("received from remote: %s", string(buf[:n]))
					close(done)
					return
				}
			}
		}
	}()

	// Wait for connection or timeout
	select {
	case <-done:
		// Convert UDP to TCP-like connection
		return wrapUDP(conn, remoteAddr)
	case <-time.After(hp.timeout):
		conn.Close()
		return nil, fmt.Errorf("hole punch timeout")
	}
}

func wrapUDP(conn *net.UDPConn, remote *net.UDPAddr) (net.Conn, error) {
	// For simplicity, we'll use TCP after hole punching
	// In a real implementation, you'd use a custom Conn wrapper
	conn.Close()
	return net.DialTimeout("tcp", remote.String(), 5*time.Second)
}
