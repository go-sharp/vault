// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main // "github.com/go-sharp/vault/vault-cli"
import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-sharp/vault"
	"github.com/go-sharp/vault/vault-cli/output"
)

// "io/ioutil"

// "github.com/go-sharp/vault"

func main() {
	g := vault.NewGenerator("../../../../../../../Downloads/test", "./output",
		vault.RecursiveOption(true),
		//vault.RelativePathOption("./etc"),
		//vault.ResourceNameOption("cool"),
		//vault.PackageNameOption("myPack"),
		//vault.IncludeFilesOption("NEXUS.*.pdf$", ".*HandBrake-1.0.7.dmg$"),
		//vault.ExcludeFilesOption("templ", "/.git/*"),
	)
	//g.Run()
	_ = g

	loader := output.NewTestLoader()

	f, err := loader.Load("/murmur3collisions/murmurhash3.h")
	if err != nil {
		log.Fatalln(err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintln(os.Stdout, string(b))

	// r := strings.NewReader(output.VaultAssetBinDownloads)
	// b, _ := ioutil.ReadAll(r)
	// fmt.Println(len(b))
	// pdf := output.VaultAssetBinDownloads[0:13011479]

	// //io.Copy(os.Stdout, strings.NewReader(pdf))
	// bb, _ := ioutil.ReadAll(strings.NewReader(pdf))
	// fmt.Println(len(pdf), len(bb), len(output.VaultAssetBinDownloads), 94490)
	// zr, err := zlib.NewReader(strings.NewReader(pdf))
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// pb, err := ioutil.ReadAll(zr)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// ioutil.WriteFile("bla.dmg", pb, 0755)

}
