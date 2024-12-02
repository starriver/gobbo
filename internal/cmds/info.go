package cmds

import (
	"fmt"
	"runtime/debug"

	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/godot"
)

const infoDesc = `
Prints version information for Gobbo, Godot and the specified project to
stdout, then exits immediately.
`

var Info = charli.Command{
	Name:        "info",
	Headline:    "Show version & environment information",
	Description: infoDesc,
	Options: []charli.Option{
		opts.Project,
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)
		store := opts.StoreSetup(r)

		project := opts.ProjectSetup(r, false)

		var godot *godot.Official
		if project != nil {
			godot = project.Godot
		} else {
			godot = opts.GodotSetup(r, store, opts.Never, false)
		}

		if r.Fail {
			return
		}

		bi, ok := debug.ReadBuildInfo()
		if !ok {
			panic("Failed to read build info")
		}

		gv := "n/a"
		if godot != nil {
			gv = godot.String()
		}

		fmt.Printf("Gobbo version:   %s\n", bi.Main.Version)
		fmt.Printf("Godot version:   %s\n", gv)
		fmt.Printf("Project version: %s\n", project.Version)
		fmt.Printf("Store path:      %s\n", store.Root)

		glog.Debugf("Project struct: %+v", project)
	},
}
