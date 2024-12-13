package export

import (
	"embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/starriver/gobbo/pkg/glog"
)

// Very bespoke Docker Compose CLI client.

const Tag = "starriver.run/gobbo:v1"

//go:embed config/*
var config embed.FS

// This is cached as a teensy tiny optimization in CheckEnvironment().
var cliPath = "docker"

// Create a Docker command and show stderr.
func command(args ...string) *exec.Cmd {
	cmd := exec.Command(cliPath, args...)
	cmd.Stderr = os.Stderr
	return cmd
}

func CheckEnvironment() error {
	// Ensure CLI is in PATH. cliPath is cached here.
	var err error
	cliPath, err = exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("Docker CLI not available: %v", err)
	}

	// Check server connectivity (with the cheapest command I can think of)
	cmd := command("ps", "-ql")
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Docker server connectivity check failed: %v", err)
	}

	// Finally, check Compose plugin availability.
	cmd = command("compose", "version")
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("Docker Compose CLI check failed: %v", err)
	}

	return nil
}

func BuildImage() error {
	const tagRef = "reference=" + Tag

	cmd := command("images", "-qf", tagRef)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if len(output) != 0 {
		glog.Debugf(
			"Image with tag '%s' already built: %s",
			tagRef, string(output),
		)
		return nil
	}

	cmd = command("build", "-t", tagRef, "-")

	// We know this file is embedded, so ignoring error here.
	f, _ := config.Open("Dockerfile")
	cmd.Stdin = f

	return cmd.Run()
}
