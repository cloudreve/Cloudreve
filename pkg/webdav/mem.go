package webdav

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// NewMemFS returns a new in-memory FileSystem implementation.
func NewMemFS() FileSystem {
	return &memFS{
		root: memFSNode{
			children: make(map[string]*memFSNode),
			mode:     0660 | os.ModeDir,
			modTime:  time.Now(),
		},
	}
}

// A memFS implements FileSystem, storing all metadata and actual file data
// in-memory. No limits on filesystem size are used, so it is not recommended
// this be used where the clients are untrusted.
//
// Concurrent access is permitted. The tree structure is protected by a mutex,
// and each node's contents and metadata are protected by a per-node mutex.
//
// TODO: Enforce file permissions.
type memFS struct {
	mu   sync.Mutex
	root memFSNode
}

// TODO: clean up and rationalize the walk/find code.

// walk walks the directory tree for the fullname, calling f at each step. If f
// returns an error, the walk will be aborted and return that same error.
//
// dir is the directory at that step, frag is the name fragment, and final is
// whether it is the final step. For example, walking "/foo/bar/x" will result
// in 3 calls to f:
//   - "/", "foo", false
//   - "/foo/", "bar", false
//   - "/foo/bar/", "x", true
// The frag argument will be empty only if dir is the root node and the walk
// ends at that root node.
func (fs *memFS) walk(op, fullname string, f func(dir *memFSNode, frag string, final bool) error) error {
	original := fullname
	fullname = slashClean(fullname)

	// Strip any leading "/"s to make fullname a relative path, as the walk
	// starts at fs.root.
	if fullname[0] == '/' {
		fullname = fullname[1:]
	}
	dir := &fs.root

	for {
		frag, remaining := fullname, ""
		i := strings.IndexRune(fullname, '/')
		final := i < 0
		if !final {
			frag, remaining = fullname[:i], fullname[i+1:]
		}
		if frag == "" && dir != &fs.root {
			panic("webdav: empty path fragment for a clean path")
		}
		if err := f(dir, frag, final); err != nil {
			return &os.PathError{
				Op:   op,
				Path: original,
				Err:  err,
			}
		}
		if final {
			break
		}
		child := dir.children[frag]
		if child == nil {
			return &os.PathError{
				Op:   op,
				Path: original,
				Err:  os.ErrNotExist,
			}
		}
		if !child.mode.IsDir() {
			return &os.PathError{
				Op:   op,
				Path: original,
				Err:  os.ErrInvalid,
			}
		}
		dir, fullname = child, remaining
	}
	return nil
}

// find returns the parent of the named node and the relative name fragment
// from the parent to the child. For example, if finding "/foo/bar/baz" then
// parent will be the node for "/foo/bar" and frag will be "baz".
//
// If the fullname names the root node, then parent, frag and err will be zero.
//
// find returns an error if the parent does not already exist or the parent
// isn't a directory, but it will not return an error per se if the child does
// not already exist. The error returned is either nil or an *os.PathError
// whose Op is op.
func (fs *memFS) find(op, fullname string) (parent *memFSNode, frag string, err error) {
	err = fs.walk(op, fullname, func(parent0 *memFSNode, frag0 string, final bool) error {
		if !final {
			return nil
		}
		if frag0 != "" {
			parent, frag = parent0, frag0
		}
		return nil
	})
	return parent, frag, err
}

func (fs *memFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir, frag, err := fs.find("mkdir", name)
	if err != nil {
		return err
	}
	if dir == nil {
		// We can't create the root.
		return os.ErrInvalid
	}
	if _, ok := dir.children[frag]; ok {
		return os.ErrExist
	}
	dir.children[frag] = &memFSNode{
		children: make(map[string]*memFSNode),
		mode:     perm.Perm() | os.ModeDir,
		modTime:  time.Now(),
	}
	return nil
}

func (fs *memFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (File, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir, frag, err := fs.find("open", name)
	if err != nil {
		return nil, err
	}
	var n *memFSNode
	if dir == nil {
		// We're opening the root.
		if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
			return nil, os.ErrPermission
		}
		n, frag = &fs.root, "/"

	} else {
		n = dir.children[frag]
		if flag&(os.O_SYNC|os.O_APPEND) != 0 {
			// memFile doesn't support these flags yet.
			return nil, os.ErrInvalid
		}
		if flag&os.O_CREATE != 0 {
			if flag&os.O_EXCL != 0 && n != nil {
				return nil, os.ErrExist
			}
			if n == nil {
				n = &memFSNode{
					mode: perm.Perm(),
				}
				dir.children[frag] = n
			}
		}
		if n == nil {
			return nil, os.ErrNotExist
		}
		if flag&(os.O_WRONLY|os.O_RDWR) != 0 && flag&os.O_TRUNC != 0 {
			n.mu.Lock()
			n.data = nil
			n.mu.Unlock()
		}
	}

	children := make([]os.FileInfo, 0, len(n.children))
	for cName, c := range n.children {
		children = append(children, c.stat(cName))
	}
	return &memFile{
		n:                n,
		nameSnapshot:     frag,
		childrenSnapshot: children,
	}, nil
}

func (fs *memFS) RemoveAll(ctx context.Context, name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir, frag, err := fs.find("remove", name)
	if err != nil {
		return err
	}
	if dir == nil {
		// We can't remove the root.
		return os.ErrInvalid
	}
	delete(dir.children, frag)
	return nil
}

func (fs *memFS) Rename(ctx context.Context, oldName, newName string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	oldName = slashClean(oldName)
	newName = slashClean(newName)
	if oldName == newName {
		return nil
	}
	if strings.HasPrefix(newName, oldName+"/") {
		// We can't rename oldName to be a sub-directory of itself.
		return os.ErrInvalid
	}

	oDir, oFrag, err := fs.find("rename", oldName)
	if err != nil {
		return err
	}
	if oDir == nil {
		// We can't rename from the root.
		return os.ErrInvalid
	}

	nDir, nFrag, err := fs.find("rename", newName)
	if err != nil {
		return err
	}
	if nDir == nil {
		// We can't rename to the root.
		return os.ErrInvalid
	}

	oNode, ok := oDir.children[oFrag]
	if !ok {
		return os.ErrNotExist
	}
	if oNode.children != nil {
		if nNode, ok := nDir.children[nFrag]; ok {
			if nNode.children == nil {
				return errNotADirectory
			}
			if len(nNode.children) != 0 {
				return errDirectoryNotEmpty
			}
		}
	}
	delete(oDir.children, oFrag)
	nDir.children[nFrag] = oNode
	return nil
}

func (fs *memFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dir, frag, err := fs.find("stat", name)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		// We're stat'ting the root.
		return fs.root.stat("/"), nil
	}
	if n, ok := dir.children[frag]; ok {
		return n.stat(path.Base(name)), nil
	}
	return nil, os.ErrNotExist
}

// A memFSNode represents a single entry in the in-memory filesystem and also
// implements os.FileInfo.
type memFSNode struct {
	// children is protected by memFS.mu.
	children map[string]*memFSNode

	mu        sync.Mutex
	data      []byte
	mode      os.FileMode
	modTime   time.Time
	deadProps map[xml.Name]Property
}

func (n *memFSNode) stat(name string) *memFileInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	return &memFileInfo{
		name:    name,
		size:    int64(len(n.data)),
		mode:    n.mode,
		modTime: n.modTime,
	}
}

func (n *memFSNode) DeadProps() (map[xml.Name]Property, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.deadProps) == 0 {
		return nil, nil
	}
	ret := make(map[xml.Name]Property, len(n.deadProps))
	for k, v := range n.deadProps {
		ret[k] = v
	}
	return ret, nil
}

func (n *memFSNode) Patch(patches []Proppatch) ([]Propstat, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	pstat := Propstat{Status: http.StatusOK}
	for _, patch := range patches {
		for _, p := range patch.Props {
			pstat.Props = append(pstat.Props, Property{XMLName: p.XMLName})
			if patch.Remove {
				delete(n.deadProps, p.XMLName)
				continue
			}
			if n.deadProps == nil {
				n.deadProps = map[xml.Name]Property{}
			}
			n.deadProps[p.XMLName] = p
		}
	}
	return []Propstat{pstat}, nil
}

type memFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (f *memFileInfo) Name() string       { return f.name }
func (f *memFileInfo) Size() int64        { return f.size }
func (f *memFileInfo) Mode() os.FileMode  { return f.mode }
func (f *memFileInfo) ModTime() time.Time { return f.modTime }
func (f *memFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f *memFileInfo) Sys() interface{}   { return nil }

// A memFile is a File implementation for a memFSNode. It is a per-file (not
// per-node) read/write position, and a snapshot of the memFS' tree structure
// (a node's name and children) for that node.
type memFile struct {
	n                *memFSNode
	nameSnapshot     string
	childrenSnapshot []os.FileInfo
	// pos is protected by n.mu.
	pos int
}

// A *memFile implements the optional DeadPropsHolder interface.
var _ DeadPropsHolder = (*memFile)(nil)

func (f *memFile) DeadProps() (map[xml.Name]Property, error)     { return f.n.DeadProps() }
func (f *memFile) Patch(patches []Proppatch) ([]Propstat, error) { return f.n.Patch(patches) }

func (f *memFile) Close() error {
	return nil
}

func (f *memFile) Read(p []byte) (int, error) {
	f.n.mu.Lock()
	defer f.n.mu.Unlock()
	if f.n.mode.IsDir() {
		return 0, os.ErrInvalid
	}
	if f.pos >= len(f.n.data) {
		return 0, io.EOF
	}
	n := copy(p, f.n.data[f.pos:])
	f.pos += n
	return n, nil
}

func (f *memFile) Readdir(count int) ([]os.FileInfo, error) {
	f.n.mu.Lock()
	defer f.n.mu.Unlock()
	if !f.n.mode.IsDir() {
		return nil, os.ErrInvalid
	}
	old := f.pos
	if old >= len(f.childrenSnapshot) {
		// The os.File Readdir docs say that at the end of a directory,
		// the error is io.EOF if count > 0 and nil if count <= 0.
		if count > 0 {
			return nil, io.EOF
		}
		return nil, nil
	}
	if count > 0 {
		f.pos += count
		if f.pos > len(f.childrenSnapshot) {
			f.pos = len(f.childrenSnapshot)
		}
	} else {
		f.pos = len(f.childrenSnapshot)
		old = 0
	}
	return f.childrenSnapshot[old:f.pos], nil
}

func (f *memFile) Seek(offset int64, whence int) (int64, error) {
	f.n.mu.Lock()
	defer f.n.mu.Unlock()
	npos := f.pos
	// TODO: How to handle offsets greater than the size of system int?
	switch whence {
	case os.SEEK_SET:
		npos = int(offset)
	case os.SEEK_CUR:
		npos += int(offset)
	case os.SEEK_END:
		npos = len(f.n.data) + int(offset)
	default:
		npos = -1
	}
	if npos < 0 {
		return 0, os.ErrInvalid
	}
	f.pos = npos
	return int64(f.pos), nil
}

func (f *memFile) Stat() (os.FileInfo, error) {
	return f.n.stat(f.nameSnapshot), nil
}

func (f *memFile) Write(p []byte) (int, error) {
	lenp := len(p)
	f.n.mu.Lock()
	defer f.n.mu.Unlock()

	if f.n.mode.IsDir() {
		return 0, os.ErrInvalid
	}
	if f.pos < len(f.n.data) {
		n := copy(f.n.data[f.pos:], p)
		f.pos += n
		p = p[n:]
	} else if f.pos > len(f.n.data) {
		// Write permits the creation of holes, if we've seek'ed past the
		// existing end of file.
		if f.pos <= cap(f.n.data) {
			oldLen := len(f.n.data)
			f.n.data = f.n.data[:f.pos]
			hole := f.n.data[oldLen:]
			for i := range hole {
				hole[i] = 0
			}
		} else {
			d := make([]byte, f.pos, f.pos+len(p))
			copy(d, f.n.data)
			f.n.data = d
		}
	}

	if len(p) > 0 {
		// We should only get here if f.pos == len(f.n.data).
		f.n.data = append(f.n.data, p...)
		f.pos = len(f.n.data)
	}
	f.n.modTime = time.Now()
	return lenp, nil
}
