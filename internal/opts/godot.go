package opts

import cli "github.com/starriver/charli"

var Godot = cli.Option{
	Short:    'g',
	Long:     "godot",
	Metavar:  "VERSION",
	Headline: "Specify Godot version",
}
