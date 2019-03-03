// Copyright © 2019 The Vault Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:generate vault-cli -s -n gen ./testdata/assets ./testdata/gen

package vault

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/go-sharp/vault/testdata/gen"
)

func TestDirectoryContent(t *testing.T) {
	type fi struct {
		name  string
		size  int64
		isDir bool
	}
	testCases := []struct {
		desc string
		path string
		want []fi
	}{
		{
			desc: "Verify root directory",
			path: "/",
			want: []fi{
				{name: "bin", size: 46768, isDir: true},
				{name: "data", size: 133265, isDir: true},
				{name: ".somespecialfile", size: 2945, isDir: false},
				{name: "gopher.jpeg", size: 4664, isDir: false},
				{name: "text.txt", size: 645, isDir: false},
			},
		},
		{
			desc: "Verify bin directory",
			path: "/bin",
			want: []fi{
				{name: "structure.sql", size: 1618, isDir: false},
				{name: "umlet.jar", size: 45150, isDir: false},
			},
		},
		{
			desc: "Verify /data directory",
			path: "/data",
			want: []fi{
				{name: "json", size: 484, isDir: true},
				{name: "css.css", size: 107415, isDir: false},
				{name: "golang-header.jpg", size: 25366, isDir: false},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fs := gen.NewGenLoader()
			f, err := fs.Open(tc.path)
			if err != nil {
				t.Fatalf("Open: missing file %v error: %v\n", tc.path, err)
			}
			defer f.Close()

			fi, err := f.Readdir(0)
			if err != nil {
				t.Fatalf("Open: file %v error: %v\n", tc.path, err)
			}

			if len(fi) != len(tc.want) {
				t.Fatalf("Open: files count get: %v want: %v\n", len(fi), len(tc.want))
			}

			for i := range fi {
				if fi[i].Name() != tc.want[i].name {
					t.Fatalf("Open: name get: %v want: %v\n", fi[i].Name(), tc.want[i].name)
				}

				if fi[i].IsDir() != tc.want[i].isDir {
					t.Fatalf("Open: name = %v, isDir get: %v want: %v\n", fi[i].Name(), fi[i].IsDir(), tc.want[i].isDir)
				}

				if fi[i].Size() != tc.want[i].size {
					t.Fatalf("Open: name= %v, size get: %v want: %v\n", fi[i].Name(), fi[i].Size(), tc.want[i].size)
				}
			}
		})
	}
}

func TestSeekInvalidOffset(t *testing.T) {
	fname := "/text.txt"
	fs := gen.NewGenLoader()
	f, err := fs.Open("/text.txt")
	if err != nil {
		t.Fatalf("Seek: missing file %v error: %v\n", fname, err)
	}
	defer f.Close()

	buf := make([]byte, 145)
	if _, err := f.Read(buf); err != nil {
		t.Fatalf("Seek: %v\n", err)
	}

	if offset, err := f.Seek(-200, io.SeekCurrent); err == nil ||
		err.Error() != "Seek: invalid offset" || offset != 145 {
		t.Fatalf("Seek: get %v,  want = 145 -> error: %v, want = 'Seek: invalid offset'\n", offset, err)
	}
}

func TestSeekFile(t *testing.T) {
	fname := "/text.txt"
	fs := gen.NewGenLoader()
	f, err := fs.Open("/text.txt")
	if err != nil {
		t.Fatalf("Seek: missing file %v error: %v\n", fname, err)
	}
	defer f.Close()

	fheader1 := make([]byte, 145)
	if n, err := f.Read(fheader1); err != nil || n != 145 {
		t.Fatalf("Seek: get %v, want = 145 -> error: %v\n", n, err)
	}

	if n, err := f.Seek(200, io.SeekCurrent); err != nil || n != 345 {
		t.Fatalf("Seek: newpos %v, want = 345 -> error: %v\n", n, err)
	}

	fmiddle := make([]byte, 255)
	if n, err := f.Read(fmiddle); err != nil || n != 255 {
		t.Fatalf("Seek: get %v, want = 300 -> error: %v\n", n, err)
	}

	if n, err := f.Seek(145, io.SeekStart); err != nil || n != 145 {
		t.Fatalf("Seek: newpos %v, want = 145 -> error: %v\n", n, err)
	}

	fheader2 := make([]byte, 200)
	if n, err := f.Read(fheader2); err != nil || n != 200 {
		t.Fatalf("Seek: get %v, want = 200 -> error: %v\n", n, err)
	}

	if n, err := f.Seek(-45, io.SeekEnd); err != nil || n != 600 {
		t.Fatalf("Seek: newpos %v, want = 600 -> error: %v\n", n, err)
	}

	fend := make([]byte, 45)
	if n, err := f.Read(fend); err != io.EOF || n != 45 {
		t.Fatalf("Seek: get %v, want = 45 -> error: %v , want = io.EOF\n", n, err)
	}

	var buf bytes.Buffer
	buf.Write(fheader1)
	buf.Write(fheader2)
	buf.Write(fmiddle)
	buf.Write(fend)

	hash := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
	if hash != "7a25377731ea6bbd36b526ad53eca30b" {
		t.Fatalf("Seek: get %v, want = '7a25377731ea6bbd36b526ad53eca30b'\n", hash)
	}
}

func TestFileGeneration(t *testing.T) {
	type file struct {
		name string
		size int64
		modT int64
		md5  string
	}
	testCases := []struct {
		desc string
		path string
		want file
	}{
		{
			desc: "Check /text.txt",
			path: "/text.txt",
			want: file{
				name: "text.txt",
				size: 645,
				modT: 1551297978,
				md5:  "7a25377731ea6bbd36b526ad53eca30b",
			},
		},
		{
			desc: "Check /.somespecialfile",
			path: "/.somespecialfile",
			want: file{
				name: ".somespecialfile",
				size: 2945,
				modT: 1551298070,
				md5:  "c1ba6b78faf926c5dc8e6080c2cbfcdb",
			},
		},
		{
			desc: "Check /gopher.jpeg",
			path: "/gopher.jpeg",
			want: file{
				name: "gopher.jpeg",
				size: 4664,
				modT: 1551298135,
				md5:  "320b0375be0752fd9f9b6c6831de5002",
			},
		},
		{
			desc: "Check /bin/structure.sql",
			path: "/bin/structure.sql",
			want: file{
				name: "structure.sql",
				size: 1618,
				modT: 1485462479,
				md5:  "7da2dde5dfcb1d1ae99e445ed99bdd39",
			},
		},
		{
			desc: "Check /bin/umlet.jar",
			path: "/bin/umlet.jar",
			want: file{
				name: "umlet.jar",
				size: 45150,
				modT: 1459688258,
				md5:  "27c970063621fea97962a91ce321919b",
			},
		},
		{
			desc: "Check /data/css.css",
			path: "/data/css.css",
			want: file{
				name: "css.css",
				size: 107415,
				modT: 1486965487,
				md5:  "8ee28c0ea319689048bdf6dee1dbde43",
			},
		},
		{
			desc: "Check /data/golang-header.jpg",
			path: "/data/golang-header.jpg",
			want: file{
				name: "golang-header.jpg",
				size: 25366,
				modT: 1551298513,
				md5:  "d7e021175d992db20f6220bcb7c3fa11",
			},
		},
		{
			desc: "Check /data/json/appsettings.json",
			path: "/data/json/appsettings.json",
			want: file{
				name: "appsettings.json",
				size: 313,
				modT: 1486375025,
				md5:  "c0c18477dac05f6c883d8b78ccde6ba2",
			},
		},
		{
			desc: "Check /data/json/readme.md",
			path: "/data/json/readme.md",
			want: file{
				name: "readme.md",
				size: 171,
				modT: 1486377234,
				md5:  "37b5efe0778e08d0012ce73367e8ffee",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fs := gen.NewGenLoader()
			f, err := fs.Open(tc.path)
			if err != nil {
				t.Fatalf("FileGeneration: missing file %v error: %v\n", tc.path, err)
			}
			defer f.Close()

			finfo, err := f.Stat()
			if err != nil {
				t.Fatalf("FileGeneration: stat for file %v error: %v\n", tc.want.name, err)
			}

			if tc.want.name != finfo.Name() {
				t.Fatalf("FileGeneration: name got: %v want = %v\n", finfo.Name(), tc.want.name)
			}

			if tc.want.size != finfo.Size() {
				t.Fatalf("FileGeneration: size got: %v want = %v\n", finfo.Size(), tc.want.size)
			}

			if tc.want.modT != finfo.ModTime().Unix() {
				t.Fatalf("FileGeneration: time got: %v want = %v\n", finfo.ModTime().Unix(), tc.want.modT)
			}

			data, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatalf("FileGeneration: failed to read data for %v: %v \n", finfo.Name(), err)
			}

			hash := fmt.Sprintf("%x", md5.Sum(data))
			if hash != tc.want.md5 {
				t.Fatalf("FileGeneration: md5 got: %v want = %v\n", hash, tc.want.md5)
			}
		})
	}
}
