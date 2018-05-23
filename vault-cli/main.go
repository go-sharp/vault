// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main // "github.com/go-sharp/vault/vault-cli"
import (
	"github.com/go-sharp/vault"
)

// "io/ioutil"

// "github.com/go-sharp/vault"

func main() {
	g := vault.NewGenerator("../../../../../../../Downloads/test", "./output",
		vault.RecursiveOption(false),
		vault.CompressOption(false),
		//vault.RelativePathOption("./etc"),
		//vault.ResourceNameOption("cool"),
		//vault.PackageNameOption("myPack"),
		//vault.IncludeFilesOption("NEXUS.*.pdf$", ".*HandBrake-1.0.7.dmg$"),
		vault.ExcludeFilesOption(".*.cc$"),
	)
	g.Run()
	_ = g

	/*
		loader := output.NewTestLoader()

		f, err := loader.Load("BT-K Feedback Aufgabe 5.1 MFR - Samuel Weidmann.docx")
		if err != nil {
			log.Fatalln(err)
		}

		_, err = ioutil.ReadAll(f)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintln(os.Stdout, f.Name(), f.ModTime(), f.Path(), f.Size())
	*/
	//ioutil.WriteFile("bla.docx", b, 0755)

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
