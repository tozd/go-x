package x

import (
	"io/fs"
	"path"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// emptyFS returns an error on all operations.
type emptyFS struct{}

var (
	_ fs.FS         = emptyFS{}
	_ fs.ReadDirFS  = emptyFS{}
	_ fs.ReadFileFS = emptyFS{}
	_ fs.ReadLinkFS = emptyFS{}
	_ fs.GlobFS     = emptyFS{}
)

// Open implements fs.FS.
func (f emptyFS) Open(name string) (fs.File, error) {
	return nil, errors.WithStack(&fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist})
}

// ReadDir implements fs.ReadDirFS.
func (f emptyFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, errors.WithStack(&fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist})
}

// ReadFile implements fs.ReadFileFS.
func (f emptyFS) ReadFile(name string) ([]byte, error) {
	return nil, errors.WithStack(&fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist})
}

// ReadLink implements fs.ReadLinkFS.
func (f emptyFS) ReadLink(name string) (string, error) {
	return "", errors.WithStack(&fs.PathError{Op: "readlink", Path: name, Err: fs.ErrNotExist})
}

// Lstat implements fs.ReadLinkFS.
func (f emptyFS) Lstat(name string) (fs.FileInfo, error) {
	return nil, errors.WithStack(&fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist})
}

// Glob implements fs.GlobFS.
func (f emptyFS) Glob(pattern string) ([]string, error) {
	_, err := path.Match(pattern, "")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return nil, nil
}

// Sub implements fs.SubFS.
func (f emptyFS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, errors.WithStack(&fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid})
	}

	return f, nil
}

// FilteredFS wraps an fs.FS and filters out specific paths.
type FilteredFS struct {
	fs      fs.FS
	exclude []string
}

var (
	_ fs.FS         = (*FilteredFS)(nil)
	_ fs.ReadDirFS  = (*FilteredFS)(nil)
	_ fs.ReadFileFS = (*FilteredFS)(nil)
	_ fs.ReadLinkFS = (*FilteredFS)(nil)
	_ fs.GlobFS     = (*FilteredFS)(nil)
)

func (f *FilteredFS) isExcluded(name string) bool {
	for _, excludePath := range f.exclude {
		if name == excludePath || strings.HasPrefix(name, excludePath+"/") {
			return true
		}
	}
	return false
}

// MakeFilteredFS creates a new FilteredFS.
func MakeFilteredFS(fsys fs.FS, exclude ...string) (fs.FS, errors.E) {
	if len(exclude) == 0 {
		return fsys, nil
	}

	for _, name := range exclude {
		if !fs.ValidPath(name) || name == "." {
			return nil, errors.WithStack(&fs.PathError{Op: "filter", Path: name, Err: fs.ErrInvalid})
		}
	}

	return &FilteredFS{
		fs:      fsys,
		exclude: exclude,
	}, nil
}

// Open implements fs.FS.
func (f *FilteredFS) Open(name string) (fs.File, error) {
	if f.isExcluded(name) {
		return nil, errors.WithStack(&fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist})
	}
	file, err := f.fs.Open(name)
	return file, errors.WithStack(err)
}

// ReadDir implements fs.ReadDirFS.
func (f *FilteredFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if f.isExcluded(name) {
		return nil, errors.WithStack(&fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist})
	}

	entries, err := fs.ReadDir(f.fs, name)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Filter out excluded paths.
	filtered := make([]fs.DirEntry, 0, len(entries))
	for _, entry := range entries {
		var entryPath string
		if name == "." {
			entryPath = entry.Name()
		} else {
			entryPath = path.Join(name, entry.Name())
		}
		if !f.isExcluded(entryPath) {
			filtered = append(filtered, entry)
		}
	}

	return filtered, nil
}

// ReadFile implements fs.ReadFileFS.
func (f *FilteredFS) ReadFile(name string) ([]byte, error) {
	if f.isExcluded(name) {
		return nil, errors.WithStack(&fs.PathError{Op: "readfile", Path: name, Err: fs.ErrNotExist})
	}

	data, err := fs.ReadFile(f.fs, name)
	return data, errors.WithStack(err)
}

// ReadLink implements fs.ReadLinkFS.
func (f *FilteredFS) ReadLink(name string) (string, error) {
	if f.isExcluded(name) {
		return "", errors.WithStack(&fs.PathError{Op: "readlink", Path: name, Err: fs.ErrNotExist})
	}

	dest, err := fs.ReadLink(f.fs, name)
	return dest, errors.WithStack(err)
}

// Lstat implements fs.ReadLinkFS.
func (f *FilteredFS) Lstat(name string) (fs.FileInfo, error) {
	if f.isExcluded(name) {
		return nil, errors.WithStack(&fs.PathError{Op: "lstat", Path: name, Err: fs.ErrNotExist})
	}

	info, err := fs.Lstat(f.fs, name)
	return info, errors.WithStack(err)
}

// Glob implements fs.GlobFS.
func (f *FilteredFS) Glob(pattern string) ([]string, error) {
	list, err := fs.Glob(f.fs, pattern)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Filter out excluded paths.
	filtered := make([]string, 0, len(list))
	for _, name := range list {
		if !f.isExcluded(name) {
			filtered = append(filtered, name)
		}
	}

	return filtered, nil
}

// Sub implements fs.SubFS.
func (f *FilteredFS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, errors.WithStack(&fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid})
	}

	if dir == "." {
		return f, nil
	}

	if f.isExcluded(dir) {
		// dir is excluded, so we return a FS which errors on all operations.
		return emptyFS{}, nil
	}

	sub, err := fs.Sub(f.fs, dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Adjust exclude paths to be relative to the subdirectory.
	newExclude := []string{}
	dirPrefix := dir + "/"

	for _, excludePath := range f.exclude {
		// Remove the directory prefix to make it relative.
		p := strings.TrimPrefix(excludePath, dirPrefix)
		if p != excludePath {
			// excludePath is under dirPrefix, so we add p.
			newExclude = append(newExclude, p)
		}
	}

	// If there are no exclude paths in the subdirectory, return unwrapped.
	if len(newExclude) == 0 {
		return sub, nil
	}

	return &FilteredFS{fs: sub, exclude: newExclude}, nil
}
