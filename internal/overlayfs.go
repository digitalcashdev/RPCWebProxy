package internal

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/digitalcashdev/rpcproxy/static"
)

type OverlayFS struct {
	LocalFS     http.FileSystem
	EmbedFS     http.FileSystem
	WebRoot     string
	WebRootOnly bool
}

func (mfs *OverlayFS) ForceLocalOrEmbedOpen(name string) (http.File, error) {
	localPath := filepath.Join(mfs.WebRoot, name)
	info, err := os.Stat(localPath)
	if err == nil && !info.IsDir() {
		return mfs.LocalFS.Open(name)
	}

	path := path.Join(static.Prefix, name)
	return mfs.EmbedFS.Open(path)
}

func (mfs *OverlayFS) Open(name string) (http.File, error) {
	if len(mfs.WebRoot) > 0 {
		localPath := filepath.Join(mfs.WebRoot, name)
		info, err := os.Stat(localPath)
		if err == nil && !info.IsDir() {
			return mfs.LocalFS.Open(name)
		}
	}

	if !mfs.WebRootOnly {
		path := path.Join(static.Prefix, name)
		return mfs.EmbedFS.Open(path)
	}

	return nil, os.ErrNotExist
}
