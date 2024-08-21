package main

import (
	"github.com/hippo-an/sync-net/pkg/discovery"
	"github.com/hippo-an/sync-net/pkg/transfer"
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

	ds := discovery.NewServer()
	go ds.Listen()

	b := discovery.NewBroadcaster()
	go b.Broadcast()

	client := transfer.NewClient(w, ds)
	go client.HandleEvents()

	ts := transfer.NewServer()
	go ts.Listen()

	wg.Wait()
}
