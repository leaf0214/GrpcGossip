package gossip

import "math/rand"

//---------- Var -----------

var SelectNonePolicy = func(Peer) bool {
	return false
}

var SelectAllPolicy = func(Peer) bool {
	return true
}

//---------- Define -----------

type RoutingFilter func(Peer) bool

//---------- Public Methods -----------

func CombineRoutingFilters(filters ...RoutingFilter) RoutingFilter {
	return func(peer Peer) bool {
		for _, filter := range filters {
			if !filter(peer) {
				return false
			}
		}
		return true
	}
}

func SelectPeers(k int, peerPool []Peer, filter RoutingFilter) []*RemotePeer {
	var res []*RemotePeer
	rand.Seed(int64(RandomUInt64()))
	for _, index := range rand.Perm(len(peerPool)) {
		if len(res) == k {
			break
		}
		peer := peerPool[index]
		if !filter(peer) {
			continue
		}
		p := &RemotePeer{PKIID: peer.PKIid, Endpoint: peer.PreferredEndpoint()}
		res = append(res, p)
	}
	return res
}
