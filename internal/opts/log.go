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
		// opt.Value is guaranteed valid at this point.
		_ = glog.ParseLevel(opt.Value)
	}
}
