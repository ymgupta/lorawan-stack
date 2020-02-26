// Copyright © 2020 The Things Industries B.V.

package cluster

import "context"

func SetResolver(cluster Cluster, resolver dnsResolver) {
	cluster.(*dnsCluster).resolver = resolver
}

func UpdatePeers(ctx context.Context, cluster Cluster) {
	cluster.(*dnsCluster).updatePeers(ctx)
}
