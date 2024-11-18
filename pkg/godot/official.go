package godot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-github/v63/github"
	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/platform"
)

type Official struct {
	Minor  uint8
	Patch  uint8
	Suffix string
	Mono   bool
}

func (g Official) BinaryPath(p *platform.Platform) string {
	str := g.String() // Heh
	path := filepath.Join("official", str, fmt.Sprintf("godot-%s", str))

	switch p.OS {
	case platform.MacOS:
		return path + ".app"
	case platform.Windows:
		return path + ".exe"
	}

	// Linux has no extension.

	return path
}

func (g Official) String() string {
	return g.StringEx(false, false, true)
}

func (g Official) StringEx(dot, stable, mono bool) (str string) {
	// TODO: use strings.Builder
	str = fmt.Sprintf("4.%d", g.Minor)

	if g.Patch > 0 {
		str += fmt.Sprintf(".%d", g.Patch)
	}

	sep := "-"
	if dot {
		sep = "."
	}
	if g.Suffix == "" {
		if stable {
			str += sep + "stable"
		}
	} else {
		str += sep + g.Suffix
	}

	if mono && g.Mono {
		str += "_mono"
	}

	return
}

const org = "godotengine"
const stableRepo = "godot"
const latestRepo = "godot-builds"

func (g *Official) DownloadURL(p *platform.Platform) (url string) {
	// TODO: use strings.Builder

	url = "https://github.com/" + org + "/"

	// Unstable releases are in the 'godot-builds' repo
	if g.Suffix == "" {
		url += stableRepo
	} else {
		url += latestRepo
	}

	gStr := g.StringEx(false, true, false)
	url += "/releases/download/" + gStr + "/Godot_v" + gStr + "_"

	if g.Mono {
		url += "mono_"
	}

	switch p.OS {
	case platform.Windows:
		switch p.Arch {
		case platform.X86_32:
			url += "win32"
		case platform.X86_64:
			url += "win64"
		case platform.ARM_32:
			// NOTE: this is anticipatory.
			// Godot don't currently publish ARM32 Windows builds.
			url += "windows_arm32"
		case platform.ARM_64:
			url += "windows_arm64"
		}

		// Sigh
		if !g.Mono {
			url += ".exe"
		}

		url += ".zip"

	case platform.Linux:
		url += "linux"

		// SIGH
		if g.Mono {
			url += "_"
		} else {
			url += "."
		}

		switch p.Arch {
		case platform.X86_32:
			url += "x86_32"
		case platform.X86_64:
			url += "x86_64"
		case platform.ARM_32:
			url += "arm32"
		case platform.ARM_64:
			url += "arm64"
		}

		url += ".zip"

	case platform.MacOS:
		url += "macos.universal.zip"
	}
	// NOTE: nil purposefully missed out for ExportTemplatesURL.

	return
}

func (g *Official) ExportTemplatesURL() string {
	return g.DownloadURL(nil) + "export_templates.tpz"
}

func CurrentRelease(latest bool) (*Official, error) {
	streamStr := "stable"
	if latest {
		streamStr = "latest"
	}
	glog.Infof("Checking %s Godot release...", streamStr)

	repoName := stableRepo
	if latest {
		repoName = latestRepo
	}

	glog.Debugf("Fetching releases from repo '%s/%s'", org, repoName)
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		org,
		repoName,
		// We can reasonably expect there to have been a 4.x release within the
		// last 5 releases, so:
		&github.ListOptions{PerPage: 5},
	)
	if err != nil {
		glog.Errorf(
			"Couldn't fetch releases from repo '%s/%s': %v",
			org, repoName, err,
		)
		return nil, err
	}

	for i, release := range releases {
		name := *release.Name
		glog.Debugf("Release %d is: '%s'", i, name)

		if name[0] != '4' {
			glog.Debugf("Release's major version isn't 4 - skipping.")
			continue
		}

		official, err := Parse(name)
		if err != nil {
			glog.Errorf("Couldn't parse release '%s': %v", name, err)
			return nil, err
		}

		glog.Infof("=> %s", official.String())
		return official, nil
	}

	return nil, fmt.Errorf("no recent 4.x releases available")
}
