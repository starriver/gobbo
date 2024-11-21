package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/godot"
	"github.com/starriver/gobbo/pkg/platform"
)

func (s *Store) IsGodotInstalled(g *godot.Official) (bool, error) {
	_, err := os.Stat(s.Join("bin", g.BinaryPath(&s.Platform)))
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

func streamString(latest bool) string {
	if latest {
		return "latest"
	}
	return "stable"
}

func (s *Store) CachedGodotRelease(latest bool) (*godot.Official, error) {
	ss := streamString(latest)
	path := s.Join("cache", ss)
	glog.Debugf("Checking for cached Godot release from '%s'", path)

	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if st.ModTime().AddDate(0, 0, 1).Before(time.Now()) {
		glog.Debug("Cache is more than a day old - busting")
		return nil, nil
	}

	// The cache files should be tiny, so this should be fine.
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(b) < 2 {
		// Not written yet. Single rune is newline.
		glog.Debug("Nothing cached")
		return nil, nil
	}

	str := string(b[:len(b)-1]) // Trim newline
	glog.Debugf("Cached %s Godot is: %s", ss, str)
	return godot.Parse(str)
}

func (s *Store) SetCachedGodotRelease(latest bool, g *godot.Official) error {
	path := s.Join("cache", streamString(latest))

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(g.String() + "\n") // POSIXly correct
	return err
}

func (s *Store) InstallGodot(g *godot.Official) error {
	glog.Infof("Downloading Godot %s...", g.String())
	url := g.DownloadURL(&s.Platform)

	zip, err := s.Download(url)
	if err != nil {
		return err
	}

	glog.Info("Extracting...")
	extracted, err := s.Unzip(zip)
	if err != nil {
		return err
	}

	err = os.Remove(zip)
	if err != nil {
		glog.Warnf("Couldn't remove '%s': %v", zip, err)
	}

	err = normalize(s, g, extracted)
	if err != nil {
		return err
	}

	dest := s.Join("bin", "official", g.String())
	_, err = os.Stat(dest)
	if err == nil {
		err = os.RemoveAll(dest)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	err = os.Rename(extracted, dest)
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
			to := filepath.Join(tmp, fmt.Sprintf("godot-%s", g.String()))
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
		to := filepath.Join(tmp, fmt.Sprintf("godot-%s.app", g.String()))
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
				to = filepath.Join(
					tmp,
					fmt.Sprintf("godot-%s_console.exe", g.String()),
				)
				okConsole = true
			} else {
				to = filepath.Join(tmp, fmt.Sprintf("godot-%s.exe", g.String()))
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
