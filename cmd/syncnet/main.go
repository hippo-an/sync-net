package main

import (
	"github.com/hippo-an/sync-net/pkg/config"
	"github.com/hippo-an/sync-net/pkg/discovery"
	"github.com/hippo-an/sync-net/pkg/transfer"
	"github.com/hippo-an/sync-net/pkg/watcher"
	"log"
	"sync"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		log.Fatal("application configuration error", err)
	}

	w, err := watcher.NewWatcher(conf)
	if err != nil {
		log.Fatal("application watcher error", err)
	}

	defer w.TearDown()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go watcher.StartWatch(w)

	ds := discovery.NewServer(conf)
	go ds.Listen()

	b := discovery.NewBroadcaster(conf)
	go b.Broadcast()

	client := transfer.NewClient(conf, w, ds)
	go client.HandleEvents()

	ts := transfer.NewServer(conf)
	go ts.ListenAndConnect(9000)

	wg.Wait()
}
