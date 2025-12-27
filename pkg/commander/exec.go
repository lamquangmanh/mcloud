package commander

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
)

// ExecCommand runs an external command and returns its output or an error
func ExecCommand(name string, args ...string) (string, error) {
	// define command and arguments
	cmd := exec.Command(name, args...)

	// capture output and error
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// run command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %s: %s", err.Error(), stderr.String())
	}

	return out.String(), nil
}

func CheckCommandExists(cmd string) error {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmd)
	}
	return nil
}

func CheckPortAvailable(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("port %d is not available", port)
	}
	_ = ln.Close()
	return nil
}

func CheckDiskExists(path string) error {
	_, err := exec.Command("lsblk", path).Output(); 
	if err != nil {
		return fmt.Errorf("disk not found or not accessible: %s", path)
	}
	return nil
}