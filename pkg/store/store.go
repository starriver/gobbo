package store

import (
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	Root string
}

type dir map[string]interface{}

// Blank string means "a file should exist here".
// Non-blank string means "a file should exist here with these contents".
var schema = dir{
	"version":       "1",
	"release-cache": "",
	"bin": dir{
		"official": dir{},
	},
}

func Load(path string) (store *Store, errs []error) {
	errs = walk(schema, "")
	if len(errs) == 0 {
		store = &Store{path}
	}
	return
}

func walk(d dir, path string) (errs []error) {
	s, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			errs = []error{err}
			return
		}

		os.Mkdir(path, os.ModeDir)
	} else if !s.IsDir() {
		errs = []error{
			fmt.Errorf("store: '%s' should be a directory", path),
		}
	}

	for k, v := range d {
		subpath := filepath.Join(path, k)

		subdir, isDir := v.(dir)
		if isDir {
			e := walk(subdir, subpath)
			errs = append(errs, e...)
			continue
		}

		contents, isFile := v.(string)
		if isFile {
			var f *os.File
			s, err := os.Stat(subpath)
			if err != nil {
				if !os.IsNotExist(err) {
					errs = append(errs, err)
					continue
				}

				f, err = os.Create(subpath)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				f.WriteString(contents)
				continue
			}

			if s.IsDir() {
				err := fmt.Errorf(
					"expected file '%s', got directory",
					subdir,
				)
				errs = append(errs, err)
				continue
			}

			if contents == "" {
				continue
			}

			// TODO stream this
			b, err := os.ReadFile(subpath)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			actual := string(b)
			if actual != (contents + "\n") {
				err = fmt.Errorf(
					"expected file '%s' to contain '%s', got '%s'",
					subpath,
					contents,
					actual,
				)
				errs = append(errs, err)
				continue
			}

			continue
		}

		panic("Store schema misconfigured")
	}

	return
}
