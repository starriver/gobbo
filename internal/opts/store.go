package opts

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/starriver/charli"
	"github.com/starriver/gobbo/pkg/glog"
	"github.com/starriver/gobbo/pkg/store"
)

var Store = charli.Option{
	Short:    's',
	Long:     "store",
	Metavar:  "PATH",
	Headline: "Override Gobbo store path",
}

func StoreSetup(r *charli.Result) *store.Store {
	path := filepath.Join(xdg.DataHome, "gobbo")
	opt := r.Options["s"]
	if opt.IsSet {
		path = opt.Value
	}

	if r.Fail {
		// Short-circuit store creation/walking
		return nil
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
