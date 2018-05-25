// Copyright © 2018 The Vault Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package vault

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"go/format"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// Version is the current vault version.
	Version            = "1.0.1"
	ttSharedTypesTempl = "sharedTypes"
	ttDebugFileTempl   = "debugFile"
	ttReleaseFileTempl = "releaseFile"
	ttFileHeaderTempl  = "fileHeaderTempl"
)

var ttRepo *template.Template

func execTempl(w io.Writer, name string, data interface{}) {
	if err := ttRepo.ExecuteTemplate(w, name, data); err != nil {
		log.Fatalf("failed to execute templ '%v': %v\n", name, err)
	}
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		log.Fatalln("failed to write: ", err)
	}
}

func init() {
	ttRepo = template.New("Repo").Funcs(template.FuncMap{
		"join": func(s ...string) string {
			return path.Clean(strings.Join(s, "/"))
		},
	})
	ttRepo = template.Must(ttRepo.New(ttDebugFileTempl).Parse(debugFileTemp))
	ttRepo = template.Must(ttRepo.New(ttSharedTypesTempl).Parse(sharedTypesTempl))
	ttRepo = template.Must(ttRepo.New(ttReleaseFileTempl).Parse(releaseFileTempl))
	ttRepo = template.Must(ttRepo.New(ttFileHeaderTempl).Parse(fileHeaderTempl))

}

type patterns []string

func (p patterns) matches(s string) bool {
	var ok bool
	var err error
	for _, pat := range p {
		ok, err = regexp.MatchString(pat, s)
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		if ok {
			return true
		}
	}

	return false
}

// Generator creates a vault with embedded files.
type Generator struct {
	config      GeneratorConfig
	sharedFile  string
	debugFile   string
	releaseFile string
}

// Run starts the vault generation, may panic if an error occurs.
func (g *Generator) Run() {
	log.Println("starting vault generation...")
	if err := os.MkdirAll(g.config.dest, 0755); err != nil {
		log.Fatalln("failed to create destination folder: ", err)
	}

	// Create shared and debug files
	g.createStaticFile(g.sharedFile,
		func(w io.Writer) { execTempl(w, ttFileHeaderTempl, g.config.pkgName) },
		func(w io.Writer) { execTempl(w, ttSharedTypesTempl, nil) })

	g.createStaticFile(g.debugFile,
		func(w io.Writer) { fprintf(w, "// +build debug\n\n") },
		func(w io.Writer) { execTempl(w, ttFileHeaderTempl, g.config.pkgName) },
		func(w io.Writer) {
			execTempl(w, ttDebugFileTempl, map[string]string{
				"Suffix": strings.Title(g.config.name),
				"Base":   getBasePath(g.config),
			})
		})

	g.createVault(walkSrcDirectory(g.config))
}

func (g *Generator) createVault(ch <-chan fileItem) {
	file, err := os.OpenFile(g.releaseFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	// Write build tags
	fprintf(file, "// +build !debug\n\n")
	// Execute header template
	execTempl(file, ttFileHeaderTempl, g.config.pkgName)
	// Write imports
	fprintf(file, releaseImportTempl)
	// Write binary data
	var files = processFiles(strings.Title(g.config.name), g.config.cmpLvl, file, ch)
	// Write release file template
	execTempl(file, ttReleaseFileTempl, map[string]interface{}{
		"Suffix": strings.Title(g.config.name),
		"Files":  files,
	})

	if err := file.Close(); err != nil {
		log.Fatalf("failed to close file: %v", err)
	}
}

func processFiles(assetName string, cmpLvl int, w io.Writer, ch <-chan fileItem) []fileModel {
	var files []fileModel
	var offset int64

	fprintf(w, "\nvar vaultAssetBin%v = \"", assetName)

	for f := range ch {
		log.Printf("processing file '%v'...\n", f.fullpath)
		of, err := os.Open(f.fullpath)
		if err != nil {
			log.Fatalf("failed to read file: %v\n", err)
		}

		// create a binary to string literal writer
		sw := &binToStrWriter{w: w}
		zw, err := zlib.NewWriterLevel(sw, cmpLvl)
		if err != nil {
			log.Fatalf("failed to create zlib writer: %v\n", err)
		}

		// read source file into byte slice
		b, err := ioutil.ReadAll(of)
		if err != nil {
			log.Fatalf("failed to read file '%v': %v", f.fullpath, err)
		}

		// write and close the zlib writer
		zw.Write(b)
		if err = zw.Close(); err != nil {
			log.Fatalf("failed to close zlib writer: %v\n", err)
		}

		files = append(files, fileModel{
			Name:    f.fi.Name(),
			Path:    getPath(f.path),
			Size:    f.fi.Size(),
			ModTime: f.fi.ModTime(),
			Offset:  offset,
			Length:  sw.length,
		})

		offset += sw.length
	}

	fprintf(w, "\"\n")
	return files
}

func getPath(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx <= 0 {
		return "/"
	}

	return path.Clean("/" + p[:idx])
}

func walkSrcDirectory(cfg GeneratorConfig) <-chan fileItem {
	ch := make(chan fileItem, 10)

	go func() {
		err := filepath.Walk(cfg.src, func(p string, fi os.FileInfo, err error) error {
			p = filepath.ToSlash(path.Clean(p))
			src := filepath.ToSlash(path.Clean(cfg.src))

			// Do not process the source directory
			if p == src {
				return nil
			}

			// Skip any directory if recursive is set to false (default)
			if fi.IsDir() {
				if !cfg.withSubdirs {
					log.Printf("skipping directory '%v'...\n", p)
					return filepath.SkipDir
				}
				return nil
			}

			vaultPath := strings.TrimPrefix(p, src)
			// If include is set, then only process matching files
			var skip bool
			if len(cfg.incl) > 0 {
				skip = !cfg.incl.matches(vaultPath) || cfg.excl.matches(vaultPath)
			} else {
				skip = cfg.excl.matches(vaultPath)
			}

			if skip {
				log.Printf("skipping file '%v'...\n", vaultPath)
				return nil
			}

			ch <- fileItem{path: vaultPath, fi: fi, fullpath: p}
			return nil
		})
		if err != nil {
			log.Fatalf("failed to walk source directory '%v': %v", cfg.src, err)
		}
		close(ch)
	}()

	return ch
}

func (g *Generator) createStaticFile(fi string, fns ...func(w io.Writer)) {
	log.Printf("creating file '%v'...", fi)

	var buf bytes.Buffer
	for i := range fns {
		fns[i](&buf)
	}

	ff, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("failed to format file: %v\n%s\n", err, buf.Bytes())
	}

	sf, err := os.OpenFile(fi, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("failed to create vault file for resource '%v': %v\n", g.config.name, err)
	}
	defer func() {
		if err := sf.Close(); err != nil {
			log.Fatalf("failed to close file: %v", err)
		}
	}()

	if _, err := sf.Write(ff); err != nil {
		log.Fatalf("failed to write to file: %v\n", err)
	}
}

// GeneratorConfig configures the vault generator.
type GeneratorConfig struct {
	src         string
	dest        string
	relPath     string
	name        string
	pkgName     string
	excl        patterns
	incl        patterns
	withSubdirs bool
	cmpLvl      int
}

// GeneratorOption configures the vault generator.
type GeneratorOption func(g *GeneratorConfig)

// NewGenerator creates a new generator instance with the given options.
func NewGenerator(src, dest string, options ...GeneratorOption) Generator {
	cfg := GeneratorConfig{src: src, dest: dest, cmpLvl: zlib.BestCompression}
	for i := range options {
		options[i](&cfg)
	}
	initGeneratorConfig(&cfg)
	g := Generator{config: cfg}

	g.sharedFile = cleanSlashedPath(dest, fmt.Sprintf("shared_%v_vault.go", cfg.name))
	g.debugFile = cleanSlashedPath(dest, fmt.Sprintf("debug_%v_vault.go", cfg.name))
	g.releaseFile = cleanSlashedPath(dest, fmt.Sprintf("release_%v_vault.go", cfg.name))
	return g
}

func cleanSlashedPath(s ...string) string {
	return filepath.ToSlash(path.Clean(
		strings.Join(s, "/")))
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

	// check if identifiers are valid
	if _, err := format.Source([]byte("package " + cfg.pkgName)); err != nil {
		log.Fatalf("'%v' is an invalid package name: try to set a valid package name manually", cfg.pkgName)
	}

	if _, err := format.Source([]byte("var " + cfg.name + " string")); err != nil {
		log.Fatalf("'%v' is an invalid resource name: try to set a valid resource name manually", cfg.name)
	}
}

func lastPath(p string) string {
	p = filepath.ToSlash(p)
	idx := strings.LastIndex(p, "/")
	if idx == -1 {
		return ""
	}
	return p[idx+1:]
}

// CompressOption if set to true all files will be compressed,
// otherwise no compression is used.
func CompressOption(compress bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		if compress {
			c.cmpLvl = zlib.BestCompression
		} else {
			c.cmpLvl = zlib.NoCompression
		}
	}
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
// Pattern matching follows the rules of regexp.Match (see https://golang.org/pkg/regexp/#Match).
func ExcludeFilesOption(name ...string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.excl = append(c.excl, name...)
	}
}

// IncludeFilesOption sets the files to include in the generation process.
// Only specified files and files not matching any exclusion pattern will be included in the generation process.
// Only relative paths will be checked, so pattern must not include the fullpath.
// Pattern matching follows the rules of regexp.Match (see https://golang.org/pkg/regexp/#Match).
func IncludeFilesOption(name ...string) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.incl = append(c.incl, name...)
	}
}

// WithSubdirsOption if set to true, the generator will walk down
// the folder tree with the source directory as root.
func WithSubdirsOption(withSubdirs bool) GeneratorOption {
	return func(c *GeneratorConfig) {
		c.withSubdirs = withSubdirs
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

func getBasePath(cfg GeneratorConfig) string {
	if cfg.relPath == "" {
		return path.Clean(filepath.ToSlash(cfg.src))
	}

	return path.Clean(filepath.ToSlash(cfg.relPath))
}

type fileModel struct {
	Name, Path           string
	Size, Offset, Length int64
	ModTime              time.Time
}

type fileItem struct {
	path, fullpath string
	fi             os.FileInfo
}

type binToStrWriter struct {
	w      io.Writer
	length int64
}

func (bw *binToStrWriter) Write(p []byte) (n int, err error) {
	var buf bytes.Buffer
	for _, b := range p {
		bw.length++

		switch {
		case b == '\n':
			fprintf(&buf, `\n`)
		case b == '\\':
			fprintf(&buf, `\\`)
		case b == '"':
			fprintf(&buf, `\"`)
		case b == '\t':
			fallthrough
		case (b >= 32 && b <= 126):
			fprintf(&buf, "%c", b)
		default:
			fprintf(&buf, "\\x%02x", b)
		}
	}

	if _, err := bw.w.Write(buf.Bytes()); err != nil {
		log.Fatalf("failed to write buffer to file: %v", err)
	}

	return buf.Len(), nil
}
