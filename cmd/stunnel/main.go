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
	"runtime"
	"sync"
	"time"

	"stunnel/internal/signaling"

	"github.com/spf13/cobra"
)

var version = "0.5.0"

func main() {
	var listen bool
	var secret string
	var port string
	var shell bool
	var generate bool
	var signalingServer string

	rootCmd := &cobra.Command{
		Use:   "stunnel",
		Short: "Connect like there is no firewall",
		Long:  "Expose local services to the internet. No VPS needed.",
		Run: func(cmd *cobra.Command, args []string) {
			if generate {
				fmt.Println(generateSecret())
				return
			}

			if secret == "" {
				secret = generateSecret()
			}

			if listen {
				runServer(secret, port, shell, signalingServer)
			} else {
				runClient(secret, shell, signalingServer)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&listen, "listen", "l", false, "Listen mode (server)")
	rootCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret")
	rootCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to expose")
	rootCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell")
	rootCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate secret")
	rootCmd.Flags().StringVar(&signalingServer, "signaling", "http://93.177.100.9:8080", "Signaling server address")

	rootCmd.Execute()
}

func generateSecret() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func runServer(secret, port string, shell bool, signalingServer string) {
	// Get public IP
	ip := getPublicIP()

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL SERVER                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Port:   %s\n", port)
	fmt.Println()
	fmt.Println("  Registering with signaling server...")
	fmt.Println()

	// Register with signaling server
	client := signaling.NewSignalingClient(signalingServer)
	err := client.Register(secret, ip, port)
	if err != nil {
		log.Printf("Warning: Failed to register: %v", err)
		log.Printf("You can still use direct connection")
	} else {
		fmt.Println("  Registered successfully!")
	}

	fmt.Println()
	fmt.Println("  Waiting for client connection...")
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println()

	// Listen for incoming connection
	addr := fmt.Sprintf(":%s", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	fmt.Printf("  Listening on %s\n", addr)

	// Accept connection
	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("Failed to accept: %v", err)
	}
	defer conn.Close()

	fmt.Println("  Client connected!")

	if shell {
		handleShell(conn, true)
	} else {
		handlePipe(conn)
	}
}

func runClient(secret string, shell bool, signalingServer string) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL CLIENT                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Println()
	fmt.Println("  Looking up server...")
	fmt.Println()

	// Lookup server
	client := signaling.NewSignalingClient(signalingServer)
	info, err := client.Lookup(secret)
	if err != nil {
		log.Fatalf("Failed to find server: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", info.IP, info.Port)
	fmt.Printf("  Found server at %s\n", addr)
	fmt.Println("  Connecting...")
	fmt.Println()

	// Connect to server
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	fmt.Println("  Connected!")

	if shell {
		handleShell(conn, false)
	} else {
		handlePipe(conn)
	}
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
