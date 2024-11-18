package opts

import (
	"github.com/starriver/charli"
	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/godot"
	"github.com/starriver/gobbo/pkg/store"
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
	var err error

	opt := r.Options["g"]

	isStream := godot.IsStream(opt.Value) || (!opt.IsSet && defaultStable)
	isLatest := opt.Value == "latest"
	if isStream {
		g, err = s.CachedGodotRelease(isLatest)
		if err != nil {
			glog.Warnf("couldn't check for cached release: %v", err)
		}
	}

	if g == nil {
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

		if g != nil && isStream {
			s.SetCachedGodotRelease(opt.Value == "latest", g)
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
			err := s.InstallGodot(g)
			if err != nil {
				r.Error(err)
			}
		}

	case Always:
		err := s.InstallGodot(g)
		if err != nil {
			r.Error(err)
		}
	}
}
