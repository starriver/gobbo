package export

import (
	"github.com/starriver/gobbo/pkg/project"
)

// Transforms a project into a Compose config for exporting.

type Filter [][2]string

type ComposeConfig struct {
	Services map[string]struct {
		Image       string
		Volumes     []string
		Environment struct {
			scriptPre  string
			scriptPost string
		}
	}
	Volumes map[string]map[any]any
}

func Configure(p *project.Project, filter Filter) (c ComposeConfig) {
	if len(p.Export.Variants) == 0 {
		presets := p.Export.Presets

		hasOnly := len(p.Export.Only) != 0
		hasFilter := len(filter) != 0

		// Short-circuit
		if hasOnly || hasFilter {
			presetsMap := make(map[string]bool)

		}

		// only := make([]string, len(p.Export.Only) + len(filter))
		// if len(p.Export.Only) != 0 {
		// 	restrict = true
		// 	copy(only, p.Export.Only)
		// }
		// if len(filter) != 0 {
		// 	restrict = true
		// }
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
