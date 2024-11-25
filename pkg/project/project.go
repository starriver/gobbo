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
		Presets []string
		Only    []string
		Dist    string
		// add Volumes, Scripts?
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

	// Load export config (note we load Godot's export presets later).
	// Using a closure here so we can easily short-circuit.
	(func() {
		v, ok := pop("export")
		if !ok {
			// Nothing to do.
			return
		}

		ex, ok := v.(map[string]any)
		if !ok {
			pushErrorf("'export': expected table, got %T", v)
			return
		}

		for k, v := range ex {
			// If this is a table, it's an export variant.
			vt, ok := v.(map[string]any)
			if ok {
				variant := Variant{}

				pop := func(key string) (v any, ok bool) {
					v, ok = vt[key]
					delete(t, key)
					return
				}

				vv, ok := pop("only")
				if ok {
					only, ok := vv.([]string)
					if !ok {
						pushErrorf(
							"'export.%s.only': expected string array, got %T",
							k, vv,
						)
					} else {
						variant.Only = only
					}
				}

				vv, ok = pop("volumes")
				if ok {
					volumes, ok := vv.(map[string]string)
					if !ok {
						pushErrorf(
							"'export.%s.volumes': expected string-string table, got %T",
							k, vv,
						)
					} else {
						variant.Volumes = volumes
					}
				}

				vv, ok = pop("scripts")
				if ok {
					scripts, ok := vv.(map[string]any)
					if !ok {
						pushErrorf(
							"'export.%s.scripts': expected table, got %T",
							k, vv,
						)
					} else {
						vvv, ok := scripts["pre"]
						if ok {
							pre, ok := vvv.(string)
							if !ok {
								pushErrorf(
									"'export.%s.scripts.pre': expected string, got %T",
									k, vvv,
								)
							} else {
								variant.Scripts.Pre = pre
							}
						}

						vvv, ok = scripts["post"]
						if ok {
							post, ok := vvv.(string)
							if !ok {
								pushErrorf(
									"'export.%s.scripts.post': expected string, got %T",
									k, vvv,
								)
							} else {
								variant.Scripts.Post = post
							}
						}

						delete(scripts, "pre")
						delete(scripts, "post")
						if len(scripts) != 0 {
							pushErrorf("'export.%s.scripts: unknown keys", k)
						}
					}
				}

				vv, ok = pop("elective")
				if ok {
					elective, ok := vv.(bool)
					if !ok {
						pushErrorf(
							"'export.%s.elective': expected bool, got %T",
							k, vv,
						)
					} else {
						variant.Elective = elective
					}
				}

				p.Export.Variants[k] = variant
				continue
			}

			switch k {
			case "only":
				only, ok := v.([]string)
				if !ok {
					pushErrorf(
						"'export.only': expected string array, got %T",
						v,
					)
				} else {
					p.Export.Only = only
				}

			case "dist":
				dist, ok := v.(string)
				if !ok {
					pushErrorf("'export.dist': expected string, got %T", v)
				} else {
					p.Export.Dist = dist
				}

			default:
				pushErrorf("'export.%s': expected table, got %T", v)
			}
		}
	})()

	// Error on remaining keys.
	for k := range t {
		pushErrorf("'%s': unknown key", k)
	}

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
