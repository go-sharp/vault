// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

const sharedTypesTempl = `
import (
	"errors"
	"io"
	"time"
)

// ErrNotFound is returned if the requested file was not found.
var ErrNotFound = errors.New("file not found")

// File is the vault abstraction of a file.
type File interface {
	io.ReadCloser
	// Size returns the size of the file.
	Size() int64
	// Name returns the name of the file.
	Name() string
	// ModTime returns the modification time.
	ModTime() time.Time
	// Path is the registered path within the vault.
	Path() string
}

// AssetLoader implements a function to load an asset from the vault
type AssetLoader interface {
	// Load loads a file from the vault.
	Load(name string) (File, error)
}

`

const releaseImportTempl = `
import (
	"compress/zlib"
	"strings"
	"io"
	"time"
)

`

const releaseFileTempl = `
type memFile struct {
	r		io.ReadCloser
	offset	int64
	name    string
	modTime time.Time
	path    string
	length  int64
	size    int64
}

func (m memFile) Close() error {
	return m.r.Close()
}
func (m memFile) Read(p []byte) (n int, err error) {
	return m.r.Read(p)
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
	if m.path == "" {
		return "/"
	}

	return m.path
}

type loader struct {
	fm map[string]memFile
}

func (l loader) Load(name string) (File, error) {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	if v, ok := l.fm[name]; ok {
		r, err := zlib.NewReader(strings.NewReader(vaultAssetBin{{.Suffix}}[v.offset : v.offset+v.length]))
		if err != nil {
			return nil, err
		}
		v.r = r
		return &v, nil
	}
	return nil, ErrNotFound
}

// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	loader := &loader{
		fm: map[string]memFile{
		{{- range  $el := .Files }}
			"{{join $el.Path $el.Name}}": memFile{offset: {{$el.Offset}},
				name: "{{$el.Name}}",
				modTime: time.Unix({{$el.ModTime.Unix}}, 0),
				path: "{{$el.Path}}",
				size: {{$el.Size}},
				length: {{$el.Length}},
				},
		{{- end}}
		},
	}
	return loader
}

`

const debugFileTemp = `
import (
	"fmt"
	"os"
	"path"
    "strings"
    "time"
)

type memFile struct {
	f    *os.File
	stat os.FileInfo
	path string
	base string
}

func (m memFile) Size() int64 {
	return m.stat.Size()
}

func (m memFile) Name() string {
	return m.stat.Name()
}

func (m memFile) ModTime() time.Time {
	return m.stat.ModTime()
}

func (m memFile) Path() string {
	return m.path
}

func (m memFile) Read(p []byte) (n int, err error) {
	return m.f.Read(p)
}

func (m memFile) Close() error {
	return m.f.Close()
}

type debugLoader struct {
	base string
}

func (d debugLoader) Load(name string) (File, error) {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	f, err := os.Open(getFullPath(d.base, name))
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &memFile{base: d.base, path: path.Clean(strings.TrimSuffix(name, stat.Name())), f: f, stat: stat}, nil
}

func getFullPath(b, p string) string {
	return path.Clean(fmt.Sprintf("%v/%v", b, path.Clean("/"+p)))
}

// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	return &debugLoader{base: "{{.Base}}"}
}

`

const fileHeaderTempl = `// This file is generated by the vault-cli command line utility.
// It offers a easy way to embed binary resources into a go executable.
// DO NOT EDIT this file, it will be overwritten on the next run of the vault-cli utility.

package {{.}}
`
