package fs

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/robertkrimen/otto"
)

type Dir struct {
	Node
	Entries map[string]Entity
}

func NewDir(parent *Dir, name string, uri string, wfs *WFS, size uint64, mtime time.Time) *Dir {
	dir := Dir{Node: Node{ID: wfs.NextID(), Name: name, WFS: wfs, Parent: parent, Uri: uri, Size: size, Mtime: mtime}, Entries: make(map[string]Entity)}
	dir.FuseType = fuse.DT_Dir
	dir.Loaded = false
	parent.Entries[name] = &dir
	var uid, gid uint32
	if user, err := user.Current(); err == nil {
		uid = id(user.Uid)
		gid = id(user.Gid)
	}
	dir.Attrs = fuse.Attr{
		Valid:     1 * time.Minute,
		Mode:      os.ModeDir | 0555,
		Atime:     time.Now(),
		Ctime:     time.Now(),
		Mtime:     mtime,
		Crtime:    time.Now(),
		Uid:       uid,
		Gid:       gid,
		Size:      size,
		Blocks:    (size + 511) / 512,
		BlockSize: 4096,
		Nlink:     1,
	}
	dir.Attrs.Inode = uint64(dir.ID)
	return &dir
}

func (d *Dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	script := d.Config()
	if !strings.HasPrefix(req.Name, ".") && d.GetDepth() == 1 && d.Parent != nil && strings.Contains(fmt.Sprintf("%s", script), "search(query)") {
		d.Search(req.Name)
	}
	if ent, ok := d.Entries[req.Name]; ok {
		if ent.IsDir() {
			return ent.(*Dir), nil
		}
		return ent.(*File), nil
	}
	return nil, fuse.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.List()
	contents := make([]fuse.Dirent, 0, len(d.Entries))
	for _, entity := range d.Entries {
		node := entity.GetNode()
		dirent := fuse.Dirent{
			Inode: node.ID,
			Type:  node.FuseType,
			Name:  node.Name,
		}
		contents = append(contents, dirent)
	}
	return contents, nil
}

func (d *Dir) GetNode() *Node {
	return &d.Node
}

func (d *Dir) GetDepth() uint32 {
	depth := 0
	current := d
	for current != d.WFS.RootDir {
		depth++
		current = current.Parent
	}
	return uint32(depth)
}

func id(s string) uint32 {
	x, _ := strconv.Atoi(s)
	return uint32(x)
}

func (d *Dir) GetSourceRoot() string {
	switch {
	case d.GetDepth() == 1:
		return d.Name
	case d.GetDepth() > 1:
		return strings.Split(d.Path(), "/")[1]
	}
	return ""
}

func (d *Dir) GetPrefix(depth ...int) string {
	splitted := strings.Split(d.Path(), "/")
	if len(depth) == 0 {
		return strings.Join(splitted[3:], "/")
	}
	return strings.Join(splitted[depth[0]:], "/")
}

func (d *Dir) List() {
	vm := d.WFS.VM

	script := d.Config()

	if _, err := vm.Run(fmt.Sprintf("%s", script)); err != nil {
		panic(err)
	}

	var r otto.Value
	var err error
	var export interface{}

	if d.Uri == "" && strings.Contains(fmt.Sprintf("%s", script), "root()") {
		r, err = vm.Call("root", nil, nil)
		if err != nil {
			panic(fmt.Sprintf("%s, %s", r, err))
		}
		export, err = r.Export()
		if err != nil {
			panic(err)
		}

	} else if d.Uri != "" && !d.Loaded && strings.Contains(fmt.Sprintf("%s", script), "list(url)") {
		r, err = vm.Call("list", nil, d.Uri)
		if err != nil {
			panic(fmt.Sprintf("%s, %s", r, err))
		}
		export, err = r.Export()
		if err != nil {
			panic(err)
		}
		d.Loaded = true
	}

	if export != nil {
		found := export.([]map[string]interface{})

		if found != nil {
			for _, entity := range export.([]map[string]interface{}) {
				if entity["type"].(string) == "file" {
					size, _ := strconv.ParseUint(entity["size"].(string), 10, 64)
					file := NewFile(d, entity["name"].(string), d.WFS, size, time.Now())
					if entity["url"] != nil {
						file.Uri = fmt.Sprintf("%s", entity["url"])
					}
					if entity["data"] != nil {
						file.Data = []byte(fmt.Sprintf("%s", entity["data"]))
					}
					if file == nil {
						log.Printf("Can't create file.")
					}
				}
				if entity["type"].(string) == "dir" {
					dir := NewDir(d, entity["name"].(string), entity["url"].(string), d.WFS, 4096, time.Now())
					if dir == nil {
						log.Printf("Can't create dir.")
					}
				}
			}
		}
	}
}

func (d *Dir) Search(query string) {

	vm := d.WFS.VM

	script := d.Config()

	if _, err := vm.Run(fmt.Sprintf("%s", script)); err != nil {
		panic(err)
	}

	r, err := vm.Call("search", nil, query)
	if err != nil {
		panic(err)
	}

	export, _ := r.Export()
	if err != nil {
		panic(err)
	}

	if fmt.Sprintf("%s", export) != "[]" {
		found := export.([]map[string]interface{})

		if found != nil {
			result := NewDir(d, query, "", d.WFS, 4096, time.Now())
			if result == nil {
				panic(err)
			}

			for _, entity := range export.([]map[string]interface{}) {
				if entity["type"].(string) == "file" {
					size, _ := strconv.ParseUint(entity["size"].(string), 10, 64)
					file := NewFile(result, entity["name"].(string), d.WFS, size, time.Now())
					file.Uri = fmt.Sprintf("%s", entity["url"])
					if file == nil {
						panic(err)
					}
				}
				if entity["type"].(string) == "dir" {
					dir := NewDir(result, entity["name"].(string), entity["url"].(string), d.WFS, 4096, time.Now())
					if dir == nil {
						log.Printf("Can't create dir.")
					}
				}
			}
		}
	}
}

func (d *Dir) Config() []byte {
	c := d.WFS.Configs[d.GetSourceRoot()]
	if c != nil {
		return c
	}
	return nil
}
