// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main // "github.com/go-sharp/vault/vault-cli"

import (
	"github.com/go-sharp/vault"
)

func main() {
	g := vault.NewGenerator("./examples", "./output")
	g.Run()
}
