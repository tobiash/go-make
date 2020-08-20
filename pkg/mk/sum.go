package mk

import (
	"bufio"
	"context"
	"golang.org/x/mod/sumdb/storage"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type YamlSumStorageFile struct {
	Path string
	Perm os.FileMode
}

type yamlStorageFileTransaction struct {
	sum             map[string]string
	dirty, readOnly bool
	lock            sync.Mutex
}

func (y *yamlStorageFileTransaction) ReadValue(ctx context.Context, key string) (value string, err error) {
	y.lock.Lock()
	defer y.lock.Unlock()
	return y.sum[key], nil
}

func (y *yamlStorageFileTransaction) ReadValues(ctx context.Context, keys []string) (values []string, err error) {
	y.lock.Lock()
	defer y.lock.Unlock()
	result := make([]string, len(keys))
	for i := range keys {
		result[i] = y.sum[keys[i]]
	}
	return result, nil
}

func (y *yamlStorageFileTransaction) BufferWrites(writes []storage.Write) error {
	y.lock.Lock()
	defer y.lock.Unlock()
	if len(writes) == 0 {
		return nil
	}
	y.dirty = true
	for i := range writes {
		y.sum[writes[i].Key] = writes[i].Value
	}
	return nil
}

func (j *YamlSumStorageFile) read() (*yamlStorageFileTransaction, error) {
	var tr yamlStorageFileTransaction
	file, err := os.Open(j.Path)
	if os.IsNotExist(err) {
		tr.sum = make(map[string]string)
		return &tr, nil
	} else if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	rdr := bufio.NewReader(file)
	if _, err = rdr.Peek(1); err == io.EOF {
		return &tr, nil
	} else if err != nil {
		return nil, err
	}
	if err := yaml.NewDecoder(rdr).Decode(&tr.sum); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (j *YamlSumStorageFile) ReadOnly(ctx context.Context, f func(context.Context, storage.Transaction) error) error {
	tr, err := j.read()
	if err != nil {
		return err
	}
	return f(ctx, tr)
}

func (j *YamlSumStorageFile) ReadWrite(ctx context.Context, f func(context.Context, storage.Transaction) error) error {
	tr, err := j.read()
	if err != nil {
		return err
	}
	if err := f(ctx, tr); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if !tr.dirty {
		return nil
	}
	dir := filepath.Dir(j.Path)
	if dir != "" {
		_ = os.MkdirAll(dir, 0700)
	}
	file, err := os.OpenFile(j.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, j.Perm)
	if err != nil {
		return err
	}
	if err := yaml.NewEncoder(file).Encode(tr.sum); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
