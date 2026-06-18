package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"stunnel/internal/cloudflare"
	"stunnel/internal/daemon"
	"stunnel/internal/netcat"
	"stunnel/internal/relay"
	"stunnel/internal/transfer"

	"github.com/spf13/cobra"
)

var version = "0.2.0"

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
		newTunnelCmd(),
		newNetcatCmd(),
		newFileCmd(),
		newStopCmd(),
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

func newTunnelCmd() *cobra.Command {
	var localAddr string

	cmd := &cobra.Command{
		Use:   "tunnel",
		Short: "Expose local service via Cloudflare Tunnel (no VPS needed)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println()
			fmt.Println("  ╔══════════════════════════════════════╗")
			fmt.Println("  ║    STUNNEL CLOUDFLARE TUNNEL         ║")
			fmt.Println("  ╚══════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  Exposing: %s\n", localAddr)
			fmt.Println("  Starting Cloudflare Tunnel...")
			fmt.Println()

			t := cloudflare.NewTunnel(localAddr)
			if err := t.Start(); err != nil {
				log.Fatal(err)
			}
			defer t.Stop()

			fmt.Println("  ╔══════════════════════════════════════╗")
			fmt.Println("  ║         TUNNEL ACTIVE                ║")
			fmt.Println("  ╚══════════════════════════════════════╝")
			fmt.Println()
			fmt.Printf("  URL: %s\n", t.URL())
			fmt.Println()
			fmt.Println("  Share this URL to access your service")
			fmt.Println("  Press Ctrl+C to stop")
			fmt.Println()

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			<-sigCh

			fmt.Println("\n  Stopping tunnel...")
		},
	}

	cmd.Flags().StringVar(&localAddr, "local", "localhost:3000", "Local service to expose")

	return cmd
}

func newNetcatCmd() *cobra.Command {
	var listen bool
	var execCmd string

	cmd := &cobra.Command{
		Use:   "nc [address]",
		Short: "Netcat mode - connect or listen",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]

			nc := netcat.New(address, address, listen, execCmd)
			if err := nc.Run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().BoolVarP(&listen, "listen", "l", false, "Listen mode")
	cmd.Flags().StringVarP(&execCmd, "exec", "e", "", "Execute command on connection")

	return cmd
}

func newFileCmd() *cobra.Command {
	var listen bool
	var dest string

	cmd := &cobra.Command{
		Use:   "file [address] [path]",
		Short: "File transfer mode",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]
			path := args[1]

			ft := transfer.New(address, dest, listen)
			if listen {
				if err := ft.Receive(); err != nil {
					log.Fatal(err)
				}
			} else {
				ft.SetLocalPath(path)
				if err := ft.Send(); err != nil {
					log.Fatal(err)
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&listen, "listen", "l", false, "Listen mode (receive)")
	cmd.Flags().StringVarP(&dest, "dest", "d", "", "Destination directory")

	return cmd
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop running daemon",
		Run: func(cmd *cobra.Command, args []string) {
			if err := daemon.StopDaemon(); err != nil {
				log.Fatal(err)
			}
		},
	}
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
