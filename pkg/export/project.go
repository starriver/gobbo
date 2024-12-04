package export

import (
	"slices"
	"sort"

	"github.com/adrg/xdg"
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
	Type string
	Source string
	Target string
	ReadOnly bool
	Bind Bind `yaml:",omitempty"`
}

type Bind struct {
	CreateHostPath `yaml:",omitempty"`
	SELinux string `yaml:"selinux,omitempty"`
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
					Type: "bind",
					Source: godotSource,
					Target: godotTarget,
					ReadOnly: true,
				},
				{
					Type: "bind",
					Source: exportTemplateSource,
					Target: "/root/.local/share/godot/export_templates",
					ReadOnly: true,
				},
				{
					Type: "bind",
					Source: p.Src,
					Target: "/srv/src-ro",
					ReadOnly: true,
				},
				{
					Type: "bind",
					Source: p.Export.Dist,
					Target: "/srv/dist",
					Bind: Bind{
						CreateHostPath: true,
						SELinux: "z",
					},
				},
			}

			for _, volume // TODO

			s.Environment = Environment{
				ExportPreset:  pr,
				ExportVariant: "",
				ScriptPre: p.Export.Scripts.Pre,
				ScriptPost: p.Export.Scripts.Post,
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
}
