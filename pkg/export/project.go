package export

import (
	"slices"
	"sort"
	"strings"

	"github.com/starriver/gobbo/pkg/project"
	"github.com/starriver/gobbo/pkg/store"
)

// Transforms a project into a Compose config for exporting.

type Filter [][2]string

type ComposeConfig struct {
	Services map[string]Service
	Volumes  map[string]map[string]any
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

func Configure(store *store.Store, p *project.Project, filter Filter) (c ComposeConfig) {
	presets := make([]string, len(p.Export.Presets))
	copy(presets, p.Export.Presets)

	godotSource := store.Join("bin", p.Godot.BinaryPath(&store.Platform))
	godotTarget := "/opt/godot-" + p.Godot.String()
	exportTemplateSource := p.Godot.ExportTemplatesPath()

	if len(p.Export.Variants) == 0 {
		hasOnly := len(p.Export.Only) != 0
		hasFilter := len(filter) != 0

		if hasOnly || hasFilter {
			slices.Sort(presets)

			if hasOnly {
				presets = intersect(presets, p.Export.Only)
			}
			if hasFilter {
				f := make([]string, len(filter))
				for i, ff := range filter {
					f[i] = ff[1]
				}
				presets = intersect(presets, f)
			}
		}

		c.Services = make(map[string]Service, len(presets))
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

			// TODO proper config here - atm we can only append short-format
			// bind-mount volumes. In order to do this, we'll need to implement
			// array handling in the project TOML loader and move the Volume and
			// Bind structs there.
			for _, str := range p.Export.Volumes {
				volume := Volume{
					Type: "bind",
					Bind: Bind{
						CreateHostPath: true,
					},
				}

				split := strings.Split(str, ":")

				switch len(split) {
				case 2, 3: // OK
				default:
					panic([]any{"Unknown volume format", str})
				}

				if len(split) == 3 {
					flags := strings.Split(split[2], ",")
					for _, flag := range flags {
						switch flag {
						case "z", "Z":
							volume.Bind.SELinux = flag
						case "ro":
							volume.ReadOnly = true
						default:
							panic([]any{"Unimplemented volume flag", flag})
						}
					}
				}

				volume.Source = split[0]
				switch volume.Source {
				case ".", "/": // OK
				default:
					panic([]any{"Non bind mount volumes not yet implemented"})
				}

				volume.Target = split[1]
				if volume.Target[0] != '/' {
					panic([]any{"Bind mount target must be an absolute path", volume.Target})
				}

				s.Volumes = append(s.Volumes, volume)
			}

			s.Environment = Environment{
				ExportPreset:  pr,
				ExportVariant: "",
				ScriptPre:     p.Export.Scripts.Pre,
				ScriptPost:    p.Export.Scripts.Post,
			}
			c.Services[pr] = s
		}
	}

	// for k, v := range p.Export.Variants {
	// 	only := make(map[string]bool)
	// 	useOnly := false
	// 	if len(filter) != 0 {
	// 		useOnly = true
	// 		for _, e := range filter {
	// 			only
	// 		}
	// 	}
	// }

	// TODO
	return
}
