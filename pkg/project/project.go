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
	Godot   *godot.Official
	Src     string
	Version string

	Export struct {
		Presets  []string
		Only     []string
		Dist     string
		Variants map[string]Variant
	}
}

type Variant struct {
	Only    []string
	Volumes map[string]string
	Scripts struct {
		Pre  string
		Post string
	}
	Elective bool
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

	// TODO: read export config.

	// ---
	// Now we start reading Godot config from src.

	// Get project version from project.godot
	configPath := p.GodotConfigPath()
	f, err = os.Open(configPath)
	if err != nil {
		pushErrorf("Couldn't open '%s': %v", configPath, err)
	} else {
		defer f.Close()

		s := bufio.NewScanner(f)
		appSection := false
		p.Version = "unspecified"

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
					v := t[from:to]
					if v != "" {
						p.Version = v
					}
					break
				}
			} else if t == "[application]" {
				appSection = true
			}
		}
	}

	// TODO: Get export presets.
	// TODO: Check preset names match those in Gobbo config, WARN otherwise

	return
}

func (p *Project) GodotConfigPath() string {
	return filepath.Join(p.Src, "project.godot")
}

func (p *Project) GodotCachePath() string {
	return filepath.Join(p.Src, ".godot")
}
