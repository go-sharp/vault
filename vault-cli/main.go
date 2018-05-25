// Copyright © 2018 The Vault Authors.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main // "github.com/go-sharp/vault/vault-cli"
import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-sharp/vault"
)

type arrayFlag []string

func (a *arrayFlag) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlag) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func main() {
	// Flag declarations
	var relpath, name, pkgName string
	var subdirs, nocomp bool
	var incl, excl arrayFlag

	flag.StringVar(&relpath, "rp", "", "Set relative path from the executing binary "+
		"to the source directory (for debug use only)")
	flag.StringVar(&name, "n", "", "Set the name of the embedded resources (default: source folder name)")
	flag.StringVar(&pkgName, "p", "", "Set the package name for the generated files (default: destination folder name)")
	flag.BoolVar(&subdirs, "s", false, "Include files in subdirectories")
	flag.BoolVar(&nocomp, "no-comp", false, "Do not compress files")
	flag.Var(&incl, "i", "Set files to include into the generated resource file (a list with regexp)")
	flag.Var(&excl, "e", "Set files to exclude from the generated resource file (a list with regexp)")

	flag.Usage = func() {
		fmt.Printf("vault-cli V%v © The Vault Authors\n\n", vault.Version)
		fmt.Println("Usage of vault-cli:")
		fmt.Println("vault-cli [options] source destination")
		flag.PrintDefaults()
	}

	flag.Parse()
	src := flag.Arg(0)
	dst := flag.Arg(1)

	if src == "" || dst == "" {
		flag.Usage()
		os.Exit(2)
	}

	generator := vault.NewGenerator(src, dst,
		vault.RelativePathOption(relpath),
		vault.PackageNameOption(pkgName),
		vault.ResourceNameOption(name),
		vault.WithSubdirsOption(subdirs),
		vault.CompressOption(nocomp),
		vault.IncludeFilesOption(incl...),
		vault.ExcludeFilesOption(excl...))

	generator.Run()
}
