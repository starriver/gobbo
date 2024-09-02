package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"gitlab.com/starriver/gobbo/pkg/glog"
)

// Replace the currently running process. Used to run Godot in the foreground.
// If the syscall fails, exits the program immediately.
func Execv(binPath string, args []string) {
	es := execStr(binPath, args)
	glog.Debugf("execv: %s", es)

	env := os.Environ()
	err := syscall.Exec(binPath, args, env)
	if err != nil {
		glog.Errorf("Couldn't execute %s: %v", es, err)
		os.Exit(1)
	}
}

// Run a process. If it continues running after a few seconds, returns nil -
// otherwise show its output and error.
func Runway(binPath string, args []string) error {
	es := execStr(binPath, args)
	glog.Debugf("Attempting to execute: %s", es)

	cmd := exec.Command(binPath, args...)

	// On Linux, Godot can start behaving very weirdly if its pipes to stdout
	// or stderr are broken. So, just give it tempfiles.
	stdout, err := os.CreateTemp("", "gobbo-runway-*")
	if err != nil {
		return err
	}
	stderr, err := os.CreateTemp("", "gobbo-runway-*")
	if err != nil {
		return err
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		glog.Errorf("Couldn't execute %s: %v", es, err)
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	timer := time.NewTimer(2 * time.Second)

	select {
	case <-timer.C:
		cmd.Process.Release()
		glog.Debug("Godot seems OK, exiting")
		// Note that the stdout/stderr tempfiles will unfortunately remain.
		return nil
	case <-done:
		// Process exited.
	}

	glog.Errorf("Exited %d from '%s'. Output:", cmd.ProcessState.ExitCode(), es)

	displayAndClose := func(from, to *os.File) error {
		_, err := from.Seek(0, 0)
		if err != nil {
			return err
		}
		_, err = to.ReadFrom(from)
		if err != nil {
			return err
		}
		err = from.Close()
		if err != nil {
			return err
		}
		err = os.Remove(from.Name())
		if err != nil {
			glog.Warn(err)
		}

		return nil
	}

	displayAndClose(stdout, os.Stdout)
	displayAndClose(stderr, os.Stderr)

	return errors.New("process exited too fast")
}

func execStr(binPath string, args []string) string {
	if len(args) == 0 {
		return fmt.Sprintf("'%s'", binPath)
	}
	return fmt.Sprintf("'%s' '%s'", binPath, strings.Join(args, "' '"))
}
