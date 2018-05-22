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
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	ttSharedTypesTempl = "sharedTypes"
	ttDebugFileTempl   = "debugFile"
	ttReleaseFileTempl = "releaseFile"
	ttFileHeaderTempl  = "fileHeaderTempl"
)

var ttRepo *template.Template

func init() {
	ttRepo = template.Must(template.New(ttDebugFileTempl).Parse(debugFileTemp))
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
			log.Println(err)
			continue
		}

		if ok {
			return true
		}
	}

	return false
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
	if err := os.MkdirAll(g.config.dest, 0755); err != nil {
		log.Fatalln("failed to create destination folder: ", err)
	}

	// Create shared and debug files
	g.createStaticFile(g.sharedFile,
		func(buf *bytes.Buffer) { ttRepo.ExecuteTemplate(buf, ttFileHeaderTempl, g.config.pkgName) },
		func(buf *bytes.Buffer) { ttRepo.ExecuteTemplate(buf, ttSharedTypesTempl, nil) })

	basePath := getBasePath(g.config)
	g.createStaticFile(g.debugFile,
		func(buf *bytes.Buffer) { fmt.Fprintf(buf, "// +build debug\n\n") },
		func(buf *bytes.Buffer) { ttRepo.ExecuteTemplate(buf, ttFileHeaderTempl, g.config.pkgName) },
		func(buf *bytes.Buffer) {
			ttRepo.ExecuteTemplate(buf, ttDebugFileTempl, map[string]string{
				"Suffix": strings.Title(g.config.name),
				"Base":   basePath,
			})
		})

	g.createVault(walkSrcDirectory(basePath, g.config))
}

func (g *Generator) createVault(ch <-chan fileItem) {
	file, err := os.OpenFile(g.releaseFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	w := &writer{f: file}

	// Write build tags
	fmt.Fprintf(w, "// +build !debug\n\n")
	// Execute header template
	ttRepo.ExecuteTemplate(w, ttFileHeaderTempl, g.config.pkgName)
	// Write binary data
	var files = processFiles(strings.Title(g.config.name), w, ch)
	// Write release file template
	ttRepo.ExecuteTemplate(w, ttReleaseFileTempl, map[string]interface{}{
		"Suffix": strings.Title(g.config.name),
		"Files":  files,
	})

	w.Close()
}

func processFiles(assetName string, w io.Writer, ch <-chan fileItem) []fileModel {
	var files []fileModel
	var offset int64

	fmt.Fprintf(w, "\nvar vaultAssetBin%v = \"", assetName)

	for f := range ch {
		log.Printf("processing file '%v'...\n", f.fullpath)
		of, err := os.Open(f.fullpath)
		if err != nil {
			log.Fatalf("failed to read file: %v\n", err)
		}

		// create a binary to string literal writer
		sw := &binToStrWriter{w: w}
		zw, err := zlib.NewWriterLevel(sw, zlib.BestCompression)
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
			Name:     f.fi.Name(),
			Path:     getPath(f.path),
			Size:     f.fi.Size(),
			ModTime:  f.fi.ModTime(),
			Offset:   offset,
			Length:   sw.length,
			fullpath: f.fullpath,
		})

		offset += sw.length
	}

	fmt.Fprintln(w, "\"")
	return files
}

func getPath(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx == -1 {
		return ""
	}

	return p[:idx]
}

func walkSrcDirectory(src string, cfg GeneratorConfig) <-chan fileItem {
	ch := make(chan fileItem, 10)

	go func() {
		err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
			// Do not process the source directory
			if path == src {
				return nil
			}

			// Skip any directory if recursive is set to false (default)
			if !cfg.recursiv && fi.IsDir() {
				log.Printf("skipping directory '%v'...\n", path)
				return filepath.SkipDir
			} else if fi.IsDir() {
				return nil
			}

			vaultPath := filepath.Clean("/" + filepath.ToSlash(strings.TrimLeft(path, src)))
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

			if !fi.IsDir() {
				ch <- fileItem{path: vaultPath, fi: fi, fullpath: path}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("failed to walk source directory '%v': %v", src, err)
		}
		close(ch)
	}()

	return ch
}

func (g *Generator) createStaticFile(fi string, fns ...func(b *bytes.Buffer)) {
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
	src      string
	dest     string
	relPath  string
	name     string
	pkgName  string
	excl     patterns
	incl     patterns
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
	g.releaseFile = cleanSlashedPath(dest, fmt.Sprintf("release_%v_vault.go", cfg.name))
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

func getBasePath(cfg GeneratorConfig) string {
	if cfg.relPath == "" {
		return filepath.Clean(filepath.ToSlash(cfg.src))
	}

	return filepath.Clean(filepath.ToSlash(cfg.relPath))
}

type fileModel struct {
	Name, Path           string
	Size, Offset, Length int64
	ModTime              time.Time
	fullpath             string
}

type fileItem struct {
	path, fullpath string
	fi             os.FileInfo
}

type writer struct {
	f *os.File
}

func (w *writer) Write(b []byte) (n int, err error) {
	n, err = w.f.Write(b)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
	return n, err
}

func (w *writer) Close() {
	if err := w.f.Close(); err != nil {
		log.Fatalf("failed to close file: %v", err)
	}
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
			_, err = fmt.Fprintf(&buf, `\n`)
		case b == '\\':
			_, err = fmt.Fprintf(&buf, `\\`)
		case b == '"':
			_, err = fmt.Fprintf(&buf, `\"`)
		case b == '\t':
			fallthrough
		case (b >= 32 && b <= 126):
			_, err = fmt.Fprintf(&buf, "%c", b)
		default:
			_, err = fmt.Fprintf(&buf, "\\x%02x", b)
		}

		if err != nil {
			log.Fatalf("failed to write to buffer: %v", err)
		}
	}

	if _, err := bw.w.Write(buf.Bytes()); err != nil {
		log.Fatalf("failed to write buffer to file: %v", err)
	}

	return buf.Len(), nil
}
