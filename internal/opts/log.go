package opts

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

var logOpt = charli.Option{
	Short:    'l',
	Long:     "log-level",
	Choices:  []string{"debug", "info", "warn", "error"},
	Metavar:  "LEVEL",
	Headline: "Set logging level",
}

func logSetup(r *charli.Result) {
	opt := r.Options["l"]
	if opt.IsSet {
		if err := glog.ParseLevel(opt.Value); err != nil {
			glog.Error(err)
		}
	}
}
