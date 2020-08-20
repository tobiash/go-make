package mk

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/mod/sumdb/dirhash"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

// A target on the filesystem (directory or file)
type FileTarget struct {
	Dir string
	Path string
}

func init() {
	RegisterTarget("file", func(u url.URL) Target {
		return &FileTarget{Path: u.Path}
	})
}

func (f *FileTarget) Name() string {
	return (&url.URL{Path: filepath.ToSlash(f.Path), Scheme: "file"}).String()
}

func (f *FileTarget) Check(digest string) (TargetStatus, error) {
	cDigest, exists, err := f.digest()
	return TargetStatus{
		UpToDate:      (digest == "" || digest == cDigest) && exists,
		Exists:        exists,
		CurrentDigest: cDigest,
	}, err
}

func (f *FileTarget) digest() (string, bool, error) {
	p := filepath.Join(f.Dir, f.Path)
	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		return "", false, nil
	} else if err != nil {
		return "", false, errors.Wrapf(err, "cannot digest %s", f.Path)
	}
	if fi.IsDir() {
		d, err := dirhash.HashDir(p, "", dirhash.DefaultHash)
		return d, true, err
	} else {
		h, err := fileHash(p)
		if err != nil {
			return "", true, err
		}
		return fmt.Sprintf("f: %s", base64.StdEncoding.EncodeToString(h)), true, nil
	}
}

func fileHash(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
