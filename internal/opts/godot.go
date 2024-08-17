package opts

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
	"gitlab.com/starriver/gobbo/pkg/godot"
)

var GodotOpt = charli.Option{
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
			glog.Error(err)
		}
	} else if defaultStable {
		g, err = godot.CurrentRelease(false)
		if err != nil {
			glog.Error(err)
		}
	}

	return
}
