package fs

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/fsnotify/fsnotify"
	"github.com/robertkrimen/otto"
	flag "github.com/spf13/pflag"
	"github.com/webfilesystem/wfs/config"
	script "github.com/webfilesystem/wfs/vm"
)

type WFS struct {
	MountPoint string
	RootDir    *Dir
	NodeID     uint64
	Configs    map[string][]byte
	VM         *otto.Otto
}

func (wfs WFS) Root() (fs.Node, error) {
	return wfs.RootDir, nil
}

var wfs *WFS

func (wfs WFS) Shutdown() error {
	if err := fuse.Unmount(wfs.MountPoint); err != nil {
		return err
	}
	return nil
}

func (wfs *WFS) NextID() uint64 {
	wfs.NodeID++
	return wfs.NodeID
}

func (wfs *WFS) Statfs(ctx context.Context, req *fuse.StatfsRequest, resp *fuse.StatfsResponse) error {
	resp.Blocks = 1000000000 // Total data blocks in file system.
	resp.Bfree = 1000000000  // Free blocks in file system.
	resp.Bavail = 1000000000 // Free blocks in file system if you're not root.
	resp.Files = 100000      // Total files in file system.
	resp.Ffree = 100000      // Free files in file system.
	resp.Bsize = 64 * 1024   // Block size
	resp.Namelen = 256       // Maximum file name length?
	resp.Frsize = 1          // Fragment size, smallest addressable data size in the file system.
	return nil
}

func NewFS(w io.Writer, mountpoint string) {
	fmt.Print("Loading the Webfilesystem...")
	fuse.Unmount(mountpoint)
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("Webfilesystem"),
		fuse.Subtype("wfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("wfs"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	wfs := WFS{MountPoint: mountpoint, NodeID: 0}
	rootdir := &Dir{Node: Node{ID: wfs.NextID(), Name: "rootnode", WFS: &wfs}, Entries: make(map[string]Entity)}
	rootdir.FuseType = fuse.DT_Dir
	wfs.RootDir = rootdir
	configs, err := config.LoadConfigs()
	if err != nil {
		log.Fatal("Can't load configs from HOME/.wfs: %s\n", err)
	}
	wfs.Configs = configs
	go watchConfig(&wfs)
	cfiles, err := config.GetConfigFiles()
	if err != nil {
		log.Fatal("Can't load configs from HOME/.wfs: %s\n", err)
	}
	for _, config := range cfiles {
		dir := NewDir(rootdir, strings.Replace(config.Name(), ".js", "", -1), "", &wfs, uint64(4096), time.Now())
		if dir == nil {
			log.Fatal(err)
		}
	}
	wfs.VM = script.NewVM()
	fmt.Print("\033[G\033[K")
	fmt.Println("Webfilesystem is ready...")
	fmt.Println("\r\r")
	err = fs.Serve(c, wfs)
	if err != nil {
		log.Fatal(err)
	}
	go signalHandler()
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
	if err := fuse.Unmount(flag.Arg(0)); err != nil {
		log.Fatal(err)
	}
}

func signalHandler() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	<-sigchan
	defer os.Exit(0)
	err := wfs.Shutdown()
	if err != nil {
		os.Exit(1)
	}
}

func watchConfig(wfs *WFS) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//fmt.Printf("event:", event)
				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
					//fmt.Printf(fmt.Sprintf("\nmodified file: %s\n", event.Name))

					configs, err := config.LoadConfigs()
					if err != nil {
						log.Fatal("Can't load configs from HOME/.wfs: %s\n", err)
					}
					wfs.Configs = configs
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("error:", err)
			}
		}
	}()

	err = watcher.Add(filepath.Join(os.Getenv("HOME"), ".wfs"))
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
