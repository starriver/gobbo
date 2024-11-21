package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/starriver/gobbo/pkg/godot"
)

type Project struct {
	Godot *godot.Official
	Src   string
}

func Load(path string, ignoreStream bool) (p *Project, errs []error) {
	var err error
	f, err := os.Open(path)
	if err != nil {
		errs = []error{err}
		return
	}

	var t map[string]any

	pushError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	pushErrorf := func(format string, a ...any) {
		errs = append(errs, fmt.Errorf(format, a...))
	}

	pop := func(key string) (v any, ok bool) {
		v, ok = t[key]
		delete(t, key)
		return
	}

	checkString := func(key string, v any) (s string, ok bool) {
		s, ok = v.(string)
		if !ok {
			pushErrorf("'%s': expected string, got %T", key, v)
		}
		return
	}

	err = toml.NewDecoder(f).Decode(&t)
	pushError(err)

	p = &Project{}

	var str string
	v, ok := pop("godot")
	if ok {
		str, ok = checkString("godot", v)
		if ok {
			p.Godot, err = godot.ParseWithStream(str, ignoreStream)
			pushError(err)
		}
	} else {
		pushErrorf("'godot' isn't set")
	}

	str = ""
	v, ok = pop("src")
	if ok {
		str, ok = checkString("src", v)
	} else {
		str = "src"
		ok = true
	}

	if ok {
		p.Src = filepath.Join(filepath.Dir(path), str)
		_, err := os.Stat(p.GodotConfigPath())
		if err != nil {
			if os.IsNotExist(err) {
				pushErrorf("src '%s' doesn't exist", str)
			} else {
				errs = append(errs, err)
			}
		}
	}

	// Error on remaining keys.
	for k := range t {
		pushErrorf("'%s': unknown key", k)
	}

	return
}

func (p *Project) GodotConfigPath() string {
	return filepath.Join(p.Src, "project.godot")
}

func (p *Project) GodotCachePath() string {
	return filepath.Join(p.Src, ".godot")
}

// Reads the project version from 'config/version' in project.godot.
func (p *Project) Version() (string, error) {
	path := p.GodotConfigPath()
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	appSection := false
	for s.Scan() {
		t := s.Text()
		if (len(t) == 0) || (t[0] == ';') {
			continue
		}

		if appSection {
			if t[0] == '[' {
				// Starting another section - don't bother scanning further.
				break
			}

			if strings.HasPrefix(t, "config/version=\"") {
				from := strings.Index(t, "\"") + 1
				to := strings.LastIndex(t, "\"")
				version := t[from:to]
				if version == "" {
					version = "unspecified"
				}
				return version, nil
			}
		} else if t == "[application]" {
			appSection = true
		}
	}

	return "unspecified", nil
}
