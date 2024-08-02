package watcher

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	*fsnotify.Watcher
	BasePath        string
	CreateEventChan chan *Event
	ModifyEventChan chan *Event
	DeleteEventChan chan *Event
	ErrorChan       chan error
	DoneChan        chan struct{}
	StopChan        chan struct{}
	wg              sync.WaitGroup
}

type FileType int

const (
	File FileType = iota
	Directory
	Deleted
)

type EventType int

const (
	Create EventType = iota
	Modify
	Delete
)

type Event struct {
	Name       string
	Path       string
	FullPath   string
	FileType   FileType
	EventType  EventType
	ModifiedAt time.Time
}

func (w *Watcher) TearDown() error {
	err := w.Watcher.Close()
	// TODO TBD nice way to tear down
	return err
}

func (w *Watcher) AddAll(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return w.Add(path)
		}
		return nil
	})
}

func NewWatcher(path string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Watcher:         watcher,
		BasePath:        path,
		CreateEventChan: make(chan *Event),
		ModifyEventChan: make(chan *Event),
		DeleteEventChan: make(chan *Event),
		ErrorChan:       make(chan error),
		DoneChan:        make(chan struct{}),
		StopChan:        make(chan struct{}),
	}

	err = w.AddAll(path)

	if err != nil {
		return nil, err
	}
	return w, nil
}

func StartWatch(w *Watcher) {
	defer func() {
		w.DoneChan <- struct{}{}
	}()

	for {
		select {
		case event, ok := <-w.Watcher.Events:
			if !ok {
				log.Println("watch event channel closed")
				return
			}
			err := w.handleEvent(event)
			if err != nil {
				w.ErrorChan <- err
			}
		case err, ok := <-w.Watcher.Errors:
			if !ok {
				log.Println("watch error channel closed")
				return
			}

			w.ErrorChan <- err
		case <-w.StopChan:
			log.Println("received stop signal from outside of event loop")
			return
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) error {

	var eventType EventType
	if event.Op&fsnotify.Create == fsnotify.Create {
		eventType = Create
	} else if event.Op&fsnotify.Write == fsnotify.Write {
		eventType = Modify
	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		eventType = Delete
	} else {
		return nil
	}

	fullPath := event.Name
	e, err := getEvent(eventType, fullPath)
	if err != nil {
		return err
	}

	w.addToWatcher(e)
	w.SendToChan(e)

	return nil
}

func (w *Watcher) addToWatcher(e *Event) {
	if e.FileType == Directory && e.EventType == Create {
		err := w.AddAll(e.FullPath)
		if err != nil {
			w.ErrorChan <- err
		}
	}
}

func (w *Watcher) SendToChan(e *Event) {
	switch e.EventType {
	case Create:
		w.CreateEventChan <- e
	case Modify:
		w.ModifyEventChan <- e
	case Delete:
		w.DeleteEventChan <- e
	}
}

func getEvent(eventType EventType, fullPath string) (*Event, error) {
	var name, path string
	var fileType FileType
	var modifiedAt time.Time

	switch eventType {
	case Create, Modify:
		info, err := os.Stat(fullPath)

		if err != nil {
			return nil, err
		}

		name = info.Name()
		path = strings.TrimSuffix(fullPath, "/"+name)
		if info.IsDir() {
			fileType = Directory
		} else {
			fileType = File
		}
		modifiedAt = info.ModTime()
	case Delete:
		path, name = filepath.Split(fullPath)
		path = strings.TrimSuffix(path, "/")
		fileType = Deleted
		modifiedAt = time.Now()
	}

	return &Event{
		Name:       name,
		Path:       path,
		FullPath:   fullPath,
		EventType:  eventType,
		FileType:   fileType,
		ModifiedAt: modifiedAt,
	}, nil
}
