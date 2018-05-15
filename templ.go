// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

const importTempl = `
import (
	"bytes"
	"errors"
	"io"
	"time"
)
`

const debugImportTempl = `
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"
)
`

const interfaceTempl = `
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
`

const memReaderTempl = `
type memReader struct {
	bytes.Reader
}

func (m *memReader) Close() error {
	return nil
}
`

const memFileTempl = `
type memFile struct {
	idx     int
	name    string
	modTime time.Time
	path    string
	base    string
	size    int64
}
`

// Template for the in memory memFile methods.
const inMemoryFileMethodTempl = `
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
`

const debugFileMethodTemp = `
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
`
const vaultAssetBinTempl = `var vaultAssetBin{{.Suffix}} = [][]byte{}`

const errorConstantTempl = `
// ErrNotFound is returned if the requested file was not found.
var ErrNotFound = errors.New("file not found")
`

const memLoaderTempl = `
type loader struct {
	fm map[string]File
}

func (l loader) Load(name string) (File, error) {
	if v, ok := l.fm[name]; ok {
		return v, nil
	}
	return nil, ErrNotFound
}
`

const memNewLoaderTempl = `
// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	loader := &loader{
		fm: map[string]memFile{
		{{- range $idx, $el := .Files }}
			"{{$el.Path}}/{{$el.Name}}": File{idx: {{$idx}}, name: "{{$el.Name}}", modTime: time.Unix({{$el.ModTime.Unix}}, 0), path: "{{$el.Path}}", size: {{$el.Size}}},
		{{- end}}
		}
	}
	return loader
}
`

const debugLoaderTempl = `
type debugLoader struct {
	base string
}

func (d debugLoader) Load(name string) (File, error) {
	fi, err := os.Stat(getFullPath(d.base, name))
	if err != nil {
		return nil, err
	}
	return &memFile{base: d.base, path: name}, nil
}

func getFullPath(b, p string) string {
	return path.Clean(fmt.Sprintf("%v%v%v", b, os.PathSeparator, p))
}
`

const debugNewLoaderTempl = `
// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	return &debugLoader{base: {{.Base}}}
}
`
