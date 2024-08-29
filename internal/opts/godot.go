package opts

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/godot"
	"gitlab.com/starriver/gobbo/pkg/store"
)

var Godot = charli.Option{
	Short:    'g',
	Long:     "godot",
	Metavar:  "VERSION",
	Headline: "Specify Godot version",
}

type InstallMode uint8

const (
	Never InstallMode = iota
	IfAbsent
	Always
)

func GodotSetup(r *charli.Result, s *store.Store, mode InstallMode, defaultStable bool) (g *godot.Official) {
	opt := r.Options["g"]

	var err error
	if opt.IsSet {
		g, err = godot.ParseWithStream(opt.Value, r.Fail)
		if err != nil {
			r.Error(err)
		}
	} else if defaultStable && !r.Fail {
		g, err = godot.CurrentRelease(false)
		if err != nil {
			r.Error(err)
		}
	}

	InstallGodot(r, s, g, mode)
	return
}

func InstallGodot(r *charli.Result, s *store.Store, g *godot.Official, mode InstallMode) {
	if r.Fail {
		return
	}

	switch mode {
	case Never:
		// Nothing to do,

	case IfAbsent:
		if s == nil {
			break
		}

		isInstalled, err := s.IsGodotInstalled(g)
		if err != nil {
			r.Error(err)
			break
		}
		if !isInstalled {
			s.InstallGodot(g)
		}

	case Always:
		err := s.InstallGodot(g)
		if err != nil {
			r.Error(err)
		}
	}
}
