package opts

import (
	cli "github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

var Log = cli.Option{
	Short:    'l',
	Long:     "log-level",
	Choices:  []string{"debug", "info", "warn", "error"},
	Metavar:  "LEVEL",
	Headline: "Set logging level",
}

func LogSetup(r *cli.Result) {
	opt := r.Options["l"]
	if opt.IsSet {
		glog.ParseLevel(opt.Value)
	}
}
