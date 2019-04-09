package gossip

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/mock"
)

//---------- Const -----------

const (
	defDigestWaitTime   = time.Duration(1000) * time.Millisecond
	defRequestWaitTime  = time.Duration(1500) * time.Millisecond
	defResponseWaitTime = time.Duration(2000) * time.Millisecond
)

const (
	HelloMsgType    MsgType = iota
	DigestMsgType
	RequestMsgType
	ResponseMsgType
)

//---------- Var -----------

var DefaultOrgInChannelA = OrgIdType("ORG1")

//---------- Interface -----------

type Sender interface {
	Send(*SignedGossipMessage, ...*RemotePeer)
}

type MembershipService interface {
	GetMembership() []Peer
}

//---------- Struct -----------

type OrgCryptoService struct{}

type JoinChanMsg struct {
	Members2AnchorPeers map[string][]AnchorPeer
}

type GossipChannel struct {
	*GossipAdapter
	sync.RWMutex
	shouldGossipStateInfo     int32
	mcs                       *NaiveMessageCryptoService
	pkiID                     PKIidType
	selfOrg                   OrgIdType
	stopChan                  chan struct{}
	stateInfoMsg              *SignedGossipMessage
	orgs                      []OrgIdType
	joinMsg                   *JoinChanMsg
	blockMsgStore             *MessageStore
	stateInfoMsgStore         *stateInfoCache
	leaderMsgStore            *MessageStore
	chainID                   ChainID
	blocksPuller              *PullMediator
	stateInfoPublishScheduler *time.Ticker
	stateInfoRequestScheduler *time.Ticker
	memFilter                 *membershipFilter
	ledgerHeight              uint64
	incTime                   uint64
	leftChannel               int32
}

type GossipAdapter struct {
	*GossipService
	*GossipDiscovery
}

type ChannelState struct {
	Stopping      int32
	sync.RWMutex
	Channels      map[string]*GossipChannel
	GossipService *GossipService
}

type ChannelDeMultiplexer struct {
	channels []*channel
	lock     *sync.RWMutex
	closed   bool
}

// 消息被添加到BatchingEmitter中，它们会定期转发T次
type BatchingEmitter struct {
	iterations int                 // 每条消息转发的次数T
	burstSize  int                 // 存储最大消息数
	delay      time.Duration       // 连续消息推送之间的最大延迟
	cb         func([]interface{}) // 进行转发的回调
	lock       *sync.Mutex
	buff       []*batchedMessage
	stopFlag   int32
}

type EmittedGossipMessage struct {
	*SignedGossipMessage
	Filter func(id PKIidType) bool
}

type PullMediator struct {
	sync.RWMutex
	*PullAdapter
	msgType2Hook map[MsgType][]messageHook
	config       PullMediatorConfig
	itemID2Msg   map[string]*SignedGossipMessage
	engine       *pullEngine
}

type PullMediatorConfig struct {
	ID                string
	PullInterval      time.Duration
	Channel           ChainID
	PeerCountToSelect int
	Tag               GossipMessage_Tag
	MsgType           PullMsgType
}

type PullAdapter struct {
	Sender           Sender
	MemSvc           MembershipService
	IdExtractor      func(*SignedGossipMessage) string
	MsgConsumer      func(message *SignedGossipMessage)
	EgressDigFilter  func(helloMsg *ReceivedMessage) func(digestItem string) bool
	IngressDigFilter func(digestMsg *DataDigest) *DataDigest
}

type ChannelStoreConfig struct {
	ID                          string
	PublishStateInfoInterval    time.Duration
	MaxBlockCountToStore        int
	PullPeerNum                 int
	PullInterval                time.Duration
	RequestStateInfoInterval    time.Duration
	BlockExpirationInterval     time.Duration
	StateInfoCacheSweepInterval time.Duration
}

type AnchorPeer struct {
	Host string
	Port int
}

type SecurityAdvisor struct {
	mock.Mock
}

type channel struct {
	pred MessageAcceptor
	ch   chan interface{}
}

type batchedMessage struct {
	data           interface{}
	iterationsLeft int
}

type pullEngine struct {
	*PullMediator
	stopFlag           int32
	state              *Set
	item2owners        map[string][]string
	peers2nonces       map[string]uint64
	nonces2peers       map[uint64]string
	acceptingDigests   int32
	acceptingResponses int32
	lock               sync.Mutex
	outgoingNONCES     *Set
	incomingNONCES     *Set
	digestFilter       func(context interface{}) func(digestItem string) bool

	digestWaitTime   time.Duration
	requestWaitTime  time.Duration
	responseWaitTime time.Duration
}

type membershipFilter struct {
	adapter *GossipAdapter
	*GossipChannel
}

type stateInfoCache struct {
	verify   func(msg *SignedGossipMessage, orgs ...OrgIdType) bool
	*MembershipStore
	*MessageStore
	stopChan chan struct{}
}

//---------- Define -----------

type MsgType int

type messageHook func([]string, []*SignedGossipMessage, *ReceivedMessage)

//---------- Struct Methods -----------

func (*OrgCryptoService) OrgByPeerId(id PeerIdType) OrgIdType {
	return DefaultOrgInChannelA
}

func (jcm *JoinChanMsg) SequenceNumber() uint64 {
	return uint64(time.Now().UnixNano())
}

func (jcm *JoinChanMsg) Peers() []OrgIdType {
	if jcm.Members2AnchorPeers == nil {
		return []OrgIdType{DefaultOrgInChannelA}
	}
	peers := make([]OrgIdType, len(jcm.Members2AnchorPeers))
	i := 0
	for org := range jcm.Members2AnchorPeers {
		peers[i] = OrgIdType(org)
		i++
	}
	return peers
}

// 返回指定组织中的AnchorPeer
func (jcm *JoinChanMsg) AnchorPeersOf(org OrgIdType) []AnchorPeer {
	if jcm.Members2AnchorPeers == nil {
		return []AnchorPeer{}
	}
	return jcm.Members2AnchorPeers[string(org)]
}

func (cs *ChannelState) LookupChannelForMsg(msg *ReceivedMessage) *GossipChannel {
	if msg.SignedGossipMessage.GetStateInfoPullReq() != nil {
		sipr := msg.SignedGossipMessage.GetStateInfoPullReq()
		mac := sipr.Channel_MAC
		pkiID := msg.ConnInfo.ID
		return cs.getGossipChannelByMAC(mac, pkiID)
	}
	return cs.LookupChannelForGossipMsg(msg.SignedGossipMessage.GossipMessage)
}

func (cs *ChannelState) LookupChannelForGossipMsg(msg *GossipMessage) *GossipChannel {
	if msg.GetStateInfo() == nil {
		return cs.GetGossipChannelByChainID(msg.Channel)
	}
	stateInfMsg := msg.GetStateInfo()
	return cs.getGossipChannelByMAC(stateInfMsg.Channel_MAC, stateInfMsg.PkiId)
}

func (cs *ChannelState) GetGossipChannelByChainID(chainID ChainID) *GossipChannel {
	if cs.isStopping() {
		return nil
	}
	cs.RLock()
	defer cs.RUnlock()
	return cs.Channels[string(chainID)]
}

func (cs *ChannelState) JoinChannel(joinMsg *JoinChanMsg, chainID ChainID) {
	if cs.isStopping() {
		return
	}
	cs.Lock()
	defer cs.Unlock()
	gc, exists := cs.Channels[string(chainID)]
	if !exists {
		pkiID := cs.GossipService.Comm.PKIID
		ga := &GossipAdapter{GossipService: cs.GossipService, GossipDiscovery: cs.GossipService.Discovery}
		gc := newGossipChannel(pkiID, cs.GossipService.SelfOrg, cs.GossipService.Mcs, chainID, ga, joinMsg)
		cs.Channels[string(chainID)] = gc
		return
	}
	gc.ConfigureChannel(joinMsg)
}

func (cs *ChannelState) Stop() {
	if cs.isStopping() {
		return
	}
	atomic.StoreInt32(&cs.Stopping, int32(1))
	cs.Lock()
	defer cs.Unlock()
	for _, gc := range cs.Channels {
		gc.Stop()
	}
}

func (cs *ChannelState) getGossipChannelByMAC(receivedMAC []byte, pkiID PKIidType) *GossipChannel {
	cs.RLock()
	defer cs.RUnlock()
	for chanName, gc := range cs.Channels {
		mac := GenerateMAC(pkiID, ChainID(chanName))
		if bytes.Equal(mac, receivedMAC) {
			return gc
		}
	}
	return nil
}

func (cs *ChannelState) isStopping() bool {
	return atomic.LoadInt32(&cs.Stopping) == int32(1)
}

func (m *ChannelDeMultiplexer) Close() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.closed = true
	for _, ch := range m.channels {
		close(ch.ch)
	}
	m.channels = nil
}

func (m *ChannelDeMultiplexer) AddChannel(predicate MessageAcceptor) chan interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()
	ch := &channel{ch: make(chan interface{}, 10), pred: predicate}
	m.channels = append(m.channels, ch)
	return ch.ch
}

func (m *ChannelDeMultiplexer) DeMultiplex(msg interface{}) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.closed {
		return
	}
	for _, ch := range m.channels {
		if ch.pred(msg) {
			ch.ch <- msg
		}
	}
}

func (m *SecurityAdvisor) OrgByPeerId(a0 PeerIdType) OrgIdType {
	ret := m.Called(a0)
	var r0 OrgIdType
	if rf, ok := ret.Get(0).(func(PeerIdType) OrgIdType); ok {
		r0 = rf(a0)
		return r0
	}
	if ret.Get(0) != nil {
		r0 = ret.Get(0).(OrgIdType)
	}
	return r0
}

// 获取发布数据的Peer列表
func (gc *GossipChannel) GetPeers() []Peer {
	var peers []Peer
	if gc.hasLeftChannel() {
		return peers
	}
	for _, peer := range gc.GetMembership() {
		if !gc.EligibleForChannel(peer) {
			continue
		}
		stateInf := gc.stateInfoMsgStore.GetMsgByID(peer.PKIid)
		if stateInf == nil {
			continue
		}
		props := stateInf.GetStateInfo().Properties
		if props != nil && props.LeftChannel {
			continue
		}
		peer.Properties = stateInf.GetStateInfo().Properties
		peer.Envelope = stateInf.Envelope
		peers = append(peers, peer)
	}
	return peers
}

// 检查指定Peer是否有资格进入Channel
func (gc *GossipChannel) IsMemberInChan(peer Peer) bool {
	org := gc.GetOrgOfPeer(peer.PKIid)
	if org == nil {
		return false
	}
	return gc.IsOrgInChannel(org)
}

// 更新Peer发布给Channel中其他Peer的Ledger高度
func (gc *GossipChannel) UpdateLedgerHeight(height uint64) {
	gc.Lock()
	defer gc.Unlock()
	var chaincodes []*Chaincode
	var leftChannel bool
	if prevMsg := gc.stateInfoMsg; prevMsg != nil {
		leftChannel = prevMsg.GetStateInfo().Properties.LeftChannel
		chaincodes = prevMsg.GetStateInfo().Properties.Chaincodes
	}
	gc.updateProperties(height, chaincodes, leftChannel)
}

// 检查指定组织是否存在Channel中
func (gc *GossipChannel) IsOrgInChannel(membersOrg OrgIdType) bool {
	gc.RLock()
	defer gc.RUnlock()
	for _, orgOfChan := range gc.orgs {
		if bytes.Equal(orgOfChan, membersOrg) {
			return true
		}
	}
	return false
}

// 检查指定Peer是否符合条件获取Channel的块
func (gc *GossipChannel) EligibleForChannel(peer Peer) bool {
	peerId := gc.GetIdByPKIID(peer.PKIid)
	if len(peerId) == 0 {
		return false
	}
	msg := gc.stateInfoMsgStore.GetMsgByID(peer.PKIid)
	if msg == nil {
		return false
	}
	return true
}

// 处理远程Peer发送的消息
func (gc *GossipChannel) HandleMessage(msg *ReceivedMessage) {
	if !gc.verifyMsg(msg) {
		return
	}
	m := msg.SignedGossipMessage
	if !m.IsChannelRestricted() {
		return
	}
	orgID := gc.GetOrgOfPeer(msg.ConnInfo.ID)
	if len(orgID) == 0 {
		return
	}
	if !gc.IsOrgInChannel(orgID) {
		return
	}
	if m.GetStateInfoPullReq() != nil {
		msg.Respond(gc.createStateInfoSnapshot(orgID))
		return
	}
	if m.GetStateSnapshot() != nil {
		gc.handleStateInfoSnapshot(m.GossipMessage, msg.ConnInfo.ID)
		return
	}
	if m.GetDataMsg() != nil || m.GetStateInfo() != nil {
		added := false
		if m.GetDataMsg() != nil {
			if m.GetDataMsg().Payload == nil {
				return
			}
			if !gc.blockMsgStore.CheckValid(msg.SignedGossipMessage) {
				return
			}
			if !gc.verifyBlock(m.GossipMessage, msg.ConnInfo.ID) {
				return
			}
			added = gc.blockMsgStore.Add(msg.SignedGossipMessage)
		} else {
			added = gc.stateInfoMsgStore.add(msg.SignedGossipMessage)
		}
		if added {
			gc.Forward(msg)
			gc.DeMultiplex(m)
			if m.GetDataMsg() != nil {
				gc.blocksPuller.Add(msg.SignedGossipMessage)
			}
		}
		return
	}
	if m.IsPullMsg() && m.GetPullMsgType() == PullMsgType_BLOCK_MSG {
		if gc.hasLeftChannel() {
			return
		}
		if gc.stateInfoMsgStore.GetMsgByID(msg.ConnInfo.ID) == nil {
			return
		}
		if !gc.eligibleForChannelAndSameOrg(Peer{PKIid: msg.ConnInfo.ID}) {
			return
		}
		if m.GetDataUpdate() != nil {
			var filteredEnvelopes []*Envelope
			for _, item := range m.GetDataUpdate().Data {
				gMsg, err := item.ToGossipMessage()
				if err != nil {
					return
				}
				if !bytes.Equal(gMsg.Channel, []byte(gc.chainID)) {
					return
				}
				if !gc.blockMsgStore.CheckValid(msg.SignedGossipMessage) {
					return
				}
				if !gc.verifyBlock(gMsg.GossipMessage, msg.ConnInfo.ID) {
					return
				}
				added := gc.blockMsgStore.Add(gMsg)
				if !added {
					continue
				}
				filteredEnvelopes = append(filteredEnvelopes, item)
			}
			m.GetDataUpdate().Data = filteredEnvelopes
		}
		gc.blocksPuller.HandleMessage(msg)
	}
	if m.GetLeadershipMsg() != nil && gc.leaderMsgStore.Add(m) {
		gc.DeMultiplex(m)
	}
}

func (gc *GossipChannel) AddToMsgStore(msg *SignedGossipMessage) {
	if msg.GetDataMsg() != nil {
		gc.blockMsgStore.Add(msg)
		gc.blocksPuller.Add(msg)
	}
	if msg.GetStateInfo() != nil {
		gc.stateInfoMsgStore.add(msg)
	}
}

func (gc *GossipChannel) ConfigureChannel(joinMsg *JoinChanMsg) {
	gc.Lock()
	defer gc.Unlock()
	if len(joinMsg.Peers()) == 0 {
		return
	}
	if gc.joinMsg == nil {
		gc.joinMsg = joinMsg
	}
	if gc.joinMsg.SequenceNumber() > (joinMsg.SequenceNumber()) {
		return
	}
	gc.orgs = joinMsg.Peers()
	gc.joinMsg = joinMsg
	cache := gc.stateInfoMsgStore
	for _, m := range cache.Get() {
		msg := m.(*SignedGossipMessage)
		if !cache.verify(msg, joinMsg.Peers()...) {
			cache.Purge(func(o interface{}) bool {
				pkiID := o.(*SignedGossipMessage).GetStateInfo().PkiId
				return bytes.Equal(pkiID, msg.GetStateInfo().PkiId)
			})
			cache.Remove(msg.GetStateInfo().PkiId)
		}
	}
}

func (gc *GossipChannel) LeaveChannel() {
	gc.Lock()
	defer gc.Unlock()
	atomic.StoreInt32(&gc.leftChannel, 1)
	var chaincodes []*Chaincode
	var height uint64
	if prevMsg := gc.stateInfoMsg; prevMsg != nil {
		chaincodes = prevMsg.GetStateInfo().Properties.Chaincodes
		height = prevMsg.GetStateInfo().Properties.LedgerHeight
	}
	gc.updateProperties(height, chaincodes, true)
}

func (gc *GossipChannel) Stop() {
	gc.stopChan <- struct{}{}
	gc.blocksPuller.Stop()
	gc.stateInfoPublishScheduler.Stop()
	gc.stateInfoRequestScheduler.Stop()
	gc.leaderMsgStore.Stop()
	gc.stateInfoMsgStore.Stop()
	gc.blockMsgStore.Stop()
}

func (gc *GossipChannel) requestStateInfo() {
	req, err := gc.createStateInfoRequest()
	if err != nil {
		return
	}
	endpoints := SelectPeers(gc.GetConf().PullPeerNum, gc.GetMembership(), gc.IsMemberInChan)
	gc.Send(req, endpoints...)
}

func (gc *GossipChannel) publishStateInfo() {
	if atomic.LoadInt32(&gc.shouldGossipStateInfo) == int32(0) {
		return
	}
	gc.RLock()
	stateInfoMsg := gc.stateInfoMsg
	gc.RUnlock()
	gc.Gossip(stateInfoMsg)
	if len(gc.GetMembership()) > 0 {
		atomic.StoreInt32(&gc.shouldGossipStateInfo, int32(0))
	}
}

func (gc *GossipChannel) createBlockPuller() *PullMediator {
	conf := PullMediatorConfig{
		MsgType:           PullMsgType_BLOCK_MSG,
		Channel:           []byte(gc.chainID),
		ID:                gc.GetConf().ID,
		PeerCountToSelect: gc.GetConf().PullPeerNum,
		PullInterval:      gc.GetConf().PullInterval,
		Tag:               GossipMessage_CHAN_AND_ORG,
	}
	seqNumFromMsg := func(msg *SignedGossipMessage) string {
		dataMsg := msg.GetDataMsg()
		if dataMsg == nil || dataMsg.Payload == nil {
			return ""
		}
		return fmt.Sprintf("%d", dataMsg.Payload.SeqNum)
	}
	adapter := &PullAdapter{
		Sender:      gc,
		MemSvc:      gc.memFilter,
		IdExtractor: seqNumFromMsg,
		MsgConsumer: func(msg *SignedGossipMessage) {
			gc.DeMultiplex(msg)
		},
	}
	adapter.IngressDigFilter = func(digestMsg *DataDigest) *DataDigest {
		gc.RLock()
		height := gc.ledgerHeight
		gc.RUnlock()
		digests := digestMsg.Digests
		digestMsg.Digests = nil
		for i := range digests {
			seqNum, err := strconv.ParseUint(string(digests[i]), 10, 64)
			if err != nil {
				continue
			}
			if seqNum >= height {
				digestMsg.Digests = append(digestMsg.Digests, digests[i])
			}

		}
		return digestMsg
	}
	return NewPullMediator(conf, adapter)
}

func (gc *GossipChannel) createStateInfoRequest() (*SignedGossipMessage, error) {
	return (&GossipMessage{
		Tag:   GossipMessage_CHAN_OR_ORG,
		Nonce: 0,
		Content: &GossipMessage_StateInfoPullReq{
			StateInfoPullReq: &StateInfoPullRequest{
				Channel_MAC: GenerateMAC(gc.pkiID, gc.chainID),
			},
		},
	}).NoopSign()
}

func (gc *GossipChannel) createStateInfoSnapshot(requestersOrg OrgIdType) *GossipMessage {
	sameOrg := bytes.Equal(gc.selfOrg, requestersOrg)
	rawElements := gc.stateInfoMsgStore.Get()
	var elements []*Envelope
	for _, rawEl := range rawElements {
		msg := rawEl.(*SignedGossipMessage)
		orgOfCurrentMsg := gc.GetOrgOfPeer(msg.GetStateInfo().PkiId)
		if sameOrg || !bytes.Equal(orgOfCurrentMsg, gc.selfOrg) {
			elements = append(elements, msg.Envelope)
			continue
		}
		if netMember := gc.Lookup(msg.GetStateInfo().PkiId); netMember == nil || netMember.Endpoint == "" {
			continue
		}
		elements = append(elements, msg.Envelope)
	}
	return &GossipMessage{
		Channel: gc.chainID,
		Tag:     GossipMessage_CHAN_OR_ORG,
		Nonce:   0,
		Content: &GossipMessage_StateSnapshot{
			StateSnapshot: &StateInfoSnapshot{
				Elements: elements,
			},
		},
	}
}

func (gc *GossipChannel) handleStateInfoSnapshot(m *GossipMessage, sender PKIidType) {
	for _, envelope := range m.GetStateSnapshot().Elements {
		stateInf, err := envelope.ToGossipMessage()
		if err != nil {
			return
		}
		if stateInf.GetStateInfo() == nil {
			return
		}
		si := stateInf.GetStateInfo()
		orgID := gc.GetOrgOfPeer(si.PkiId)
		if orgID == nil {
			return
		}
		if !gc.IsOrgInChannel(orgID) {
			return
		}
		expectedMAC := GenerateMAC(si.PkiId, gc.chainID)
		if !bytes.Equal(si.Channel_MAC, expectedMAC) {
			return
		}
		if err = gc.GossipService.ValidateStateInfoMsg(stateInf); err != nil {
			return
		}
		if gc.Lookup(si.PkiId) == nil {
			continue
		}
		gc.stateInfoMsgStore.add(stateInf)
	}
}

func (gc *GossipChannel) updateProperties(ledgerHeight uint64, chaincodes []*Chaincode, leftChannel bool) {
	stateInfMsg := &StateInfo{
		Channel_MAC: GenerateMAC(gc.pkiID, gc.chainID),
		PkiId:       gc.pkiID,
		Timestamp: &PeerTime{
			IncNum: gc.incTime,
			SeqNum: uint64(time.Now().UnixNano()),
		},
		Properties: &Properties{
			LeftChannel:  leftChannel,
			LedgerHeight: ledgerHeight,
			Chaincodes:   chaincodes,
		},
	}
	m := &GossipMessage{
		Nonce: 0,
		Tag:   GossipMessage_CHAN_OR_ORG,
		Content: &GossipMessage_StateInfo{
			StateInfo: stateInfMsg,
		},
	}
	msg, err := gc.Sign(m)
	if err != nil {
		return
	}
	gc.stateInfoMsgStore.add(msg)
	gc.ledgerHeight = msg.GetStateInfo().Properties.LedgerHeight
	gc.stateInfoMsg = msg
	atomic.StoreInt32(&gc.shouldGossipStateInfo, int32(1))
}

func (gc *GossipChannel) hasLeftChannel() bool {
	return atomic.LoadInt32(&gc.leftChannel) == 1
}

func (gc *GossipChannel) periodicalInvocation(fn func(), c <-chan time.Time) {
	for {
		select {
		case <-c:
			fn()
		case <-gc.stopChan:
			gc.stopChan <- struct{}{}
			return
		}
	}
}

func (gc *GossipChannel) eligibleForChannelAndSameOrg(peer Peer) bool {
	sameOrg := func(networkMember Peer) bool {
		return bytes.Equal(gc.GetOrgOfPeer(networkMember.PKIid), gc.selfOrg)
	}
	return CombineRoutingFilters(gc.EligibleForChannel, sameOrg)(peer)
}

func (gc *GossipChannel) verifyMsg(msg *ReceivedMessage) bool {
	if msg == nil {
		return false
	}
	m := msg.SignedGossipMessage
	if m == nil {
		return false
	}
	if msg.ConnInfo.ID == nil {
		return false
	}
	if m.GetStateInfo() != nil {
		si := m.GetStateInfo()
		expectedMAC := GenerateMAC(si.PkiId, gc.chainID)
		if !bytes.Equal(expectedMAC, si.Channel_MAC) {
			return false
		}
		return true
	}
	if m.GetStateInfoPullReq() != nil {
		sipr := m.GetStateInfoPullReq()
		expectedMAC := GenerateMAC(msg.ConnInfo.ID, gc.chainID)
		if !bytes.Equal(expectedMAC, sipr.Channel_MAC) {
			return false
		}
		return true
	}
	if !bytes.Equal(m.Channel, []byte(gc.chainID)) {
		return false
	}
	return true
}

func (gc *GossipChannel) verifyBlock(msg *GossipMessage, sender PKIidType) bool {
	if msg.GetDataMsg() == nil {
		return false
	}
	payload := msg.GetDataMsg().Payload
	if payload == nil {
		return false
	}
	seqNum := payload.SeqNum
	rawBlock := payload.Data
	if err := gc.mcs.VerifyBlock(msg.Channel, seqNum, rawBlock); err != nil {
		return false
	}
	return true
}

func (ga *GossipAdapter) Sign(msg *GossipMessage) (*SignedGossipMessage, error) {
	signer := func(msg []byte) ([]byte, error) {
		return ga.Mcs.Sign(msg)
	}
	sMsg := &SignedGossipMessage{
		GossipMessage: msg,
	}
	e, err := sMsg.Sign(signer)
	if err != nil {
		return nil, err
	}
	return &SignedGossipMessage{
		Envelope:      e,
		GossipMessage: msg,
	}, nil
}

func (ga *GossipAdapter) GetConf() ChannelStoreConfig {
	return ChannelStoreConfig{
		ID:                          ga.Conf.ID,
		MaxBlockCountToStore:        ga.Conf.MaxBlockCountToStore,
		PublishStateInfoInterval:    ga.Conf.PublishStateInfoInterval,
		PullInterval:                ga.Conf.PullInterval,
		PullPeerNum:                 ga.Conf.PullPeerNum,
		RequestStateInfoInterval:    ga.Conf.RequestStateInfoInterval,
		BlockExpirationInterval:     ga.Conf.PullInterval * 100,
		StateInfoCacheSweepInterval: ga.Conf.PullInterval * 5,
	}
}

func (ga *GossipAdapter) Gossip(msg *SignedGossipMessage) {
	ga.GossipService.Emitter.Add(&EmittedGossipMessage{
		SignedGossipMessage: msg,
		Filter: func(_ PKIidType) bool {
			return true
		},
	})
}

func (ga *GossipAdapter) Send(msg *SignedGossipMessage, peers ...*RemotePeer) {
	ga.GossipService.Comm.Send(msg, peers...)
}

func (ga *GossipAdapter) Forward(msg *ReceivedMessage) {
	ga.GossipService.Emitter.Add(&EmittedGossipMessage{
		SignedGossipMessage: msg.SignedGossipMessage,
		Filter:              msg.ConnInfo.ID.IsNotSameFilter,
	})
}

func (ga *GossipAdapter) GetOrgOfPeer(PKIID PKIidType) OrgIdType {
	return ga.GossipService.GetOrgOfPeer(PKIID)
}

func (ga *GossipAdapter) GetIdByPKIID(pkiID PKIidType) PeerIdType {
	id, err := ga.IdMapper.Get(pkiID)
	if err != nil {
		return nil
	}
	return id
}

// 添加要批处理的消息
func (p *BatchingEmitter) Add(message interface{}) {
	if p.iterations == 0 {
		return
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	p.buff = append(p.buff, &batchedMessage{data: message, iterationsLeft: p.iterations})
	if len(p.buff) >= p.burstSize {
		p.emit()
	}
}

// 要发出的待处理消息量
func (p *BatchingEmitter) Size() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	return len(p.buff)
}

func (p *BatchingEmitter) Stop() {
	atomic.StoreInt32(&(p.stopFlag), int32(1))
}

func (p *BatchingEmitter) periodicEmit() {
	for !p.toDie() {
		time.Sleep(p.delay)
		p.lock.Lock()
		p.emit()
		p.lock.Unlock()
	}
}

func (p *BatchingEmitter) emit() {
	if p.toDie() {
		return
	}
	if len(p.buff) == 0 {
		return
	}
	msgs2beEmitted := make([]interface{}, len(p.buff))
	for i, v := range p.buff {
		msgs2beEmitted[i] = v.data
	}
	p.cb(msgs2beEmitted)
	n := len(p.buff)
	for i := 0; i < n; i++ {
		msg := p.buff[i]
		msg.iterationsLeft--
		if msg.iterationsLeft == 0 {
			p.buff = append(p.buff[:i], p.buff[i+1:]...)
			n--
			i--
		}
	}
}

func (p *BatchingEmitter) toDie() bool {
	return atomic.LoadInt32(&(p.stopFlag)) == int32(1)
}

func (p *PullMediator) RegisterMsgHook(pullMsgType MsgType, hook messageHook) {
	p.Lock()
	defer p.Unlock()
	p.msgType2Hook[pullMsgType] = append(p.msgType2Hook[pullMsgType], hook)
}

func (p *PullMediator) Add(msg *SignedGossipMessage) {
	p.Lock()
	defer p.Unlock()
	itemID := p.IdExtractor(msg)
	p.itemID2Msg[itemID] = msg
	p.engine.state.Add(itemID)
}

// 处理远程Peer的消息
func (p *PullMediator) HandleMessage(m *ReceivedMessage) {
	if m.SignedGossipMessage == nil || !m.SignedGossipMessage.IsPullMsg() {
		return
	}
	msg := m.SignedGossipMessage
	msgType := msg.GetPullMsgType()
	if msgType != p.config.MsgType {
		return
	}
	var itemIDs []string
	var items []*SignedGossipMessage
	var pullMsgType MsgType
	if helloMsg := msg.GetHello(); helloMsg != nil {
		pullMsgType = HelloMsgType
		p.engine.onHello(helloMsg.Nonce, m)
	}
	if digest := msg.GetDataDig(); digest != nil {
		d := p.PullAdapter.IngressDigFilter(digest)
		itemIDs = BytesToStrings(d.Digests)
		pullMsgType = DigestMsgType
		p.engine.onDigest(itemIDs, d.Nonce, m)
	}
	if req := msg.GetDataReq(); req != nil {
		itemIDs = BytesToStrings(req.Digests)
		pullMsgType = RequestMsgType
		p.engine.onReq(itemIDs, req.Nonce, m)
	}
	if res := msg.GetDataUpdate(); res != nil {
		itemIDs = make([]string, len(res.Data))
		items = make([]*SignedGossipMessage, len(res.Data))
		pullMsgType = ResponseMsgType
		for i, pulledMsg := range res.Data {
			msg, err := pulledMsg.ToGossipMessage()
			if err != nil {
				return
			}
			p.MsgConsumer(msg)
			itemIDs[i] = p.IdExtractor(msg)
			items[i] = msg
			p.Lock()
			p.itemID2Msg[itemIDs[i]] = msg
			p.Unlock()
		}
		p.engine.onRes(itemIDs, res.Nonce)
	}
	for _, h := range p.hooksByMsgType(pullMsgType) {
		h(itemIDs, items, m)
	}
}

func (p *PullMediator) Remove(digest string) {
	p.Lock()
	defer p.Unlock()
	delete(p.itemID2Msg, digest)
	for _, seq := range digest {
		p.engine.state.Remove(seq)
	}
}

func (p *PullMediator) Stop() {
	p.engine.Stop()
}

func (p *PullMediator) SelectPeers() []string {
	k := p.config.PeerCountToSelect
	peerPool := p.MemSvc.GetMembership()
	if len(peerPool) < k {
		k = len(peerPool)
	}
	indices := GetRandomIndices(k, len(peerPool)-1)
	remotePeers := make([]*RemotePeer, len(indices))
	for i, j := range indices {
		remotePeers[i] = &RemotePeer{Endpoint: peerPool[j].PreferredEndpoint(), PKIID: peerPool[j].PKIid}
	}
	endpoints := make([]string, len(remotePeers))
	for i, peer := range remotePeers {
		endpoints[i] = peer.Endpoint
	}
	return endpoints
}

func (p *PullMediator) Hello(dest string, nonce uint64) {
	helloMsg := &GossipMessage{
		Channel: p.config.Channel,
		Tag:     p.config.Tag,
		Content: &GossipMessage_Hello{
			Hello: &GossipHello{
				Nonce:    nonce,
				Metadata: nil,
				MsgType:  p.config.MsgType,
			},
		},
	}
	sMsg, err := helloMsg.NoopSign()
	if err != nil {
		return
	}
	p.Sender.Send(sMsg, p.peersWithEndpoints(dest)...)
}

// 将摘要发送到远程PullEngine
func (p *PullMediator) SendDigest(digest []string, nonce uint64, context interface{}) {
	digMsg := &GossipMessage{
		Channel: p.config.Channel,
		Tag:     p.config.Tag,
		Nonce:   0,
		Content: &GossipMessage_DataDig{
			DataDig: &DataDigest{
				MsgType: p.config.MsgType,
				Nonce:   nonce,
				Digests: StringsToBytes(digest),
			},
		},
	}
	context.(*ReceivedMessage).Respond(digMsg)
}

func (p *PullMediator) SendReq(dest string, items []string, nonce uint64) {
	req := &GossipMessage{
		Channel: p.config.Channel,
		Tag:     p.config.Tag,
		Nonce:   0,
		Content: &GossipMessage_DataReq{
			DataReq: &DataRequest{
				MsgType: p.config.MsgType,
				Nonce:   nonce,
				Digests: StringsToBytes(items),
			},
		},
	}
	sMsg, err := req.NoopSign()
	if err != nil {
		return
	}
	p.Sender.Send(sMsg, p.peersWithEndpoints(dest)...)
}

func (p *PullMediator) SendRes(items []string, context interface{}, nonce uint64) {
	var items2return []*Envelope
	p.RLock()
	defer p.RUnlock()
	for _, item := range items {
		if msg, exists := p.itemID2Msg[item]; exists {
			items2return = append(items2return, msg.Envelope)
		}
	}
	returnedUpdate := &GossipMessage{
		Channel: p.config.Channel,
		Tag:     p.config.Tag,
		Nonce:   0,
		Content: &GossipMessage_DataUpdate{
			DataUpdate: &DataUpdate{
				MsgType: p.config.MsgType,
				Nonce:   nonce,
				Data:    items2return,
			},
		},
	}
	context.(*ReceivedMessage).Respond(returnedUpdate)
}

func (p *PullMediator) peersWithEndpoints(endpoints ...string) []*RemotePeer {
	var peers []*RemotePeer
	for _, peer := range p.MemSvc.GetMembership() {
		for _, endpoint := range endpoints {
			if peer.PreferredEndpoint() == endpoint {
				peers = append(peers, &RemotePeer{Endpoint: peer.PreferredEndpoint(), PKIID: peer.PKIid})
			}
		}
	}
	return peers
}

func (p *PullMediator) hooksByMsgType(msgType MsgType) []messageHook {
	p.RLock()
	defer p.RUnlock()
	var returnedHooks []messageHook
	for _, h := range p.msgType2Hook[msgType] {
		returnedHooks = append(returnedHooks, h)
	}
	return returnedHooks
}

func (engine *pullEngine) Stop() {
	atomic.StoreInt32(&(engine.stopFlag), int32(1))
}

func (engine *pullEngine) onDigest(digest []string, nonce uint64, context interface{}) {
	if !engine.isAcceptingDigests() || !engine.outgoingNONCES.Exists(nonce) {
		return
	}
	engine.lock.Lock()
	defer engine.lock.Unlock()
	for _, n := range digest {
		if engine.state.Exists(n) {
			continue
		}
		if _, exists := engine.item2owners[n]; !exists {
			engine.item2owners[n] = make([]string, 0)
		}
		engine.item2owners[n] = append(engine.item2owners[n], engine.nonces2peers[nonce])
	}
}

func (engine *pullEngine) onHello(nonce uint64, context interface{}) {
	engine.incomingNONCES.Add(nonce)
	time.AfterFunc(engine.requestWaitTime, func() {
		engine.incomingNONCES.Remove(nonce)
	})
	a := engine.state.ToArray()
	var digest []string
	filter := engine.digestFilter(context)
	for _, item := range a {
		dig := item.(string)
		if !filter(dig) {
			continue
		}
		digest = append(digest, dig)
	}
	if len(digest) == 0 {
		return
	}
	engine.SendDigest(digest, nonce, context)
}

func (engine *pullEngine) onReq(items []string, nonce uint64, context interface{}) {
	if !engine.incomingNONCES.Exists(nonce) {
		return
	}
	engine.lock.Lock()
	defer engine.lock.Unlock()
	filter := engine.digestFilter(context)
	var items2Send []string
	for _, item := range items {
		if engine.state.Exists(item) && filter(item) {
			items2Send = append(items2Send, item)
		}
	}
	if len(items2Send) == 0 {
		return
	}
	go engine.SendRes(items2Send, context, nonce)
}

func (engine *pullEngine) onRes(items []string, nonce uint64) {
	if !engine.outgoingNONCES.Exists(nonce) || !engine.isAcceptingResponses() {
		return
	}
	for _, seq := range items {
		engine.state.Add(seq)
	}
}

func (engine *pullEngine) newNONCE() uint64 {
	n := uint64(0)
	for {
		n = RandomUInt64()
		if !engine.outgoingNONCES.Exists(n) {
			return n
		}
	}
}

func (engine *pullEngine) isAcceptingDigests() bool {
	return atomic.LoadInt32(&(engine.acceptingDigests)) == int32(1)
}

func (engine *pullEngine) isAcceptingResponses() bool {
	return atomic.LoadInt32(&(engine.acceptingResponses)) == int32(1)
}

func (engine *pullEngine) initiatePull() {
	engine.lock.Lock()
	defer engine.lock.Unlock()
	atomic.StoreInt32(&(engine.acceptingDigests), int32(1))
	for _, peer := range engine.SelectPeers() {
		nonce := engine.newNONCE()
		engine.outgoingNONCES.Add(nonce)
		engine.nonces2peers[nonce] = peer
		engine.peers2nonces[peer] = nonce
		engine.Hello(peer, nonce)
	}
	time.AfterFunc(engine.digestWaitTime, func() {
		atomic.StoreInt32(&(engine.acceptingDigests), int32(0))
		engine.lock.Lock()
		defer engine.lock.Unlock()
		requestMapping := make(map[string][]string)
		for n, sources := range engine.item2owners {
			source := sources[RandomInt(len(sources))]
			if _, exists := requestMapping[source]; !exists {
				requestMapping[source] = make([]string, 0)
			}
			requestMapping[source] = append(requestMapping[source], n)
		}
		atomic.StoreInt32(&(engine.acceptingResponses), int32(1))
		for dest, seqsToReq := range requestMapping {
			engine.SendReq(dest, seqsToReq, engine.peers2nonces[dest])
		}
		time.AfterFunc(engine.responseWaitTime, engine.endPull)
	})
}

func (engine *pullEngine) endPull() {
	engine.lock.Lock()
	defer engine.lock.Unlock()
	atomic.StoreInt32(&(engine.acceptingResponses), int32(0))
	engine.outgoingNONCES.Clear()
	engine.item2owners = make(map[string][]string)
	engine.peers2nonces = make(map[string]uint64)
	engine.nonces2peers = make(map[uint64]string)
}

func (engine *pullEngine) toDie() bool {
	return atomic.LoadInt32(&(engine.stopFlag)) == int32(1)
}

func (mf *membershipFilter) GetMembership() []Peer {
	if mf.hasLeftChannel() {
		return nil
	}
	var peers []Peer
	for _, mem := range mf.adapter.GetMembership() {
		if mf.eligibleForChannelAndSameOrg(mem) {
			peers = append(peers, mem)
		}
	}
	return peers
}

func (cache *stateInfoCache) add(msg *SignedGossipMessage) bool {
	if !cache.MessageStore.CheckValid(msg) {
		return false
	}
	if !cache.verify(msg) {
		return false
	}
	added := cache.MessageStore.Add(msg)
	if added {
		pkiID := msg.GetStateInfo().PkiId
		cache.MembershipStore.Put(pkiID, msg)
	}
	return added
}

func (cache *stateInfoCache) Stop() {
	cache.stopChan <- struct{}{}
}

//---------- Public Methods -----------

func NewChannelDemultiplexer() *ChannelDeMultiplexer {
	return &ChannelDeMultiplexer{
		channels: make([]*channel, 0),
		lock:     &sync.RWMutex{},
	}
}

func NewPullMediator(config PullMediatorConfig, adapter *PullAdapter) *PullMediator {
	egressDigFilter := adapter.EgressDigFilter
	acceptAllFilter := func(_ *ReceivedMessage) func(string) bool {
		return func(string) bool {
			return true
		}
	}
	if egressDigFilter == nil {
		egressDigFilter = acceptAllFilter
	}
	p := &PullMediator{
		PullAdapter:  adapter,
		msgType2Hook: make(map[MsgType][]messageHook),
		config:       config,
		itemID2Msg:   make(map[string]*SignedGossipMessage),
	}
	engine := &pullEngine{
		PullMediator:       p,
		stopFlag:           int32(0),
		state:              NewSet(),
		item2owners:        make(map[string][]string),
		peers2nonces:       make(map[string]uint64),
		nonces2peers:       make(map[uint64]string),
		acceptingDigests:   int32(0),
		acceptingResponses: int32(0),
		incomingNONCES:     NewSet(),
		outgoingNONCES:     NewSet(),
		digestWaitTime:     GetDurationOrDefault("peer.gossip.digestWaitTime", defDigestWaitTime),
		requestWaitTime:    GetDurationOrDefault("peer.gossip.requestWaitTime", defRequestWaitTime),
		responseWaitTime:   GetDurationOrDefault("peer.gossip.responseWaitTime", defResponseWaitTime),
		digestFilter: func(context interface{}) func(digestItem string) bool {
			return func(digestItem string) bool {
				return egressDigFilter(context.(*ReceivedMessage))(digestItem)
			}
		},
	}
	go func() {
		for !engine.toDie() {
			time.Sleep(config.PullInterval)
			if engine.toDie() {
				return
			}
			engine.initiatePull()
		}
	}()
	p.engine = engine
	if adapter.IngressDigFilter == nil {
		adapter.IngressDigFilter = func(digestMsg *DataDigest) *DataDigest {
			return digestMsg
		}
	}
	return p
}

//---------- Private Methods -----------

func newGossipChannel(pkiID PKIidType, org OrgIdType, mcs *NaiveMessageCryptoService,
	chainID ChainID, adapter *GossipAdapter, joinMsg *JoinChanMsg) *GossipChannel {
	gc := &GossipChannel{
		incTime:                   uint64(time.Now().UnixNano()),
		selfOrg:                   org,
		pkiID:                     pkiID,
		mcs:                       mcs,
		GossipAdapter:             adapter,
		stopChan:                  make(chan struct{}, 1),
		shouldGossipStateInfo:     int32(0),
		stateInfoPublishScheduler: time.NewTicker(adapter.GetConf().PublishStateInfoInterval),
		stateInfoRequestScheduler: time.NewTicker(adapter.GetConf().RequestStateInfoInterval),
		orgs:                      []OrgIdType{},
		chainID:                   chainID,
	}
	gc.memFilter = &membershipFilter{adapter: gc.GossipAdapter, GossipChannel: gc}
	comparator := NewGossipMessageComparator(adapter.GetConf().MaxBlockCountToStore)
	gc.blocksPuller = gc.createBlockPuller()
	seqNumFromMsg := func(m interface{}) string {
		return fmt.Sprintf("%d", m.(*SignedGossipMessage).GetDataMsg().Payload.SeqNum)
	}
	gc.blockMsgStore = NewMessageStoreExpirable(comparator, func(m interface{}) {
		gc.blocksPuller.Remove(seqNumFromMsg(m))
	}, gc.GetConf().BlockExpirationInterval, nil, nil, func(m interface{}) {
		gc.blocksPuller.Remove(seqNumFromMsg(m))
	})
	hashPeerExpiredInMembership := func(o interface{}) bool {
		pkiID := o.(*SignedGossipMessage).GetStateInfo().PkiId
		return gc.Lookup(pkiID) == nil
	}
	verifyStateInfoMsg := func(msg *SignedGossipMessage, orgs ...OrgIdType) bool {
		si := msg.GetStateInfo()
		if bytes.Equal(gc.pkiID, si.PkiId) {
			return true
		}
		peerId := adapter.GetIdByPKIID(si.PkiId)
		if len(peerId) == 0 {
			return false
		}
		isOrgInChan := func(org OrgIdType) bool {
			if len(orgs) == 0 {
				if !gc.IsOrgInChannel(org) {
					return false
				}
				return true
			}
			found := false
			for _, chanMember := range orgs {
				if bytes.Equal(chanMember, org) {
					found = true
					break
				}
			}
			return found
		}
		org := gc.GetOrgOfPeer(si.PkiId)
		if !isOrgInChan(org) {
			return false
		}
		if err := gc.mcs.VerifyByChannel(chainID, peerId, msg.Signature, msg.Payload); err != nil {
			return false
		}
		return true
	}
	membershipStore := NewMembershipStore()
	p := NewGossipMessageComparator(0)
	s := &stateInfoCache{
		verify:          verifyStateInfoMsg,
		MembershipStore: membershipStore,
		stopChan:        make(chan struct{}),
	}
	invalidationTrigger := func(m interface{}) {
		pkiID := m.(*SignedGossipMessage).GetStateInfo().PkiId
		membershipStore.Remove(pkiID)
	}
	s.MessageStore = NewMessageStore(p, invalidationTrigger)
	go func() {
		for {
			select {
			case <-s.stopChan:
				return
			case <-time.After(gc.GetConf().StateInfoCacheSweepInterval):
				s.Purge(hashPeerExpiredInMembership)
			}
		}
	}()
	gc.stateInfoMsgStore = s
	ttl := GetDurationOrDefault("peer.gossip.election.leaderAliveThreshold", time.Second*10) * 10
	policy := NewGossipMessageComparator(0)
	gc.leaderMsgStore = NewMessageStoreExpirable(policy, Noop, ttl, nil, nil, nil)
	gc.ConfigureChannel(joinMsg)
	//定期发布，请求状态信息
	go gc.periodicalInvocation(gc.publishStateInfo, gc.stateInfoPublishScheduler.C)
	go gc.periodicalInvocation(gc.requestStateInfo, gc.stateInfoRequestScheduler.C)
	return gc
}
