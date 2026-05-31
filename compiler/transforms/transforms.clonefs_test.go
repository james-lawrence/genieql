package transforms_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/james-lawrence/genieql/compiler/transforms"
	"github.com/stretchr/testify/require"
)

// vanishingDirFS wraps a MapFS and makes one named directory return ENOENT on Open
// and ReadDir, simulating a directory that appears in a parent listing but is deleted
// before WalkDir can recurse into it (the WriteMapper mkcache temp-dir race).
type vanishingDirFS struct {
	fstest.MapFS
	vanishDir string
}

func (v vanishingDirFS) Open(name string) (fs.File, error) {
	if name == v.vanishDir {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return v.MapFS.Open(name)
}

func (v vanishingDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == v.vanishDir {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return v.MapFS.ReadDir(name)
}

// errDirFS wraps a MapFS and makes one named directory return a non-ENOENT error.
type errDirFS struct {
	fstest.MapFS
	errDir string
	err    error
}

func (e errDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == e.errDir {
		return nil, e.err
	}
	return e.MapFS.ReadDir(name)
}

func TestCloneFS(t *testing.T) {
	t.Run("flat files", func(t *testing.T) {
		src := fstest.MapFS{
			"a.txt": &fstest.MapFile{Data: []byte("aaa")},
			"b.txt": &fstest.MapFile{Data: []byte("bbb")},
		}

		dst := t.TempDir()
		require.NoError(t, transforms.CloneFS(dst, ".", src))

		for name, f := range src {
			content, err := os.ReadFile(filepath.Join(dst, name))
			require.NoError(t, err)
			require.Equal(t, string(f.Data), string(content))
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		src := fstest.MapFS{
			"sub/deep/file.txt": &fstest.MapFile{Data: []byte("nested")},
		}

		dst := t.TempDir()
		require.NoError(t, transforms.CloneFS(dst, ".", src))

		content, err := os.ReadFile(filepath.Join(dst, "sub", "deep", "file.txt"))
		require.NoError(t, err)
		require.Equal(t, "nested", string(content))
	})

	t.Run(".cache directory is skipped", func(t *testing.T) {
		src := fstest.MapFS{
			"real.txt":        &fstest.MapFile{Data: []byte("yes")},
			".cache/skip.txt": &fstest.MapFile{Data: []byte("no")},
		}

		dst := t.TempDir()
		require.NoError(t, transforms.CloneFS(dst, ".", src))

		_, err := os.ReadFile(filepath.Join(dst, "real.txt"))
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dst, ".cache"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("non-ENOENT directory error is propagated", func(t *testing.T) {
		sentinel := errors.New("permission denied")
		src := errDirFS{
			MapFS:  fstest.MapFS{"broken/file.txt": &fstest.MapFile{Data: []byte("x")}},
			errDir: "broken",
			err:    sentinel,
		}

		dst := t.TempDir()
		require.ErrorIs(t, transforms.CloneFS(dst, ".", src), sentinel)
	})

	t.Run("mkcache directory is skipped", func(t *testing.T) {
		src := fstest.MapFS{
			"real.txt":               &fstest.MapFile{Data: []byte("yes")},
			"mkcache.1234/file.yaml": &fstest.MapFile{Data: []byte("no")},
		}

		dst := t.TempDir()
		require.NoError(t, transforms.CloneFS(dst, ".", src))

		_, err := os.ReadFile(filepath.Join(dst, "real.txt"))
		require.NoError(t, err)

		_, err = os.Stat(filepath.Join(dst, "mkcache.1234"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("vanishing directory", func(t *testing.T) {
		src := vanishingDirFS{
			MapFS: fstest.MapFS{
				"real.txt":              &fstest.MapFile{Data: []byte("hello")},
				"vanishing/ignored.txt": &fstest.MapFile{Data: []byte("world")},
			},
			vanishDir: "vanishing",
		}

		dst := t.TempDir()
		require.NoError(t, transforms.CloneFS(dst, ".", src))

		content, err := os.ReadFile(filepath.Join(dst, "real.txt"))
		require.NoError(t, err)
		require.Equal(t, "hello", string(content))

		// the vanishing directory may be created before it disappears, but its contents
		// are never read so nothing inside it should be cloned.
		_, err = os.Stat(filepath.Join(dst, "vanishing", "ignored.txt"))
		require.ErrorIs(t, err, os.ErrNotExist)
	})
}
