package logger

import (
	"log"
	"os"

	"golang.org/x/term"
)

/*
	========================
	Terminal / Color Detect
	========================
*/

// isTerminal checks if the standard output is connected to a terminal (TTY).
// This is used to determine whether to enable colored output.
// Colors are only enabled when output goes to a terminal, not when piped or redirected to a file.
//
// Returns:
//   - true if stdout is a terminal (supports ANSI colors)
//   - false if stdout is redirected to a file or pipe
//
// Example Output (Terminal):
//   true  // Running in terminal: ./app
//
// Example Output (Piped):
//   false // Running with pipe: ./app | tee output.log
//
// Example Output (Redirected):
//   false // Running with redirect: ./app > output.log
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

var (
	reset  = "" // ANSI reset code (clears all formatting)
	red    = "" // ANSI red color code for errors
	yellow = "" // ANSI yellow color code for warnings
	green  = "" // ANSI green color code for info
	cyan   = "" // ANSI cyan color code for debug
)

// initColors initializes ANSI color codes if output is to a terminal.
// If output is not a terminal (e.g., piped or redirected), colors remain empty strings.
// This prevents ANSI escape codes from appearing in log files.
//
// Side Effect:
//   Sets global color variables:
//   - reset  = "\033[0m"  (if terminal)
//   - red    = "\033[31m" (if terminal)
//   - yellow = "\033[33m" (if terminal)
//   - green  = "\033[32m" (if terminal)
//   - cyan   = "\033[36m" (if terminal)
//
// Example (Terminal):
//   Before: reset = "", red = "", yellow = "", green = "", cyan = ""
//   After:  reset = "\033[0m", red = "\033[31m", yellow = "\033[33m", etc.
//
// Example (Non-Terminal):
//   Before: reset = "", red = "", yellow = "", green = "", cyan = ""
//   After:  reset = "", red = "", yellow = "", green = "", cyan = "" (unchanged)
func initColors() {
	// Only enable colors when output is a terminal
	if !isTerminal() {
		return
	}

	// Set ANSI escape codes for colors
	reset = "\033[0m"   // Reset all attributes
	red = "\033[31m"    // Red text
	yellow = "\033[33m" // Yellow text
	green = "\033[32m"  // Green text
	cyan = "\033[36m"   // Cyan text
}

/*
	========================
	Logger (Production Safe)
	========================
*/

var (
	infoLog  *log.Logger // Logger for informational messages (green prefix)
	warnLog  *log.Logger // Logger for warning messages (yellow prefix)
	errorLog *log.Logger // Logger for error messages (red prefix)
	debugLog *log.Logger // Logger for debug messages (cyan prefix)
)

// initLogger initializes all logger instances with colored prefixes.
// Each logger is configured with timestamps and appropriate output streams:
//   - INFO and DEBUG: stdout (standard output)
//   - WARN and ERROR: stderr (standard error)
//
// This function must be called before using any logging functions (Info, Warn, Error, Debug).
// Colors are automatically disabled if output is not a terminal.
//
// Side Effect:
//   Initializes global logger variables:
//   - infoLog:  [INFO]  prefix in green  → stdout
//   - warnLog:  [WARN]  prefix in yellow → stderr
//   - errorLog: [ERROR] prefix in red    → stderr
//   - debugLog: [DEBUG] prefix in cyan   → stdout
//
// Example Output (Terminal):
//   infoLog  prints: "\033[32m[INFO] \033[0m2026-01-02 10:30:45 message"
//   warnLog  prints: "\033[33m[WARN] \033[0m2026-01-02 10:30:45 message"
//   errorLog prints: "\033[31m[ERROR] \033[0m2026-01-02 10:30:45 message"
//   debugLog prints: "\033[36m[DEBUG] \033[0m2026-01-02 10:30:45 message"
//
// Example Output (Non-Terminal):
//   infoLog  prints: "[INFO] 2026-01-02 10:30:45 message"
//   warnLog  prints: "[WARN] 2026-01-02 10:30:45 message"
//   errorLog prints: "[ERROR] 2026-01-02 10:30:45 message"
//   debugLog prints: "[DEBUG] 2026-01-02 10:30:45 message"
func InitLogger() {
	// Initialize color codes first
	initColors()

	// Configure logger flags: timestamp + message prefix
	flags := log.LstdFlags | log.Lmsgprefix

	// Create INFO logger (stdout, green prefix)
	infoLog = log.New(
		os.Stdout,
		green+"[INFO] "+reset,
		flags,
	)

	// Create WARN logger (stderr, yellow prefix)
	warnLog = log.New(
		os.Stderr,
		yellow+"[WARN] "+reset,
		flags,
	)

	// Create ERROR logger (stderr, red prefix)
	errorLog = log.New(
		os.Stderr,
		red+"[ERROR] "+reset,
		flags,
	)

	// Create DEBUG logger (stdout, cyan prefix)
	debugLog = log.New(
		os.Stdout,
		cyan+"[DEBUG] "+reset,
		flags,
	)
}

/*
	========================
	Public Logging API
	========================
*/

// Info logs an informational message to stdout with a green [INFO] prefix.
// Automatically formats the message using fmt.Sprintf if format specifiers are present.
// Uses the global infoLog logger which must be initialized via InitLogger() function.
//
// Parameters:
//   msg string - Format string (as in fmt.Printf)
//   v ...any - Variable arguments for format string
//
// Output: Writes to stdout with format:
//   [INFO] YYYY-MM-DD HH:MM:SS formatted_message
//
// Example Input 1:
//   Info("Cluster initialized successfully")
//
// Example Output 1 (Terminal):
//   \033[32m[INFO] \033[0m2026-01-02 10:30:45 Cluster initialized successfully
//
// Example Output 1 (Non-Terminal):
//   [INFO] 2026-01-02 10:30:45 Cluster initialized successfully
//
// Example Input 2:
//   Info("Server listening on %s:%d", "127.0.0.1", 9028)
//
// Example Output 2:
//   [INFO] 2026-01-02 10:30:45 Server listening on 127.0.0.1:9028
func Info(msg string, v ...any) {
	infoLog.Printf(msg, v...)
}

// Warn logs a warning message to stderr with a yellow [WARN] prefix.
// Use for non-critical issues that should be noticed but don't prevent execution.
// Automatically formats the message using fmt.Sprintf if format specifiers are present.
//
// Parameters:
//   msg string - Format string (as in fmt.Printf)
//   v ...any - Variable arguments for format string
//
// Output: Writes to stderr with format:
//   [WARN] YYYY-MM-DD HH:MM:SS formatted_message
//
// Example Input 1:
//   Warn("LXD not available, using mock client")
//
// Example Output 1 (Terminal):
//   \033[33m[WARN] \033[0m2026-01-02 10:30:45 LXD not available, using mock client
//
// Example Output 1 (Non-Terminal):
//   [WARN] 2026-01-02 10:30:45 LXD not available, using mock client
//
// Example Input 2:
//   Warn("Failed to detect LAN interface, falling back to %s", "127.0.0.1")
//
// Example Output 2:
//   [WARN] 2026-01-02 10:30:45 Failed to detect LAN interface, falling back to 127.0.0.1
func Warn(msg string, v ...any) {
	warnLog.Printf(msg, v...)
}

// Error logs an error message to stderr with a red [ERROR] prefix.
// Use for errors that cause operations to fail but don't crash the program.
// Automatically formats the message using fmt.Sprintf if format specifiers are present.
//
// Parameters:
//   msg string - Format string (as in fmt.Printf)
//   v ...any - Variable arguments for format string
//
// Output: Writes to stderr with format:
//   [ERROR] YYYY-MM-DD HH:MM:SS formatted_message
//
// Example Input 1:
//   Error("Failed to initialize database: %v", err)
//   // where err = errors.New("file locked")
//
// Example Output 1 (Terminal):
//   \033[31m[ERROR] \033[0m2026-01-02 10:30:45 Failed to initialize database: file locked
//
// Example Output 1 (Non-Terminal):
//   [ERROR] 2026-01-02 10:30:45 Failed to initialize database: file locked
//
// Example Input 2:
//   Error("Connection refused on %s:%d", "127.0.0.1", 9028)
//
// Example Output 2:
//   [ERROR] 2026-01-02 10:30:45 Connection refused on 127.0.0.1:9028
func Error(msg string, v ...any) {
	errorLog.Printf(msg, v...)
}

// Debug logs a debug message to stdout with a cyan [DEBUG] prefix.
// Use for detailed diagnostic information during development and troubleshooting.
// Automatically formats the message using fmt.Sprintf if format specifiers are present.
//
// Parameters:
//   msg string - Format string (as in fmt.Printf)
//   v ...any - Variable arguments for format string
//
// Output: Writes to stdout with format:
//   [DEBUG] YYYY-MM-DD HH:MM:SS formatted_message
//
// Example Input 1:
//   Debug("Processing request with ID: %s", "req-12345")
//
// Example Output 1 (Terminal):
//   \033[36m[DEBUG] \033[0m2026-01-02 10:30:45 Processing request with ID: req-12345
//
// Example Output 1 (Non-Terminal):
//   [DEBUG] 2026-01-02 10:30:45 Processing request with ID: req-12345
//
// Example Input 2:
//   Debug("Transaction started, isolation level: %s", "READ COMMITTED")
//
// Example Output 2:
//   [DEBUG] 2026-01-02 10:30:45 Transaction started, isolation level: READ COMMITTED
func Debug(msg string, v ...any) {
	debugLog.Printf(msg, v...)
}
