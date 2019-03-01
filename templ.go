// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

const sharedTypesTempl = `
import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// ErrNotFound is returned if the requested file was not found.
var ErrNotFound = errors.New("file not found")

// AssetLoader implements a function to load an asset from the vault
type AssetLoader interface {
	// Open loads a file from the vault.
	Open(name string) (http.File, error)
}

// assetMap holds all information about the embedded files
type assetMap map[string]memFile

func createDirFile(path string, assets assetMap) http.File {
	md := memDir{dir: path}
	dirs := map[string]*memDir{}

	for k, v := range assets {
		if v.path == md.dir {
			md.files = append(md.files, &v)
			continue
		}
		if strings.HasPrefix(k, md.dir) {
			p := strings.TrimLeft(k, md.dir)
			if p[0] == '/' {
				p = p[1:]
			}

			idx := len(p)
			for i := 0; i < len(p); i++ {
				if p[i] == '/' {
					idx = i
					break
				}
			}
			p = p[:idx]
			if dir, ok := dirs[p]; ok {
				dir.size += v.size
			} else {
				newDir := memDir{dir: fmt.Sprintf("%v/%v", md.dir, p), size: v.size}
				md.files = append(md.files, newDir)
				dirs[p] = &newDir
			}
		}
	}

	sort.Slice(md.files, func(i int, j int) bool {
		return md.files[i].Name() < md.files[j].Name()
	})
	return &md
}
`

const releaseImportTempl = `
import (
	"compress/zlib"
	"errors"
	"io"
	"os"
	"strings"
	"time"
	"net/http"
)

`

const releaseFileTempl = `
type assetReader interface {
	io.ReadCloser
	zlib.Resetter
}

type memFile struct {
	r       assetReader
	rOffset int64
	offset  int64
	name    string
	modTime time.Time
	path    string
	length  int64
	size    int64
}

// Readdir see os.File Readdir function
func (m memFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, io.EOF
}

func (m memFile) Close() error {
	if m.r == nil {
		return nil
	}
	return m.r.Close()
}

func (m *memFile) resetReader() (err error) {
	if m.r == nil {
		var r io.ReadCloser
		r, err = zlib.NewReader(strings.NewReader(vaultAssetBin{{.Suffix}}[m.offset : m.offset+m.length]))
		m.r = r.(assetReader)
	} else {
		err = m.r.Reset(strings.NewReader(vaultAssetBin{{.Suffix}}[m.offset:m.offset+m.length]), nil)
	}

	if err != nil {
		return err
	}

	m.rOffset = 0
	return nil
}

func (m *memFile) Read(p []byte) (n int, err error) {
	if m.r == nil {
		if err := m.resetReader(); err != nil {
			return 0, err
		}
	}

	n, err = m.r.Read(p)
	m.rOffset += int64(n)
	return n, err
}

func (m *memFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		offset += m.rOffset
	case io.SeekStart:
	case io.SeekEnd:
		offset += m.size
	default:
		return 0, errors.New("Seek: invalid whence")

	}

	if offset < 0 {
		return m.rOffset, errors.New("Seek: invalid offset")
	}

	if offset < m.rOffset {
		if err := m.resetReader(); err != nil {
			return m.rOffset, err
		}
	}

	buf := make([]byte, offset - m.rOffset)
	_, err := m.Read(buf)
	return m.rOffset, err
}

func (m memFile) Stat() (os.FileInfo, error) {
	return m, nil
}

func (m memFile) Name() string {
	return m.name
}

func (m memFile) Size() int64 {
	return m.size
}

func (m memFile) Mode() os.FileMode {
	return 0444
}

func (m memFile) ModTime() time.Time {
	return m.modTime
}

func (m memFile) IsDir() bool {
	return false
}

func (m memFile) Sys() interface{} {
	return nil
}


type memDir struct {
	dir   string
	files []os.FileInfo
	size  int64
}

func (m memDir) Close() error {
	return nil
}

func (m memDir) Read(p []byte) (n int, err error) {
	return 0, errors.New("Read: invalid operation on directory")
}

func (m memDir) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("Seek: invalid operation on directory")
}

func (m *memDir) Readdir(count int) ([]os.FileInfo, error) {
	defer func() {
		if count <= 0 || count >= len(m.files) {
			m.files = m.files[0:0]
		} else {
			m.files = m.files[:count]
		}
	}()

	if count <= 0 {
		return m.files[:], nil
	} else if count >= len(m.files) {
		return m.files[:], io.EOF
	}

	return m.files[:count], nil
}

func (m memDir) Stat() (os.FileInfo, error) {
	return m, nil
}

func (m memDir) Name() string {
	if m.dir == "/" {
		return "/"
	}
	return m.dir[strings.LastIndex(m.dir, "/")+1:]
}

func (m memDir) Size() int64 {
	return m.size
}

func (m memDir) Mode() os.FileMode {
	return os.FileMode(0555)
}

func (m memDir) ModTime() time.Time {
	// Until now no directory information is stored
	// in the asset data, so for now we return the current
	// time.
	return time.Now()
}

func (m memDir) IsDir() bool {
	return true
}

func (m memDir) Sys() interface{} {
	return nil
}

type loader struct {
	fm assetMap
}

func (l loader) Open(name string) (http.File, error) {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	if len(name) > 1 && name[len(name)-1] == '/' {
		name = strings.TrimRight(name, "/")
	}

	if v, ok := l.fm[name]; ok {
		return &v, nil
	}

	for _, v := range l.fm {
		if strings.HasPrefix(v.path, name) {
			return createDirFile(name, l.fm), nil
		}
	}

	return nil, os.ErrNotExist
}

// New{{.Suffix}}Loader returns a new AssetLoader for the {{.Suffix}} resources.
func New{{.Suffix}}Loader() AssetLoader {
	loader := &loader{
		fm: assetMap{
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
	"net/http"
)

type debugLoader struct {
	base string
}

func (d debugLoader) Open(name string) (http.File, error) {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	return os.Open(getFullPath(d.base, name))
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
