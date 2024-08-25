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

func GodotSetup(r *charli.Result, s *store.Store, defaultStable bool) (g *godot.Official) {
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

	return
}
