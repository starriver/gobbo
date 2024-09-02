package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gitlab.com/starriver/gobbo/pkg/godot"
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
