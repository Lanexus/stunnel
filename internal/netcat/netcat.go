package netcat

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Netcat struct {
	remoteAddr string
	localAddr  string
	listen     bool
	execCmd    string
}

func New(remoteAddr, localAddr string, listen bool, execCmd string) *Netcat {
	return &Netcat{
		remoteAddr: remoteAddr,
		localAddr:  localAddr,
		listen:     listen,
		execCmd:    execCmd,
	}
}

func (nc *Netcat) Run() error {
	if nc.listen {
		return nc.runServer()
	}
	return nc.runClient()
}

func (nc *Netcat) runServer() error {
	ln, err := net.Listen("tcp", nc.localAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("listening on %s", nc.localAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}

		go nc.handleConnection(conn)
	}
}

func (nc *Netcat) handleConnection(conn net.Conn) {
	defer conn.Close()

	if nc.execCmd != "" {
		nc.handleExec(conn)
		return
	}

	nc.handlePipe(conn)
}

func (nc *Netcat) handleExec(conn net.Conn) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", nc.execCmd)
	} else {
		cmd = exec.Command("sh", "-c", nc.execCmd)
	}

	cmd.Stdin = conn
	cmd.Stdout = conn
	cmd.Stderr = conn

	if err := cmd.Run(); err != nil {
		log.Printf("exec: %v", err)
	}
}

func (nc *Netcat) handlePipe(conn net.Conn) {
	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	io.Copy(conn, os.Stdin)
	<-done
}

func (nc *Netcat) runClient() error {
	conn, err := net.DialTimeout("tcp", nc.remoteAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	if nc.execCmd != "" {
		nc.handleExec(conn)
		return nil
	}

	done := make(chan struct{})
	go func() {
		io.Copy(os.Stdout, conn)
		close(done)
	}()
	io.Copy(conn, os.Stdin)
	<-done

	return nil
}
