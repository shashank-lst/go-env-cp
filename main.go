package main

import (
	"context"
	fileprocessor "envoy-cp/processors/fileprocessor"
	filewatcher "envoy-cp/processors/filewatcher"
	server "envoy-cp/processors/server"
	"flag"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	log "github.com/sirupsen/logrus"
)

var (
	l log.FieldLogger

	watchDirectoryFileName string
	port                   uint
	basePort               uint
	mode                   string

	nodeID string
)

func init() {
	l = log.New()
	log.SetLevel(log.DebugLevel)

	// The port that this xDS server listens on
	flag.UintVar(&port, "port", 9002, "xDS management server port")

	// Tell Envoy to use this Node ID
	flag.StringVar(&nodeID, "nodeID", "test-id", "Node ID")

	// Define the directory to watch for Envoy configuration files
	flag.StringVar(&watchDirectoryFileName, "watchDirectoryFileName", "config/config.json", "full path to directory to watch for files")
}

func main() {
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	// Create a processor
	proc := fileprocessor.NewProcessor(
		cache, nodeID, log.WithField("context", "processor"))

	// Create initial snapshot from file
	proc.ProcessFile(filewatcher.NotifyMessage{
		Operation: filewatcher.Create,
		FilePath:  watchDirectoryFileName,
	})

	// Notify channel for file system events
	notifyCh := make(chan filewatcher.NotifyMessage)

	go func() {
		// Watch for file changes
		filewatcher.Watch(watchDirectoryFileName, notifyCh)
	}()

	go func() {
		// Run the xDS server
		ctx := context.Background()
		srv := serverv3.NewServer(ctx, cache, nil)
		server.RunServer(ctx, srv, port)
	}()

	for {
		select {
		case msg := <-notifyCh:
			proc.ProcessFile(msg)
		}
	}
}
