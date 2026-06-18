package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"stunnel/internal/client"
	"stunnel/internal/server"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "stunnel",
		Short: "A TCP tunnel tool",
		Long:  "Expose local services through a relay server",
	}

	rootCmd.AddCommand(
		newServeCmd(),
		newConnectCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateSecret() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func encodeKey(host, port, secret string) string {
	raw := fmt.Sprintf("%s:%s:%s", host, port, secret)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeKey(key string) (host, port, secret string, err error) {
	raw, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid key")
	}
	parts := strings.SplitN(string(raw), ":", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid key format")
	}
	return parts[0], parts[1], parts[2], nil
}

func newServeCmd() *cobra.Command {
	var addr string
	var publicAddr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start tunnel server and print connection key",
		Run: func(cmd *cobra.Command, args []string) {
			secret := generateSecret()

			// Get public IP
			host := getPublicIP()
			pubPort := extractPort(publicAddr)

			key := encodeKey(host, pubPort, secret)

			srv := server.New(addr, publicAddr, secret)

			fmt.Println()
			fmt.Println("  ╔══════════════════════════════════════╗")
			fmt.Println("  ║       STUNNEL SERVER STARTED         ║")
			fmt.Println("  ╚══════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  Key: %s\n", key)
			fmt.Println()
			fmt.Println("  On your local machine, run:")
			fmt.Printf("  stunnel connect %s --local :PORT\n", key)
			fmt.Println()

			if err := srv.Start(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":7000", "Server listen address")
	cmd.Flags().StringVar(&publicAddr, "public-addr", ":8080", "Public listener address")

	return cmd
}

func newConnectCmd() *cobra.Command {
	var localAddr string

	cmd := &cobra.Command{
		Use:   "connect [key]",
		Short: "Connect to tunnel server using key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			host, port, secret, err := decodeKey(key)
			if err != nil {
				log.Fatalf("invalid key: %v", err)
			}

			serverAddr := net.JoinHostPort(host, port)

			fmt.Printf("Connecting to %s ...\n", serverAddr)
			fmt.Printf("Exposing local %s\n", localAddr)

			c := client.New(serverAddr, secret, localAddr)
			if err := c.Connect(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&localAddr, "local", "localhost:3000", "Local service address to expose")

	return cmd
}

func getPublicIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func extractPort(addr string) string {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "8080"
	}
	return port
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("stunnel v%s\n", version)
		},
	}
}
