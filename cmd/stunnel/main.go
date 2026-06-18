package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var (
	version = "0.3.0"
	defaultRelay = "relay.stunnel.io:7000"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "stunnel",
		Short: "Connect like there is no firewall. Securely.",
		Long: `Stunnel allows two users behind NAT/Firewall to establish a TCP connection.
Both users use the same secret to find each other through the relay network.`,
	}

	var secret string
	var listen bool
	var relay string
	var shell bool
	var generate bool
	var port string
	var verbose bool

	rootCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret (password)")
	rootCmd.Flags().BoolVarP(&listen, "listen", "l", false, "Listen mode (server)")
	rootCmd.Flags().StringVarP(&relay, "relay", "r", defaultRelay, "Relay server address")
	rootCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell mode")
	rootCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate random secret")
	rootCmd.Flags().StringVarP(&port, "port", "p", "", "Local port to forward")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if generate {
			fmt.Println(generateSecret())
			return
		}

		if secret == "" {
			fmt.Println("Error: --secret (-s) is required")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println("  stunnel -s <secret> -l          # Listen (server)")
			fmt.Println("  stunnel -s <secret>             # Connect (client)")
			fmt.Println("  stunnel -s <secret> -l --shell  # Listen with shell")
			fmt.Println("  stunnel -s <secret> --shell     # Connect with shell")
			fmt.Println("  stunnel -g                      # Generate secret")
			os.Exit(1)
		}

		if listen {
			runServer(secret, relay, shell, port, verbose)
		} else {
			runClient(secret, relay, shell, port, verbose)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateSecret() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func runServer(secret, relay string, shell bool, port string, verbose bool) {
	if verbose {
		log.Printf("Connecting to relay %s...", relay)
	}

	conn, err := net.DialTimeout("tcp", relay, 10*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to relay: %v", err)
	}
	defer conn.Close()

	// Register as server
	msg := fmt.Sprintf("REGISTER %s\n", secret)
	if _, err := conn.Write([]byte(msg)); err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	response := string(buf[:n])
	if verbose {
		log.Printf("Relay response: %s", response)
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL SERVER ACTIVE          ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Println()
	fmt.Println("  Waiting for connection...")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Wait for client connection
	n, err = conn.Read(buf)
	if err != nil {
		log.Fatalf("Connection lost: %v", err)
	}

	response = string(buf[:n])
	if verbose {
		log.Printf("Client connected: %s", response)
	}

	fmt.Println("  Client connected!")

	if shell {
		handleShell(conn, true)
	} else if port != "" {
		handlePortForward(conn, port, true)
	} else {
		handlePipe(conn)
	}
}

func runClient(secret, relay string, shell bool, port string, verbose bool) {
	if verbose {
		log.Printf("Connecting to relay %s...", relay)
	}

	conn, err := net.DialTimeout("tcp", relay, 10*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to relay: %v", err)
	}
	defer conn.Close()

	// Connect as client
	msg := fmt.Sprintf("CONNECT %s\n", secret)
	if _, err := conn.Write([]byte(msg)); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	response := string(buf[:n])
	if verbose {
		log.Printf("Relay response: %s", response)
	}

	fmt.Println("  Connected to server!")

	if shell {
		handleShell(conn, false)
	} else if port != "" {
		handlePortForward(conn, port, false)
	} else {
		handlePipe(conn)
	}
}

func handleShell(conn net.Conn, isServer bool) {
	fmt.Println("  Starting interactive shell...")
	fmt.Println("  Type 'exit' to quit")
	fmt.Println()

	if isServer {
		// Server: execute commands
		handleShellServer(conn)
	} else {
		// Client: send commands
		handleShellClient(conn)
	}
}

func handleShellServer(conn net.Conn) {
	// Create shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := createShellCommand(shell)
	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	if err := cmd.Run(); err != nil {
		log.Printf("Shell exited: %v", err)
	}
}

func handleShellClient(conn net.Conn) {
	done := make(chan struct{})
	
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	
	io.Copy(conn, os.Stdin)
	<-done
}

func handlePortForward(conn net.Conn, port string, isServer bool) {
	addr := fmt.Sprintf("localhost:%s", port)
	
	if isServer {
		// Listen on local port
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("Cannot listen on %s: %v", addr, err)
		}
		defer ln.Close()

		fmt.Printf("  Forwarding port %s through tunnel\n", port)

		for {
			localConn, err := ln.Accept()
			if err != nil {
				continue
			}
			go func() {
				defer localConn.Close()
				bridge(conn, localConn)
			}()
		}
	} else {
		// Connect to local port
		localConn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			log.Fatalf("Cannot connect to %s: %v", addr, err)
		}
		defer localConn.Close()

		fmt.Printf("  Forwarding port %s through tunnel\n", port)
		bridge(conn, localConn)
	}
}

func handlePipe(conn net.Conn) {
	done := make(chan struct{})
	
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	
	io.Copy(conn, os.Stdin)
	<-done
}

func bridge(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(b, a)
	}()

	go func() {
		defer wg.Done()
		io.Copy(a, b)
	}()

	wg.Wait()
}
