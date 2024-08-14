package opts

import cli "github.com/starriver/charli"

var Store = cli.Option{
	Short:    's',
	Long:     "store",
	Metavar:  "PATH",
	Headline: "Override Gobbo store path",
}

func StoreSetup(r *cli.Result) {

}
