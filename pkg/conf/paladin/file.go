package paladin

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	defaultChSize = 10
)

var _ Client = &file{}

// file is file config client.
type file struct {
	values *Map
	rawVal map[string]*Value

	watchCh chan Event
	mx      sync.Mutex
	wg      sync.WaitGroup

	base string
	done chan struct{}
}

func isHiddenFile(name string) bool {
	// TODO: support windows.
	return strings.HasPrefix(filepath.Base(name), ".")
}

func readAllPaths(base string) (cfgFilePaths []string, err error) {
	fi, err := os.Stat(base)
	if err != nil {
		return nil, fmt.Errorf("check local config file fail! error: %s", err)
	}
	// dirs or file to paths
	if !fi.IsDir() {
		return
	}
	files, err := ioutil.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("read dir %s error: %s", base, err)
	}
	for _, f := range files {
		if !f.IsDir() && !isHiddenFile(f.Name()) {
			cfgFilePaths = append(cfgFilePaths, path.Join(base, f.Name()))
		}
	}
	return
}

func loadValuesFromPaths(paths []string) (map[string]*Value, error) {
	// laod config file to values
	var err error
	values := make(map[string]*Value, len(paths))
	for _, fpath := range paths {
		if values[path.Base(fpath)], err = loadValue(fpath); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func loadValue(fpath string) (*Value, error) {
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return &Value{raw: data}, nil
}

// set config path. e.g. -conf "xx/cfg1,zz/cfg2"  will read all file in xx/cfg1 and zz/cfg2
func NewFile(base string) (Client, error) {
	lists := strings.Split(base, ",")
	log.Printf("conf base: %v", base)
	log.Printf("conf lists: %v", lists)
	var filePaths []string

	for _, v := range lists {
		str := filepath.FromSlash(v)
		tmp, err := readAllPaths(str)
		if err != nil {
			return nil, err
		}
		filePaths = append(filePaths, tmp...)
	}

	if len(filePaths) == 0 {
		return nil, fmt.Errorf("empty config path")
	}

	rawVal, err := loadValuesFromPaths(filePaths)
	if err != nil {
		return nil, err
	}

	valMap := &Map{}
	valMap.Store(rawVal)
	fc := &file{
		values:  valMap,
		rawVal:  rawVal,
		watchCh: make(chan Event, len(filePaths)),

		base: base,
		done: make(chan struct{}, 1),
	}

	fc.wg.Add(1)
	go fc.daemon()

	return fc, nil
}

// Get return value by key.
func (f *file) Get(key string) *Value {
	return f.values.Get(key)
}

// GetAll return value map.
func (f *file) GetAll() *Map {
	return f.values
}

// WatchEvent watch multi key.
func (f *file) WatchEvent(ctx context.Context, keys ...string) <-chan Event {
	return f.watchCh
}

// Close close watcher.
func (f *file) Close() error {
	f.done <- struct{}{}
	f.wg.Wait()
	return nil
}

// file config daemon to watch file modification
func (f *file) daemon() {
	defer f.wg.Done()
	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("create file watcher fail! reload function will lose efficacy error: %s", err)
		return
	}
	lists := strings.Split(f.base, ",")
	for _, v := range lists {
		if err = fswatcher.Add(v); err != nil {
			log.Printf("create fsnotify for base path %s fail %s, reload function will lose efficacy", f.base, err)
			return
		}
	}

	//if err = fswatcher.Add(f.base); err != nil {
	//	log.Printf("create fsnotify for base path %s fail %s, reload function will lose efficacy", f.base, err)
	//	return
	//}
	log.Printf("start watch filepath: %s", f.base)
	for event := range fswatcher.Events {
		switch event.Op {
		// use vim edit config will trigger rename
		case fsnotify.Write, fsnotify.Create:
			f.reloadFile(event.Name)
		case fsnotify.Chmod:
		default:
			log.Printf("unsupport event %s ingored", event)
		}
	}
}

func (f *file) reloadFile(name string) {
	log.Printf("reloadFile %s", name)
	if isHiddenFile(name) {
		log.Printf("reloadFile isHiddenFile %s", name)
		return
	}
	// NOTE: in some case immediately read file content after receive event
	// will get old content, sleep 100ms make sure get correct content.
	time.Sleep(200 * time.Millisecond)
	val, err := loadValue(name)
	if err != nil {
		log.Printf("load file %s error: %s, skipped", name, err)
		return
	}
	key := filepath.Base(name)
	f.rawVal[key] = val
	f.values.Store(f.rawVal)

	f.mx.Lock()
	f.mx.Unlock()

	select {
	case f.watchCh <- Event{Event: EventUpdate, Key: key, Value: val.raw}:
	default:
		log.Printf("event channel full discard file %s update event", name)
	}

}
