package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("usage: sship [user@]host")
		fmt.Println()
		fmt.Println("  Copy your SSH public key to a remote host.")
		fmt.Println("  Uses the system ssh client with all your existing config.")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -h, --help    Show this help message")
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sship [user@]host")
		fmt.Fprintln(os.Stderr, "  Copy your SSH public key to a remote host.")
		fmt.Fprintln(os.Stderr, "  Uses the system ssh client with all your existing config.")
		os.Exit(1)
	}
	target := os.Args[1]

	pubPath, err := pickPubKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sship: %v\n", err)
		os.Exit(1)
	}

	key, err := os.ReadFile(pubPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sship: read %s: %v\n", pubPath, err)
		os.Exit(1)
	}
	key = bytes.TrimSpace(key)

	// Build the remote script: mkdir, chmod, append key.
	// Escape single quotes to prevent shell injection via the key string.
	escaped := strings.ReplaceAll(string(key), "'", "'\\''")
	remote := "mkdir -p ~/.ssh && chmod 700 ~/.ssh && echo '" + escaped + "' >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys && echo 'sship: key installed ✓'"

	cmd := exec.Command("ssh", target, remote)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "sship: ssh failed: %v\n", err)
		os.Exit(1)
	}
}

func pickPubKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home dir: %w", err)
	}
	sshDir := filepath.Join(home, ".ssh")

	// Preferred key types, in order
	candidates := []string{"id_ed25519.pub", "id_ecdsa.pub", "id_rsa.pub"}
	var found []string
	for _, name := range candidates {
		path := filepath.Join(sshDir, name)
		if _, err := os.Stat(path); err == nil {
			found = append(found, path)
		}
	}

	if len(found) == 0 {
		fmt.Print("No SSH key found. Generate one? [id_ed25519 / id_rsa / skip]: ")
		r := bufio.NewReader(os.Stdin)
		choice, _ := r.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if slices.Contains([]string{"skip", "n", "no", ""}, choice) {
			return "", fmt.Errorf("no key to ship")
		}
		if choice == "id_rsa" || choice == "rsa" {
			choice = "rsa"
		} else {
			choice = "ed25519"
		}
		keyPath := filepath.Join(sshDir, "id_"+choice)
		fmt.Printf("Running: ssh-keygen -t %s -f %s\n", choice, keyPath)
		cmd := exec.Command("ssh-keygen", "-t", choice, "-f", keyPath, "-N", "")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("ssh-keygen failed: %w", err)
		}
		return keyPath + ".pub", nil
	}

	if len(found) == 1 {
		return found[0], nil
	}

	// Multiple keys — let user pick
	fmt.Println("Multiple SSH keys found:")
	for i, p := range found {
		fmt.Printf("  [%d] %s\n", i+1, filepath.Base(p))
	}
	fmt.Print("Choose: ")
	var n int
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	fmt.Sscanf(line, "%d", &n)
	if n < 1 || n > len(found) {
		return "", fmt.Errorf("invalid choice")
	}
	return found[n-1], nil
}
