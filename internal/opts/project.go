package opts

import (
	"os"
	"path/filepath"

	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
	"gitlab.com/starriver/gobbo/pkg/godot"
	"gitlab.com/starriver/gobbo/pkg/project"
	"gitlab.com/starriver/gobbo/pkg/store"
)

var Project = charli.Option{
	Short:    'p',
	Long:     "project",
	Metavar:  "PATH",
	Headline: "Specify a project directory or config file",
}

func ProjectSetup(r *charli.Result, required bool) *project.Project {
	opt := r.Options["p"]
	path := "."
	if opt.IsSet {
		path = opt.Value
	}

	checkPath := func() (os.FileInfo, bool) {
		st, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if required || opt.IsSet {
					r.Errorf(
						"not a Gobbo project directory or config file: '%s'",
						path,
					)
				}
			} else {
				r.Error(err)
			}
			return nil, false
		}
		return st, true
	}

	st, ok := checkPath()
	if !ok {
		return nil
	}

	// If this is a directory, change the path to the config file and check
	// again
	if st.IsDir() {
		path = filepath.Join("gobbo.toml")
		_, ok = checkPath()
		if !ok {
			return nil
		}
	}

	p, errs := project.Load(path, r.Fail)
	for _, err := range errs {
		r.Error(err)
	}

	return p
}

func ProjectGodotSetup(
	r *charli.Result,
	s *store.Store,
	mode InstallMode,
	projectRequired bool,
) (p *project.Project, g *godot.Official) {
	pOpt := r.Options["p"]
	gOpt := r.Options["g"]

	p = ProjectSetup(r, projectRequired)
	if p == nil {
		// Special error for when a Godot version is required and the (implicit)
		// project dir failed to load
		if !pOpt.IsSet && (projectRequired || !gOpt.IsSet) {
			glog.Error("this doesn't look like a Gobbo project directory.")
			godotMention := ""
			if !projectRequired {
				godotMention = " or -g/--godot"
			}
			glog.Errorf(
				"change to a project directory, or use -p/--project%s.",
				godotMention,
			)
			r.Fail = true
		}
	}

	if gOpt.IsSet {
		g = GodotSetup(r, s, mode, false)
	} else if p != nil {
		g = p.Godot
		InstallGodot(r, s, g, mode)
	} else {
		glog.Error("no project or Godot version supplied.")
		glog.Error(
			"change to a project directory, or use -p/--project or -g/--godot.",
		)
		r.Fail = true
	}

	return
}
