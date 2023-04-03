package inotify

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

func WatchPath(path string) {
	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer w.Close()

	// Start listening for events.
	go dedupLoop(w)

	err = w.Add(path)
	if err != nil {
		panic(err)
	}

	printTime("ready; press ^C to exit")
}

func dedupLoop(w *fsnotify.Watcher) {
	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		printEvent = func(e fsnotify.Event) {
			printTime(e.String())

			// Don't need to remove the timer if you don't have a lot of files.
			mu.Lock()
			delete(timers, e.Name)
			mu.Unlock()
		}
	)

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			printTime("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// We just want to watch for file creation, so ignore everything
			// outside of Create and Write.
			if !e.Has(fsnotify.Create) && !e.Has(fsnotify.Write) {
				continue
			}

			// Get timer.
			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			// No timer yet, so create one.
			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() { printEvent(e) })
				t.Stop()

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			}

			// Reset the timer for this path, so it will start from 100ms again.
			t.Reset(waitFor)
		}
	}
}

func printTime(s string, args ...interface{}) {
	fmt.Printf(time.Now().Format("15:04:05.0000")+" "+s+"\n", args...)
}
