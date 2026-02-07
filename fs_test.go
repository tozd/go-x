package x_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

// createTestFS creates a test filesystem with a standard structure.
func createTestFS() fstest.MapFS {
	return fstest.MapFS{
		"file1.txt": {Data: []byte("content1")},
		"file2.txt": {Data: []byte("content2")},
		"dir1/file3.txt": {
			Data: []byte("content3"),
		},
		"dir1/file4.txt": {
			Data: []byte("content4"),
		},
		"dir1/subdir/file5.txt": {
			Data: []byte("content5"),
		},
		"dir2/file6.txt": {
			Data: []byte("content6"),
		},
		"dir2/subdir/file7.txt": {
			Data: []byte("content7"),
		},
		"secret/password.txt": {
			Data: []byte("secret"),
		},
		"secret/nested/key.txt": {
			Data: []byte("key"),
		},
	}
}

func TestMakeFilteredFS(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()

	t.Run("no exclusions", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS)
		require.NoError(t, errE, "% -+#.1v", errE)

		// Should return the original FS when no exclusions.
		assert.Equal(t, testFS, filtered)
	})

	t.Run("single exclusion", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, "secret")
		require.NoError(t, errE, "% -+#.1v", errE)
		assert.NotNil(t, filtered)
	})

	t.Run("multiple exclusions", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, "secret", "dir1/subdir")
		require.NoError(t, errE, "% -+#.1v", errE)
		assert.NotNil(t, filtered)
	})

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, "../invalid")
		require.Error(t, errE)
		assert.Nil(t, filtered)
		assert.ErrorIs(t, errE, fs.ErrInvalid)
	})

	t.Run("dot path", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, ".")
		require.Error(t, errE)
		assert.Nil(t, filtered)
		assert.ErrorIs(t, errE, fs.ErrInvalid)
	})

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, "/absolute")
		require.Error(t, errE)
		assert.Nil(t, filtered)
		assert.ErrorIs(t, errE, fs.ErrInvalid)
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()

		filtered, errE := x.MakeFilteredFS(testFS, "")
		require.Error(t, errE)
		assert.Nil(t, filtered)
		assert.ErrorIs(t, errE, fs.ErrInvalid)
	})
}

func TestFilteredFS_Open(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret", "dir1/subdir")
	require.NoError(t, err)

	t.Run("open regular file", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("file1.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())
	})

	t.Run("open file in allowed directory", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("dir1/file3.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())
	})

	t.Run("open excluded file", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("secret/password.txt")
		require.Error(t, err)
		assert.Nil(t, file)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("open excluded directory", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("secret")
		require.Error(t, err)
		assert.Nil(t, file)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("open file in excluded subdirectory", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("dir1/subdir/file5.txt")
		require.Error(t, err)
		assert.Nil(t, file)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("open nonexistent file", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("nonexistent.txt")
		require.Error(t, err)
		assert.Nil(t, file)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestFilteredFS_ReadDir(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret", "dir1/subdir")
	require.NoError(t, err)

	t.Run("read root directory", func(t *testing.T) {
		t.Parallel()

		entries, err := fs.ReadDir(filtered, ".")
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		// Should not include "secret" directory.
		assert.Contains(t, names, "file1.txt")
		assert.Contains(t, names, "file2.txt")
		assert.Contains(t, names, "dir1")
		assert.Contains(t, names, "dir2")
		assert.NotContains(t, names, "secret")
	})

	t.Run("read directory with exclusion", func(t *testing.T) {
		t.Parallel()

		entries, err := fs.ReadDir(filtered, "dir1")
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		// Should not include "subdir" because it's excluded.
		assert.Contains(t, names, "file3.txt")
		assert.Contains(t, names, "file4.txt")
		assert.NotContains(t, names, "subdir")
	})

	t.Run("read directory without exclusions", func(t *testing.T) {
		t.Parallel()

		entries, err := fs.ReadDir(filtered, "dir2")
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		// Should include all entries.
		assert.Contains(t, names, "file6.txt")
		assert.Contains(t, names, "subdir")
	})

	t.Run("read excluded directory", func(t *testing.T) {
		t.Parallel()

		_, err := fs.ReadDir(filtered, "secret")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("read nonexistent directory", func(t *testing.T) {
		t.Parallel()

		_, err := fs.ReadDir(filtered, "nonexistent")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestFilteredFS_ReadFile(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret")
	require.NoError(t, err)

	t.Run("read regular file", func(t *testing.T) {
		t.Parallel()

		data, err := fs.ReadFile(filtered, "file1.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content1"), data)
	})

	t.Run("read file in subdirectory", func(t *testing.T) {
		t.Parallel()

		data, err := fs.ReadFile(filtered, "dir1/file3.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content3"), data)
	})

	t.Run("read excluded file", func(t *testing.T) {
		t.Parallel()

		data, err := fs.ReadFile(filtered, "secret/password.txt")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("read nonexistent file", func(t *testing.T) {
		t.Parallel()

		data, err := fs.ReadFile(filtered, "nonexistent.txt")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}

func TestFilteredFS_Glob(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret")
	require.NoError(t, err)

	t.Run("glob all txt files", func(t *testing.T) {
		t.Parallel()

		matches, err := fs.Glob(filtered, "*.txt")
		require.NoError(t, err)

		assert.Contains(t, matches, "file1.txt")
		assert.Contains(t, matches, "file2.txt")
	})

	t.Run("glob all files recursively", func(t *testing.T) {
		t.Parallel()

		matches, err := fs.Glob(filtered, "*/*.txt")
		require.NoError(t, err)

		assert.Contains(t, matches, "dir1/file3.txt")
		assert.Contains(t, matches, "dir1/file4.txt")
		assert.Contains(t, matches, "dir2/file6.txt")
		// Should not include excluded paths.
		assert.NotContains(t, matches, "secret/password.txt")
	})

	t.Run("glob nested files", func(t *testing.T) {
		t.Parallel()

		matches, err := fs.Glob(filtered, "*/*/*.txt")
		require.NoError(t, err)

		assert.Contains(t, matches, "dir1/subdir/file5.txt")
		assert.Contains(t, matches, "dir2/subdir/file7.txt")
		// Should not include excluded paths.
		assert.NotContains(t, matches, "secret/nested/key.txt")
	})

	t.Run("glob in specific directory", func(t *testing.T) {
		t.Parallel()

		matches, err := fs.Glob(filtered, "dir1/*.txt")
		require.NoError(t, err)

		assert.Contains(t, matches, "dir1/file3.txt")
		assert.Contains(t, matches, "dir1/file4.txt")
		assert.Len(t, matches, 2)
	})

	t.Run("glob with no matches", func(t *testing.T) {
		t.Parallel()

		matches, err := fs.Glob(filtered, "*.pdf")
		require.NoError(t, err)
		assert.Empty(t, matches)
	})
}

func TestFilteredFS_Sub(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "dir1/subdir", "secret")
	require.NoError(t, err)

	t.Run("sub with dot returns self", func(t *testing.T) {
		t.Parallel()

		sub, err := fs.Sub(filtered, ".")
		require.NoError(t, err)
		assert.Equal(t, filtered, sub)
	})

	t.Run("sub on regular directory", func(t *testing.T) {
		t.Parallel()

		sub, err := fs.Sub(filtered, "dir2")
		require.NoError(t, err)
		assert.NotNil(t, sub)

		// Should be able to read files in the subdirectory.
		data, err := fs.ReadFile(sub, "file6.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content6"), data)
	})

	t.Run("sub on directory with exclusions", func(t *testing.T) {
		t.Parallel()

		sub, err := fs.Sub(filtered, "dir1")
		require.NoError(t, err)
		assert.NotNil(t, sub)

		// Should be able to read allowed files.
		data, err := fs.ReadFile(sub, "file3.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content3"), data)

		// Should not be able to access excluded subdirectory.
		_, err = fs.ReadFile(sub, "subdir/file5.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("sub on excluded directory", func(t *testing.T) {
		t.Parallel()

		// Sub should not error, but operations later on should.
		// See: https://github.com/golang/go/issues/77447
		sub, err := fs.Sub(filtered, "secret")
		require.NoError(t, err)
		assert.NotNil(t, sub)

		entries, err := fs.ReadDir(sub, ".")
		require.Error(t, err)
		assert.Nil(t, entries)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("sub with invalid path", func(t *testing.T) {
		t.Parallel()

		sub, err := fs.Sub(filtered, "../invalid")
		require.Error(t, err)
		assert.Nil(t, sub)
		assert.ErrorIs(t, err, fs.ErrInvalid)
	})

	t.Run("sub on nonexistent directory", func(t *testing.T) {
		t.Parallel()

		// Sub should not error, but operations later on should.
		// See: https://github.com/golang/go/issues/77447
		sub, err := fs.Sub(filtered, "nonexistent")
		require.NoError(t, err)
		assert.NotNil(t, sub)

		entries, err := fs.ReadDir(sub, ".")
		require.Error(t, err)
		assert.Nil(t, entries)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("nested sub calls", func(t *testing.T) {
		t.Parallel()

		sub1, err := fs.Sub(filtered, "dir2")
		require.NoError(t, err)

		sub2, err := fs.Sub(sub1, "subdir")
		require.NoError(t, err)

		data, err := fs.ReadFile(sub2, "file7.txt")
		require.NoError(t, err)
		assert.Equal(t, []byte("content7"), data)
	})
}

func TestFilteredFS_PathPrefixMatching(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"test/file.txt": {
			Data: []byte("test"),
		},
		"test2/file.txt": {
			Data: []byte("test2"),
		},
		"testdir/file.txt": {
			Data: []byte("testdir"),
		},
	}

	filtered, err := x.MakeFilteredFS(testFS, "test")
	require.NoError(t, err)

	t.Run("exact match is excluded", func(t *testing.T) {
		t.Parallel()

		_, err := filtered.Open("test/file.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("prefix without slash not excluded", func(t *testing.T) {
		t.Parallel()

		// "test2" should not be excluded even though it starts with "test".
		file, err := filtered.Open("test2/file.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())

		// "testdir" should not be excluded.
		file, err = filtered.Open("testdir/file.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())
	})
}

func TestFilteredFS_TestFS(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret")
	require.NoError(t, err)

	// Use the standard library's fstest.TestFS to verify the FS implementation.
	require.NoError(t, fstest.TestFS(filtered,
		"file1.txt",
		"file2.txt",
		"dir1/file3.txt",
		"dir1/file4.txt",
		"dir1/subdir/file5.txt",
		"dir2/file6.txt",
		"dir2/subdir/file7.txt",
	))
}

func TestFilteredFS_MultipleExclusions(t *testing.T) {
	t.Parallel()

	testFS := createTestFS()
	filtered, err := x.MakeFilteredFS(testFS, "secret", "dir1/subdir", "dir2")
	require.NoError(t, err)

	t.Run("all exclusions are applied", func(t *testing.T) {
		t.Parallel()

		// "secret" should be excluded.
		_, err := filtered.Open("secret/password.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)

		// "dir1/subdir" should be excluded.
		_, err = filtered.Open("dir1/subdir/file5.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)

		// "dir2" should be excluded.
		_, err = filtered.Open("dir2/file6.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)

		// "dir1" files should be accessible.
		file, err := filtered.Open("dir1/file3.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())
	})

	t.Run("readdir reflects all exclusions", func(t *testing.T) {
		t.Parallel()

		entries, err := fs.ReadDir(filtered, ".")
		require.NoError(t, err)

		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}

		assert.Contains(t, names, "file1.txt")
		assert.Contains(t, names, "file2.txt")
		assert.Contains(t, names, "dir1")
		assert.NotContains(t, names, "dir2")
		assert.NotContains(t, names, "secret")
	})
}

func TestFilteredFS_NestedExclusions(t *testing.T) {
	t.Parallel()

	testFS := fstest.MapFS{
		"a/b/c/d/file.txt": {
			Data: []byte("nested"),
		},
		"a/b/file.txt": {
			Data: []byte("ab"),
		},
		"a/file.txt": {
			Data: []byte("a"),
		},
	}

	filtered, err := x.MakeFilteredFS(testFS, "a/b/c")
	require.NoError(t, err)

	t.Run("nested path is excluded", func(t *testing.T) {
		t.Parallel()

		_, err := filtered.Open("a/b/c/d/file.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("parent paths are accessible", func(t *testing.T) {
		t.Parallel()

		file, err := filtered.Open("a/file.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())

		file, err = filtered.Open("a/b/file.txt")
		require.NoError(t, err)
		assert.NotNil(t, file)
		require.NoError(t, file.Close())
	})
}
