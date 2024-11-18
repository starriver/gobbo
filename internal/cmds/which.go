package cmds

import (
	"fmt"

	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
)

const whichDesc = `
Prints the path to the Godot executable for the given project to stdout,
then exits immediately.

{-g}/{--godot} can be used to specify an arbitrary Godot version.

If the Godot version in question isn't installed, Gobbo will exit with an
error status.
`

var Which = charli.Command{
	Name:        "which",
	Headline:    "Show Godot executable path",
	Description: whichDesc,
	Options: []charli.Option{
		opts.Project,
	},

	Run: func(r *charli.Result) {
		store := opts.StoreSetup(r)
		_, godot := opts.ProjectGodotSetup(r, store, opts.Never, false)

		if godot != nil {
			installed, err := store.IsGodotInstalled(godot)
			if err != nil {
				r.Error(err)
			} else if !installed {
				r.Errorf("Godot %s not installed", godot.String())
			}
		}

		if r.Fail {
			return
		}

		fmt.Println(store.Join("bin", godot.BinaryPath(&store.Platform)))
	},
}
