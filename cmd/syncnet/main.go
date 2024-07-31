package main

import (
	"github.com/hippo-an/sync-net/pkg/utils"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"log"
	"sync"
)

func main() {
	homeDir := utils.PathJoinWithHome("/test")

	w, err := watcher.NewWatcher(homeDir)
	if err != nil {
		log.Fatal(err)
	}

	defer w.TearDown()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go watcher.StartWatch(w)

	go func() {
		for {
			select {
			case ce := <-w.CreateEventChan:
				log.Printf("Create event: %+v", ce)
			case me := <-w.ModifyEventChan:
				log.Printf("Modify event: %+v", me)
			case de := <-w.DeleteEventChan:
				log.Printf("Delete event: %+v", de)
			case <-w.Done:
				wg.Done()
				break
			case err := <-w.ErrorChan:
				log.Printf("Error: %+v", err)
			}
		}
	}()

	wg.Wait()
}
