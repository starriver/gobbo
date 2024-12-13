package export

import (
	"io"

	"gopkg.in/yaml.v3"
)

func Run(c *ComposeConfig) error {
	cmd := command("compose", "-f", "-", "up")
	reader, writer := io.Pipe()

	go func() {
		enc := yaml.NewEncoder(writer)
		defer writer.Close()
		enc.Encode(c)
	}()

	cmd.Stdin = reader

	err := cmd.Run()

	return err
}
