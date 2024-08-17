package opts

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

var Log = charli.Option{
	Short:    'l',
	Long:     "log-level",
	Choices:  []string{"debug", "info", "warn", "error"},
	Metavar:  "LEVEL",
	Headline: "Set logging level",
}

func LogSetup(r *charli.Result) {
	opt := r.Options["l"]
	if opt.IsSet {
		if err := glog.ParseLevel(opt.Value); err != nil {
			glog.Error(err)
		}
	}
}
