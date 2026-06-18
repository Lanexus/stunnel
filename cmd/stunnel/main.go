package main

import (
	"fmt"
	"log"
	"os"

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
		newServerCmd(),
		newClientCmd(),
		newVersionCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newServerCmd() *cobra.Command {
	var addr string
	var publicAddr string
	var secret string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start tunnel server",
		Run: func(cmd *cobra.Command, args []string) {
			srv := server.New(addr, publicAddr, secret)
			log.Printf("starting server on %s", addr)
			if err := srv.Start(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":7000", "Server listen address")
	cmd.Flags().StringVar(&publicAddr, "public-addr", ":8080", "Public listener address for user connections")
	cmd.Flags().StringVar(&secret, "secret", "", "Shared secret (required)")
	cmd.MarkFlagRequired("secret")

	return cmd
}

func newClientCmd() *cobra.Command {
	var serverAddr string
	var secret string
	var localAddr string

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Connect to tunnel server",
		Run: func(cmd *cobra.Command, args []string) {
			c := client.New(serverAddr, secret, localAddr)
			log.Printf("connecting to %s", serverAddr)
			if err := c.Connect(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringVar(&serverAddr, "server", "", "Server address (host:port)")
	cmd.Flags().StringVar(&secret, "secret", "", "Shared secret")
	cmd.Flags().StringVar(&localAddr, "local", "localhost:3000", "Local service address")
	cmd.MarkFlagRequired("server")
	cmd.MarkFlagRequired("secret")

	return cmd
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
