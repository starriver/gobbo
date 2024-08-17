package template

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"gitlab.com/starriver/gobbo/pkg/godot"
)

//go:embed default/*
var defaultTemplate embed.FS

func Generate(src, dest string, godot *godot.Official, bare bool) error {
	var srcFS fs.FS = defaultTemplate
	srcRoot := "default"
	if src != "" {
		srcFS = os.DirFS(src)
		srcRoot = "."
	}

	tempDir, err := os.MkdirTemp("", "gobbo-template")
	if err != nil {
		return fmt.Errorf("couldn't create temporary directory: %v", err)
	}
	// TODO:
	// defer os.RemoveAll(tempDir)

	data := map[string]interface{}{
		"Project": path.Base(dest),
		"Godot":   godot.String(),
		"Bare":    bare,
	}

	err = fs.WalkDir(srcFS, srcRoot, func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(srcRoot, srcPath)
		if err != nil {
			return err
		}

		destPath := filepath.Join(tempDir, relative)

		// Ignore directories - these are created via os.MkdirAll before writes.
		if d.IsDir() {
			return nil
		}

		isTemplate := filepath.Ext(destPath) == ".tmpl"
		if isTemplate {
			// Remove '.tmpl' extension.
			destPath = destPath[:len(destPath)-5]
		}

		os.MkdirAll(filepath.Dir(destPath), 0755)

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		// For non-templated files, short-circuit to a straight copy.
		if !isTemplate {
			srcFile, err := srcFS.Open(srcPath)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return err
			}

			return nil
		}

		content, err := fs.ReadFile(srcFS, srcPath)
		if err != nil {
			return err
		}

		tmplName := filepath.Base(destPath)
		tmpl, err := template.New(tmplName).Parse(string(content))
		if err != nil {
			return err
		}

		// Now actually template the file.
		if err = tmpl.Execute(destFile, data); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("templating failed: %v", err)
	}

	if err = os.Rename(tempDir, dest); err != nil {
		return fmt.Errorf("couldn't move '%s' to '%s': %v", tempDir, dest, err)
	}

	return nil
}
