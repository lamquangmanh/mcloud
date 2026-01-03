// Package installer provides system-level installation and setup for the mcloudd daemon.
// It handles copying the mcloudd binary to the system path, creating systemd service units,
// and managing the daemon lifecycle (enable/start).
package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Installation constants defining paths and service names
const (
	binaryName = "mcloudd"                           // Systemd service name
	binaryDst  = "/usr/local/bin/mcloudd"            // Destination path for the mcloudd binary
	unitPath   = "/etc/systemd/system/mcloudd.service" // Systemd unit file location
)

// Init installs the mcloudd daemon as a systemd service and starts it.
// This is the main entry point for daemon installation during cluster initialization.
//
// The function performs the following steps:
//   1. Check for root privileges (required for system-level installation)
//   2. Copy the mcloudd binary to /usr/local/bin/
//   3. Create systemd unit file at /etc/systemd/system/mcloudd.service
//   4. Reload systemd daemon to recognize the new service
//   5. Enable mcloudd to start on boot
//   6. Start the mcloudd service immediately
//
// Returns:
//   - nil if installation succeeds
//   - error if any step fails (insufficient permissions, file I/O errors, systemd errors)
//
// Example Input:
//   Called during: mcloudctl init --name my-cluster
//   Process UID: 0 (root)
//   Current executable: /home/user/mcloud/mcloudd
//
// Example Output (Success):
//   Console output:
//     ✔ copied mcloudd → /usr/local/bin/mcloudd
//     ✅ mcloudd installed and started
//   Side effects:
//     - Binary copied to /usr/local/bin/mcloudd with mode 0755
//     - Unit file created at /etc/systemd/system/mcloudd.service
//     - Service enabled: systemctl enable mcloudd
//     - Service started: systemctl start mcloudd
//     - Service status: active (running)
//
// Example Output (Error - Not Root):
//   Returns: error("must run as root")
//   Current UID: 1000 (non-root user)
//
// Example Output (Error - Binary Copy Failed):
//   Returns: error("open /usr/local/bin/mcloudd: permission denied")
func Init() error {
	// Step 1: Verify root privileges (UID 0 required)
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root")
	}

	// Step 2: Copy mcloudd binary to system path
	if err := installBinary(); err != nil {
		return err
	}

	// Step 3: Create systemd unit file
	if err := writeUnitFile(); err != nil {
		return err
	}

	// Step 4: Reload systemd to recognize new service
	if err := run("systemctl", "daemon-reload"); err != nil {
		return err
	}

	// Step 5: Enable service to start on boot
	if err := run("systemctl", "enable", binaryName); err != nil {
		return err
	}

	// Step 6: Start service immediately
	if err := run("systemctl", "start", binaryName); err != nil {
		return err
	}

	fmt.Println("✅ mcloudd installed and started")
	return nil
}

// installBinary copies the mcloudd executable to the system binary directory.
// It resolves symlinks, checks if already installed, and sets proper permissions.
//
// The function performs the following operations:
//   1. Determine the path of the current executable
//   2. Resolve any symlinks to get the real binary path
//   3. Check if binary is already installed at destination
//   4. Copy binary from source to /usr/local/bin/mcloudd
//   5. Set executable permissions (0755)
//
// Returns:
//   - nil if installation succeeds or binary already installed
//   - error if file operations fail
//
// Example Input 1 (First Installation):
//   Current executable: /home/user/mcloud/cmd/mcloudctl/mcloudctl
//   Symlink resolution: /home/user/mcloud/build/mcloudd
//   Destination exists: false
//
// Example Output 1:
//   Console: ✔ copied mcloudd → /usr/local/bin/mcloudd
//   File created: /usr/local/bin/mcloudd (mode 0755, executable by all)
//   File size: matches source binary
//
// Example Input 2 (Already Installed):
//   Current executable: /usr/local/bin/mcloudd
//   Source path: /usr/local/bin/mcloudd
//   Destination: /usr/local/bin/mcloudd
//
// Example Output 2:
//   Console: binary already installed
//   Returns: nil (no copy performed)
//
// Example Input 3 (Permission Denied):
//   Current user: non-root (UID 1000)
//   Destination: /usr/local/bin/mcloudd
//
// Example Output 3:
//   Returns: error("open /usr/local/bin/mcloudd: permission denied")
func installBinary() error {
	// Step 1: Get path of currently running executable
	src, err := os.Executable()
	if err != nil {
		return err
	}
	// Step 2: Resolve symlinks to get real binary path
	src, _ = filepath.EvalSymlinks(src)

	// Step 3: Check if binary is already installed at destination
	if src == binaryDst {
		fmt.Println("binary already installed")
		return nil
	}

	// Step 4a: Open source binary for reading
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	// Step 4b: Create destination file
	out, err := os.Create(binaryDst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Step 4c: Copy binary content from source to destination
	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	// Step 5: Set executable permissions (rwxr-xr-x)
	if err := out.Chmod(0755); err != nil {
		return err
	}

	fmt.Println("✔ copied mcloudd →", binaryDst)
	return nil
}

// writeUnitFile creates a systemd unit file for the mcloudd daemon.
// The unit file configures the daemon to start after network is available,
// restart automatically on failure, and start on boot.
//
// Unit file configuration:
//   [Unit] section:
//     - Description: Human-readable service description
//     - After: Wait for network.target before starting
//     - Wants: Prefer network-online.target (non-blocking)
//
//   [Service] section:
//     - Type: simple (process runs in foreground)
//     - ExecStart: Command to execute (/usr/local/bin/mcloudd)
//     - Restart: always (restart on any exit, including success)
//     - RestartSec: 5 seconds delay before restart
//
//   [Install] section:
//     - WantedBy: multi-user.target (start during normal boot)
//
// Returns:
//   - nil if file is created successfully
//   - error if write fails (permissions, disk full, etc.)
//
// Example Input:
//   Unit path: /etc/systemd/system/mcloudd.service
//   File exists: false (or will be overwritten)
//   User: root (UID 0)
//
// Example Output (Success):
//   File created: /etc/systemd/system/mcloudd.service
//   File mode: 0644 (rw-r--r--)
//   File content:
//     [Unit]
//     Description=mcloud daemon
//     After=network.target
//     Wants=network-online.target
//     
//     [Service]
//     Type=simple
//     ExecStart=/usr/local/bin/mcloudd
//     Restart=always
//     RestartSec=5
//     LimitNOFILE=1048576
//     
//     # Security (optional but should have)
//     NoNewPrivileges=true
//     PrivateTmp=true
//     
//     [Install]
//     WantedBy=multi-user.target
//
// Example Output (Error):
//   Returns: error("open /etc/systemd/system/mcloudd.service: permission denied")
//   Cause: Non-root user or /etc/systemd/system not writable
func writeUnitFile() error {
	// Define systemd unit file content
	// [Unit]: Service metadata and dependencies
	// [Service]: Execution configuration and restart policy
	// [Install]: Boot-time behavior
	content := `[Unit]
Description=mcloud daemon
After=network.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mcloudd
Restart=always
RestartSec=5
LimitNOFILE=1048576

# Security (optional but should have)
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
`
	// Write unit file with mode 0644 (readable by all, writable by owner)
	return os.WriteFile(unitPath, []byte(content), 0644)
}

// run executes a system command and streams its output to the current process's stdout/stderr.
// This is a helper function for executing systemctl and other system commands during installation.
//
// Parameters:
//   cmd string - The command to execute (e.g., "systemctl", "systemd-analyze")
//   args ...string - Variable arguments passed to the command
//
// Returns:
//   - nil if command exits with status code 0
//   - error if command fails or exits with non-zero status
//
// Example Input 1 (Daemon Reload):
//   cmd: "systemctl"
//   args: ["daemon-reload"]
//
// Example Output 1 (Success):
//   Command executed: systemctl daemon-reload
//   Exit code: 0
//   Returns: nil
//   Side effect: systemd reloads all unit files
//
// Example Input 2 (Enable Service):
//   cmd: "systemctl"
//   args: ["enable", "mcloudd"]
//
// Example Output 2 (Success):
//   Command executed: systemctl enable mcloudd
//   Console output:
//     Created symlink /etc/systemd/system/multi-user.target.wants/mcloudd.service
//   Exit code: 0
//   Returns: nil
//
// Example Input 3 (Start Service):
//   cmd: "systemctl"
//   args: ["start", "mcloudd"]
//
// Example Output 3 (Success):
//   Command executed: systemctl start mcloudd
//   Exit code: 0
//   Returns: nil
//   Side effect: mcloudd daemon process started (PID assigned)
//
// Example Input 4 (Service Not Found):
//   cmd: "systemctl"
//   args: ["start", "nonexistent"]
//
// Example Output 4 (Error):
//   Command executed: systemctl start nonexistent
//   Console stderr: Failed to start nonexistent.service: Unit not found.
//   Exit code: 5
//   Returns: error("exit status 5")
func run(cmd string, args ...string) error {
	// Create command with arguments
	c := exec.Command(cmd, args...)
	// Pipe command output to current process stdout
	c.Stdout = os.Stdout
	// Pipe command errors to current process stderr
	c.Stderr = os.Stderr
	// Execute command and wait for completion
	return c.Run()
}
