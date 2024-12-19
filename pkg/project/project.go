package project

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/godot"
)

type Project struct {
	Godot   *godot.Official
	Src     string
	Name    string
	Version string

	Export struct {
		Presets  []Preset
		Only     []string
		Dist     string
		Zip      bool
		Volumes  []string
		Scripts  Scripts
		Variants map[string]*Variant
	}
}

type Preset struct {
	Name     string
	Platform string
}

type Variant struct {
	Only     []string
	Volumes  []string
	Scripts  Scripts
	Elective bool
}

type Scripts struct {
	Pre  string
	Post string
}

func popFunc[T any](m map[string]any, pushErrorf func(string, ...any)) func(string, bool) (T, bool) {
	return func(path string, require bool) (T, bool) {
		var t T
		mm := m
		segs := strings.Split(path, ".") // heh

		for i, seg := range segs {
			location := strings.Join(segs[0:i], ".")
			v, ok := mm[seg]
			if !ok {
				if require {
					pushErrorf("required '%s' isn't set", location)
				}
				return t, false
			}

			// Traverse deeper?
			if i != len(segs)-1 {
				// Expect a table.
				table, ok := v.(map[string]any)
				if !ok {
					pushErrorf("'%s': expected %T, got %T", location, table, v)
					return t, false
				}
				mm = table
				continue
			}

			t, ok = v.(T)
			if !ok {
				pushErrorf("'%s': expected %T, got %T", t, v)
			}
			// Delete this element from its parent if it isn't a table.
			// map[any]any is (possibly) a quicker coersive check than
			// map[string]any?
			if _, isTable := v.(map[any]any); !isTable {
				delete(mm, seg)
			}
		}

		return t, true
	}
}

func scanKeys(m map[string]any, path string) []string {
	unknown := []string{}
	for k, v := range m {
		p := path + k
		mm, ok := v.(map[string]any)
		if !ok {
			unknown = append(unknown, p)
		} else {
			unknown = append(unknown, scanKeys(mm, p+".")...)
		}
	}
	return unknown
}

func Load(path string, ignoreStream bool) (p *Project, errs []error) {
	var err error
	f, err := os.Open(path)
	if err != nil {
		errs = []error{err}
		return
	}

	var root map[string]any

	pushErrorf := func(format string, a ...any) {
		errs = append(errs, fmt.Errorf(format, a...))
	}

	err = toml.NewDecoder(f).Decode(&root)
	if err != nil {
		errs = append(errs, err)
		return
	}

	popString := popFunc[string](root, pushErrorf)
	popBool := popFunc[bool](root, pushErrorf)
	popStringArray := popFunc[[]string](root, pushErrorf)
	// popStringMap := popFunc[map[string]string](root, pushErrorf)

	p = &Project{}

	s, ok := popString("godot", true)
	if ok {
		p.Godot, err = godot.ParseWithStream(s, ignoreStream)
		if err != nil {
			errs = append(errs, err)
		}
	}

	s, ok = popString("src", false)
	if !ok {
		s = "src"
	}
	p.Src = filepath.Join(filepath.Dir(path), s)
	_, err = os.Stat(p.GodotConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			pushErrorf("source directory doesn't exist: '%s'", p.Src)
		} else {
			errs = append(errs, err)
		}
	}

	p.Export.Only, _ = popStringArray("export.only", false)

	s, ok = popString("export.dist", false)
	if !ok {
		s = "dist"
	}
	p.Export.Dist = filepath.Join(filepath.Dir(path), s)
	// Directory doesn't need to exist before export.

	p.Export.Zip, _ = popBool("export.zip", false)

	p.Export.Volumes, _ = popStringArray("export.volumes", false)

	p.Export.Scripts.Pre, _ = popString("export.scripts.pre", false)
	p.Export.Scripts.Post, _ = popString("export.scripts.post", false)

	// The remaining export.* keys will be read as variants.
	popVariants := popFunc[map[string]any](root, pushErrorf)
	variants, _ := popVariants("export", false)

	for k, table := range variants {
		// Check this is actually a table first.
		// This prevent error spam from the below pop calls.
		_, ok := table.(map[string]any)
		if !ok {
			pushErrorf("'export.%s': expected variant config, got %T", k, table)
			continue
		}

		v := &Variant{}
		prefix := fmt.Sprintf("export.%s.", k)

		v.Only, _ = popStringArray(prefix+"only", false)
		v.Volumes, _ = popStringArray(prefix+"volumes", false)
		v.Scripts.Pre, _ = popString(prefix+"scripts.pre", false)
		v.Scripts.Post, _ = popString(prefix+"scripts.post", false)
		v.Elective, _ = popBool(prefix+"elective", false)

		p.Export.Variants[k] = v
	}

	// Error on remaining keys, if anything still exists that isn't an empty
	// table (recursively).
	unknown := scanKeys(root, "")
	for u := range unknown {
		pushErrorf("'%s': unknown key", u)
	}

	// ---
	// Now we start reading Godot config from src.

	// Query the Godot project config
	gcPath := p.GodotConfigPath()
	gc, err := godot.Query(gcPath, godot.Q{
		"application": {
			"config/name",
			"config/version",
		},
	})
	if err != nil {
		pushErrorf("couldn't load '%s': %v", gcPath, err)
	} else {
		// Must exist:
		app := gc["application"]
		p.Name, _ = app["config/name"]
		p.Version, _ = app["config/version"]
	}

	// Query export presets
	epPath := p.ExportPresetsPath()
	presetMap := map[string]bool{}

	_, err = os.Stat(epPath)
	if err != nil {
		if !os.IsNotExist(err) {
			pushErrorf("couldn't stat '%s': %v", epPath, err)
		}
		// Else, export_presets.cfg doesn't exist, so no presets are defined.
	} else {
		// export_presets.cfg exists, load it:
		ep, err := godot.Query(epPath, godot.Q{
			"presets.*": {
				"name",
				"platform",
			},
		})
		if err != nil {
			pushErrorf("couldn't load '%s': %v", epPath, err)
		} else {
			presets := make([]Preset, len(ep))
			i := 0
			for _, pr := range ep {
				presets[i].Name, _ = pr["name"]
				presets[i].Platform, _ = pr["platform"]
				i++
			}
			slices.SortFunc(presets, func(a, b Preset) int {
				if a.Name < b.Name {
					return -1
				} else if a.Name > b.Name {
					return 1
				}
				return 0
			})
			p.Export.Presets = presets
		}
	}

	// If referenced presets don't exist, warn, and remove from in-memory config
	only := make([]string, len(p.Export.Only))
	for _, preset := range p.Export.Only {
		if _, ok := presetMap[preset]; !ok {
			glog.Warnf("export.only: missing preset '%s'", preset)
		} else {
			only = append(only, preset)
		}
	}
	p.Export.Only = only

	for k, variant := range p.Export.Variants {
		only := make([]string, len(variant.Only))
		for _, preset := range variant.Only {
			if _, ok := presetMap[preset]; !ok {
				glog.Warnf("export.%s.only: missing preset '%s'", k, preset)
			} else {
				only = append(only, preset)
			}
		}
		variant.Only = only
	}

	return
}

func (p *Project) GodotConfigPath() string {
	return filepath.Join(p.Src, "project.godot")
}

func (p *Project) ExportPresetsPath() string {
	return filepath.Join(p.Src, "export_presets.cfg")
}

func (p *Project) GodotCachePath() string {
	return filepath.Join(p.Src, ".godot")
}
