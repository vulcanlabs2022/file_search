package inotify

// import (
// 	"fmt"

// 	fsevents "github.com/tywkeene/go-fsevents"
// )

// var FsWatcher Watcher

// type Watcher struct {
// 	dir string
// }

// func Start(dir string) {
// 	FsWatcher = Watcher{dir: dir}
// 	FsWatcher.Start()
// }

// func handleEvents(w *fsevents.Watcher) error {
// 	// Watch for events
// 	go w.Watch()
// 	fmt.Println("Waiting for events...")

// 	for {
// 		select {
// 		case event := <-w.Events:
// 			// Contextual metadata is stored in the event object as well as a pointer to the WatchDescriptor that event belongs to
// 			fmt.Printf("Event Name: %s Event Path: %s Event Descriptor: %v", event.Name, event.Path, event.Descriptor)
// 			// A Watcher keeps a running atomic counter of all events it sees
// 			fmt.Println("Watcher Event Count:", w.GetEventCount())
// 			fmt.Println("Running descriptors:", w.GetRunningDescriptors())

// 			if event.IsDirCreated() == true {
// 				fmt.Println("Directory created:", event.Path)
// 				// A Watcher can be used dynamically in response to events to add/modify/delete WatchDescriptors
// 				d, err := w.AddDescriptor(event.Path, fsevents.DirCreatedEvent)
// 				if err != nil {
// 					fmt.Printf("Error adding descriptor for path %q: %s\n", event.Path, err)
// 					break
// 				}
// 				// WatchDescriptors can be started and stopped at any time and in response to events
// 				if err := d.Start(); err != nil {
// 					fmt.Printf("Error starting descriptor for path %q: %s\n", event.Path, err)
// 					break
// 				}
// 				fmt.Printf("Watch started for newly created directory %q\n", event.Path)
// 			}

// 			if event.IsDirRemoved() == true {
// 				fmt.Println("Directory removed:", event.Path)
// 				if err := w.RemoveDescriptor(event.Path); err != nil {
// 					fmt.Printf("Error removing descriptor for path %q: %s\n", event.Path, err)
// 					break
// 				}
// 			}

// 			if event.IsFileCreated() == true {
// 				fmt.Println("File created: ", event.Name)
// 			}
// 			if event.IsFileRemoved() == true {
// 				fmt.Println("File removed: ", event.Name)
// 			}
// 			break
// 		case err := <-w.Errors:
// 			fmt.Println(err)
// 			break
// 		}
// 	}
// }

// func (w *Watcher) Start() {
// 	var mask uint32 = fsevents.DirCreatedEvent | fsevents.DirRemovedEvent |
// 		fsevents.FileCreatedEvent | fsevents.FileRemovedEvent | fsevents.FileChangedEvent

// 	wa, err := fsevents.NewWatcher()
// 	if err != nil {
// 		panic(err)
// 	}

// 	d, err := wa.AddDescriptor(w.dir, mask)
// 	if err != nil {
// 		panic(err)
// 	}

// 	if err := d.Start(); err != nil {
// 		panic(err)
// 	}

// 	if err := handleEvents(wa); err != nil {
// 		fmt.Printf("Error handling events: %s", err.Error())
// 	}
// }
