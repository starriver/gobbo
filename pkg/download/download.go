package download

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

const prefix = "gobbo-download-"

func Download(url string) (string, error) {
	hash := sha256.Sum256([]byte(url))
	b64 := base64.URLEncoding.EncodeToString(hash[:])
	tmp := filepath.Join(os.TempDir(), prefix+b64)

	var offset int64

	if stat, err := os.Stat(tmp); !os.IsNotExist(err) {
		if err != nil {
			return "", err
		}
		offset = stat.Size()
	}
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Discard the last 4KiB to be reasonably safe.
	if offset >= 4096 {
		offset -= 4096
		f.Truncate(offset)
		f.Seek(offset, io.SeekStart)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Offset", fmt.Sprintf("bytes=%d-", offset))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %s", url, res.Status)
	}

	bar := progressbar.DefaultBytes(offset + res.ContentLength)
	err = bar.Set64(offset)
	if err != nil {
		glog.Warnf("couldn't set progress bar position: %v", err)
	}

	_, err = io.Copy(io.MultiWriter(f, bar), res.Body)
	if err != nil {
		return "", err
	}

	return tmp, nil
}
