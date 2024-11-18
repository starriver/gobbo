package store

import (
	"path/filepath"

	"github.com/starriver/gobbo/pkg/platform"
)

type Store struct {
	Root     string
	Platform platform.Platform
}

type dir map[string]any

// Blank string means "a file should exist here".
// Non-blank string means "a file should exist here with these contents".
var schema = dir{
	"version": "1",
	"cache": dir{
		"stable": "",
		"latest": "",
	},
	"bin": dir{
		"official": dir{},
	},
	"tmp": dir{},
}

func (s *Store) Join(elem ...string) string {
	return filepath.Join(append([]string{s.Root}, elem...)...)
}
