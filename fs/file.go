package fs

import (
	"fmt"
	"os/user"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

type File struct {
	Node
	Data []byte
}

func NewFile(parent *Dir, name string, wfs *WFS, size uint64, mtime time.Time) *File {
	file := File{Node: Node{ID: wfs.NextID(), Name: name, WFS: wfs, Size: size, Mtime: mtime, Parent: parent}}
	file.FuseType = fuse.DT_File
	parent.Entries[name] = &file
	var uid, gid uint32
	if user, err := user.Current(); err == nil {
		uid = id(user.Uid)
		gid = id(user.Gid)
	}
	file.Attrs = fuse.Attr{
		Valid:     1 * time.Minute,
		Mode:      0666,
		Atime:     time.Now(),
		Ctime:     time.Now(),
		Mtime:     time.Now(),
		Crtime:    time.Now(),
		Uid:       uid,
		Gid:       gid,
		Size:      size,
		Blocks:    (size + 511) / 512,
		BlockSize: 4096,
		Nlink:     1,
	}
	file.Attrs.Inode = uint64(file.ID)
	return &file
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if f.Data == nil {
		f.Data = f.Download(ctx)
	}
	fuseutil.HandleRead(req, resp, f.Data)
	return nil
}

func (f *File) Download(ctx context.Context) []byte {

	vm := f.WFS.VM

	script := f.WFS.Configs[f.Parent.GetSourceRoot()]

	if _, err := vm.Run(fmt.Sprintf("%s", script)); err != nil {
		panic(err)
	}

	r, err := vm.Call("download", nil, f.Uri)
	if err != nil {
		panic(err)
	}

	export, _ := r.ToString()
	if err != nil {
		panic(err)
	}

	f.Size = uint64(len(export))
	return []byte(export)
}

func (f *File) GetSourceRoot() string {
	switch {
	case f.Parent.GetDepth() == 2:
		return f.Parent.Name
	case f.Parent.GetDepth() > 2:
		return strings.Split(f.Parent.Path(), "/")[2]
	}
	return ""
}

func (f *File) GetNode() *Node {
	return &f.Node
}
