package embedpkging

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gobuffalo/here"
	"github.com/markbates/pkger/pkging"
)

// testInfo returns a here.Info for the fake module used in tests.
func testInfo() here.Info {
	return here.Info{
		Dir:        "/fake/root",
		ImportPath: "example.com/fake",
		Name:       "fake",
		Module: here.Module{
			Path: "example.com/fake",
		},
	}
}

// testFS builds an in-memory fs.FS with a handful of test files.
func testFS() fs.FS {
	m := fstest.MapFS{
		"hello.txt": &fstest.MapFile{Data: []byte("hello, world")},
		"sub/a.txt": &fstest.MapFile{Data: []byte("sub-a")},
		"sub/b.txt": &fstest.MapFile{Data: []byte("sub-b")},
	}
	return m
}

func newPkger(t *testing.T) pkging.Pkger {
	t.Helper()
	p := New(testFS(), testInfo())
	return p
}

// ---------------------------------------------------------------------------
// Current / Info / Parse
// ---------------------------------------------------------------------------

func TestCurrent(t *testing.T) {
	p := newPkger(t)
	info, err := p.Current()
	if err != nil {
		t.Fatalf("Current: unexpected error: %v", err)
	}
	if info.Module.Path != "example.com/fake" {
		t.Errorf("Current: got module path %q, want %q", info.Module.Path, "example.com/fake")
	}
}

func TestInfo(t *testing.T) {
	p := newPkger(t)

	for _, name := range []string{"example.com/fake", "example.com/fake", ""} {
		info, err := p.Info(name)
		if err != nil {
			t.Errorf("Info(%q): unexpected error: %v", name, err)
			continue
		}
		if info.Module.Path != "example.com/fake" {
			t.Errorf("Info(%q): got module path %q, want %q", name, info.Module.Path, "example.com/fake")
		}
	}

	_, err := p.Info("some.other/pkg")
	if err == nil {
		t.Error("Info(unknown pkg): expected error, got nil")
	}
}

func TestParse(t *testing.T) {
	p := newPkger(t)

	for _, tc := range []struct {
		input    string
		wantPkg  string
		wantName string
	}{
		{"/hello.txt", "example.com/fake", "/hello.txt"},
		{"example.com/fake:/hello.txt", "example.com/fake", "/hello.txt"},
		{"/sub/a.txt", "example.com/fake", "/sub/a.txt"},
	} {
		pt, err := p.Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if pt.Pkg != tc.wantPkg {
			t.Errorf("Parse(%q): pkg = %q, want %q", tc.input, pt.Pkg, tc.wantPkg)
		}
		if pt.Name != tc.wantName {
			t.Errorf("Parse(%q): name = %q, want %q", tc.input, pt.Name, tc.wantName)
		}
	}
}

// ---------------------------------------------------------------------------
// Open / Read / Seek
// ---------------------------------------------------------------------------

func TestOpen(t *testing.T) {
	p := newPkger(t)

	f, err := p.Open("/hello.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "hello, world" {
		t.Errorf("content = %q, want %q", data, "hello, world")
	}
}

func TestOpenWithPackagePrefix(t *testing.T) {
	p := newPkger(t)

	f, err := p.Open("example.com/fake:/hello.txt")
	if err != nil {
		t.Fatalf("Open with pkg prefix: %v", err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "hello, world" {
		t.Errorf("content = %q, want %q", data, "hello, world")
	}
}

func TestOpenNotFound(t *testing.T) {
	p := newPkger(t)
	_, err := p.Open("/does-not-exist.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSeek(t *testing.T) {
	p := newPkger(t)

	f, err := p.Open("/hello.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	// Read first 5 bytes.
	buf := make([]byte, 5)
	if _, err := io.ReadFull(f, buf); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}
	if string(buf) != "hello" {
		t.Errorf("first read = %q, want %q", buf, "hello")
	}

	// Seek back to start.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("Seek: %v", err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll after seek: %v", err)
	}
	if string(data) != "hello, world" {
		t.Errorf("after seek = %q, want %q", data, "hello, world")
	}
}

// ---------------------------------------------------------------------------
// Stat
// ---------------------------------------------------------------------------

func TestStat(t *testing.T) {
	p := newPkger(t)

	fi, err := p.Stat("/hello.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if fi.Name() != "hello.txt" {
		t.Errorf("Name = %q, want %q", fi.Name(), "hello.txt")
	}
	if fi.IsDir() {
		t.Error("IsDir = true, want false")
	}
	if fi.Size() != int64(len("hello, world")) {
		t.Errorf("Size = %d, want %d", fi.Size(), len("hello, world"))
	}
}

func TestStatNotFound(t *testing.T) {
	p := newPkger(t)
	_, err := p.Stat("/nope.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// Walk
// ---------------------------------------------------------------------------

func TestWalk(t *testing.T) {
	p := newPkger(t)

	var visited []string
	err := p.Walk("/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	// All three files plus the root "." and "sub" directory should appear.
	want := map[string]struct{}{
		"example.com/fake:/":          struct{}{},
		"example.com/fake:/hello.txt": struct{}{},
		"example.com/fake:/sub":       struct{}{},
		"example.com/fake:/sub/a.txt": struct{}{},
		"example.com/fake:/sub/b.txt": struct{}{},
	}
	for _, v := range visited {
		delete(want, v)
	}
	if len(want) > 0 {
		missing := make([]string, 0, len(want))
		for k := range want {
			missing = append(missing, k)
		}
		t.Errorf("Walk did not visit: %v", missing)
	}
}

func TestWalkSubdir(t *testing.T) {
	p := newPkger(t)

	var visited []string
	err := p.Walk("/sub", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk /sub: %v", err)
	}

	for _, v := range visited {
		if !strings.HasPrefix(v, "example.com/fake:/sub") {
			t.Errorf("Walk /sub: unexpected path outside sub: %q", v)
		}
	}
	if len(visited) != 3 { // sub itself + a.txt + b.txt
		t.Errorf("Walk /sub: visited %d paths, want 3: %v", len(visited), visited)
	}
}

// ---------------------------------------------------------------------------
// Name / Path / Info on File
// ---------------------------------------------------------------------------

func TestFileName(t *testing.T) {
	p := newPkger(t)

	f, err := p.Open("/hello.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()

	if f.Name() != "example.com/fake:/hello.txt" {
		t.Errorf("Name = %q, want %q", f.Name(), "example.com/fake:/hello.txt")
	}
	if f.Path().Name != "/hello.txt" {
		t.Errorf("Path().Name = %q, want %q", f.Path().Name, "/hello.txt")
	}
	if f.Info().Module.Path != "example.com/fake" {
		t.Errorf("Info().Module.Path = %q, want %q", f.Info().Module.Path, "example.com/fake")
	}
}

// ---------------------------------------------------------------------------
// Readdir
// ---------------------------------------------------------------------------

func TestReaddir(t *testing.T) {
	p := newPkger(t)

	dir, err := p.Open("/sub")
	if err != nil {
		t.Fatalf("Open /sub: %v", err)
	}
	defer dir.Close()

	infos, err := dir.Readdir(-1)
	if err != nil {
		t.Fatalf("Readdir: %v", err)
	}
	if len(infos) != 2 {
		t.Errorf("Readdir: got %d entries, want 2", len(infos))
	}
	names := map[string]bool{}
	for _, fi := range infos {
		names[fi.Name()] = true
	}
	for _, want := range []string{"a.txt", "b.txt"} {
		if !names[want] {
			t.Errorf("Readdir: missing entry %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// Unsupported write operations
// ---------------------------------------------------------------------------

func TestCreateUnsupported(t *testing.T) {
	p := newPkger(t)
	_, err := p.Create("/new.txt")
	if err == nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("Create: expected ErrUnsupported, got %v", err)
	}
}

func TestMkdirAllUnsupported(t *testing.T) {
	p := newPkger(t)
	err := p.MkdirAll("/newdir", 0755)
	if err == nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("MkdirAll: expected ErrUnsupported, got %v", err)
	}
}

func TestRemoveUnsupported(t *testing.T) {
	p := newPkger(t)
	err := p.Remove("/hello.txt")
	if err == nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("Remove: expected ErrUnsupported, got %v", err)
	}
}

func TestRemoveAllUnsupported(t *testing.T) {
	p := newPkger(t)
	err := p.RemoveAll("/sub")
	if err == nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("RemoveAll: expected ErrUnsupported, got %v", err)
	}
}

func TestWriteUnsupported(t *testing.T) {
	p := newPkger(t)
	f, err := p.Open("/hello.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	_, err = f.Write([]byte("oops"))
	if err == nil || !errors.Is(err, errors.ErrUnsupported) {
		t.Errorf("Write: expected ErrUnsupported, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// http.File via Open(name) on a File
// ---------------------------------------------------------------------------

func TestFileHTTPOpen(t *testing.T) {
	p := newPkger(t)

	dir, err := p.Open("/sub")
	if err != nil {
		t.Fatalf("Open /sub: %v", err)
	}
	defer dir.Close()

	httpFile, err := dir.Open("a.txt")
	if err != nil {
		t.Fatalf("dir.Open(a.txt): %v", err)
	}
	defer httpFile.Close()

	data, err := io.ReadAll(httpFile)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "sub-a" {
		t.Errorf("content = %q, want %q", data, "sub-a")
	}
}
