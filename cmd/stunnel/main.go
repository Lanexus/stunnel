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
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var version = "0.6.0"

func main() {
	var secret string
	var port string
	var shell bool
	var generate bool
	var install bool
	var uninstall bool
	var serverAddr string

	rootCmd := &cobra.Command{
		Use:   "stunnel",
		Short: "Connect like there is no firewall",
		Long:  "Persistent tunnel server. Install once, connect anytime.",
	}

	// Server command
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run server (persistent)",
		Run: func(cmd *cobra.Command, args []string) {
			if generate {
				fmt.Println(generateSecret())
				return
			}

			if install {
				installService(secret, port, shell)
				return
			}

			if uninstall {
				uninstallService()
				return
			}

			if secret == "" {
				secret = generateSecret()
			}

			runServer(secret, port, shell)
		},
	}

	serverCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret")
	serverCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to expose")
	serverCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell")
	serverCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate secret")
	serverCmd.Flags().BoolVar(&install, "install", false, "Install as system service")
	serverCmd.Flags().BoolVar(&uninstall, "uninstall", false, "Uninstall service")

	// Client command
	clientCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to server",
		Run: func(cmd *cobra.Command, args []string) {
			if secret == "" {
				fmt.Println("Error: -s <secret> required")
				fmt.Println()
				fmt.Println("Usage:")
				fmt.Println("  stunnel connect -s <secret>")
				fmt.Println("  stunnel connect -s <secret> --shell")
				os.Exit(1)
			}

			runClient(secret, shell, serverAddr)
		},
	}

	clientCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret (required)")
	clientCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell")
	clientCmd.Flags().StringVarP(&serverAddr, "addr", "a", "", "Server address (e.g., 1.2.3.4:3000)")
	clientCmd.MarkFlagRequired("secret")

	rootCmd.AddCommand(serverCmd, clientCmd)
	rootCmd.Execute()
}

func generateSecret() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func runServer(secret, port string, shell bool) {
	// Get public IP
	ip := getPublicIP()
	addr := fmt.Sprintf("%s:%s", ip, port)

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL SERVER                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Address: %s\n", addr)
	fmt.Println()
	fmt.Println("  Client command:")
	fmt.Printf("    stunnel connect -s %s\n", secret)
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Listen for incoming connection
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	fmt.Printf("  Listening on :%s\n", port)

	// Handle shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n  Shutting down...")
		ln.Close()
		os.Exit(0)
	}()

	// Accept connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		go handleServerConnection(conn, secret, shell)
	}
}

func handleServerConnection(conn net.Conn, secret string, shell bool) {
	defer conn.Close()

	// Read secret from client
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	clientSecret := string(buf[:n])
	if clientSecret != secret {
		conn.Write([]byte("ERROR Invalid secret"))
		return
	}

	// Send OK
	conn.Write([]byte("OK"))

	fmt.Println("  Client connected!")

	if shell {
		handleShell(conn, true)
	} else {
		handlePipe(conn)
	}
}

func runClient(secret string, shell bool, serverAddr string) {
	// For now, connect directly to server
	// In a real implementation, we'd use signaling server
	// For simplicity, we'll ask for server address
	
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL CLIENT                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Println()

	// Get server address
	if serverAddr == "" {
		fmt.Println("  Enter server address (e.g., 1.2.3.4:3000):")
		fmt.Print("  > ")
		fmt.Scanln(&serverAddr)
	}

	fmt.Println()
	fmt.Printf("  Connecting to %s...\n", serverAddr)

	// Connect to server
	conn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send secret
	if _, err := conn.Write([]byte(secret)); err != nil {
		log.Fatalf("Failed to send secret: %v", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	response := string(buf[:n])
	if response != "OK" {
		log.Fatalf("Authentication failed: %s", response)
	}

	fmt.Println("  Connected!")

	if shell {
		handleShell(conn, false)
	} else {
		handlePipe(conn)
	}
}

func installService(secret, port string, shell bool) {
	if secret == "" {
		secret = generateSecret()
	}

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       INSTALLING STUNNEL             ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Port: %s\n", port)
	fmt.Println()

	// Create systemd service
	serviceContent := fmt.Sprintf(`[Unit]
Description=Stunnel Server
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/stunnel server -s %s -p %s
Environment=SHELL=/bin/bash

[Install]
WantedBy=multi-user.target
`, secret, port)

	if shell {
		serviceContent = fmt.Sprintf(`[Unit]
Description=Stunnel Server
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/usr/local/bin/stunnel server -s %s -p %s --shell
Environment=SHELL=/bin/bash

[Install]
WantedBy=multi-user.target
`, secret, port)
	}

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
	fmt.Println("    systemctl stop stunnel")
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

func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func handleShell(conn net.Conn, isServer bool) {
	if isServer {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd.exe")
		} else {
			cmd = exec.Command(shell, "-i")
		}
		
		cmd.Stdin = conn
		cmd.Stdout = conn
		cmd.Stderr = conn
		cmd.Run()
	} else {
		done := make(chan struct{})
		go func() {
			io.Copy(os.Stdout, conn)
			close(done)
		}()
		io.Copy(conn, os.Stdin)
		<-done
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
	go func() { defer wg.Done(); io.Copy(b, a) }()
	go func() { defer wg.Done(); io.Copy(a, b) }()
	wg.Wait()
}
