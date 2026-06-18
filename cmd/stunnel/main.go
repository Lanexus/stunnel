package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"

	"stunnel/internal/relay"

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
		newRelayCmd(),
		newServeCmd(),
		newConnectCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateSecret() string {
	b := make([]byte, 8)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func newRelayCmd() *cobra.Command {
	var addr string

	cmd := &cobra.Command{
		Use:   "relay",
		Short: "Start relay server (run on VPS)",
		Run: func(cmd *cobra.Command, args []string) {
			r := relay.New(addr)
			log.Printf("starting relay on %s", addr)
			if err := r.Start(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":7000", "Relay listen address")

	return cmd
}

func newServeCmd() *cobra.Command {
	var relayAddr string
	var localAddr string
	var secret string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Expose local service through relay",
		Run: func(cmd *cobra.Command, args []string) {
			if secret == "" {
				secret = generateSecret()
			}

			fmt.Println()
			fmt.Println("  ╔══════════════════════════════════════╗")
			fmt.Println("  ║       STUNNEL SERVE STARTED          ║")
			fmt.Println("  ╚══════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  Secret: %s\n", secret)
			fmt.Println()
			fmt.Println("  On another machine, run:")
			fmt.Printf("  stunnel connect --relay %s --secret %s\n", relayAddr, secret)
			fmt.Println()

			sc := relay.NewServeClient(relayAddr, secret, localAddr)
			if err := sc.Connect(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&relayAddr, "relay", "localhost:7000", "Relay server address")
	cmd.Flags().StringVar(&localAddr, "local", "localhost:3000", "Local service to expose")
	cmd.Flags().StringVar(&secret, "secret", "", "Shared secret (auto-generated if empty)")

	return cmd
}

func newConnectCmd() *cobra.Command {
	var relayAddr string
	var secret string

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a served tunnel",
		Run: func(cmd *cobra.Command, args []string) {
			if secret == "" {
				log.Fatal("--secret is required")
			}

			fmt.Printf("Connecting to relay %s ...\n", relayAddr)

			cc := relay.NewConnectClient(relayAddr, secret)
			if err := cc.Connect(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&relayAddr, "relay", "localhost:7000", "Relay server address")
	cmd.Flags().StringVar(&secret, "secret", "", "Shared secret (required)")
	cmd.MarkFlagRequired("secret")

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

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("stunnel v%s\n", version)
		},
	}
}
