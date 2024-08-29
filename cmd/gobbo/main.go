package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/cmds"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

var description = `
If in a project directory, {COMMAND} defaults to {edit}.
`

var app = charli.App{
	Description: description,
	Commands: []charli.Command{
		cmds.New,
		cmds.Install,
		cmds.Edit,
		cmds.Run,
		// cmds.Export,
		cmds.Clean,
		// cmds.Add,
		// cmds.Remove,
		// cmds.Upgrade,
		cmds.Which,
		cmds.Info,
	},
	GlobalOptions: []charli.Option{
		opts.Log,
		opts.Store,
		opts.Godot,
	},
	ErrorHandler: func(err error) {
		glog.Error(err)
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
		"%s v%s - CLI toolchain for Godot 4.x\n",
		title,
		bi.Main.Version,
	)

	// If gobbo.toml exists, edit by default (fail silently on FS issue here).
	if _, err := os.Stat("gobbo.toml"); err == nil {
		app.DefaultCommand = "edit"
	}

	r := app.Parse(os.Args)

	switch r.Action {
	case charli.Proceed:
		r.RunCommand()
	case charli.Help:
		r.PrintHelp()
	}

	if r.Fail {
		os.Exit(1)
	}
}
