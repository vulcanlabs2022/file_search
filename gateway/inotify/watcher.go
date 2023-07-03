package inotify

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
	"wzinc/parser"
	"wzinc/rpc"

	"bytetrade.io/web3os/fs-lib/jfsnotify"

	// "github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

var watcher *jfsnotify.Watcher

func WatchPath(path string) {
	// Create a new watcher.
	var err error
	watcher, err = jfsnotify.NewWatcher("myWatcher")
	if err != nil {
		panic(err)
	}

	// Start listening for events.
	go dedupLoop(watcher)
	log.Info().Msgf("watching path %s", path)

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				fmt.Println("watcher add error:", err)
				return err
			}
		} else {
			err = updateOrInputDoc(path)
			if err != nil {
				log.Error().Msgf("udpate or input doc err %v", err)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	printTime("ready; press ^C to exit")
}

func dedupLoop(w *jfsnotify.Watcher) {
	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 1000 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		printEvent = func(e jfsnotify.Event) {
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
				log.Warn().Msg("watcher event channel closed")
				return
			}
			log.Debug().Msgf("pending event %v", e)
			// Get timer.
			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			// No timer yet, so create one.
			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() {
					printEvent(e)
					err := handleEvent(e)
					if err != nil {
						log.Error().Msgf("handle watch file event error %s", err.Error())
					}
				})
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

func handleEvent(e jfsnotify.Event) error {
	if e.Has(jfsnotify.Remove) || e.Has(jfsnotify.Rename) {
		res, err := rpc.RpcServer.ZincQueryByPath(rpc.FileIndex, e.Name)
		if err != nil {
			return err
		}
		docs, err := rpc.GetFileQueryResult(res)
		if err != nil {
			return err
		}
		for _, doc := range docs {
			_, err = rpc.RpcServer.ZincDelete(doc.DocId, rpc.FileIndex)
			if err != nil {
				log.Error().Msgf("zinc delete error %s", err.Error())
			}
			log.Debug().Msgf("delete doc id %s path %s", doc.DocId, e.Name)
		}
		return nil
	}

	if e.Has(jfsnotify.Create) {
		err := filepath.Walk(e.Name, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				//add dir to watch list
				err = watcher.Add(path)
				if err != nil {
					log.Error().Msgf("watcher add error:%v", err)
				}
			} else {
				//input zinc file
				err = updateOrInputDoc(path)
				if err != nil {
					log.Error().Msgf("update or input doc error %v", err)
				}
			}
			return nil
		})
		if err != nil {
			log.Error().Msgf("handle create file error %v", err)
		}
		return nil
	}

	if e.Has(jfsnotify.Write) || e.Has(jfsnotify.Chmod) {
		return updateOrInputDoc(e.Name)
	}
	return nil
}

func updateOrInputDoc(filepath string) error {
	log.Debug().Msg("try update or input" + filepath)
	res, err := rpc.RpcServer.ZincQueryByPath(rpc.FileIndex, filepath)
	if err != nil {
		return err
	}
	docs, err := rpc.GetFileQueryResult(res)
	if err != nil {
		return err
	}
	// update doc if path exist
	// if len(docs) > 0 {
	// 	fileType := parser.GetTypeFromName(filepath)
	// 	if _, ok := parser.ParseAble[fileType]; ok {
	// 		f, err := os.Open(filepath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defer f.Close()
	// 		content, err := parser.ParseDoc(f, filepath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		log.Debug().Msgf("try update content from old doc id %s path %s", docs[0].DocId, filepath)
	// 		_, err = rpc.RpcServer.UpdateFileContentFromOldDoc(rpc.FileIndex, content, docs[0])
	// 		return err
	// 	}
	// 	return nil
	// }

	//delete all
	for _, doc := range docs {
		log.Debug().Msgf("try delete docid %s path %s", doc.DocId, doc.Where)
		_, err := rpc.RpcServer.ZincDelete(doc.DocId, rpc.FileIndex)
		if err != nil {
			log.Error().Msgf("zinc delete error %v", err)
		}
	}

	//input new doc if path not exist
	fileType := parser.GetTypeFromName(filepath)
	content := ""
	if _, ok := parser.ParseAble[fileType]; ok {
		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		data, _ := ioutil.ReadAll(f)
		f.Close()
		r := bytes.NewReader(data)
		content, err = parser.ParseDoc(r, filepath)
		if err != nil {
			return err
		}
	}

	filename := path.Base(filepath)
	size := 0
	fileInfo, err := os.Stat(filepath)
	if err == nil {
		size = int(fileInfo.Size())
	}
	doc := map[string]interface{}{
		"name":        filename,
		"where":       filepath,
		"content":     content,
		"size":        size,
		"created":     time.Now().Unix(),
		"updated":     time.Now().Unix(),
		"format_name": rpc.FormatFilename(filename),
	}
	id, err := rpc.RpcServer.ZincInput(rpc.FileIndex, doc)
	log.Debug().Msgf("zinc input doc id %s path %s", id, filepath)
	return err
}

func printTime(s string, args ...interface{}) {
	log.Info().Msgf(time.Now().Format("15:04:05.0000")+" "+s+"\n", args...)
}
