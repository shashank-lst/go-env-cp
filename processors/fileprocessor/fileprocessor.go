package fileprocessor

import (
	"context"
	fileparser "envoy-cp/processors/fileparser"
	"envoy-cp/processors/filewatcher"
	"envoy-cp/processors/resources"
	"envoy-cp/processors/xdscache"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/sirupsen/logrus"
)
type Processor struct {
	cache  cache.SnapshotCache
	nodeID string

	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64
	logrus.FieldLogger
	xdsCache xdscache.XDSCache
}

func NewProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger) *Processor {
	return &Processor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		FieldLogger:     log,
		xdsCache: xdscache.XDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
	}
}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (p *Processor) newSnapshotVersion() string {

	// Reset the snapshotVersion if it ever hits max size.
	if p.snapshotVersion == math.MaxInt64 {
		p.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	p.snapshotVersion++
	return strconv.FormatInt(p.snapshotVersion, 10)
}

func (p *Processor) ProcessFile(file filewatcher.NotifyMessage) {
	envoyConfig, err := fileparser.ParseJson(file.FilePath)
	if err != nil {
		fmt.Errorf("error parsing yaml file: %+v", err)
		return
	}
	// Parse Listeners
	for _, l := range envoyConfig.Listeners {
		var lRoutes []string
		for _, lr := range l.Routes {
			lRoutes = append(lRoutes, lr.Name)
		}

		p.xdsCache.AddListener(l.Name, lRoutes, l.Address, l.Port)

		for _, r := range l.Routes {
			p.xdsCache.AddRoute(r.Name, r.Prefix, r.ClusterNames)
		}
	}

	// Parse Clusters
	for _, c := range envoyConfig.Clusters {
		p.xdsCache.AddCluster(c.Name)

		// Parse endpoints
		for _, e := range c.Endpoints {
			p.xdsCache.AddEndpoint(c.Name, e.Address, e.Port)
		}
	}

	// Create a resource map keyed off the type URL of a resource,
	// followed by the slice of resource objects.
	resources := map[resource.Type][]types.Resource{
		resource.EndpointType: p.xdsCache.EndpointsContents(),
		resource.ClusterType:  p.xdsCache.ClusterContents(),
		resource.RouteType:    p.xdsCache.RouteContents(),
		resource.ListenerType: p.xdsCache.ListenerContents(),
	}
	// Create the snapshot that we'll serve to Envoy
	snapshot, err := cache.NewSnapshot(
		p.newSnapshotVersion(), // version
		resources,
	)
	if err != nil {
		p.Errorf("error generating new snapshot: %v", err)
		return
	}

	if err := snapshot.Consistent(); err != nil {
		p.Errorf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		return
	}
	p.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := p.cache.SetSnapshot(context.Background(), p.nodeID, snapshot); err != nil {
		p.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}