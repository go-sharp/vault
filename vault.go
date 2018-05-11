// Copyright © 2018 The Vault Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

type Vault struct {
	src      string
	dest     string
	pkgName  string
	excl     []string
	incl     []string
	recursiv bool
}
