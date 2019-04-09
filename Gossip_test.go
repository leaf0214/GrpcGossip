package p2p

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	. "xianhetian.com/tendermint/p2p/gossip"
)

//---------- Var -----------

var (
	testWG  = sync.WaitGroup{}
	timeout = time.Second * time.Duration(180)
)

var tests = []func(t *testing.T){
	TestDissemination,
	TestDisseminateAll2All,
	TestPull,
	TestConnectToAnchorPeers,
	TestMembership,
	TestMembershipConvergence,
	TestMembershipRequestSpoofing,
	TestDataLeakage,
	TestLeaveChannel,
	TestIdExpiration,
	TestSendByCriteria,
	TestNoMessagesSelfLoop,
}

//---------- Public Methods -----------

func init() {
	rand.Seed(int64(time.Now().Second()))
	aliveTimeInterval := time.Duration(1000) * time.Millisecond
	SetAliveTimeInterval(aliveTimeInterval)
	SetAliveExpirationCheckInterval(aliveTimeInterval)
	SetAliveExpirationTimeout(aliveTimeInterval * 10)
	SetReconnectInterval(aliveTimeInterval)
	SetMaxConnAttempts(5)
	for range tests {
		testWG.Add(1)
	}
}

func TestDissemination(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 3610
	t1 := time.Now()
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	n := 4
	msgsCount2Send := 4
	boot := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	boot.JoinChan(&JoinChanMsg{}, ChainID("A"))
	boot.UpdateLedgerHeight(1, ChainID("A"))
	peers := make([]Gossip, n)
	receivedMessages := make([]int, n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 1; i <= n; i++ {
		pI := DefaultGossipInstance(portPrefix, i, 100, &NaiveMessageCryptoService{}, 0)
		peers[i-1] = pI
		pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
		pI.UpdateLedgerHeight(1, ChainID("A"))
		acceptChan, _ := pI.Receive(acceptData, false)
		go func(index int, acceptChan <-chan *GossipMessage) {
			defer wg.Done()
			for j := 0; j < msgsCount2Send; j++ {
				<-acceptChan
				receivedMessages[index]++
			}
		}(i-1, acceptChan)
		if i == n {
			pI.UpdateLedgerHeight(2, ChainID("A"))
		}
	}
	membershipTime := time.Now()
	waitUntilOrFail(t, checkPeersMembership(t, peers, n))
	t.Log("Membership establishment took", time.Since(membershipTime))
	for i := 2; i <= msgsCount2Send+1; i++ {
		boot.Gossip(createDataMsg(uint64(i), []byte{}, ChainID("A")))
	}
	t2 := time.Now()
	waitUntilOrFailBlocking(t, wg.Wait)
	t.Log("Block dissemination took", time.Since(t2))
	t2 = time.Now()
	var lastPeer = fmt.Sprintf("localhost:%d", n+portPrefix)
	metaDataUpdated := func() bool {
		if 2 != heightOfPeer(boot.PeersOfChannel(ChainID("A")), lastPeer) {
			return false
		}
		for i := 0; i < n-1; i++ {
			if 2 != heightOfPeer(peers[i].PeersOfChannel(ChainID("A")), lastPeer) {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, metaDataUpdated)
	t.Log("Metadata dissemination took", time.Since(t2))
	for i := 0; i < n; i++ {
		assert.Equal(t, msgsCount2Send, receivedMessages[i])
	}
	t.Log("Stopping peers")
	stop := func() { stopPeers(append(peers, boot)) }
	stopTime := time.Now()
	waitUntilOrFailBlocking(t, stop)
	t.Log("Stop took", time.Since(stopTime))
	t.Log("Took", time.Since(t1))
	atomic.StoreInt32(&stopped, int32(1))
}

func TestDisseminateAll2All(t *testing.T) {
	t.Parallel()
	portPrefix := 6610
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	totalPeers := []int{0, 1, 2, 3, 4, 5, 6}
	n := len(totalPeers)
	peers := make([]Gossip, n)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			totPeers := append([]int(nil), totalPeers[:i]...)
			bootPeers := append(totPeers, totalPeers[i+1:]...)
			pI := DefaultGossipInstance(portPrefix, i, 100, &NaiveMessageCryptoService{}, bootPeers...)
			pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
			pI.UpdateLedgerHeight(1, ChainID("A"))
			peers[i] = pI
			wg.Done()
		}(i)
	}
	wg.Wait()
	waitUntilOrFail(t, checkPeersMembership(t, peers, n-1))
	bMutex := sync.WaitGroup{}
	bMutex.Add(10 * n * (n - 1))
	wg = sync.WaitGroup{}
	wg.Add(n)
	reader := func(msgChan <-chan *GossipMessage, i int) {
		wg.Done()
		for range msgChan {
			bMutex.Done()
		}
	}
	for i := 0; i < n; i++ {
		msgChan, _ := peers[i].Receive(acceptData, false)
		go reader(msgChan, i)
	}
	wg.Wait()
	for i := 0; i < n; i++ {
		go func(i int) {
			for j := 0; j < 10; j++ {
				blockSeq := uint64(j + i*10)
				peers[i].Gossip(createDataMsg(blockSeq, []byte{}, ChainID("A")))
			}
		}(i)
	}
	waitUntilOrFailBlocking(t, bMutex.Wait)
	stop := func() {
		stopPeers(peers)
	}
	waitUntilOrFailBlocking(t, stop)
	atomic.StoreInt32(&stopped, int32(1))
	testWG.Done()
}

func TestPull(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 5610
	t1 := time.Now()
	shortenedWaitTime := time.Duration(200) * time.Millisecond
	SetDigestWaitTime(shortenedWaitTime)
	SetRequestWaitTime(shortenedWaitTime)
	SetResponseWaitTime(shortenedWaitTime)
	defer func() {
		SetDigestWaitTime(time.Duration(1) * time.Second)
		SetRequestWaitTime(time.Duration(1) * time.Second)
		SetResponseWaitTime(time.Duration(2) * time.Second)
	}()
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	n := 5
	msgsCount2Send := 10
	peers := make([]Gossip, n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 1; i <= n; i++ {
		go func(i int) {
			defer wg.Done()
			pI := newGossipInstanceWithOnlyPull(portPrefix, i, 100, 0)
			pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
			pI.UpdateLedgerHeight(1, ChainID("A"))
			peers[i-1] = pI
		}(i)
	}
	wg.Wait()
	time.Sleep(time.Second)
	boot := newGossipInstanceWithOnlyPull(portPrefix, 0, 100)
	boot.JoinChan(&JoinChanMsg{}, ChainID("A"))
	boot.UpdateLedgerHeight(1, ChainID("A"))
	knowAll := func() bool {
		for i := 1; i <= n; i++ {
			neighborCount := len(peers[i-1].Peers())
			if n != neighborCount {
				return false
			}
		}
		return true
	}
	receivedMessages := make([]int, n)
	wg = sync.WaitGroup{}
	wg.Add(n)
	for i := 1; i <= n; i++ {
		go func(i int) {
			acceptChan, _ := peers[i-1].Receive(acceptData, false)
			go func(index int, ch <-chan *GossipMessage) {
				defer wg.Done()
				for j := 0; j < msgsCount2Send; j++ {
					<-ch
					receivedMessages[index]++
				}
			}(i-1, acceptChan)
		}(i)
	}
	for i := 1; i <= msgsCount2Send; i++ {
		boot.Gossip(createDataMsg(uint64(i), []byte{}, ChainID("A")))
	}
	waitUntilOrFail(t, knowAll)
	waitUntilOrFailBlocking(t, wg.Wait)
	receivedAll := func() bool {
		for i := 0; i < n; i++ {
			if msgsCount2Send != receivedMessages[i] {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, receivedAll)
	stop := func() {
		stopPeers(append(peers, boot))
	}
	waitUntilOrFailBlocking(t, stop)
	t.Log("Took", time.Since(t1))
	atomic.StoreInt32(&stopped, int32(1))
}

func TestLeaveChannel(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 4500
	p0 := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{}, 2)
	p0.JoinChan(&JoinChanMsg{}, ChainID("A"))
	p0.UpdateLedgerHeight(1, ChainID("A"))
	defer p0.Stop()
	p1 := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{}, 0)
	p1.JoinChan(&JoinChanMsg{}, ChainID("A"))
	p1.UpdateLedgerHeight(1, ChainID("A"))
	defer p1.Stop()
	p2 := DefaultGossipInstance(portPrefix, 2, 100, &NaiveMessageCryptoService{}, 1)
	p2.JoinChan(&JoinChanMsg{}, ChainID("A"))
	p2.UpdateLedgerHeight(1, ChainID("A"))
	defer p2.Stop()
	countMembership := func(g Gossip, expected int) func() bool {
		return func() bool {
			peers := g.PeersOfChannel(ChainID("A"))
			return len(peers) == expected
		}
	}
	waitUntilOrFail(t, countMembership(p0, 2))
	waitUntilOrFail(t, countMembership(p1, 2))
	waitUntilOrFail(t, countMembership(p2, 2))
	p2.LeaveChan(ChainID("A"))
	waitUntilOrFail(t, countMembership(p0, 1))
	waitUntilOrFail(t, countMembership(p1, 1))
	waitUntilOrFail(t, countMembership(p2, 0))
}

func TestConnectToAnchorPeers(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 8610
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	n := 10
	anchorPeercount := 3
	jcm := &JoinChanMsg{Members2AnchorPeers: map[string][]AnchorPeer{string(DefaultOrgInChannelA): {}}}
	for i := 0; i < anchorPeercount; i++ {
		ap := AnchorPeer{
			Port: portPrefix + i,
			Host: "localhost",
		}
		jcm.Members2AnchorPeers[string(DefaultOrgInChannelA)] = append(jcm.Members2AnchorPeers[string(DefaultOrgInChannelA)], ap)
	}
	peers := make([]Gossip, n)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			peers[i] = DefaultGossipInstance(portPrefix, i+anchorPeercount, 100, &NaiveMessageCryptoService{})
			peers[i].JoinChan(jcm, ChainID("A"))
			peers[i].UpdateLedgerHeight(1, ChainID("A"))
			wg.Done()
		}(i)
	}
	waitUntilOrFailBlocking(t, wg.Wait)
	time.Sleep(time.Second * 5)
	anchorPeer := DefaultGossipInstance(portPrefix, rand.Intn(anchorPeercount), 100, &NaiveMessageCryptoService{})
	anchorPeer.JoinChan(jcm, ChainID("A"))
	anchorPeer.UpdateLedgerHeight(1, ChainID("A"))
	defer anchorPeer.Stop()
	waitUntilOrFail(t, checkPeersMembership(t, peers, n))
	channelMembership := func() bool {
		for _, peer := range peers {
			if len(peer.PeersOfChannel(ChainID("A"))) != n {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, channelMembership)
	stop := func() {
		stopPeers(peers)
	}
	waitUntilOrFailBlocking(t, stop)
	atomic.StoreInt32(&stopped, int32(1))
}

func TestMembership(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 4610
	t1 := time.Now()
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	n := 10
	boot := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	boot.JoinChan(&JoinChanMsg{}, ChainID("A"))
	boot.UpdateLedgerHeight(1, ChainID("A"))
	peers := make([]Gossip, n)
	wg := sync.WaitGroup{}
	wg.Add(n)
	for i := 1; i <= n; i++ {
		go func(i int) {
			defer wg.Done()
			pI := DefaultGossipInstance(portPrefix, i, 100, &NaiveMessageCryptoService{}, 0)
			peers[i-1] = pI
			pI.JoinChan(&JoinChanMsg{}, ChainID("A"))
			pI.UpdateLedgerHeight(1, ChainID("A"))
		}(i)
	}
	waitUntilOrFailBlocking(t, wg.Wait)
	t.Log("Peers started")
	seeAllNeighbors := func() bool {
		for i := 1; i <= n; i++ {
			neighborCount := len(peers[i-1].Peers())
			if neighborCount != n {
				return false
			}
		}
		return true
	}
	membershipEstablishTime := time.Now()
	waitUntilOrFail(t, seeAllNeighbors)
	t.Log("Membership established in", time.Since(membershipEstablishTime))
	stop := func() {
		stopPeers(append(peers, boot))
	}
	stopTime := time.Now()
	waitUntilOrFailBlocking(t, stop)
	t.Log("Stop took", time.Since(stopTime))
	t.Log("Took", time.Since(t1))
	atomic.StoreInt32(&stopped, int32(1))
}

func TestNoMessagesSelfLoop(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 17610
	boot := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	boot.JoinChan(&JoinChanMsg{}, ChainID("A"))
	boot.UpdateLedgerHeight(1, ChainID("A"))
	peer := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{}, 0)
	peer.JoinChan(&JoinChanMsg{}, ChainID("A"))
	peer.UpdateLedgerHeight(1, ChainID("A"))
	waitUntilOrFail(t, checkPeersMembership(t, []Gossip{peer}, 1))
	_, commCh := boot.Receive(func(msg interface{}) bool {
		return msg.(*ReceivedMessage).SignedGossipMessage.GetDataMsg() != nil
	}, true)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func(ch <-chan *ReceivedMessage) {
		defer wg.Done()
		for {
			select {
			case msg := <-ch:
				if msg.SignedGossipMessage.GetDataMsg() != nil {
					t.Fatal("Should not receive data message back, got", msg)
				}
			case <-time.After(2 * time.Second):
				return
			}
		}
	}(commCh)
	peerCh, _ := peer.Receive(acceptData, false)
	go func(ch <-chan *GossipMessage) {
		defer wg.Done()
		<-ch
	}(peerCh)
	boot.Gossip(createDataMsg(uint64(2), []byte{}, ChainID("A")))
	waitUntilOrFailBlocking(t, wg.Wait)
	stop := func() {
		stopPeers([]Gossip{peer, boot})
	}
	waitUntilOrFailBlocking(t, stop)
}

func TestMembershipConvergence(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 2610
	t1 := time.Now()
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	boot0 := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	boot1 := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{})
	boot2 := DefaultGossipInstance(portPrefix, 2, 100, &NaiveMessageCryptoService{})
	peers := []Gossip{boot0, boot1, boot2}
	for i := 3; i < 7; i++ {
		pI := DefaultGossipInstance(portPrefix, i, 100, &NaiveMessageCryptoService{}, i%3)
		peers = append(peers, pI)
	}
	t.Log("Sets of peers connected successfully")
	connectorPeer := DefaultGossipInstance(portPrefix, 7, 100, &NaiveMessageCryptoService{}, 0, 1, 2)
	fullKnowledge := func() bool {
		for i := 0; i < 7; i++ {
			if 7 != len(peers[i].Peers()) {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, fullKnowledge)
	t.Log("Stopping connector...")
	waitUntilOrFailBlocking(t, connectorPeer.Stop)
	t.Log("Stopped")
	time.Sleep(time.Duration(5) * time.Second)
	ensureForget := func() bool {
		for i := 0; i < 7; i++ {
			if 6 != len(peers[i].Peers()) {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, ensureForget)
	connectorPeer = DefaultGossipInstance(portPrefix, 7, 100, &NaiveMessageCryptoService{}, 0, 1, 2)
	t.Log("Started connector")
	ensureResync := func() bool {
		for i := 0; i < 7; i++ {
			if 7 != len(peers[i].Peers()) {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, ensureResync)
	waitUntilOrFailBlocking(t, connectorPeer.Stop)
	t.Log("Stopping peers")
	stop := func() {
		stopPeers(peers)
	}
	waitUntilOrFailBlocking(t, stop)
	atomic.StoreInt32(&stopped, int32(1))
	t.Log("Took", time.Since(t1))
}

func TestMembershipRequestSpoofing(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 2000
	g1 := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	g2 := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{}, 2)
	g3 := DefaultGossipInstance(portPrefix, 2, 100, &NaiveMessageCryptoService{}, 1)
	defer g1.Stop()
	defer g2.Stop()
	defer g3.Stop()
	waitUntilOrFail(t, checkPeersMembership(t, []Gossip{g2, g3}, 1))
	_, aliveMsgChan := g2.Receive(func(o interface{}) bool {
		msg := o.(*ReceivedMessage).SignedGossipMessage
		return msg.GetAliveMsg() != nil && bytes.Equal(msg.GetAliveMsg().Membership.PkiId, []byte("localhost:2002"))
	}, true)
	aliveMsg := <-aliveMsgChan
	_, g1ToG2 := g2.Receive(func(o interface{}) bool {
		connInfo := o.(*ReceivedMessage).ConnInfo
		return bytes.Equal([]byte("localhost:2000"), connInfo.ID)
	}, true)
	_, g1ToG3 := g3.Receive(func(o interface{}) bool {
		connInfo := o.(*ReceivedMessage).ConnInfo
		return bytes.Equal([]byte("localhost:2000"), connInfo.ID)
	}, true)
	memRequestSpoofFactory := func(aliveMsgEnv *Envelope) *SignedGossipMessage {
		sMsg, _ := (&GossipMessage{
			Tag:   GossipMessage_EMPTY,
			Nonce: uint64(0),
			Content: &GossipMessage_MemReq{
				MemReq: &MembershipRequest{
					SelfInformation: aliveMsgEnv,
					Known:           [][]byte{},
				},
			},
		}).NoopSign()
		return sMsg
	}
	spoofedMemReq := memRequestSpoofFactory(aliveMsg.Envelope)
	g2.Send(spoofedMemReq.GossipMessage, &RemotePeer{Endpoint: "localhost:2000", PKIID: PKIidType("localhost:2000")})
	select {
	case <-time.After(time.Second):
		break
	case <-g1ToG2:
		assert.Fail(t, "Received response from g1 but shouldn't have")
	}
	g3.Send(spoofedMemReq.GossipMessage, &RemotePeer{Endpoint: "localhost:2000", PKIID: PKIidType("localhost:2000")})
	select {
	case <-time.After(time.Second):
		assert.Fail(t, "Didn't receive a message back from g1 on time")
	case <-g1ToG3:
		break
	}
}

func TestDataLeakage(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 1610
	totalPeers := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	mcs := &NaiveMessageCryptoService{
		AllowedPkiIDS: map[string]struct{}{
			// Channel A
			"localhost:1610": {},
			"localhost:1611": {},
			"localhost:1612": {},
			// Channel B
			"localhost:1615": {},
			"localhost:1616": {},
			"localhost:1617": {},
		},
	}
	stopped := int32(0)
	go waitForTestCompletion(&stopped, t)
	n := len(totalPeers)
	peers := make([]Gossip, n)
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			totPeers := append([]int(nil), totalPeers[:i]...)
			bootPeers := append(totPeers, totalPeers[i+1:]...)
			peers[i] = DefaultGossipInstance(portPrefix, i, 100, mcs, bootPeers...)
			wg.Done()
		}(i)
	}
	waitUntilOrFailBlocking(t, wg.Wait)
	waitUntilOrFail(t, checkPeersMembership(t, peers, n-1))
	channels := []ChainID{ChainID("A"), ChainID("B")}
	height := uint64(1)
	for i, channel := range channels {
		for j := 0; j < (n / 2); j++ {
			instanceIndex := (n/2)*i + j
			peers[instanceIndex].JoinChan(&JoinChanMsg{}, channel)
			if i != 0 {
				height = uint64(2)
			}
			peers[instanceIndex].UpdateLedgerHeight(height, channel)
			t.Log(instanceIndex, "joined", string(channel))
		}
	}
	seeChannelMetadata := func() bool {
		for i, channel := range channels {
			for j := 0; j < 3; j++ {
				instanceIndex := (n/2)*i + j
				if len(peers[instanceIndex].PeersOfChannel(channel)) < 2 {
					return false
				}
			}
		}
		return true
	}
	t1 := time.Now()
	waitUntilOrFail(t, seeChannelMetadata)
	t.Log("Metadata sync took", time.Since(t1))
	for i, channel := range channels {
		for j := 0; j < 3; j++ {
			instanceIndex := (n/2)*i + j
			assert.Len(t, peers[instanceIndex].PeersOfChannel(channel), 2)
			if i == 0 {
				assert.Equal(t, uint64(1), peers[instanceIndex].PeersOfChannel(channel)[0].Properties.LedgerHeight)
			} else {
				assert.Equal(t, uint64(2), peers[instanceIndex].PeersOfChannel(channel)[0].Properties.LedgerHeight)
			}
		}
	}
	gotMessages := func() {
		var wg sync.WaitGroup
		wg.Add(4)
		for i, channel := range channels {
			for j := 1; j < 3; j++ {
				instanceIndex := (n/2)*i + j
				go func(instanceIndex int, channel ChainID) {
					incMsgChan, _ := peers[instanceIndex].Receive(acceptData, false)
					msg := <-incMsgChan
					assert.Equal(t, []byte(channel), []byte(msg.Channel))
					wg.Done()
				}(instanceIndex, channel)
			}
		}
		wg.Wait()
	}
	t1 = time.Now()
	peers[0].Gossip(createDataMsg(2, []byte{}, channels[0]))
	peers[n/2].Gossip(createDataMsg(3, []byte{}, channels[1]))
	waitUntilOrFailBlocking(t, gotMessages)
	t.Log("Dissemination took", time.Since(t1))
	stop := func() {
		stopPeers(peers)
	}
	stopTime := time.Now()
	waitUntilOrFailBlocking(t, stop)
	t.Log("Stop took", time.Since(stopTime))
	atomic.StoreInt32(&stopped, int32(1))
}

func TestSendByCriteria(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	portPrefix := 20000
	g1 := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	g2 := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{}, 0)
	g3 := DefaultGossipInstance(portPrefix, 2, 100, &NaiveMessageCryptoService{}, 0)
	g4 := DefaultGossipInstance(portPrefix, 3, 100, &NaiveMessageCryptoService{}, 0)
	peers := []Gossip{g1, g2, g3, g4}
	for _, p := range peers {
		p.JoinChan(&JoinChanMsg{}, ChainID("A"))
		p.UpdateLedgerHeight(1, ChainID("A"))
	}
	defer stopPeers(peers)
	msg, _ := createDataMsg(1, []byte{}, ChainID("A")).NoopSign()
	criteria := SendCriteria{
		IsEligible: func(Peer) bool {
			t.Fatal("Shouldn't have called, because when max peers is 0, the operation is a no-op")
			return false
		},
		Timeout: time.Second * 1,
		MinAck:  1,
	}
	assert.NoError(t, g1.SendByCriteria(msg, criteria))
	criteria = SendCriteria{MaxPeers: 100}
	err := g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Equal(t, "Timeout should be specified", err.Error())
	criteria.Timeout = time.Second * 3
	err = g1.SendByCriteria(msg, criteria)
	assert.NoError(t, err)
	criteria.Channel = ChainID("B")
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "but no such channel exists")
	criteria.Channel = ChainID("A")
	criteria.MinAck = 10
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requested to send to at least 10 peers, but know only of")
	waitUntilOrFail(t, func() bool {
		return len(g1.PeersOfChannel(ChainID("A"))) > 2
	})
	criteria.MinAck = 3
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	assert.Contains(t, err.Error(), "3")
	acceptDataMsgs := func(m interface{}) bool {
		return m.(*ReceivedMessage).SignedGossipMessage.GetDataMsg() != nil
	}
	_, ackChan2 := g2.Receive(acceptDataMsgs, true)
	_, ackChan3 := g3.Receive(acceptDataMsgs, true)
	_, ackChan4 := g4.Receive(acceptDataMsgs, true)
	ack := func(c <-chan *ReceivedMessage) {
		msg := <-c
		msg.Ack(nil)
	}
	go ack(ackChan2)
	go ack(ackChan3)
	go ack(ackChan4)
	err = g1.SendByCriteria(msg, criteria)
	assert.NoError(t, err)
	nack := func(c <-chan *ReceivedMessage) {
		msg := <-c
		msg.Ack(fmt.Errorf("uh oh"))
	}
	go ack(ackChan2)
	go nack(ackChan3)
	go nack(ackChan4)
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uh oh")
	failOnAckRequest := func(c <-chan *ReceivedMessage, peerId int) {
		msg := <-c
		if msg == nil {
			return
		}
		t.Fatalf("%d got a message, but shouldn't have!", peerId)
	}
	g2Endpoint := fmt.Sprintf("localhost:%d", portPrefix+1)
	g3Endpoint := fmt.Sprintf("localhost:%d", portPrefix+2)
	criteria.IsEligible = func(nm Peer) bool {
		return nm.InternalEndpoint == g2Endpoint || nm.InternalEndpoint == g3Endpoint
	}
	criteria.MinAck = 1
	go failOnAckRequest(ackChan4, 3)
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	assert.Contains(t, err.Error(), "2")
	ack(ackChan2)
	ack(ackChan3)
	criteria.MaxPeers = 1
	waitForMessage := func(c <-chan *ReceivedMessage, f func()) {
		select {
		case msg := <-c:
			if msg == nil {
				return
			}
		case <-time.After(time.Second * 5):
			return
		}
		f()
	}
	var messagesSent uint32
	go waitForMessage(ackChan2, func() {
		atomic.AddUint32(&messagesSent, 1)
	})
	go waitForMessage(ackChan3, func() {
		atomic.AddUint32(&messagesSent, 1)
	})
	err = g1.SendByCriteria(msg, criteria)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
	assert.Equal(t, uint32(1), atomic.LoadUint32(&messagesSent))
}

func TestIdExpiration(t *testing.T) {
	t.Parallel()
	defer testWG.Done()
	ExpirationTimes["localhost:7004"] = time.Now().Add(time.Second * 5)
	portPrefix := 7000
	g1 := DefaultGossipInstance(portPrefix, 0, 100, &NaiveMessageCryptoService{})
	g2 := DefaultGossipInstance(portPrefix, 1, 100, &NaiveMessageCryptoService{}, 0)
	g3 := DefaultGossipInstance(portPrefix, 2, 100, &NaiveMessageCryptoService{}, 0)
	g4 := DefaultGossipInstance(portPrefix, 3, 100, &NaiveMessageCryptoService{}, 0)
	g5 := DefaultGossipInstance(portPrefix, 4, 100, &NaiveMessageCryptoService{}, 0)
	peers := []Gossip{g1, g2, g3, g4}
	time.AfterFunc(time.Second*5, func() {
		for _, p := range peers {
			p.(*GossipService).Mcs.Revoke(PKIidType("localhost:7004"))
		}
	})
	seeAllNeighbors := func() bool {
		for i := 0; i < 4; i++ {
			neighborCount := len(peers[i].Peers())
			if neighborCount != 3 {
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, seeAllNeighbors)
	revokedPeerIndex := rand.Intn(4)
	revokedPkiID := PKIidType(fmt.Sprintf("localhost:%d", portPrefix+int(revokedPeerIndex)))
	for i, p := range peers {
		if i == revokedPeerIndex {
			continue
		}
		p.(*GossipService).Mcs.Revoke(revokedPkiID)
	}
	for i := 0; i < 4; i++ {
		if i == revokedPeerIndex {
			continue
		}
		peers[i].SuspectPeers(func(_ PeerIdType) bool {
			return true
		})
	}
	ensureRevokedPeerIsIgnored := func() bool {
		for i := 0; i < 4; i++ {
			neighborCount := len(peers[i].Peers())
			expectedNeighborCount := 2
			if i == revokedPeerIndex || i == 4 {
				expectedNeighborCount = 0
			}
			if neighborCount != expectedNeighborCount {
				fmt.Println("neighbor count of", i, "is", neighborCount)
				return false
			}
		}
		return true
	}
	waitUntilOrFail(t, ensureRevokedPeerIsIgnored)
	stopPeers(peers)
	g5.Stop()
}

//---------- Private Methods -----------

func newGossipInstanceWithOnlyPull(portPrefix int, id int, maxMsgCount int, boot ...int) Gossip {
	port := id + portPrefix
	conf := &GossipConfig{
		BindPort:                   port,
		BootstrapPeers:             BootPeers(portPrefix, boot...),
		ID:                         fmt.Sprintf("p%d", id),
		MaxBlockCountToStore:       maxMsgCount,
		MaxPropagationBurstLatency: time.Duration(1000) * time.Millisecond,
		MaxPropagationBurstSize:    10,
		PropagateIterations:        0,
		PropagatePeerNum:           0,
		PullInterval:               time.Duration(1000) * time.Millisecond,
		PullPeerNum:                20,
		InternalEndpoint:           fmt.Sprintf("localhost:%d", port),
		ExternalEndpoint:           fmt.Sprintf("1.2.3.4:%d", port),
		PublishCertPeriod:          time.Duration(0) * time.Second,
		PublishStateInfoInterval:   time.Duration(1) * time.Second,
		RequestStateInfoInterval:   time.Duration(1) * time.Second,
	}
	cryptoService := &NaiveMessageCryptoService{}
	selfID := PeerIdType(conf.InternalEndpoint)
	return NewGossipServiceWithServer(conf, &OrgCryptoService{}, cryptoService, selfID, nil, nil)
}

func checkPeersMembership(t *testing.T, peers []Gossip, n int) func() bool {
	return func() bool {
		for _, peer := range peers {
			if len(peer.Peers()) != n {
				return false
			}
			for _, p := range peer.Peers() {
				assert.NotNil(t, p.InternalEndpoint)
				assert.NotEmpty(t, p.Endpoint)
			}
		}
		return true
	}
}

func waitForTestCompletion(stopFlag *int32, t *testing.T) {
	time.Sleep(timeout)
	if atomic.LoadInt32(stopFlag) == int32(1) {
		return
	}
	assert.Fail(t, "Didn't stop within a timely manner")
}

func waitUntilOrFailBlocking(t *testing.T, f func()) {
	successChan := make(chan struct{}, 1)
	go func() {
		f()
		successChan <- struct{}{}
	}()
	select {
	case <-time.NewTimer(timeout).C:
		break
	case <-successChan:
		return
	}
	assert.Fail(t, "Timeout expired!")
}

func stopPeers(peers []Gossip) {
	stoppingWg := sync.WaitGroup{}
	stoppingWg.Add(len(peers))
	for i, pI := range peers {
		go func(i int, pI Gossip) {
			defer stoppingWg.Done()
			pI.Stop()
		}(i, pI)
	}
	stoppingWg.Wait()
	time.Sleep(time.Second * time.Duration(2))
}

func heightOfPeer(peers []Peer, endpoint string) int {
	for _, peer := range peers {
		if peer.InternalEndpoint == endpoint {
			return int(peer.Properties.LedgerHeight)
		}
	}
	return -1
}

func checkPeersNum(boot Gossip, total int) func() bool {
	return func() bool {
		if len(boot.Peers()) != total {
			return false
		}
		return true
	}
}

func createDataMsg(seqnum uint64, data []byte, channel ChainID) *GossipMessage {
	return &GossipMessage{
		Channel: []byte(channel),
		Nonce:   0,
		Tag:     GossipMessage_CHAN_AND_ORG,
		Content: &GossipMessage_DataMsg{
			DataMsg: &DataMessage{
				Payload: &Payload{Data: data, SeqNum: seqnum},
			},
		},
	}
}

func acceptData(m interface{}) bool {
	if dataMsg := m.(*GossipMessage).GetDataMsg(); dataMsg != nil {
		return true
	}
	return false
}

func waitUntilOrFail(t *testing.T, pred func() bool) {
	start := time.Now()
	limit := start.UnixNano() + timeout.Nanoseconds()
	for time.Now().UnixNano() < limit {
		if pred() {
			return
		}
		time.Sleep(timeout / 60)
	}
	assert.Fail(t, "Timeout expired!")
}
