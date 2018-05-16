// Copyright © 2018 The Vault Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	ttImportTempl             = "imports"
	ttDebugImportTempl        = "debugImports"
	ttSharedTypesTempl        = "sharedTypes"
	ttInMemoryFileMethodTempl = "inMemoryFileMethod"
	ttDebugFileMethodTempl    = "debugFileMethod"
	ttVaultAssetBinTempl      = "vaultAssetBin"
	ttMemLoaderTempl          = "memLoader"
	ttMemNewLoaderTempl       = "memNewLoader"
	ttDebugLoaderTempl        = "debugLoader"
	ttDebugNewLoaderTempl     = "debugNewLoader"
	ttFileHeaderTempl         = "fileHeaderTempl"
)

var ttRepo *template.Template

func init() {
	ttRepo = template.Must(template.New(ttImportTempl).Parse(importTempl))
	ttRepo = template.Must(ttRepo.New(ttDebugImportTempl).Parse(debugImportTempl))
	ttRepo = template.Must(ttRepo.New(ttSharedTypesTempl).Parse(sharedTypesTempl))
	ttRepo = template.Must(ttRepo.New(ttInMemoryFileMethodTempl).Parse(inMemoryFileMethodTempl))
	ttRepo = template.Must(ttRepo.New(ttDebugFileMethodTempl).Parse(debugFileMethodTemp))
	ttRepo = template.Must(ttRepo.New(ttVaultAssetBinTempl).Parse(vaultAssetBinTempl))
	ttRepo = template.Must(ttRepo.New(ttMemLoaderTempl).Parse(memLoaderTempl))
	ttRepo = template.Must(ttRepo.New(ttMemNewLoaderTempl).Parse(memNewLoaderTempl))
	ttRepo = template.Must(ttRepo.New(ttDebugLoaderTempl).Parse(debugLoaderTempl))
	ttRepo = template.Must(ttRepo.New(ttDebugNewLoaderTempl).Parse(debugNewLoaderTempl))
	ttRepo = template.Must(ttRepo.New(ttFileHeaderTempl).Parse(fileHeaderTempl))

}

// Generator creates a vault with files in there binary representation.
type Generator struct {
	config      GeneratorConfig
	sharedFile  string
	debugFile   string
	releaseFile string
}

// Run starts the vault generation, may panic if an error occurs.
func (g *Generator) Run() {
	log.Println("starting vault generation...")
	g.createShared()
}

func (g *Generator) createShared() {
	sf, err := os.OpenFile(g.sharedFile, os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatalf("failed to create shared vault file for resource '%v': %v\n", g.config.name, err)
	}
	defer func() {
		if err := sf.Close(); err != nil {
			log.Fatalf("failed to close file: %v", err)
		}
	}()

	ttRepo.ExecuteTemplate(sf, ttFileHeaderTempl, nil)
	ttRepo.ExecuteTemplate(sf, ttImportTempl, nil)
	ttRepo.ExecuteTemplate(sf, ttSharedTypesTempl, nil)
}

// GeneratorConfig configures the vault generator.
type GeneratorConfig struct {
	src      string
	dest     string
	relPath  string
	name     string
	pkgName  string
	excl     []string
	incl     []string
	recursiv bool
}

// GeneratorOption configures the vault generator.
type GeneratorOption func(g *GeneratorConfig)

// NewGenerator creates a new generator instance with the given options.
func NewGenerator(src, dest string, options ...GeneratorOption) Generator {
	cfg := GeneratorConfig{src: src, dest: dest}
	for i := range options {
		options[i](&cfg)
	}
	initGeneratorConfig(&cfg)
	g := Generator{config: cfg}

	g.sharedFile = cleanSlashedPath(dest, fmt.Sprintf("shared_%v_vault.go", cfg.name))
	g.debugFile = cleanSlashedPath(dest, fmt.Sprintf("debug_%v_vault.go", cfg.name))
	g.releaseFile = cleanSlashedPath(dest, fmt.Sprintf("%v_vault.go", cfg.name))
	return g
}

func cleanSlashedPath(s ...string) string {
	return filepath.Clean(filepath.ToSlash(
		strings.Join(s, string(os.PathSeparator))))
}

func initGeneratorConfig(cfg *GeneratorConfig) {
	if cfg.pkgName == "" {
		if cfg.pkgName = lastPath(cfg.dest); cfg.pkgName == "" {
			log.Fatalln("could not determine package name: try to set package name manually")
		}
	}

	if cfg.name == "" {
		if cfg.name = lastPath(cfg.src); cfg.name == "" {
			cfg.name = cfg.pkgName
		}
	}
}

func lastPath(p string) string {
	idx := strings.LastIndex(p, string(os.PathSeparator))
	if idx == -1 {
		return ""
	}
	return p[idx+1:]
}

// PackageNameOption sets the package name of the generated vault files.
// If not set, the generator tries to deduce the correct package name.
func PackageNameOption(name string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.pkgName = name
	}
}

// ExcludeFilesOption sets the files to exclude in the generation process.
// Only relative paths will be checked, so pattern must not include the fullpath.
// Pattern matching follows the rules of filepath.Match (see https://golang.org/pkg/path/filepath/#Match).
func ExcludeFilesOption(name ...string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.excl = append(c.excl, name...)
	}
}

// IncludeFilesOption sets the files to include in the generation process.
// Only specified files and files not matching any exclusion pattern will be included in the generation process.
// Only relative paths will be checked, so pattern must not include the fullpath.
// Pattern matching follows the rules of filepath.Match (see https://golang.org/pkg/path/filepath/#Match).
func IncludeFilesOption(name ...string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.incl = append(c.incl, name...)
	}
}

// RecursiveOption sets the recursive mode for the generation process.
// If true the generator walks recurively down the folder hierarchy.
func RecursiveOption(recursive bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.recursiv = recursive
	}
}

// ResourceNameOption sets the name of the generated resources.
func ResourceNameOption(name string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.name = name
	}
}

// RelativePathOption sets the relative path for the debug asset loader.
// If not specified the generator uses the relative path from the directory
// where the generator was invoked.
func RelativePathOption(path string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.relPath = path
	}
}
