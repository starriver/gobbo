package cmds

import (
	"os"

	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

const cleanDesc = `
Removes temporary project files in .godot/ (within Godot project's source
directory).
`

var Clean = charli.Command{
	Name:        "clean",
	Headline:    "Remove temporary files",
	Description: cleanDesc,

	Options: []charli.Option{
		opts.Project,
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)

		project := opts.ProjectSetup(r, true)

		if r.Fail {
			return
		}

		path := project.GodotCachePath()
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				glog.Info("Nothing to clean.")
			} else {
				r.Error(err)
			}
			return
		}

		err = os.RemoveAll(path)
		if err != nil {
			r.Error(err)
		} else {
			glog.Infof("Deleted: '%s'", path)
		}
	},
}
