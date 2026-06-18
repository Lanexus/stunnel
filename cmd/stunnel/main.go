package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var version = "0.7.0"

func main() {
	var secret string
	var port string
	var shell bool
	var generate bool
	var relay string
	var install bool
	var uninstall bool

	rootCmd := &cobra.Command{
		Use:   "stunnel",
		Short: "Connect like there is no firewall",
		Long:  "Both sides connect OUTBOUND to relay. No ports needed.",
	}

	// Server command - connects OUTBOUND to relay
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Expose local service (connects to relay)",
		Run: func(cmd *cobra.Command, args []string) {
			if generate {
				fmt.Println(generateSecret())
				return
			}

			if install {
				installService(secret, port, relay)
				return
			}

			if uninstall {
				uninstallService()
				return
			}

			if secret == "" {
				secret = generateSecret()
			}

			runServer(secret, port, relay)
		},
	}

	serverCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret")
	serverCmd.Flags().StringVarP(&port, "port", "p", "3000", "Local port to expose")
	serverCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell")
	serverCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate secret")
	serverCmd.Flags().StringVarP(&relay, "relay", "r", "93.177.100.9:7000", "Relay server address")
	serverCmd.Flags().BoolVar(&install, "install", false, "Install as systemd service")
	serverCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Uninstall service")

	// Client command - connects OUTBOUND to relay
	clientCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to server via relay",
		Run: func(cmd *cobra.Command, args []string) {
			if secret == "" {
				fmt.Println("Error: -s <secret> required")
				fmt.Println()
				fmt.Println("Usage:")
				fmt.Println("  stunnel connect -s <secret>")
				os.Exit(1)
			}

			runClient(secret, relay)
		},
	}

	clientCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret (required)")
	clientCmd.Flags().StringVarP(&relay, "relay", "r", "93.177.100.9:7000", "Relay server address")
	clientCmd.MarkFlagRequired("secret")

	rootCmd.AddCommand(serverCmd, clientCmd)
	rootCmd.Execute()
}

func generateSecret() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func runServer(secret, port, relay string) {
	// Connect OUTBOUND to relay
	conn, err := net.DialTimeout("tcp", relay, 10*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to relay: %v", err)
	}
	defer conn.Close()

	// Register with relay
	_, err = conn.Write([]byte("REGISTER " + secret + "\n"))
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	response := string(buf[:n])
	if response[:2] != "OK" {
		log.Fatalf("Registration failed: %s", response)
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL SERVER ACTIVE           ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Port:   %s\n", port)
	fmt.Println()
	fmt.Println("  Client command:")
	fmt.Printf("    stunnel connect -s %s\n", secret)
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n  Shutting down...")
		conn.Close()
		os.Exit(0)
	}()

	// Wait for client connection
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("Connection lost: %v", err)
		}

		msg := string(buf[:n])
		if msg == "CLIENT_CONNECTED\n" {
			fmt.Println("  Client connected!")
			// Now bridge to local port
			bridgeToLocal(conn, port)
			return
		}
	}
}

func runClient(secret, relay string) {
	// Connect OUTBOUND to relay
	conn, err := net.DialTimeout("tcp", relay, 10*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to relay: %v", err)
	}
	defer conn.Close()

	// Connect with secret
	_, err = conn.Write([]byte("CONNECT " + secret + "\n"))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	response := string(buf[:n])
	if response[:2] != "OK" {
		log.Fatalf("Connection failed: %s", response)
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL CONNECTED               ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  Connected! Pipe stdin/stdout...")
	fmt.Println("  Press Ctrl+C to disconnect")
	fmt.Println()

	// Pipe stdin/stdout
	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	io.Copy(conn, os.Stdin)
	<-done
}

func bridgeToLocal(conn net.Conn, port string) {
	// Connect to local service
	localConn, err := net.DialTimeout("tcp", "localhost:"+port, 5*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to local port %s: %v", port, err)
	}
	defer localConn.Close()

	// Bridge connections
	done := make(chan struct{})
	go func() {
		io.Copy(localConn, conn)
		close(done)
	}()
	io.Copy(conn, localConn)
	<-done
}

func installService(secret, port, relay string) {
	if secret == "" {
		secret = generateSecret()
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       INSTALLING STUNNEL             ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Port:   %s\n", port)
	fmt.Printf("  Relay:  %s\n", relay)
	fmt.Println()

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=Stunnel Server
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=5
ExecStart=/usr/local/bin/stunnel server -s %s -p %s -r %s
Environment=SHELL=/bin/bash

[Install]
WantedBy=multi-user.target
`, secret, port, relay)

	// Write service file
	servicePath := "/etc/systemd/system/stunnel.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()

	// Enable and start service
	exec.Command("systemctl", "enable", "stunnel").Run()
	exec.Command("systemctl", "start", "stunnel").Run()

	fmt.Println("  ✓ Installed as systemd service")
	fmt.Println("  ✓ Enabled on boot")
	fmt.Println("  ✓ Started")
	fmt.Println()
	fmt.Println("  Your secret (save this!):")
	fmt.Printf("    %s\n", secret)
	fmt.Println()
	fmt.Println("  Connect from anywhere:")
	fmt.Printf("    stunnel connect -s %s\n", secret)
	fmt.Println()
	fmt.Println("  Manage service:")
	fmt.Println("    systemctl status stunnel")
	fmt.Println("    systemctl restart stunnel")
	fmt.Println("    journalctl -u stunnel -f")
	fmt.Println()
}

func uninstallService() {
	fmt.Println()
	fmt.Println("  Uninstalling stunnel...")
	
	exec.Command("systemctl", "stop", "stunnel").Run()
	exec.Command("systemctl", "disable", "stunnel").Run()
	os.Remove("/etc/systemd/system/stunnel.service")
	exec.Command("systemctl", "daemon-reload").Run()

	fmt.Println("  ✓ Uninstalled")
	fmt.Println()
}
