package export

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/gosimple/slug"
	"github.com/starriver/gobbo/pkg/project"
	"github.com/starriver/gobbo/pkg/store"
)

// Transforms a project into a Compose config for exporting.

type Filter [][2]string

type ComposeConfig struct {
	Services map[string]Service
	// Volumes  map[string]map[string]any // unused for now
}

type Service struct {
	Image       string
	Volumes     []Volume
	Environment Environment
}

type Volume struct {
	Type     string
	Source   string
	Target   string
	ReadOnly bool
	Bind     Bind `yaml:",omitempty"`
}

type Bind struct {
	CreateHostPath bool   `yaml:",omitempty"`
	SELinux        string `yaml:"selinux,omitempty"`
}

type Environment struct {
	ExportPreset  string `yaml:"EXPORT_PRESET"`
	ExportVariant string `yaml:"EXPORT_VARIANT,omitempty"`
	ScriptPre     string `yaml:"SCRIPT_PRE"`
	ScriptPost    string `yaml:"SCRIPT_POST"`
}

// Stolen from: https://github.com/juliangruber/go-intersect/
func intersect(a []string, b []string) []string {
	set := make([]string, len(a))

	for _, v := range a {
		idx := sort.Search(len(b), func(i int) bool {
			return b[i] == v
		})
		if idx < len(b) && b[idx] == v {
			set = append(set, v)
		}
	}

	return set
}

// TODO proper config here - atm we can only append short-format
// bind-mount volumes. In order to do this, we'll need to implement
// array handling in the project TOML loader and move the Volume and
// Bind structs there.
func parseVolume(s string) Volume {
	v := Volume{
		Type: "bind",
		Bind: Bind{
			CreateHostPath: true,
		},
	}

	split := strings.Split(s, ":")

	switch len(split) {
	case 2, 3: // OK
	default:
		panic([]any{"Unknown volume format", s})
	}

	if len(split) == 3 {
		flags := strings.Split(split[2], ",")
		for _, flag := range flags {
			switch flag {
			case "z", "Z":
				v.Bind.SELinux = flag
			case "ro":
				v.ReadOnly = true
			default:
				panic([]any{"Unimplemented volume flag", flag})
			}
		}
	}

	v.Source = split[0]
	switch v.Source {
	case ".", "/": // OK
	default:
		panic([]any{"Non bind mount volumes not yet implemented"})
	}

	v.Target = split[1]
	if v.Target[0] != '/' {
		panic([]any{"Bind mount target must be an absolute path", v.Target})
	}

	return v
}

func Configure(store *store.Store, p *project.Project, filter []string) (c ComposeConfig) {
	presets := make([]string, len(p.Export.Presets))
	copy(presets, p.Export.Presets)

	godotSource := store.Join("bin", p.Godot.BinaryPath(&store.Platform))
	godotTarget := "/opt/godot-" + p.Godot.String()
	exportTemplateSource := p.Godot.ExportTemplatesPath()

	// Start by creating a prospective service per preset.

	hasVariants := len(p.Export.Variants) != 0
	hasOnly := len(p.Export.Only) != 0
	hasFilter := !hasVariants && (len(filter) != 0)
	// We deal with filtering later if using variants.

	if hasOnly || hasFilter {
		slices.Sort(presets)

		if hasOnly {
			presets = intersect(presets, p.Export.Only)
		}
		if hasFilter {
			presets = intersect(presets, filter)
		}
	}

	presetServices := make(map[string]Service, len(presets))
	for _, pr := range presets {
		s := Service{}
		s.Image = Tag
		s.Volumes = []Volume{
			{
				Type:     "bind",
				Source:   godotSource,
				Target:   godotTarget,
				ReadOnly: true,
			},
			{
				Type:     "bind",
				Source:   exportTemplateSource,
				Target:   "/root/.local/share/godot/export_templates",
				ReadOnly: true,
			},
			{
				Type:     "bind",
				Source:   p.Src,
				Target:   "/srv/src-ro",
				ReadOnly: true,
			},
			{
				Type:   "bind",
				Source: p.Export.Dist,
				Target: "/srv/dist",
				Bind: Bind{
					CreateHostPath: true,
					SELinux:        "z",
				},
			},
		}

		for _, str := range p.Export.Volumes {
			s.Volumes = append(s.Volumes, parseVolume(str))
		}

		s.Environment = Environment{
			ExportPreset:  pr,
			ExportVariant: "",
			ScriptPre:     p.Export.Scripts.Pre,
			ScriptPost:    p.Export.Scripts.Post,
		}

		presetServices[slug.Make(pr)] = s
	}

	// At this point, we have everything we need for a no-variants config.
	if len(p.Export.Variants) == 0 {
		c.Services = presetServices
		return
	}

	// Otherwise, we have variants - so we need to create a build matrix.

	c.Services = make(
		map[string]Service,
		len(p.Export.Variants)*len(p.Export.Presets),
	)

	hasFilter = len(filter) != 0

	for variantName, variant := range p.Export.Variants {
		for preset, s := range presetServices {
			if !slices.Contains(variant.Only, preset) {
				continue
			}
			if hasFilter {
				// TODO error on filter args that don't hit anything.
				cell := fmt.Sprintf("%s:%s", variantName, preset)
				if !slices.Contains(filter, cell) {
					continue
				}
			}

			for _, str := range p.Export.Volumes {
				s.Volumes = append(s.Volumes, parseVolume(str))
			}

			s.Environment.ExportVariant = variantName
			if variant.Scripts.Pre != "" {
				s.Environment.ScriptPre = variant.Scripts.Pre
			}
			if variant.Scripts.Post != "" {
				s.Environment.ScriptPost = variant.Scripts.Post
			}

			sName := fmt.Sprintf("%s_%s", slug.Make(variantName), preset)
			c.Services[sName] = s
		}
	}

	return
}
