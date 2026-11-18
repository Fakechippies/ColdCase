package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"coldcase/pkg/server"

	"github.com/spf13/cobra"
)

// PID and log files live under ~/.coldcase so every serve mode (detached,
// foreground) is visible to the same down/status commands.
const (
	servePidFile = "serve.pid"
	serveLogFile = "serve.log"
)

func servePidPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".coldcase", servePidFile), nil
}

func serveLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".coldcase", serveLogFile), nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func confirmRemote(host string) bool {
	fmt.Printf("WARNING: binding to %q exposes the console beyond loopback.\n", host)
	fmt.Print("Anyone with network access could run forensics tools. Continue? [y/N] ")
	var reply string
	_, _ = fmt.Scanln(&reply)
	reply = strings.ToLower(strings.TrimSpace(reply))
	return reply == "y" || reply == "yes"
}

// serveState is the runtime picture used by up/status/down.
type serveState struct {
	pid       int // 0 when no pid file
	pidAlive  bool
	pidIsOurs bool
	addr      string
	listening bool
}

func readServePid() int {
	path, err := servePidPath()
	if err != nil {
		return 0
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is a fixed file under ~/.coldcase, never user input
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return 0
	}
	return pid
}

func writeServePid(pid int) error {
	path, err := servePidPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0600)
}

func removeServePid() {
	if path, err := servePidPath(); err == nil {
		_ = os.Remove(path)
	}
}

// inspectServeState reads the pid file and probes the TCP port. Both signals
// are reported so stale state can be cleaned up honestly.
func inspectServeState(host string, port int) serveState {
	st := serveState{addr: net.JoinHostPort(host, strconv.Itoa(port))}
	if pid := readServePid(); pid > 0 {
		st.pid = pid
		st.pidAlive = pidAlive(pid)
		st.pidIsOurs = !st.pidAlive || processIsColdcase(pid)
	}
	conn, err := net.DialTimeout("tcp", st.addr, 400*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		st.listening = true
	}
	return st
}

// ─── command tree ────────────────────────────────────────────────────────────

func init() {
	rootCmd.AddCommand(newServeCmd())
}

func newServeCmd() *cobra.Command {
	var (
		host        string
		port        int
		allowRemote bool
		assumeYes   bool
	)

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the local ColdCase web console",
		Long: `Start the ColdCase web console as a detached background process.

By default the console daemonizes: 'coldcase serve' returns immediately and
prints the URL, PID and log location. The console binds to 127.0.0.1 only;
binding beyond loopback is refused unless --allow-remote is given (an
interactive confirmation is still required; skip it with --yes).

Subcommands:
  coldcase serve up        Start the detached console (same as bare 'serve')
  coldcase serve run       Run in the foreground (Ctrl+C to stop)
  coldcase serve down      Gracefully stop the running console
  coldcase serve status    Report whether the console is running`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDetached(host, port, allowRemote, assumeYes)
		},
	}

	up := &cobra.Command{
		Use:   "up",
		Short: "Start the console in the background (detached)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startDetached(host, port, allowRemote, assumeYes)
		},
	}

	run := &cobra.Command{
		Use:   "run",
		Short: "Run the console in the foreground",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runForeground(host, port, allowRemote, assumeYes)
		},
	}

	down := &cobra.Command{
		Use:   "down",
		Short: "Stop the running console",
		RunE: func(cmd *cobra.Command, args []string) error {
			return stopServe()
		},
	}

	status := &cobra.Command{
		Use:   "status",
		Short: "Show whether the console is running",
		RunE: func(cmd *cobra.Command, args []string) error {
			return printServeStatus(host, port)
		},
	}

	cmd.PersistentFlags().StringVar(&host, "host", "127.0.0.1", "Interface to bind (default loopback-only)")
	cmd.PersistentFlags().IntVarP(&port, "port", "p", 8765, "Port to listen on")
	cmd.PersistentFlags().BoolVar(&allowRemote, "allow-remote", false, "Allow binding to a non-loopback interface (dangerous)")
	cmd.PersistentFlags().BoolVar(&assumeYes, "yes", false, "Skip the interactive confirmation for --allow-remote")

	cmd.AddCommand(up, run, down, status)
	return cmd
}

// validateBind refuses non-loopback binds unless explicitly overridden.
func validateBind(host string, allowRemote, assumeYes bool) error {
	if isLoopback(host) {
		return nil
	}
	if !allowRemote {
		return fmt.Errorf("refusing to bind to non-loopback address %q; pass --allow-remote to override", host)
	}
	if !assumeYes && !confirmRemote(host) {
		return errors.New("aborted")
	}
	return nil
}

// ─── up ──────────────────────────────────────────────────────────────────────

// startDetached spawns 'coldcase serve run' as a new session, detached from
// the current terminal, with output captured to ~/.coldcase/serve.log.
func startDetached(host string, port int, allowRemote, assumeYes bool) error {
	if err := validateBind(host, allowRemote, assumeYes); err != nil {
		return err
	}

	st := inspectServeState(host, port)
	if st.listening {
		if st.pid > 0 && st.pidAlive {
			return fmt.Errorf("console already running (pid %d) at http://%s", st.pid, st.addr)
		}
		return fmt.Errorf("something is already listening on %s — not started by 'coldcase serve up'", st.addr)
	}
	if st.pid > 0 && !st.pidAlive {
		removeServePid() // stale pid file
	}

	logPath, err := serveLogPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		return err
	}
	logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) //nolint:gosec // logPath is a fixed file under ~/.coldcase
	if err != nil {
		return fmt.Errorf("could not open log file: %w", err)
	}
	defer func() { _ = logFile.Close() }()

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	childArgs := []string{"serve", "run", "--host", host, "--port", strconv.Itoa(port)}
	if allowRemote {
		childArgs = append(childArgs, "--allow-remote")
	}
	if assumeYes {
		childArgs = append(childArgs, "--yes")
	}

	child := exec.Command(exe, childArgs...) //nolint:gosec // exe is os.Executable(); we re-exec ourselves in a new session
	child.Stdout = logFile
	child.Stderr = logFile
	child.Stdin = nil
	detachProcess(child)
	if err := child.Start(); err != nil {
		return fmt.Errorf("could not start console: %w", err)
	}
	pid := child.Process.Pid
	if err := writeServePid(pid); err != nil {
		return fmt.Errorf("console started but the pid file could not be written: %w", err)
	}

	// Give the child a moment to bind; fail loudly if it dies instantly.
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !pidAlive(pid) {
			removeServePid()
			return fmt.Errorf("console exited during startup — see %s", logPath)
		}
		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			fmt.Printf("[*] ColdCase console running at http://%s (pid %d)\n", addr, pid)
			fmt.Printf("    Log: %s — stop with 'coldcase serve down'\n", logPath)
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}
	fmt.Printf("[*] ColdCase console starting (pid %d); not responding yet — check %s\n", pid, logPath)
	return nil
}

// ─── run (foreground) ────────────────────────────────────────────────────────

func runForeground(host string, port int, allowRemote, assumeYes bool) error {
	if err := validateBind(host, allowRemote, assumeYes); err != nil {
		return err
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	srv, err := server.New(addr)
	if err != nil {
		return err
	}

	if err := writeServePid(os.Getpid()); err != nil {
		fmt.Printf("[!] could not write pid file: %v\n", err)
	}
	defer removeServePid()

	// Graceful shutdown on SIGINT/SIGTERM so 'coldcase serve down' also works
	// against a foreground instance.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	fmt.Printf("[*] ColdCase console listening on http://%s\n", addr)
	if isLoopback(host) {
		fmt.Println("    Local-only access. Press Ctrl+C to stop.")
	} else {
		fmt.Println("    WARNING: exposed beyond loopback. Anyone with network access can run tools.")
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	fmt.Println("[*] ColdCase console stopped.")
	return nil
}

// ─── down ────────────────────────────────────────────────────────────────────

func stopServe() error {
	pid := readServePid()
	if pid == 0 {
		fmt.Println("Console is not running (no pid file).")
		return nil
	}
	if !pidAlive(pid) {
		removeServePid()
		fmt.Printf("Stale pid file (pid %d is gone) — cleaned up. Console was not running.\n", pid)
		return nil
	}
	if !processIsColdcase(pid) {
		return fmt.Errorf("pid %d does not look like a ColdCase process; refusing to stop it", pid)
	}

	fmt.Printf("Stopping console (pid %d)…\n", pid)
	terminateProcess(pid) // SIGTERM

	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		if !pidAlive(pid) {
			removeServePid()
			fmt.Println("[*] ColdCase console stopped.")
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}

	killProcess(pid) // escalate
	time.Sleep(500 * time.Millisecond)
	removeServePid()
	if pidAlive(pid) {
		return fmt.Errorf("pid %d did not exit even after SIGKILL", pid)
	}
	fmt.Println("[*] ColdCase console stopped (forced).")
	return nil
}

// ─── status ──────────────────────────────────────────────────────────────────

func printServeStatus(host string, port int) error {
	st := inspectServeState(host, port)

	switch {
	case st.listening && st.pid > 0 && st.pidAlive:
		fmt.Printf("[+] running — http://%s (pid %d)\n", st.addr, st.pid)
	case st.listening:
		fmt.Printf("[+] listening on %s, but not managed by 'coldcase serve' (no live pid file)\n", st.addr)
	case st.pid > 0 && st.pidAlive:
		fmt.Printf("[~] pid %d is alive but %s is not answering — it may still be starting up\n", st.pid, st.addr)
	default:
		fmt.Printf("[-] not running (%s is free)\n", st.addr)
		if st.pid > 0 {
			fmt.Printf("    Removing stale pid file (pid %d).\n", st.pid)
			removeServePid()
		}
	}
	return nil
}
