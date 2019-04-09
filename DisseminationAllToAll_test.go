package p2p

import (
	"github.com/crossdock/crossdock-go/assert"
	"sync/atomic"
	"testing"
	"time"
	. "xianhetian.com/tendermint/p2p/gossip"
)

func TestGossipNode1(t *testing.T) {
	t.Parallel()
	var receivedMessages int32
	bootPeers := append([]int(nil), []int{1, 2, 3, 4}...)
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(6610, 1, 100, &NaiveMessageCryptoService{}, bootPeers...)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	go func() {
		for range acceptChan {
			atomic.AddInt32(&receivedMessages, 1)
		}
	}()
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(pI, 4))
	// 5.Gossip传播信息
	pI.Gossip(createDataMsg(uint64(1), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	// 6.校验结果是否正确
	assert.Equal(t, int32(4), receivedMessages)
	pI.Stop()
}

func TestGossipNode2(t *testing.T) {
	t.Parallel()
	var receivedMessages int32
	bootPeers := append([]int(nil), []int{0, 2, 3, 4}...)
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(6610, 2, 100, &NaiveMessageCryptoService{}, bootPeers...)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	go func() {
		for range acceptChan {
			atomic.AddInt32(&receivedMessages, 1)
		}
	}()
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(pI, 4))
	// 5.Gossip传播信息
	pI.Gossip(createDataMsg(uint64(2), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	// 6.校验结果是否正确
	assert.Equal(t, int32(4), receivedMessages)
	pI.Stop()
}

func TestGossipNode3(t *testing.T) {
	t.Parallel()
	var receivedMessages int32
	bootPeers := append([]int(nil), []int{0, 1, 3, 4}...)
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(6610, 3, 100, &NaiveMessageCryptoService{}, bootPeers...)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	go func() {
		for range acceptChan {
			atomic.AddInt32(&receivedMessages, 1)
		}
	}()
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(pI, 4))
	// 5.Gossip传播信息
	pI.Gossip(createDataMsg(uint64(3), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	// 6.校验结果是否正确
	assert.Equal(t, int32(4), receivedMessages)
	pI.Stop()
}

func TestGossipNode4(t *testing.T) {
	t.Parallel()
	var receivedMessages int32
	bootPeers := append([]int(nil), []int{0, 1, 2, 4}...)
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(6610, 4, 100, &NaiveMessageCryptoService{}, bootPeers...)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	go func() {
		for range acceptChan {
			atomic.AddInt32(&receivedMessages, 1)
		}
	}()
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(pI, 4))
	// 5.Gossip传播信息
	pI.Gossip(createDataMsg(uint64(4), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	// 6.校验结果是否正确
	assert.Equal(t, int32(4), receivedMessages)
	pI.Stop()
}

func TestGossipNode5(t *testing.T) {
	t.Parallel()
	var receivedMessages int32
	bootPeers := append([]int(nil), []int{0, 1, 2, 3}...)
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(6610, 5, 100, &NaiveMessageCryptoService{}, bootPeers...)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	go func() {
		for range acceptChan {
			atomic.AddInt32(&receivedMessages, 1)
		}
	}()
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(pI, 4))
	// 5.Gossip传播信息
	pI.Gossip(createDataMsg(uint64(5), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	// 6.校验结果是否正确
	assert.Equal(t, int32(4), receivedMessages)
	pI.Stop()
}
