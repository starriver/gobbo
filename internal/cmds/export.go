package cmds

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
	"github.com/starriver/gobbo/pkg/export"
	"github.com/starriver/gobbo/pkg/glog"
	"gopkg.in/yaml.v3"
)

const exportDesc = `
Builds project exports in parallel using Docker containers. All exports are
built unless {EXPORT}s are specified.

Docker Compose is required, and the Docker CLI must be in your {PATH}. If not,
the command will fail.

Before exporting, various prerequisites are ensured:
- Export templates are installed, if they aren't already.
- A Docker image for the exports is built.
- Project imports are run.

Gobbo builds its own Docker image for the containers, using the
{starriver.run/gobbo} tag. You can provide your own image by creating a
Dockerfile in your project's root directory.

Use '{FROM starriver.run/gobbo:latest}' in your Dockerfile to ensure you have
all of the prerequisite dependencies available (unless you know what you're
doing!). Note that Godot will be bind-mounted into the build containers, so
there's no need to install it in your Dockerfile.
`

var Export = charli.Command{
	Name:        "export",
	Headline:    "Build project exports",
	Description: exportDesc,
	Options: []charli.Option{
		opts.Project,

		// TODO: -j/--jobs
		// compose up will require that we start all containers immediately, but
		// we can signal them to actually start their builds by eg. placing a
		// file.

		{
			Short:    'n',
			Long:     "no-install",
			Flag:     true,
			Headline: "Abort if dependencies are missing",
		},
		{
			Short:    'm',
			Long:     "no-import",
			Flag:     true,
			Headline: "Don't import assets before exporting",
		},
		{
			Short:    'r',
			Long:     "rebuild-image",
			Flag:     true,
			Headline: "Always rebuild exporter Docker image",
		},
		{
			Short:    'd',
			Long:     "debug",
			Flag:     true,
			Headline: "Build debug exports",
		},
		{
			Short:    'c',
			Long:     "compose",
			Flag:     true,
			Headline: "Output Docker Compose config then exit",
		},
	},
	Args: charli.Args{
		Varadic:  true,
		Metavars: []string{"EXPORT"},
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)

		store := opts.StoreSetup(r)

		installMode := opts.IfAbsent
		if r.Options["n"].IsSet {
			installMode = opts.Never
		}

		project, godot := opts.ProjectGodotSetup(r, store, installMode, true)

		installed, err := store.IsGodotInstalled(godot)
		if err != nil {
			r.Error(err)
		} else if !installed {
			r.Errorf("Godot %s not installed", godot.String())
		}

		if r.Fail {
			return
		}

		if len(project.Export.Presets) == 0 {
			r.Errorf("No export presets configured.")
			return
		}

		err = export.CheckEnvironment()
		if err != nil {
			r.Error(err)
			return
		}

		debug := r.Options["d"].IsSet
		c := export.Configure(store, project, debug, r.Args)

		if r.Options["c"].IsSet {
			err = yaml.NewEncoder(os.Stdout).Encode(c)
			if err != nil {
				r.Error(err)
			}
			return
		}

		alwaysRebuild := r.Options["r"].IsSet
		err = export.BuildImage(alwaysRebuild)
		if err != nil {
			r.Error(err)
			return
		}

		if !r.Options["m"].IsSet {
			glog.Info("Importing assets...")
			godotPath := store.Join("bin", godot.BinaryPath(&store.Platform))
			cmd := exec.Command(godotPath, "--no-header", "--headless", "--import", project.GodotConfigPath())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			glog.Debugf("%s %v", cmd.Path, cmd.Args)

			err = cmd.Run()
			if err != nil {
				r.Errorf("Import failed, aborting.")
				return
			}
		}

		glog.Info("Starting exports...")
		err = export.Run(c)
		if err != nil {
			r.Error(err)
			return
		}

		if project.Export.Zip {
			// If we're zipping, dist subdirectories will all have just a single
			// file - move them up for convenience.

			zips := make([]string, 0, len(c.Services))
			err := filepath.WalkDir(project.Export.Dist, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if !d.IsDir() && filepath.Ext(path) == ".zip" {
					zips = append(zips, path)
				}
				return nil
			})

			if err != nil {
				glog.Warnf("Couldn't walk '%s': %v", project.Export.Dist, err)
			} else {
				for _, z := range zips {
					dir := filepath.Dir(z)
					err = os.Rename(z, filepath.Dir(dir))
					if err != nil {
						glog.Warnf("Couldn't move '%s' up: %v", z, err)
						continue
					}

					err = os.Remove(dir)
					if err != nil {
						glog.Warnf("Couldn't remove directory '%s': %v", dir, err)
					}
				}
			}
		}
	},
}
