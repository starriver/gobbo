package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/starriver/gobbo/pkg/download"
	"gitlab.com/starriver/gobbo/pkg/glog"
	"gitlab.com/starriver/gobbo/pkg/godot"
	"gitlab.com/starriver/gobbo/pkg/platform"
)

func (s *Store) DownloadGodot(g *godot.Official) error {
	url := g.DownloadURL(&s.Platform)

	zip, err := download.Download(url)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "gobbo-extract-")
	if err != nil {
		return err
	}

	err = download.Unzip(zip, tmp)
	if err != nil {
		return err
	}

	err = os.Remove(zip)
	if err != nil {
		glog.Warnf("couldn't remove '%s': %v", zip, err)
	}

	err = normalize(s, g, tmp)
	if err != nil {
		return err
	}

	dest := filepath.Join(s.Root, g.String())
	err = os.Rename(tmp, dest)
	if err != nil {
		return err
	}

	return nil
}

// This a relatively fuzzy way of normalizing the contents of downloaded release
// zips.
func normalize(s *Store, g *godot.Official, tmp string) error {
	dir, err := os.ReadDir(tmp)
	if err != nil {
		return err
	}
	switch s.Platform.OS {
	case platform.Linux:
		// Rename the first file in the directory.
		ok := false
		for _, f := range dir {
			if f.IsDir() {
				continue
			}

			from := filepath.Join(tmp, f.Name())
			to := filepath.Join(tmp, g.String())
			err = os.Rename(from, to)
			if err != nil {
				return err
			}

			err = os.Chmod(to, 0o755)
			if err != nil {
				return err
			}

			ok = true
			break
		}
		if !ok {
			return fmt.Errorf("expected regular file in '%s'", tmp)
		}

	case platform.MacOS:
		// Rename the .app.
		if len(dir) == 0 {
			return fmt.Errorf("expected a directory in '%s'", tmp)
		}

		from := filepath.Join(tmp, dir[0].Name())
		to := filepath.Join(tmp, g.String()+".app")
		err = os.Rename(from, to)
		if err != nil {
			return err
		}

	case platform.Windows:
		// Rename both the normal and console .exes.
		okNormal := false
		okConsole := false
		for _, f := range dir {
			n := f.Name()
			if !strings.HasSuffix(n, ".exe") {
				continue
			}

			from := filepath.Join(tmp, n)
			var to string
			if strings.HasSuffix(n, "_console.exe") {
				to = filepath.Join(tmp, g.String()+"_console.exe")
				okConsole = true
			} else {
				to = filepath.Join(tmp, g.String()+".exe")
				okNormal = true
			}

			err = os.Rename(from, to)
			if err != nil {
				return err
			}

			if okNormal && okConsole {
				break
			}
		}

		if !(okNormal && okConsole) {
			return fmt.Errorf(
				"expected normal and console executables in '%s'",
				tmp,
			)
		}
	}

	return nil
}
