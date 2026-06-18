package cloudflare

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const (
	cloudflaredURL = "https://github.com/cloudflare/cloudflared/releases/latest/download"
)

type Tunnel struct {
	cmd     *exec.Cmd
	url     string
	localAddr string
}

func NewTunnel(localAddr string) *Tunnel {
	return &Tunnel{
		localAddr: localAddr,
	}
}

func (t *Tunnel) Start() error {
	// Check if cloudflared is installed
	binary, err := ensureCloudflared()
	if err != nil {
		return fmt.Errorf("ensure cloudflared: %w", err)
	}

	// Start cloudflared tunnel
	t.cmd = exec.Command(binary, "tunnel", "--url", "http://"+t.localAddr)
	
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := t.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("start cloudflared: %w", err)
	}

	// Parse output to find tunnel URL
	urlCh := make(chan string, 1)
	errCh := make(chan error, 1)

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("[cloudflared] %s", line)
			
			// Look for the tunnel URL
			if strings.Contains(line, ".trycloudflare.com") {
				re := regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)
				match := re.FindString(line)
				if match != "" {
					urlCh <- match
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- err
		}
		errCh <- fmt.Errorf("cloudflared exited without providing URL")
	}()

	select {
	case url := <-urlCh:
		t.url = url
		return nil
	case err := <-errCh:
		return err
	}
}

func (t *Tunnel) URL() string {
	return t.url
}

func (t *Tunnel) Stop() {
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}
}

func ensureCloudflared() (string, error) {
	// Check if already in PATH
	if path, err := exec.LookPath("cloudflared"); err == nil {
		return path, nil
	}

	// Check local binary
	localBin := filepath.Join(".", "cloudflared")
	if runtime.GOOS == "windows" {
		localBin += ".exe"
	}
	if _, err := os.Stat(localBin); err == nil {
		return localBin, nil
	}

	// Download
	log.Printf("downloading cloudflared...")
	if err := downloadCloudflared(localBin); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	return localBin, nil
}

func downloadCloudflared(dest string) error {
	var filename string
	switch runtime.GOOS {
	case "linux":
		filename = "cloudflared-linux-amd64"
	case "darwin":
		filename = "cloudflared-darwin-amd64"
	case "windows":
		filename = "cloudflared-windows-amd64.exe"
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	url := cloudflaredURL + "/" + filename
	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		os.Chmod(dest, 0755)
	}

	return nil
}
