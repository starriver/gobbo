package godot

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-github/v63/github"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

type Official struct {
	Minor  uint
	Patch  uint
	Suffix string
	Mono   bool
}

func (g Official) BinaryPath() string {
	str := g.String() // Heh
	return filepath.Join("official", str, "godot-"+str)
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

func (g Official) DownloadURL(platform Platform) (url string) {
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

	switch platform {
	case Windows:
		url += "win64"

		// Sigh
		if !g.Mono {
			url += ".exe"
		}

		url += ".zip"

	case Linux:
		url += "linux"

		// SIGH
		if g.Mono {
			url += "_"
		} else {
			url += "."
		}

		url += "x86_64.zip"

	case MacOS:
		url += "macos.universal.zip"

	case ExportTemplates:
		url += "export_templates.tpz"
	}

	return
}

func CurrentRelease(latest bool) (Official, error) {
	streamStr := "stable"
	if latest {
		streamStr = "latest"
	}
	glog.Infof("Checking %s Godot release...", streamStr)

	repoName := stableRepo
	if latest {
		repoName = latestRepo
	}

	glog.Debugf("Fetching releases from repo %s/%s", org, repoName)
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(
		context.Background(),
		org,
		repoName,
		&github.ListOptions{PerPage: 1},
	)
	if err != nil {
		glog.Errorf(
			"Couldn't fetch releases from repo %s/%s: %v",
			org, repoName, err,
		)
		return Official{}, err
	}

	releaseTitle := releases[0].Name
	glog.Debugf("First release is: '%s'", *releaseTitle)

	official, err := Parse(*releaseTitle)
	if err != nil {
		glog.Errorf("Couldn't parse release '%s': %v", *releaseTitle, err)
		return Official{}, err
	}

	return official, nil
}
