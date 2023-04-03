package inotify

import (
	"fmt"
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
