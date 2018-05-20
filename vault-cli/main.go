// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main // "github.com/go-sharp/vault/vault-cli"
import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-sharp/vault"
	"github.com/go-sharp/vault/vault-cli/output"
)

// "io/ioutil"

// "github.com/go-sharp/vault"

func main() {
	g := vault.NewGenerator("/Users/sandro/Downloads", "./output",
		vault.RecursiveOption(false),
		vault.IncludeFilesOption("NEXUS.*.pdf$", "HandBrake-1.0.7.dmg$"),
		//vault.ExcludeFilesOption("templ", "/.git/*"),
	)
	//g.Run()
	_ = g
	l := output.NewDownloadsLoader()
	f, err := l.Load("/HandBrake-1.0.7.dmg")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Name: %v, Size: %v, Path: %v, ModTime: %v\n", f.Name(), f.Size(), f.Path(), f.ModTime())

	nf, err := os.OpenFile("test.dmg", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	fmt.Println(err)
	fmt.Println(io.Copy(nf, f.Read()))
	nf.Close()

	var s string
	fmt.Scan(&s)
	fmt.Println("finished")

}
