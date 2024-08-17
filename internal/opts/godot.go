package opts

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/godot"
)

var Godot = charli.Option{
	Short:    'g',
	Long:     "godot",
	Metavar:  "VERSION",
	Headline: "Specify Godot version",
}

func GodotSetup(r *charli.Result, defaultStable bool) (g *godot.Official) {
	opt := r.Options["g"]

	var err error
	if opt.IsSet {
		g, err = godot.Parse(opt.Value)
		if err != nil {
			r.Error(err)
		}
	} else if defaultStable {
		g, err = godot.CurrentRelease(false)
		if err != nil {
			r.Error(err)
		}
	}

	return
}
