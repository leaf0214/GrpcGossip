package p2p

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	. "xianhetian.com/tendermint/p2p/gossip"
)

func TestBootstrapNode(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	boot := DefaultGossipInstance(3610, 0, 100, &NaiveMessageCryptoService{})
	// 2.将Gossip实例加入到Channel中
	boot.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	boot.UpdateLedgerHeight(1, ChainID("A"))
	// 4.等待获取到足够的Peer数量
	waitUntilOrFail(t, checkPeersNum(boot, 4))
	// 5.Gossip传播信息
	boot.Gossip(createDataMsg(uint64(1), []byte("data"), ChainID("A")))
	time.Sleep(10 * time.Second)
	boot.Stop()
}

func TestNode1(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(3610, 1, 100, &NaiveMessageCryptoService{}, 0)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	receivedMessages := <-acceptChan
	// 5.校验结果是否正确
	assert.Equal(t, []byte("data"), receivedMessages.GetDataMsg().Payload.Data)
	pI.Stop()
}

func TestNode2(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(3610, 2, 100, &NaiveMessageCryptoService{}, 0)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	receivedMessages := <-acceptChan
	// 5.校验结果是否正确
	assert.Equal(t, []byte("data"), receivedMessages.GetDataMsg().Payload.Data)
	pI.Stop()
}

func TestNode3(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(3610, 3, 100, &NaiveMessageCryptoService{}, 0)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	receivedMessages := <-acceptChan
	// 5.校验结果是否正确
	assert.Equal(t, []byte("data"), receivedMessages.GetDataMsg().Payload.Data)
	pI.Stop()
}

func TestNode4(t *testing.T) {
	t.Parallel()
	// 1.使用默认配置创建一个Gossip实例
	pI := DefaultGossipInstance(3610, 4, 100, &NaiveMessageCryptoService{}, 0)
	// 2.将Gossip实例加入到Channel中
	pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
	// 3.更新P2P发布的高度
	pI.UpdateLedgerHeight(1, ChainID("A"))
	// 4.获取接收回复消息的只读Channel
	acceptChan, _ := pI.Receive(acceptData, false)
	receivedMessages := <-acceptChan
	// 5.校验结果是否正确
	assert.Equal(t, []byte("data"), receivedMessages.GetDataMsg().GetPayload().GetData())
	pI.Stop()
}
