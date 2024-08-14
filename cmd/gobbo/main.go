package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	cli "github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/cmds"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

var description = `
If in a project directory, {COMMAND} defaults to {edit}.
`

var app = cli.App{
	Description: description,
	Commands: []cli.Command{
		cmds.New,
		cmds.Install,
		cmds.Edit,
		cmds.Run,
		cmds.Export,
		cmds.Clean,
		cmds.Add,
		cmds.Remove,
		cmds.Upgrade,
		cmds.Which,
		cmds.Info,
	},
}

func main() {
	glog.CurrentLevel = glog.LevelInfo

	title := color.New(color.FgHiBlue, color.Bold).Sprint("Gobbo")
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build info")
	}
	app.Headline = fmt.Sprintf(
		"%s v%s - CLI toolchain for Godot 4.x",
		title,
		bi.Main.Version,
	)

	// If gobbo.toml exists, edit by default (fail silently on FS issue here).
	if _, err := os.Stat("gobbo.toml"); err == nil {
		app.DefaultCommand = "edit"
	}

	r := app.Parse(os.Args)

	ok = false
	switch r.Action {
	case cli.Proceed:
		ok = r.RunCommand()
	case cli.HelpOK:
		r.PrintHelp()
		ok = true
	case cli.HelpError:
		fmt.Fprint(os.Stderr, app.Help(os.Args[0], r.Command))
	case cli.Fatal:
		// Nothing else to do.
	}

	for _, err := range r.Errs {
		glog.Error(err)
	}

	if !ok {
		os.Exit(1)
	}
}
