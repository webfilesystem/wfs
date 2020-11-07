package fs

import (
	"os"
	"os/user"
	"path/filepath"
	"time"

	"bazil.org/fuse"
	"golang.org/x/net/context"
)

var SourceType string

type Entity interface {
	IsDir() bool
	Path() string
	GetNode() *Node
}

type Node struct {
	ID         uint64
	FuseType   fuse.DirentType
	Name       string
	Size       uint64
	Mtime      time.Time
	Attrs      fuse.Attr
	Parent     *Dir
	Uri        string
	Loaded     bool
	SourceType string
	WFS        *WFS
}

func (n *Node) Attr(ctx context.Context, attr *fuse.Attr) error {
	if user, err := user.Current(); err == nil {
		attr.Uid = id(user.Uid)
		attr.Gid = id(user.Gid)
	}
	attr.Inode = n.ID
	if n.IsDir() {
		attr.Mode = os.ModeDir | 0555
		attr.Size = n.Size
		attr.Mtime = n.Mtime
	} else {
		attr.Mode = 0666
		attr.Size = uint64(n.Size)
		attr.Mtime = n.Mtime
	}
	return nil
}

func (n *Node) IsDir() bool {
	return n.FuseType == fuse.DT_Dir
}

func (n *Node) GetParent() *Dir {
	return n.Parent
}

func (n *Node) Path() string {
	if n.Parent != nil {
		return filepath.Join(n.Parent.Path(), n.Name)
	}
	return n.Name
}

func (n *Node) GetNode() *Node {
	return n
}

func (n *Node) String() string {
	return n.Path()
}
