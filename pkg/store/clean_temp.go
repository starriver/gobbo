package store

import (
	"os"
	"path/filepath"
	"time"

	"github.com/starriver/gobbo/pkg/glog"
)

// Remove all files in store tmp/ that are older than a month.
func (s *Store) CleanTemp() {
	tmp := s.Join("tmp")
	dir, err := os.ReadDir(tmp)
	if err != nil {
		glog.Warnf("Couldn't read directory '%s': %v", tmp, err)
		return
	}

	expiry := time.Now().AddDate(0, 0, -31)
	glog.Debugf("Removing tempfiles in '%s' for mod time < '%s'", tmp, expiry.String())

	for _, entry := range dir {
		path := filepath.Join(tmp, entry.Name())
		st, err := entry.Info()
		if err != nil {
			glog.Warnf(
				"Couldn't stat '%s': %v",
				filepath.Join(tmp, path),
				err,
			)
			continue
		}

		if st.ModTime().Before(expiry) {
			glog.Debugf("Removing: %s", path)
			err = os.RemoveAll(path)
			if err != nil {
				glog.Warnf("Couldn't remove '%s': %v", path, err)
			}
		} else {
			glog.Debugf("Keeping: %s", path)
		}
	}

}
