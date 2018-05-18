// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

const sharedTypesTempl = `
import (
	"bytes"
	"errors"
	"io"
	"time"
)

// ErrNotFound is returned if the requested file was not found.
var ErrNotFound = errors.New("file not found")

// ReadSeekCloser interface
type ReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
}

// File is the vault abstraction of a file.
type File interface {
	// Size returns the size of the file.
	Size() int64
	// Name returns the name of the file.
	Name() string
	// ModTime returns the modification time.
	ModTime() time.Time
	// Path is the registered path within the vault.
	Path() string
	// Read returns a ReadSeeker for this file.
	Read() ReadSeekCloser
}

// AssetLoader implements a function to load a asset from the vault
type AssetLoader interface {
	// Load loads a file from the vault.
	Load(name string) (File, error)
}

type memReader struct {
	*bytes.Reader
}

func (m *memReader) Close() error {
	return nil
}

type memFile struct {
	idx     int
	name    string
	modTime time.Time
	path    string
	base    string
	size    int64
}

`

const vaultAssetBinTempl = `var vaultAssetBin{{.Suffix}} = [][]byte{}`

const releaseFileTempl = `
import (
	"bytes"
	"time"
)

func (m memFile) Read() ReadSeekCloser {
	return &memReader{Reader: bytes.NewReader(vaultAssetBin{{.Suffix}}[m.idx])}
}

func (m memFile) Size() int64 {
	return m.size
}

func (m memFile) Name() string {
	return m.name
}

func (m memFile) ModTime() time.Time {
	return m.modTime
}

func (m memFile) Path() string {
	return m.path
}

type loader struct {
	fm map[string]memFile
}

func (l loader) Load(name string) (File, error) {
	if v, ok := l.fm[name]; ok {
		return &v, nil
	}
	return nil, ErrNotFound
}

// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	loader := &loader{
		fm: map[string]memFile{
		{{- range $idx, $el := .Files }}
			"{{$el.Path}}/{{$el.Name}}": memFile{idx: {{$idx}}, name: "{{$el.Name}}", modTime: time.Unix({{$el.ModTime.Unix}}, 0), path: "{{$el.Path}}", size: {{$el.Size}}},
		{{- end}}
		},
	}
	return loader
}

`

const debugFileTemp = `
import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"
)

func (m memFile) Read() ReadSeekCloser {
	f, err := os.Open(getFullPath(m.base, m.path))
	if err != nil {
		return &memReader{Reader: bytes.NewReader([]byte{})}
	}
	return f
}

func (m memFile) Size() int64 {
	fi, err := os.Stat(getFullPath(m.base, m.path))
	if err != nil {
		return 0
	}
	return fi.Size()
}

func (m memFile) Name() string {
	fi, err := os.Stat(getFullPath(m.base, m.path))
	if err != nil {
		return ""
	}
	return fi.Name()
}

func (m memFile) ModTime() time.Time {
	fi, err := os.Stat(getFullPath(m.base, m.path))
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

func (m memFile) Path() string {
	return m.path
}

type debugLoader struct {
	base string
}

func (d debugLoader) Load(name string) (File, error) {
	_, err := os.Stat(getFullPath(d.base, name))
	if err != nil {
		return nil, err
	}
	return &memFile{base: d.base, path: name}, nil
}

func getFullPath(b, p string) string {
	return path.Clean(fmt.Sprintf("%v/%v", b, p))
}

// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	return &debugLoader{base: "{{.Base}}"}
}

`

const fileHeaderTempl = `// This file is generated by the vault-cli command line utility.
// It offers a easy way to embed binary resources into a go executable.
// Do not edit this file, it will be overwritten on the next run of the vault-cli utility.

package {{.}}
`
