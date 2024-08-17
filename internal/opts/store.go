package opts

import (
	"github.com/adrg/xdg"
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/pkg/glog"
	"gitlab.com/starriver/gobbo/pkg/store"
)

var Store = charli.Option{
	Short:    's',
	Long:     "store",
	Metavar:  "PATH",
	Headline: "Override Gobbo store path",
}

func StoreSetup(r *charli.Result) *store.Store {
	path := xdg.DataHome
	ro := r.Options["s"]
	if ro.IsSet {
		path = ro.Value
	}

	s, errs := store.New(path)
	if s == nil {
		glog.Error("invalid store:")
		for _, err := range errs {
			glog.Error(err)
		}
		r.Fail = true
	}
	return s
}
