package export

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gosimple/slug"
	"github.com/starriver/gobbo/pkg/godot"
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
	StopSignal  string `yaml:"stop_signal"`
}

type Volume struct {
	Type     string
	Source   string
	Target   string
	ReadOnly bool `yaml:"read_only,omitempty"`
	Bind     Bind `yaml:",omitempty"`
}

type Bind struct {
	CreateHostPath bool   `yaml:"create_host_path,omitempty"`
	SELinux        string `yaml:",omitempty"`
}

type Environment struct {
	GodotPath            string `yaml:"GODOT_PATH"`
	GodotSettingsVersion string `yaml:"GODOT_SETTINGS_VERSION"`
	ProjectName          string `yaml:"PROJECT_NAME"`
	ProjectVersion       string `yaml:"PROJECT_VERSION"`
	ExportPreset         string `yaml:"EXPORT_PRESET"`
	ExportVariant        string `yaml:"EXPORT_VARIANT"`
	ExportDebug          string `yaml:"EXPORT_DEBUG"`
	Extension            string `yaml:"EXTENSION"`
	ScriptPre            string `yaml:"SCRIPT_PRE"`
	ScriptPost           string `yaml:"SCRIPT_POST"`
	Zip                  string `yaml:"ZIP"`
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

var platformExts = make(map[string]string, 6)

func init() {
	platformExts["Android"] = "apk"
	platformExts["iOS"] = "ipa"
	platformExts["Linux"] = ""
	platformExts["macOS"] = "app"
	platformExts["Web"] = "html"
	platformExts["Windows Desktop"] = "exe"
}

func Configure(store *store.Store, p *project.Project, debug bool, filter []string) (c *ComposeConfig) {
	c = &ComposeConfig{}

	presetNames := make([]string, len(p.Export.Presets))
	for i, p := range p.Export.Presets {
		presetNames[i] = p.Name
	}

	godotSource := store.Join("bin", p.Godot.BinaryPath(&store.Platform))
	godotTarget := "/opt/godot-" + p.Godot.String()
	exportTemplateSource := godot.ExportTemplatesRoot()

	// Start by creating a prospective service per preset.

	hasVariants := len(p.Export.Variants) != 0
	hasOnly := len(p.Export.Only) != 0
	hasFilter := !hasVariants && (len(filter) != 0)
	// We deal with filtering later if using variants.

	if hasOnly || hasFilter {
		slices.Sort(presetNames)

		if hasOnly {
			presetNames = intersect(presetNames, p.Export.Only)
		}
		if hasFilter {
			presetNames = intersect(presetNames, filter)
		}
	}

	// This is the 4.x minor version string to be used for the editor
	// settings filename.
	settingsVersion := fmt.Sprintf("4.%d", p.Godot.Minor)

	zip := "0"
	if p.Export.Zip {
		zip = "1"
	}

	presetServices := make(map[string]Service, len(presetNames))
	for _, pr := range p.Export.Presets {
		if !slices.Contains(presetNames, pr.Name) {
			continue
		}

		// slugProject := slug.Make(p.Name)
		slugPreset := slug.Make(pr.Name)

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
			// nb. this may be modified in the variant config later -
			// Volumes[3] is hard-coded.
			{
				Type:   "bind",
				Source: filepath.Join(p.Export.Dist, pr.Name),
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

		debugStr := "0"
		if debug {
			debugStr = "1"
		}

		// Default to zip extension if this is some platform we've not heard of
		ext := "zip"
		if e, ok := platformExts[pr.Platform]; ok {
			ext = e
		}

		s.Environment = Environment{
			GodotPath:            godotTarget,
			GodotSettingsVersion: settingsVersion,
			ProjectName:          p.Name,
			ProjectVersion:       p.Version,
			ExportPreset:         pr.Name,
			ExportVariant:        "",
			ExportDebug:          debugStr,
			Extension:            ext,
			ScriptPre:            p.Export.Scripts.Pre,
			ScriptPost:           p.Export.Scripts.Post,
			Zip:                  zip,
		}

		// There's no point in waiting for Godot to clean up if we Ctrl-C the
		// exports, so:
		s.StopSignal = "SIGKILL"

		presetServices[slugPreset] = s
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
			if (len(variant.Only) != 0) && !slices.Contains(variant.Only, preset) {
				continue
			}

			if hasFilter {
				// TODO error on filter args that don't hit anything.
				cell := fmt.Sprintf("%s:%s", variantName, preset)
				if !slices.Contains(filter, cell) {
					continue
				}
			} else if variant.Elective {
				// Elective variants are only configured when they're explicitly
				// specified in the filter.
				continue
			}

			// Make dist one level deeper.
			s.Volumes[3].Source = filepath.Join(
				p.Export.Dist, variantName, preset,
			)

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
