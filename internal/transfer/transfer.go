package transfer

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type FileTransfer struct {
	remoteAddr string
	localPath  string
	listen     bool
}

func New(remoteAddr, localPath string, listen bool) *FileTransfer {
	return &FileTransfer{
		remoteAddr: remoteAddr,
		localPath:  localPath,
		listen:     listen,
	}
}

func (ft *FileTransfer) SetLocalPath(path string) {
	ft.localPath = path
}

func (ft *FileTransfer) Send() error {
	conn, err := net.DialTimeout("tcp", ft.remoteAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	file, err := os.Open(ft.localPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	filename := filepath.Base(ft.localPath)
	header := fmt.Sprintf("%s\n%d\n", filename, stat.Size())
	if _, err := conn.Write([]byte(header)); err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	n, err := io.Copy(conn, file)
	if err != nil {
		return fmt.Errorf("send file: %w", err)
	}

	log.Printf("sent %s (%d bytes)", filename, n)
	return nil
}

func (ft *FileTransfer) Receive() error {
	ln, err := net.Listen("tcp", ft.remoteAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Printf("listening on %s", ft.remoteAddr)

	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("accept: %w", err)
	}
	defer conn.Close()

	var filename string
	var size int64
	if _, err := fmt.Fscanf(conn, "%s\n%d\n", &filename, &size); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	if ft.localPath != "" {
		filename = filepath.Join(ft.localPath, filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	n, err := io.CopyN(file, conn, size)
	if err != nil {
		return fmt.Errorf("receive file: %w", err)
	}

	log.Printf("received %s (%d bytes)", filename, n)
	return nil
}
