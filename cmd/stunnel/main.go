package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var version = "0.4.0"

func main() {
	var listen bool
	var secret string
	var port string
	var shell bool
	var generate bool

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
				runServer(secret, port, shell)
			} else {
				runClient(secret, shell)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&listen, "listen", "l", false, "Listen mode (server)")
	rootCmd.Flags().StringVarP(&secret, "secret", "s", "", "Shared secret")
	rootCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to expose")
	rootCmd.Flags().BoolVar(&shell, "shell", false, "Interactive shell")
	rootCmd.Flags().BoolVarP(&generate, "generate", "g", false, "Generate secret")

	rootCmd.Execute()
}

func generateSecret() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func runServer(secret, port string, shell bool) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL SERVER                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Printf("  Port:   %s\n", port)
	fmt.Println()
	fmt.Println("  Starting tunnel...")
	fmt.Println()

	// Check/install cloudflared
	cloudflared, err := ensureCloudflared()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Start cloudflare tunnel
	cmd := exec.Command(cloudflared, "tunnel", "--url", "http://localhost:"+port)
	
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start tunnel: %v", err)
	}

	// Read output to find URL
	urlCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "trycloudflare.com") {
				re := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
				match := re.FindString(line)
				if match != "" {
					urlCh <- match
					return
				}
			}
		}
	}()

	select {
	case url := <-urlCh:
		fmt.Println("  ╔══════════════════════════════════════╗")
		fmt.Println("  ║       TUNNEL ACTIVE                  ║")
		fmt.Println("  ╚══════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  URL: %s\n", url)
		fmt.Println()
		fmt.Println("  Share this URL to access your service")
		fmt.Println()
		fmt.Println("  Client command:")
		fmt.Printf("    stunnel -s %s\n", secret)
		fmt.Println()
		fmt.Println("  Press Ctrl+C to stop")
		fmt.Println()
		
		// Wait for interrupt
		select {}
		
	case <-time.After(30 * time.Second):
		log.Fatal("Timeout waiting for tunnel URL")
	}
}

func runClient(secret string, shell bool) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║       STUNNEL CLIENT                 ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Secret: %s\n", secret)
	fmt.Println()
	fmt.Println("  To connect, you need the server URL")
	fmt.Println("  Ask the server operator for the URL")
	fmt.Println()
	fmt.Println("  Then open it in your browser or use:")
	fmt.Printf("    curl <server-url>\n")
	fmt.Println()
}

func ensureCloudflared() (string, error) {
	// Check if already installed
	if path, err := exec.LookPath("cloudflared"); err == nil {
		return path, nil
	}

	// Check local binary
	localBin := "./cloudflared"
	if _, err := os.Stat(localBin); err == nil {
		return localBin, nil
	}

	// Download
	log.Printf("Downloading cloudflared...")
	return downloadCloudflared()
}

func downloadCloudflared() (string, error) {
	// Detect OS
	osName := "linux"
	if _, err := exec.LookPath("cmd.exe"); err == nil {
		osName = "windows"
	}

	var url string
	switch osName {
	case "linux":
		url = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64"
	case "windows":
		url = "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.exe"
	default:
		return "", fmt.Errorf("unsupported OS: %s", osName)
	}

	// Download
	cmd := exec.Command("curl", "-sL", url, "-o", "cloudflared")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	// Make executable
	os.Chmod("cloudflared", 0755)

	return "./cloudflared", nil
}

func handleShell(conn net.Conn, isServer bool) {
	if isServer {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		cmd := exec.Command(shell, "-i")
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

func bridge(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(b, a) }()
	go func() { defer wg.Done(); io.Copy(a, b) }()
	wg.Wait()
}
