// Package embedpkging provides a read-only implementation of the pkging.Pkger
// interface backed by a Go standard library embed.FS.
package embedpkging

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/here"
	"github.com/markbates/pkger/pkging"
)

// Ensure EmbedPkger satisfies the pkging.Pkger interface at compile time.
var _ pkging.Pkger = &EmbedPkger{}

// EmbedPkger is a read-only pkging.EmbedPkger backed by a golang fs.FS.
// Write operations (Create, MkdirAll, Remove, RemoveAll) always return
// errors.ErrUnsupported because embed.FS is immutable and this
// package only aims to add support for embed.FS.
// The New constructor accepts a fs.FS so that testing is easier.
type EmbedPkger struct {
	fs   fs.FS
	info here.Info
}

// New creates a new Pkger from any fs.FS and a here.Info describing the
// module. Typically the caller provides an embed.FS together with a
// here.Info whose Module.Path is set to the Go module path declared in
// go.mod (e.g. "github.com/squat/schemasonnet").
func New(fsys fs.FS, info here.Info) *EmbedPkger {
	return &EmbedPkger{fs: fsys, info: info}
}

// Current returns the here.Info that was supplied to New.
func (e *EmbedPkger) Current() (here.Info, error) {
	return e.info, nil
}

// Info returns the here.Info for the given import path. Only the module's
// own import path is recognised; any other value returns an error.
func (e *EmbedPkger) Info(importPath string) (here.Info, error) {
	if importPath == "" || importPath == e.info.ImportPath || importPath == e.info.Module.Path {
		return e.info, nil
	}
	return here.Info{}, &os.PathError{Op: "info", Path: importPath, Err: os.ErrNotExist}
}

// Parse parses a string in here.Path format ("[pkg]:path") and returns the
// corresponding here.Path. When no package prefix is given the module's own
// import path is used.
func (e *EmbedPkger) Parse(s string) (here.Path, error) {
	return e.info.Parse(s)
}

// Open opens the named file for reading and returns a pkging.File.
// The name may be in here.Path format ("[pkg]:/path") or a plain path
// ("/path" or "path"). Leading slashes are stripped before the path is
// looked up inside the embed.FS (embed.FS paths are always relative and
// never start with "/").
func (e *EmbedPkger) Open(name string) (pkging.File, error) {
	path, err := e.Parse(name)
	if err != nil {
		return nil, err
	}
	fsPath := toFSPath(path.Name)
	fsFile, err := e.fs.Open(fsPath)
	if err != nil {
		return nil, err
	}
	info, err := fsFile.Stat()
	if err != nil {
		_ = fsFile.Close()
		return nil, err
	}
	f, err := newFile(fsFile, info, path, e.info, e)
	if err != nil {
		_ = fsFile.Close()
		return nil, err
	}
	return f, nil
}

// Stat returns a FileInfo describing the named file.
func (e *EmbedPkger) Stat(name string) (os.FileInfo, error) {
	pt, err := e.Parse(name)
	if err != nil {
		return nil, err
	}
	fsPath := toFSPath(pt.Name)
	fi, err := fs.Stat(e.fs, fsPath)
	if err != nil {
		return nil, err
	}
	return pkging.NewFileInfo(fi), nil
}

// Walk walks the file tree rooted at p, calling wf for each file or
// directory. Path strings passed to wf are in here.Path format.
func (e *EmbedPkger) Walk(root string, wf filepath.WalkFunc) error {
	pt, err := e.Parse(root)
	if err != nil {
		return err
	}
	fsRoot := toFSPath(pt.Name)

	return fs.WalkDir(e.fs, fsRoot, func(fsPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return wf(fsPath, nil, err)
		}
		fi, err := d.Info()
		if err != nil {
			return wf(fsPath, nil, err)
		}
		// Reconstruct the here.Path name: always starts with "/".
		// fs.WalkDir uses "." for the root, which maps back to "/".
		var pkgName string
		if fsPath == "." {
			pkgName = "/"
		} else {
			pkgName = "/" + strings.TrimPrefix(fsPath, "/")
		}
		pkgPath := here.Path{Pkg: pt.Pkg, Name: pkgName}
		return wf(pkgPath.String(), pkging.NewFileInfo(fi), nil)
	})
}

// Create is not supported by a read-only embed.FS implementation.
func (e *EmbedPkger) Create(name string) (pkging.File, error) {
	return nil, &os.PathError{Op: "create", Path: name, Err: errors.ErrUnsupported}
}

// MkdirAll is not supported by a read-only embed.FS implementation.
func (e *EmbedPkger) MkdirAll(dir string, perm os.FileMode) error {
	return &os.PathError{Op: "mkdir", Path: dir, Err: errors.ErrUnsupported}
}

// Remove is not supported by a read-only embed.FS implementation.
func (e *EmbedPkger) Remove(name string) error {
	return &os.PathError{Op: "remove", Path: name, Err: errors.ErrUnsupported}
}

// RemoveAll is not supported by a read-only embed.FS implementation.
func (e *EmbedPkger) RemoveAll(dir string) error {
	return &os.PathError{Op: "removeall", Path: dir, Err: errors.ErrUnsupported}
}

// toFSPath converts a here.Path Name (which may start with "/") to the
// relative path expected by embed.FS.
func toFSPath(name string) string {
	name = path.Clean(name)
	name = strings.TrimPrefix(name, "/")
	if name == "" || name == "." {
		return "."
	}
	return name
}
