package p2p

import (
	"testing"
	"time"
	. "xianhetian.com/tendermint/p2p/gossip"
)

func TestLeaveChannelNode1(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	p0 := DefaultGossipInstance(4500, 0, 100, &NaiveMessageCryptoService{}, 2)
	// 2.将Gossip实例加入到Channel中
	p0.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	p0.UpdateLedgerHeight(1, ChainID("A"))
	countMembership := func(g Gossip, expected int) func() bool {
		return func() bool {
			// 获取存活并且订阅指定Channel的Peer列表
			peers := g.PeersOfChannel(ChainID("A"))
			return len(peers) == expected
		}
	}
	// 4.等待校验结果通过
	waitUntilOrFail(t, countMembership(p0, 2))
	waitUntilOrFail(t, countMembership(p0, 1))
	time.Sleep(10 * time.Second)
	p0.Stop()
}

func TestLeaveChannelNode2(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	p1 := DefaultGossipInstance(4500, 1, 100, &NaiveMessageCryptoService{}, 0)
	// 2.将Gossip实例加入到Channel中
	p1.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	p1.UpdateLedgerHeight(1, ChainID("A"))
	countMembership := func(g Gossip, expected int) func() bool {
		return func() bool {
			// 获取存活并且订阅指定Channel的Peer列表
			peers := g.PeersOfChannel(ChainID("A"))
			return len(peers) == expected
		}
	}
	// 4.等待校验结果通过
	waitUntilOrFail(t, countMembership(p1, 2))
	waitUntilOrFail(t, countMembership(p1, 1))
	time.Sleep(10 * time.Second)
	p1.Stop()
}

func TestLeaveChannelNode3(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	p2 := DefaultGossipInstance(4500, 2, 100, &NaiveMessageCryptoService{}, 1)
	// 2.将Gossip实例加入到Channel中
	p2.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	p2.UpdateLedgerHeight(1, ChainID("A"))
	countMembership := func(g Gossip, expected int) func() bool {
		return func() bool {
			// 获取存活并且订阅指定Channel的Peer列表
			peers := g.PeersOfChannel(ChainID("A"))
			return len(peers) == expected
		}
	}
	waitUntilOrFail(t, countMembership(p2, 2))
	// 使Gossip实例离开Channel
	p2.LeaveChan(ChainID("A"))
	// 4.等待校验结果通过
	waitUntilOrFail(t, countMembership(p2, 0))
	time.Sleep(10 * time.Second)
	p2.Stop()
}
