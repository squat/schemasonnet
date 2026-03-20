package embedpkging

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"

	"github.com/gobuffalo/here"
	"github.com/markbates/pkger/pkging"
)

// Ensure File satisfies the pkging.File interface at compile time.
var _ pkging.File = &File{}

// File wraps an fs.File and satisfies the pkging.File interface.
// Because embed.FS files are not seekable, the full file contents are read
// into memory on open so that Seek works correctly.
type File struct {
	// reader holds the fully-buffered file contents and supports Seek.
	reader *bytes.Reader
	// underlying is the original fs.File, kept for Readdir and directory stat.
	underlying fs.File
	info       *pkging.FileInfo
	path       here.Path
	her        here.Info
	pkger      pkging.Pkger
}

// newFile constructs a File, eagerly buffering non-directory file contents.
func newFile(fs fs.File, info os.FileInfo, path here.Path, her here.Info, pkger *EmbedPkger) (*File, error) {
	ef := &File{
		underlying: fs,
		info:       pkging.NewFileInfo(info),
		path:       path,
		her:        her,
		pkger:      pkger,
	}
	if !info.IsDir() {
		data, err := io.ReadAll(fs)
		if err != nil {
			return nil, err
		}
		ef.reader = bytes.NewReader(data)
	}
	return ef, nil
}

// Close closes the underlying fs.File.
func (f *File) Close() error {
	return f.underlying.Close()
}

// Info returns the here.Info of the file's module.
func (f *File) Info() here.Info {
	return f.her
}

// Name returns the file name in here.Path format (e.g. "pkg:/path").
func (f *File) Name() string {
	return f.path.String()
}

// Open implements the http.FileSystem interface by delegating back to the
// owning Pkger. The path is resolved relative to this file's directory.
func (f *File) Open(name string) (http.File, error) {
	fp := path.Join(f.path.Name, name)
	hf, err := f.pkger.Open(fp)
	if err != nil {
		return nil, err
	}
	return hf, nil
}

// Path returns the here.Path of the file.
func (f *File) Path() here.Path {
	return f.path
}

// Read reads from the in-memory buffer. Returns os.ErrInvalid for directories.
func (f *File) Read(b []byte) (int, error) {
	if f.reader == nil {
		return 0, os.ErrInvalid
	}
	return f.reader.Read(b)
}

// Readdir reads directory entries. Returns os.ErrInvalid if the underlying FS
// is not a ReadDirFile.
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	rdf, ok := f.underlying.(fs.ReadDirFile)
	if !ok {
		return nil, os.ErrInvalid
	}
	entries, err := rdf.ReadDir(count)
	if err != nil && (count > 0 || !errors.Is(err, io.EOF)) {
		return nil, err
	}
	infos := make([]os.FileInfo, 0, len(entries))
	for _, e := range entries {
		fi, err := e.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, pkging.NewFileInfo(fi))
	}
	return infos, nil
}

// Seek sets the read offset on the in-memory buffer.
// Returns os.ErrInvalid for directories.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.reader == nil {
		return 0, os.ErrInvalid
	}
	return f.reader.Seek(offset, whence)
}

// Stat returns the FileInfo for the file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.info, nil
}

// Write is not supported by a read-only embed.FS implementation.
func (f *File) Write(b []byte) (int, error) {
	return 0, &os.PathError{Op: "write", Path: f.path.Name, Err: errors.ErrUnsupported}
}
