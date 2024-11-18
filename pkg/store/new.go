package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/starriver/gobbo/pkg/platform"
)

func New(path string) (store *Store, errs []error) {
	errs = walk(schema, path)

	if len(errs) == 0 {
		store = &Store{
			Root:     path,
			Platform: platform.FromRuntime(),
		}
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

		os.Mkdir(path, os.ModePerm)
	} else if !s.IsDir() {
		errs = []error{
			fmt.Errorf("'%s' should be a directory", path),
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
			s, err := os.Stat(subpath)
			if err != nil {
				if !os.IsNotExist(err) {
					errs = append(errs, err)
					continue
				}

				f, err := os.Create(subpath)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				f.WriteString(contents + "\n") // POSIX
				continue
			}

			if s.IsDir() {
				err = fmt.Errorf(
					"expected file '%s', got directory",
					subdir,
				)
				errs = append(errs, err)
				continue
			}

			if contents == "" {
				continue
			}

			// Compare file contents

			contents += "\n" // POSIX

			want := []byte(contents)
			f, err := os.Open(subpath)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			got := make([]byte, len(want))
			_, err = f.Read(got)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			if !bytes.Equal(want, got) {
				err = fmt.Errorf(
					"'%s': expected '%s', got '%s'",
					subpath,
					contents[:len(contents)-1],
					strings.Trim(string(got), " \n"),
				)
				errs = append(errs, err)
			}

			continue
		}

		panic("Store schema misconfigured")
	}

	return
}
