package cmds

import cli "github.com/starriver/charli"

var Which = cli.Command{
	Name:     "which",
	Headline: "Show Godot executable path",
}
