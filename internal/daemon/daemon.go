package daemon

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const pidFile = ".stunnel.pid"

func Daemonize() error {
	if isDaemon() {
		return nil
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}

	args := append([]string{exe}, os.Args[1:]...)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), "STUNNEL_DAEMON=1")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}

	pidPath := getPIDPath()
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		return fmt.Errorf("write pid: %w", err)
	}

	fmt.Printf("Daemon started with PID %d\n", cmd.Process.Pid)
	os.Exit(0)
	return nil
}

func isDaemon() bool {
	return os.Getenv("STUNNEL_DAEMON") == "1"
}

func getPIDPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "stunnel", pidFile)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "stunnel", pidFile)
}

func StopDaemon() error {
	pidPath := getPIDPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return fmt.Errorf("parse pid: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("kill process: %w", err)
	}

	os.Remove(pidPath)
	fmt.Printf("Daemon stopped (PID %d)\n", pid)
	return nil
}

func IsRunning() bool {
	pidPath := getPIDPath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return false
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	err = proc.Signal(os.Kill)
	return err == nil
}

func SetupLogging(logFile string) {
	if logFile == "" {
		return
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("open log file: %v", err)
		return
	}

	log.SetOutput(f)
}
