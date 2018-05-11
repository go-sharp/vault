// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

import (
	"bytes"
	"io"
	"time"
)

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
	Read() io.ReadSeeker
}

type memFile struct {
	idx     int
	name    string
	modTime time.Time
	path    string
	size    int64
}

func (m memFile) Read() io.ReadSeeker {
	return bytes.NewReader(vaultAssetBin[m.idx])
}

var vaultAssetBin = [][]byte{}
