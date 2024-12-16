package export

import (
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	glog.Debugf("Export: running %v", args)
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

func BuildImage(always bool) error {
	var cmd *exec.Cmd
	if !always {
		cmd = command("images", "-qf", "reference="+Tag)
		output, err := cmd.Output()
		if err != nil {
			return err
		}

		if len(output) != 0 {
			glog.Debugf(
				"Image with tag '%s' already built: %s",
				Tag, string(output),
			)
			return nil
		}
	}

	glog.Infof("Building export image with tag '%s'", Tag)

	// Create a temporary build context.
	ctx, err := os.MkdirTemp("", "gobbo-image-build")
	if err != nil {
		return err
	}
	defer os.RemoveAll(ctx)

	// Ignoring error here, we know this exists:
	files, _ := config.ReadDir("config")

	// Copy the embedFS into the build context.
	for _, f := range files {
		name := f.Name()

		src, _ := config.Open(filepath.Join("config", name))

		var mode os.FileMode = 0644
		if strings.HasSuffix(name, ".sh") {
			mode = 0755
		}
		dest, err := os.OpenFile(filepath.Join(ctx, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	cmd = command("build", "-t", Tag, ctx)
	return cmd.Run()
}
