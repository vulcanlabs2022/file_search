package inotify

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestWatch(t *testing.T) {
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Println("event:", event)
				if event.Has(fsnotify.Write) {
					fmt.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add("/tmp")
	if err != nil {
		panic(err)
	}

	// Block main goroutine forever.
	<-make(chan struct{})
}

func TestSplit(t *testing.T) {
	fmt.Println(len(strings.Split("/", "/")))
	for _, in := range strings.Split("/proc/sys/fs/inotify/max_user_watches", "/") {
		fmt.Printf("|%s|\n", in)
	}
}

func TestPathTrie(t *testing.T) {
	root := new(node)
	root.addFile("/data/aa/t1.txt")
	root.addFile("/data/aa/t2.txt")
	root.addFile("/data/aa/bb/t3.txt")
	root.addFile("/data/t4.txt")
	fmt.Println(root.getLeafs())
	fmt.Println(root.deletePath("/data/aa/bb/t3.txt"))
	fmt.Println(root.getLeafs())
}
