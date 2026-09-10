// Command preamble is the .pex preamble for Windows: the executable stub that please_pex
// prepends to a .pex archive so that running the .pex runs a Python interpreter on it.
//
// It is the counterpart of the C preamble used on every other platform, and reads the same
// configuration from the same place inside the archive. It is a separate program rather than a
// port of the C one because Windows has no exec(): the C runtime's _execv returns to its caller
// instead of replacing the process, which loses both the child's exit code and its attachment to
// the console. This runs the interpreter as a child process and passes its exit status on.
package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/please-build/python-rules/tools/please_pex/preamble"
)

// plzBinPathVar is the placeholder that an interpreter path may start with to mean "wherever
// Please's binary outputs are", resolved at run time by plzBinPath.
const plzBinPathVar = "$PLZ_BIN_PATH"

func main() {
	pex, err := os.Executable()
	if err != nil {
		logf(preamble.LogFatal, "Failed to get path to .pex file: %s", err)
		os.Exit(1)
	}
	// Resolve any symlinks, so that a .pex reached through one still finds itself.
	if resolved, err := filepath.EvalSymlinks(pex); err == nil {
		pex = resolved
	}

	config, err := readConfig(pex)
	if err != nil {
		logf(preamble.LogFatal, "Failed to get .pex preamble configuration: %s", err)
		os.Exit(1)
	}
	setVerbosity(config.Verbosity)

	interps, err := interpreters(config, pex)
	if err != nil {
		logf(preamble.LogFatal, "Failed to get interpreters from .pex preamble configuration: %s", err)
		os.Exit(1)
	}

	// The interpreter is invoked as `<interpreter> <configured args> <this .pex> <our args>`,
	// which is the same command line the C preamble builds.
	args := append([]string{}, config.InterpreterArgs...)
	args = append(args, pex)
	args = append(args, os.Args[1:]...)

	// Ctrl-C is delivered to every process attached to the console, so the interpreter gets it
	// too. Ignoring it here keeps this process alive long enough to report the interpreter's exit
	// status rather than dying first and leaving it running against a console we no longer own.
	signal.Ignore(os.Interrupt)

	for _, interpreter := range interps {
		logf(preamble.LogDebug, "Attempting to execute interpreter %s", interpreter)
		if code, ran := run(interpreter, args); ran {
			os.Exit(code)
		}
	}

	logf(preamble.LogFatal, "Failed to execute any interpreters in the interpreter search path")
	os.Exit(1)
}

// run runs the given interpreter and returns its exit code, and whether it ran at all. A false
// second return means the caller should try the next interpreter: not existing is an ordinary
// outcome here, since the interpreter list is a search path.
func run(interpreter string, args []string) (int, bool) {
	cmd := exec.Command(interpreter, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0, true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true
	}
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
		// Not an error case: there are legitimate reasons for any given interpreter not to be
		// installed, which is why more than one can be configured.
		logf(preamble.LogInfo, "%s does not exist or could not be executed", interpreter)
		return 0, false
	}
	logf(preamble.LogError, "Failed to execute %s: %s", interpreter, err)
	return 0, false
}

// readConfig reads and parses the preamble configuration held inside the .pex archive.
func readConfig(pex string) (*preamble.Config, error) {
	r, err := zip.OpenReader(pex)
	if err != nil {
		return nil, fmt.Errorf("open .pex: %w", err)
	}
	defer r.Close()
	f, err := r.Open(preamble.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("open .pex configuration: %w", err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read .pex configuration: %w", err)
	}
	config := &preamble.Config{}
	if err := json.Unmarshal(b, config); err != nil {
		return nil, fmt.Errorf("parse .pex configuration: %w", err)
	}
	return config, nil
}

// interpreters returns the interpreter search path from the configuration, with any
// $PLZ_BIN_PATH prefix resolved. Interpreters that cannot be resolved are dropped rather than
// being an error, so long as at least one remains.
func interpreters(config *preamble.Config, pex string) ([]string, error) {
	if len(config.Interpreters) == 0 {
		return nil, errors.New("interpreters must not be empty")
	}
	binPath, resolved := "", false
	interpreters := make([]string, 0, len(config.Interpreters))
	for _, interpreter := range config.Interpreters {
		if rest, ok := strings.CutPrefix(interpreter, plzBinPathVar+"/"); ok {
			if !resolved {
				resolved = true
				var err error
				if binPath, err = plzBinPath(pex); err != nil {
					return nil, fmt.Errorf("$PLZ_BIN_PATH resolution failure: %w", err)
				} else if binPath == "" {
					logf(preamble.LogWarn, ".pex file is not inside a Please repo; omitting interpreters prepended with '%s'", plzBinPathVar)
				} else {
					logf(preamble.LogDebug, "Resolved %s to %s", plzBinPathVar, binPath)
				}
			}
			if binPath == "" {
				logf(preamble.LogInfo, "Omitting interpreter from search path: %s", interpreter)
				continue
			}
			// The rest of the configured path is /-separated, because it came from a build
			// definition; filepath.Join is what turns it into a Windows path.
			interpreter = filepath.Join(binPath, filepath.FromSlash(rest))
		}
		logf(preamble.LogDebug, "Added interpreter to search path: %s", interpreter)
		interpreters = append(interpreters, interpreter)
	}
	if len(interpreters) == 0 {
		return nil, errors.New("interpreters list contains no resolvable paths")
	}
	return interpreters, nil
}

// plzBinPath returns the directory holding Please's binary outputs for the repo this .pex
// belongs to, or an empty string if it does not belong to one.
//
// Inside a build environment that is the environment's own root, which Please gives us as
// TMP_DIR. Outside one it is the bin/ directory of the plz-out/ this .pex was found in.
//
// N.B. these are real paths on disk, so filepath is the right package here - unlike the paths
// that come out of the archive or out of a build definition.
func plzBinPath(pex string) (string, error) {
	dir := filepath.Dir(pex)
	if os.Getenv("PLZ_ENV") != "" {
		tmpDir := os.Getenv("TMP_DIR")
		if tmpDir == "" {
			// Either TMP_DIR was unset before this ran, or PLZ_ENV is set for some unrelated
			// reason. The latter is far more likely; almost every build definition would break
			// without TMP_DIR.
			logf(preamble.LogWarn, "PLZ_ENV is defined but TMP_DIR is not; assuming this is not a Please build environment")
		} else {
			root, err := filepath.EvalSymlinks(tmpDir)
			if err != nil {
				return "", fmt.Errorf("resolve TMP_DIR: %w", err)
			}
			if dir == root || strings.HasPrefix(dir, root+string(filepath.Separator)) {
				return root, nil
			}
		}
	}
	for d := dir; ; {
		if filepath.Base(d) == "plz-out" {
			return filepath.Join(d, "bin"), nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", nil
		}
		d = parent
	}
}

// levels orders the logging levels, lowest first. A level's position in it is what logf
// compares; anything not in it is treated as error.
var levels = []preamble.Verbosity{
	preamble.LogTrace, preamble.LogDebug, preamble.LogInfo,
	preamble.LogWarn, preamble.LogError, preamble.LogFatal,
}

// errorLevel is the position of the error level in levels, which is the fallback everywhere.
const errorLevel = 4

// verbosity is the position in levels of the lowest level that logf will print.
var verbosity = errorLevel

// setVerbosity sets the minimum logging level from PLZ_PEX_PREAMBLE_VERBOSITY, falling back to
// the level from the .pex configuration and then to error.
func setVerbosity(configured preamble.Verbosity) {
	verbosity = index(configured)
	env := os.Getenv("PLZ_PEX_PREAMBLE_VERBOSITY")
	if env == "" {
		return
	}
	var level preamble.Verbosity
	if err := level.UnmarshalFlag(env); err != nil {
		logf(preamble.LogError, "Unknown logging level '%s'; defaulting to '%s'", env, levels[verbosity])
		return
	}
	verbosity = index(level)
}

// index returns the position of a level in the ordering, or that of error if it is unknown.
func index(level preamble.Verbosity) int {
	for i, l := range levels {
		if l == level {
			return i
		}
	}
	return errorLevel
}

// logf writes a message to stderr if the given level is at least the current verbosity.
func logf(level preamble.Verbosity, msg string, args ...interface{}) {
	if index(level) < verbosity {
		return
	}
	fmt.Fprintf(os.Stderr, "%s %-5s %s\n", time.Now().Format("2006-01-02 15:04:05"), strings.ToUpper(string(level)), fmt.Sprintf(msg, args...))
}
