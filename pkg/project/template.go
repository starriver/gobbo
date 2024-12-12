package project

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/starriver/gobbo/pkg/godot"
	"github.com/starriver/gobbo/pkg/store"
)

//go:embed template/*
var defaultTemplate embed.FS

func Generate(src, dest string, store *store.Store, godot *godot.Official, bare bool) error {
	var srcFS fs.FS = defaultTemplate
	srcRoot := "template"
	if src != "" {
		if bare {
			panic("can't use a custom template when bare is true")
		}

		srcFS = os.DirFS(src)
		srcRoot = "."
	}

	data := map[string]interface{}{
		"Project": path.Base(dest),
		"Godot":   godot.String(),
		"Bare":    bare,
	}

	if bare {
		f, err := os.CreateTemp("", "gobbo-template")
		if err != nil {
			return err
		}
		// Note: defer f.Close() omitted.
		filename := f.Name()

		content, err := fs.ReadFile(srcFS, "default/gobbo.toml.tmpl")
		if err != nil {
			// This should be all but impossible.
			panic(err)
		}

		tmpl, err := template.New("bare").Parse(string(content))
		if err != nil {
			// Again, the default template must be valid, so this should be
			// impossible.
			panic(err)
		}

		err = tmpl.Execute(f, data)
		f.Close()
		if err != nil {
			return err
		}

		err = os.Rename(filename, dest)
		return err
	}

	tempDir, err := os.MkdirTemp(store.Join("tmp"), "gobbo-template")
	if err != nil {
		return fmt.Errorf("couldn't create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

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
		err = tmpl.Execute(destFile, data)
		if err != nil {
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
