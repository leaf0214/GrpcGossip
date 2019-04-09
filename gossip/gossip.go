package gossip

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

//---------- Const -----------

const (
	presumedDeadChanSize = 100
	acceptChanSize       = 100
)

//---------- Interface -----------

type Gossip interface {
	Peers() []Peer                                                                  // 获取存活的Peer列表
	PeersOfChannel(ChainID) []Peer                                                  // 获取存活并且订阅指定Channel的Peer列表
	SuspectPeers(PeerSuspector)                                                     // 使Gossip实例验证可疑Peer是否存活
	JoinChan(*JoinChanMsg, ChainID)                                                 // 使Gossip实例加入到Channel
	Gossip(*GossipMessage)                                                          // 向网络中其他Peer发送消息
	Send(*GossipMessage, ...*RemotePeer)                                            // 向指定远程Peer发送消息
	SendByCriteria(*SignedGossipMessage, SendCriteria) error                        // 将指定消息发送给与指定发送条件匹配的所有Peer
	Receive(MessageAcceptor, bool) (<-chan *GossipMessage, <-chan *ReceivedMessage) // 获取接收回复消息的只读Channel
	LeaveChan(ChainID)                                                              // 使Gossip实例离开Channel
	Stop()                                                                          // 停止Gossip组件
	UpdateLedgerHeight(height uint64, chainID ChainID)                              // 更新P2P发布的高度
}

//---------- Struct -----------

type GossipService struct {
	SelfId            PeerIdType
	IncludeIdPeriod   time.Time
	CertStore         *certStore
	IdMapper          *IdMapper
	PresumedDead      chan PKIidType
	Discovery         *GossipDiscovery
	Comm              *Communicate
	SelfOrg           OrgIdType
	*ChannelDeMultiplexer
	StopSignal        *sync.WaitGroup
	Conf              *GossipConfig
	ToDieChan         chan struct{}
	StopFlag          int32
	Emitter           *BatchingEmitter
	DiscAdapter       *DiscoveryAdapter
	SecAdvisor        *OrgCryptoService
	ChanState         *ChannelState
	DisSecAdap        *DiscoverySecurityAdapter
	Mcs               *NaiveMessageCryptoService
	StateInfoMsgStore *MessageStore
	CertPuller        *PullMediator
}

type GossipConfig struct {
	BindPort                   int           // 绑定端口
	ID                         string        // 实例ID
	BootstrapPeers             []string      // 启动时连接的Peer
	PropagateIterations        int           // 消息被推送到远程Peer的次数
	PropagatePeerNum           int           // 选择将消息推送的Peer数量
	MaxBlockCountToStore       int           // 存储最大块数
	MaxPropagationBurstSize    int           // 存储最大消息数
	MaxPropagationBurstLatency time.Duration // 连续消息推送之间的最大延迟
	PullInterval               time.Duration // Pull间隔
	PullPeerNum                int           // PullPeer数量
	PublishCertPeriod          time.Duration // 发布证书周期
	PublishStateInfoInterval   time.Duration // 发布状态信息间隔
	RequestStateInfoInterval   time.Duration // 请求状态信息间隔
	InternalEndpoint           string        // 发布给组织中Peer的端点
	ExternalEndpoint           string        // Peer将此端点而不是SelfEndpoint发布给外组织
}

type SendCriteria struct {
	Timeout    time.Duration // 等待确认的时间
	MinAck     int           // 确认的Peer数量
	MaxPeers   int           // 发送消息的最大Peer数
	IsEligible RoutingFilter // 指定Peer是否有资格接收该消息
	Channel    ChainID       // 指定发送此消息的Channel，只有加入该Channel的Peer才会收到此消息
}

type certStore struct {
	selfId   PeerIdType
	idMapper *IdMapper
	puller   *PullMediator
	mcs      *NaiveMessageCryptoService
}

//---------- Define -----------

type channelRoutingFilterFactory func(*GossipChannel) RoutingFilter

//---------- Struct Methods -----------

func (g *GossipService) Gossip(msg *GossipMessage) {
	if err := msg.IsTagLegal(); err != nil {
		panic(errors.WithStack(err))
	}
	sMsg := &SignedGossipMessage{
		GossipMessage: msg,
	}
	var err error
	if sMsg.GetDataMsg() != nil {
		sMsg, err = sMsg.NoopSign()
	} else {
		_, err = sMsg.Sign(func(msg []byte) ([]byte, error) {
			return g.Mcs.Sign(msg)
		})
	}
	if err != nil {
		return
	}
	if msg.IsChannelRestricted() {
		gc := g.ChanState.GetGossipChannelByChainID(msg.Channel)
		if gc == nil {
			return
		}
		if msg.GetDataMsg() != nil {
			gc.AddToMsgStore(sMsg)
		}
	}
	if g.Conf.PropagateIterations == 0 {
		return
	}
	g.Emitter.Add(&EmittedGossipMessage{
		SignedGossipMessage: sMsg,
		Filter: func(_ PKIidType) bool {
			return true
		},
	})
}

func (g *GossipService) Send(msg *GossipMessage, peers ...*RemotePeer) {
	m, err := msg.NoopSign()
	if err != nil {
		return
	}
	g.Comm.Send(m, peers...)
}

func (g *GossipService) SendByCriteria(msg *SignedGossipMessage, criteria SendCriteria) error {
	if criteria.MaxPeers == 0 {
		return nil
	}
	if criteria.Timeout == 0 {
		return errors.New("Timeout should be specified")
	}
	if criteria.IsEligible == nil {
		criteria.IsEligible = SelectAllPolicy
	}
	membership := g.Discovery.GetMembership()
	if len(criteria.Channel) > 0 {
		gc := g.ChanState.GetGossipChannelByChainID(criteria.Channel)
		if gc == nil {
			return fmt.Errorf("requested to Send for channel %s, but no such channel exists", string(criteria.Channel))
		}
		membership = gc.GetPeers()
	}
	peers2send := SelectPeers(criteria.MaxPeers, membership, criteria.IsEligible)
	if len(peers2send) < criteria.MinAck {
		return fmt.Errorf("requested to send to at least %d peers, but know only of %d suitable peers", criteria.MinAck, len(peers2send))
	}
	results := g.Comm.SendWithAck(msg, criteria.Timeout, criteria.MinAck, peers2send...)
	for _, res := range results {
		if res.Error() == "" {
			continue
		}
	}
	if results.AckCount() < criteria.MinAck {
		return errors.New(results.String())
	}
	return nil
}

func (g *GossipService) Receive(acceptor MessageAcceptor, passThrough bool) (<-chan *GossipMessage, <-chan *ReceivedMessage) {
	if passThrough {
		return nil, g.Comm.Receive(acceptor)
	}
	acceptByType := func(o interface{}) bool {
		if o, isGossipMsg := o.(*GossipMessage); isGossipMsg {
			return acceptor(o)
		}
		if o, isSignedMsg := o.(*SignedGossipMessage); isSignedMsg {
			sMsg := o
			return acceptor(sMsg.GossipMessage)
		}
		return false
	}
	inCh := g.AddChannel(acceptByType)
	outCh := make(chan *GossipMessage, acceptChanSize)
	go func() {
		for {
			select {
			case s := <-g.ToDieChan:
				g.ToDieChan <- s
				return
			case m := <-inCh:
				if m == nil {
					return
				}
				outCh <- m.(*SignedGossipMessage).GossipMessage
			}
		}
	}()
	return outCh, nil
}

func (g *GossipService) Peers() []Peer {
	return g.Discovery.GetMembership()
}

func (g *GossipService) PeersOfChannel(channel ChainID) []Peer {
	gc := g.ChanState.GetGossipChannelByChainID(channel)
	if gc == nil {
		return nil
	}
	return gc.GetPeers()
}

func (g *GossipService) SuspectPeers(isSuspected PeerSuspector) {
	g.CertStore.idMapper.SuspectPeers(isSuspected)
}

func (g *GossipService) UpdateLedgerHeight(height uint64, chainID ChainID) {
	gc := g.ChanState.GetGossipChannelByChainID(chainID)
	if gc == nil {
		return
	}
	gc.UpdateLedgerHeight(height)
}

func (g *GossipService) JoinChan(joinMsg *JoinChanMsg, chainID ChainID) {
	g.ChanState.JoinChannel(joinMsg, chainID)
	for _, org := range joinMsg.Peers() {
		g.learnAnchorPeers(string(chainID), org, joinMsg.AnchorPeersOf(org))
	}
}

func (g *GossipService) LeaveChan(chainID ChainID) {
	gc := g.ChanState.GetGossipChannelByChainID(chainID)
	if gc == nil {
		return
	}
	gc.LeaveChannel()
}

func (g *GossipService) Stop() {
	if g.toDie() {
		return
	}
	atomic.StoreInt32(&g.StopFlag, int32(1))
	g.ChanState.Stop()
	atomic.StoreInt32(&g.DiscAdapter.Stopping, int32(1))
	close(g.DiscAdapter.IncChan)
	g.Discovery.Stop()
	g.CertStore.puller.Stop()
	g.CertStore.idMapper.Stop()
	g.ToDieChan <- struct{}{}
	g.Emitter.Stop()
	g.ChannelDeMultiplexer.Close()
	g.StateInfoMsgStore.Stop()
	g.StopSignal.Wait()
	g.Comm.Stop()
}

func (g *GossipService) GetOrgOfPeer(PKIID PKIidType) OrgIdType {
	cert, err := g.IdMapper.Get(PKIID)
	if err != nil {
		return nil
	}
	return g.SecAdvisor.OrgByPeerId(cert)
}

func (g *GossipService) ValidateStateInfoMsg(msg *SignedGossipMessage) error {
	verifier := func(id []byte, signature, message []byte) error {
		pkiID := g.IdMapper.GetPKIidOfCert(PeerIdType(id))
		if pkiID == nil {
			return errors.New("PKI-ID not found in id mapper")
		}
		return g.IdMapper.Verify(pkiID, signature, message)
	}
	ids, err := g.IdMapper.Get(msg.GetStateInfo().PkiId)
	if err != nil {
		return errors.WithStack(err)
	}
	return msg.Verify(ids, verifier)
}

func (g *GossipService) start() {
	go g.syncDiscovery()
	go g.handlePresumedDead()
	msgSelector := func(msg interface{}) bool {
		gMsg, isGossipMsg := msg.(*ReceivedMessage)
		if !isGossipMsg {
			return false
		}
		isConn := gMsg.SignedGossipMessage.GetConn() != nil
		isEmpty := gMsg.SignedGossipMessage.GetEmpty() != nil
		isPrivateData := gMsg.SignedGossipMessage.IsPrivateDataMsg()
		return !(isConn || isEmpty || isPrivateData)
	}
	incMsgs := g.Comm.Receive(msgSelector)
	go g.acceptMessages(incMsgs)
}

func (g *GossipService) syncDiscovery() {
	for !g.toDie() {
		g.Discovery.InitiateSync(g.Conf.PullPeerNum)
		time.Sleep(g.Conf.PullInterval)
	}
}

func (g *GossipService) handleMessage(m *ReceivedMessage) {
	if g.toDie() || m == nil || m.SignedGossipMessage == nil {
		return
	}
	msg := m.SignedGossipMessage
	if !g.validateMsg(m) {
		return
	}
	if msg.IsChannelRestricted() {
		gc := g.ChanState.LookupChannelForMsg(m)
		if gc == nil {
			if g.isInMyorg(Peer{PKIid: m.ConnInfo.ID}) && msg.GetStateInfo() != nil && g.StateInfoMsgStore.Add(msg) {
				g.Emitter.Add(&EmittedGossipMessage{
					SignedGossipMessage: msg,
					Filter:              m.ConnInfo.ID.IsNotSameFilter,
				})
			}
			return
		}
		if m.SignedGossipMessage.GetLeadershipMsg() != nil {
			if err := g.validateLeadershipMessage(m.SignedGossipMessage); err != nil {
				return
			}
		}
		gc.HandleMessage(m)
		return
	}
	if selectOnlyDiscoveryMessages(m) {
		if m.SignedGossipMessage.GetMemReq() != nil {
			sMsg, err := m.SignedGossipMessage.GetMemReq().SelfInformation.ToGossipMessage()
			if err != nil {
				return
			}
			if sMsg.GetAliveMsg() == nil {
				return
			}
			if !bytes.Equal(sMsg.GetAliveMsg().Membership.PkiId, m.ConnInfo.ID) {
				return
			}
		}
		g.forwardDiscoveryMsg(m)
	}
	if msg.IsPullMsg() && msg.GetPullMsgType() == PullMsgType_IDENTITY_MSG {
		if update := m.SignedGossipMessage.GetDataUpdate(); update != nil {
			for _, env := range update.Data {
				m, err := env.ToGossipMessage()
				if err != nil {
					return
				}
				if m.GetPeerIdentity() == nil {
					return
				}
				if err := g.CertStore.validateIdMsg(m); err != nil {
					return
				}
			}
		}
		g.CertStore.puller.HandleMessage(m)
	}
}

func (g *GossipService) handlePresumedDead() {
	defer g.StopSignal.Done()
	for {
		select {
		case s := <-g.ToDieChan:
			g.ToDieChan <- s
			return
		case deadEndpoint := <-g.Comm.PresumedDead():
			g.PresumedDead <- deadEndpoint
		}
	}
}

func (g *GossipService) acceptMessages(incMsgs <-chan *ReceivedMessage) {
	defer g.StopSignal.Done()
	for {
		select {
		case s := <-g.ToDieChan:
			g.ToDieChan <- s
			return
		case msg := <-incMsgs:
			g.handleMessage(msg)
		}
	}
}

func (g *GossipService) learnAnchorPeers(channel string, orgOfAnchorPeers OrgIdType, anchorPeers []AnchorPeer) {
	if len(anchorPeers) == 0 {
		return
	}
	for _, ap := range anchorPeers {
		if ap.Host == "" {
			continue
		}
		if ap.Port == 0 {
			continue
		}
		endpoint := fmt.Sprintf("%s:%d", ap.Host, ap.Port)
		if g.selfNetworkMember().Endpoint == endpoint || g.selfNetworkMember().InternalEndpoint == endpoint {
			continue
		}
		inOurOrg := bytes.Equal(g.SelfOrg, orgOfAnchorPeers)
		if !inOurOrg && g.selfNetworkMember().Endpoint == "" {
			continue
		}
		id := func() (*PeerId, error) {
			remotePeerId, err := g.Comm.Handshake(&RemotePeer{Endpoint: endpoint})
			if err != nil {
				err = errors.WithStack(err)
				return nil, err
			}
			isAnchorPeerInMyOrg := bytes.Equal(g.SelfOrg, g.SecAdvisor.OrgByPeerId(remotePeerId))
			if bytes.Equal(orgOfAnchorPeers, g.SelfOrg) && !isAnchorPeerInMyOrg {
				err := errors.Errorf("Anchor peer %s isn't in our org, but is claimed to be", endpoint)
				return nil, err
			}
			pkiID := g.Mcs.GetPKIidOfCert(remotePeerId)
			if len(pkiID) == 0 {
				return nil, errors.Errorf("Wasn't able to extract PKI-ID of remote peer with id of %v", remotePeerId)
			}
			return &PeerId{
				ID:      pkiID,
				SelfOrg: isAnchorPeerInMyOrg,
			}, nil
		}
		g.Discovery.Connect(Peer{InternalEndpoint: endpoint, Endpoint: endpoint}, id)
	}
}

func (g *GossipService) selfNetworkMember() Peer {
	self := Peer{
		Endpoint:         g.Conf.ExternalEndpoint,
		PKIid:            g.Comm.PKIID,
		Metadata:         []byte{},
		InternalEndpoint: g.Conf.InternalEndpoint,
	}
	if g.Discovery != nil {
		self.Metadata = g.Discovery.Self().Metadata
	}
	return self
}

func (g *GossipService) toDie() bool {
	return atomic.LoadInt32(&g.StopFlag) == int32(1)
}

func (g *GossipService) forwardDiscoveryMsg(msg *ReceivedMessage) {
	defer func() {
		recover()
	}()
	g.DiscAdapter.IncChan <- msg
}

func (g *GossipService) validateMsg(msg *ReceivedMessage) bool {
	if err := msg.SignedGossipMessage.IsTagLegal(); err != nil {
		return false
	}
	if msg.SignedGossipMessage.GetAliveMsg() != nil && !g.DisSecAdap.ValidateAliveMsg(msg.SignedGossipMessage) {
		return false
	}
	if msg.SignedGossipMessage.GetStateInfo() != nil {
		if err := g.ValidateStateInfoMsg(msg.SignedGossipMessage); err != nil {
			return false
		}
	}
	return true
}

func (g *GossipService) sendGossipBatch(a []interface{}) {
	msgs := make([]*EmittedGossipMessage, len(a))
	for i, e := range a {
		msgs[i] = e.(*EmittedGossipMessage)
	}
	if g.Discovery == nil {
		return
	}
	var blocks []*EmittedGossipMessage
	var stateInfoMsgs []*EmittedGossipMessage
	var orgMsgs []*EmittedGossipMessage
	var leadershipMsgs []*EmittedGossipMessage
	isABlock := func(o interface{}) bool {
		return o.(*EmittedGossipMessage).GetDataMsg() != nil
	}
	isAStateInfoMsg := func(o interface{}) bool {
		return o.(*EmittedGossipMessage).GetStateInfo() != nil
	}
	aliveMsgsWithNoEndpointAndInOurOrg := func(o interface{}) bool {
		msg := o.(*EmittedGossipMessage)
		if msg.GetAliveMsg() == nil {
			return false
		}
		peer := msg.GetAliveMsg().Membership
		return peer.Endpoint == "" && g.isInMyorg(Peer{PKIid: peer.PkiId})
	}
	isOrgRestricted := func(o interface{}) bool {
		return aliveMsgsWithNoEndpointAndInOurOrg(o) || o.(*EmittedGossipMessage).IsOrgRestricted()
	}
	isLeadershipMsg := func(o interface{}) bool {
		return o.(*EmittedGossipMessage).GetLeadershipMsg() != nil
	}
	blocks, msgs = partitionMessages(isABlock, msgs)
	g.gossipInChan(blocks, func(gc *GossipChannel) RoutingFilter {
		return CombineRoutingFilters(gc.EligibleForChannel, gc.IsMemberInChan, g.isInMyorg)
	})
	leadershipMsgs, msgs = partitionMessages(isLeadershipMsg, msgs)
	g.gossipInChan(leadershipMsgs, func(gc *GossipChannel) RoutingFilter {
		return CombineRoutingFilters(gc.EligibleForChannel, gc.IsMemberInChan, g.isInMyorg)
	})
	stateInfoMsgs, msgs = partitionMessages(isAStateInfoMsg, msgs)
	for _, stateInfMsg := range stateInfoMsgs {
		peerSelector := g.isInMyorg
		gc := g.ChanState.LookupChannelForGossipMsg(stateInfMsg.GossipMessage)
		if gc != nil && g.hasExternalEndpoint(stateInfMsg.GossipMessage.GetStateInfo().PkiId) {
			peerSelector = gc.IsMemberInChan
		}
		peerSelector = CombineRoutingFilters(peerSelector, func(peer Peer) bool {
			return stateInfMsg.Filter(peer.PKIid)
		})
		peers2Send := SelectPeers(g.Conf.PropagatePeerNum, g.Discovery.GetMembership(), peerSelector)
		g.Comm.Send(stateInfMsg.SignedGossipMessage, peers2Send...)
	}
	orgMsgs, msgs = partitionMessages(isOrgRestricted, msgs)
	peers2Send := SelectPeers(g.Conf.PropagatePeerNum, g.Discovery.GetMembership(), g.isInMyorg)
	for _, msg := range orgMsgs {
		g.Comm.Send(msg.SignedGossipMessage, g.removeSelfLoop(msg, peers2Send)...)
	}
	for _, msg := range msgs {
		if msg.GetAliveMsg() == nil {
			continue
		}
		selectByOriginOrg := g.peersByOriginOrgPolicy(Peer{PKIid: msg.GetAliveMsg().Membership.PkiId})
		selector := CombineRoutingFilters(selectByOriginOrg, func(peer Peer) bool {
			return msg.Filter(peer.PKIid)
		})
		peers2Send := SelectPeers(g.Conf.PropagatePeerNum, g.Discovery.GetMembership(), selector)
		for _, peer := range peers2Send {
			aliveMsgFromDiffOrg := msg.SignedGossipMessage.GetAliveMsg() != nil && !g.isInMyorg(Peer{PKIid: msg.SignedGossipMessage.GetAliveMsg().Membership.PkiId})
			if aliveMsgFromDiffOrg && !g.hasExternalEndpoint(peer.PKIID) {
				continue
			}
			if !g.isInMyorg(Peer{PKIid: peer.PKIID}) {
				msg.SignedGossipMessage.Envelope.SecretEnvelope = nil
			}
			g.Comm.Send(msg.SignedGossipMessage, peer)
		}
	}
}

func (g *GossipService) gossipInChan(messages []*EmittedGossipMessage, chanRoutingFactory channelRoutingFilterFactory) {
	if len(messages) == 0 {
		return
	}
	var totalChannels []ChainID
	for _, m := range messages {
		if len(m.Channel) == 0 {
			continue
		}
		sameChan := func(a interface{}, b interface{}) bool {
			return bytes.Equal(a.(ChainID), b.(ChainID))
		}
		if IndexInSlice(totalChannels, ChainID(m.Channel), sameChan) == -1 {
			totalChannels = append(totalChannels, ChainID(m.Channel))
		}
	}
	var channels ChainID
	var messagesOfChannel []*EmittedGossipMessage
	for len(totalChannels) > 0 {
		channels, totalChannels = totalChannels[0], totalChannels[1:]
		grabMsgs := func(o interface{}) bool {
			return bytes.Equal(o.(*EmittedGossipMessage).Channel, channels)
		}
		messagesOfChannel, messages = partitionMessages(grabMsgs, messages)
		if len(messagesOfChannel) == 0 {
			continue
		}
		gc := g.ChanState.GetGossipChannelByChainID(channels)
		if gc == nil {
			continue
		}
		membership := g.Discovery.GetMembership()
		var peers2Send []*RemotePeer
		if messagesOfChannel[0].GetLeadershipMsg() != nil {
			peers2Send = SelectPeers(len(membership), membership, chanRoutingFactory(gc))
		} else {
			peers2Send = SelectPeers(g.Conf.PropagatePeerNum, membership, chanRoutingFactory(gc))
		}
		for _, msg := range messagesOfChannel {
			filteredPeers := g.removeSelfLoop(msg, peers2Send)
			g.Comm.Send(msg.SignedGossipMessage, filteredPeers...)
		}
	}
}

func (g *GossipService) removeSelfLoop(msg *EmittedGossipMessage, peers []*RemotePeer) []*RemotePeer {
	var result []*RemotePeer
	for _, peer := range peers {
		if msg.Filter(peer.PKIID) {
			result = append(result, peer)
		}
	}
	return result
}

func (g *GossipService) createCertStorePuller() *PullMediator {
	conf := PullMediatorConfig{
		MsgType:           PullMsgType_IDENTITY_MSG,
		Channel:           []byte(""),
		ID:                g.Conf.InternalEndpoint,
		PeerCountToSelect: g.Conf.PullPeerNum,
		PullInterval:      g.Conf.PullInterval,
		Tag:               GossipMessage_EMPTY,
	}
	pkiIDFromMsg := func(msg *SignedGossipMessage) string {
		idMsg := msg.GetPeerIdentity()
		if idMsg == nil || idMsg.PkiId == nil {
			return ""
		}
		return fmt.Sprintf("%s", string(idMsg.PkiId))
	}
	certConsumer := func(msg *SignedGossipMessage) {
		idMsg := msg.GetPeerIdentity()
		if idMsg == nil || idMsg.Cert == nil || idMsg.PkiId == nil {
			return
		}
		g.IdMapper.Put(PKIidType(idMsg.PkiId), PeerIdType(idMsg.Cert))
	}
	adapter := &PullAdapter{
		Sender:          g.Comm,
		MemSvc:          g.Discovery,
		IdExtractor:     pkiIDFromMsg,
		MsgConsumer:     certConsumer,
		EgressDigFilter: g.sameOrgOrOurOrgPullFilter,
	}
	return NewPullMediator(conf, adapter)
}

func (g *GossipService) sameOrgOrOurOrgPullFilter(msg *ReceivedMessage) func(string) bool {
	peersOrg := g.SecAdvisor.OrgByPeerId(msg.ConnInfo.Id)
	if len(peersOrg) == 0 {
		return func(string) bool {
			return false
		}
	}
	if bytes.Equal(g.SelfOrg, peersOrg) {
		return func(string) bool {
			return true
		}
	}
	return func(item string) bool {
		pkiID := PKIidType(item)
		msgsOrg := g.GetOrgOfPeer(pkiID)
		if len(msgsOrg) == 0 {
			return false
		}
		if !g.hasExternalEndpoint(pkiID) {
			return false
		}
		return bytes.Equal(msgsOrg, g.SelfOrg) || bytes.Equal(msgsOrg, peersOrg)
	}
}

func (g *GossipService) connect2BootstrapPeers() {
	for _, endpoint := range g.Conf.BootstrapPeers {
		endpoint := endpoint
		id := func() (*PeerId, error) {
			remotePeerId, err := g.Comm.Handshake(&RemotePeer{Endpoint: endpoint})
			if err != nil {
				return nil, errors.WithStack(err)
			}
			sameOrg := bytes.Equal(g.SelfOrg, g.SecAdvisor.OrgByPeerId(remotePeerId))
			if !sameOrg {
				return nil, errors.Errorf("%s isn't in our organization, cannot be a bootstrap peer", endpoint)
			}
			pkiID := g.Mcs.GetPKIidOfCert(remotePeerId)
			if len(pkiID) == 0 {
				return nil, errors.Errorf("Wasn't able to extract PKI-ID of remote peer with id of %v", remotePeerId)
			}
			return &PeerId{ID: pkiID, SelfOrg: sameOrg}, nil
		}
		g.Discovery.Connect(Peer{
			InternalEndpoint: endpoint,
			Endpoint:         endpoint,
		}, id)
	}

}

func (g *GossipService) hasExternalEndpoint(PKIID PKIidType) bool {
	if nm := g.Discovery.Lookup(PKIID); nm != nil {
		return nm.Endpoint != ""
	}
	return false
}

func (g *GossipService) isInMyorg(peer Peer) bool {
	if peer.PKIid == nil {
		return false
	}
	if org := g.GetOrgOfPeer(peer.PKIid); org != nil {
		return bytes.Equal(g.SelfOrg, org)
	}
	return false
}

func (g *GossipService) validateLeadershipMessage(msg *SignedGossipMessage) error {
	pkiID := msg.GetLeadershipMsg().PkiId
	if len(pkiID) == 0 {
		return errors.New("Empty PKI-ID")
	}
	ids, err := g.IdMapper.Get(pkiID)
	if err != nil {
		return errors.Wrap(err, "Unable to fetch PKI-ID from id-mapper")
	}
	return msg.Verify(ids, func(_ []byte, signature, message []byte) error {
		return g.Mcs.Verify(ids, signature, message)
	})
}

func (g *GossipService) disclosurePolicy(remotePeer *Peer) (Sieve, EnvelopeFilter) {
	remotePeerOrg := g.GetOrgOfPeer(remotePeer.PKIid)
	if len(remotePeerOrg) == 0 {
		return func(*SignedGossipMessage) bool { return false }, func(msg *SignedGossipMessage) *Envelope { return msg.Envelope }
	}
	return func(msg *SignedGossipMessage) bool {
		org := g.GetOrgOfPeer(msg.GetAliveMsg().Membership.PkiId)
		if len(org) == 0 {
			return false
		}
		fromSameForeignOrg := bytes.Equal(remotePeerOrg, org)
		fromMyOrg := bytes.Equal(g.SelfOrg, org)
		if !(fromSameForeignOrg || fromMyOrg) {
			return false
		}
		return bytes.Equal(org, remotePeerOrg) || msg.GetAliveMsg().Membership.Endpoint != "" && remotePeer.Endpoint != ""
	}, func(msg *SignedGossipMessage) *Envelope {
		if !bytes.Equal(g.SelfOrg, remotePeerOrg) {
			msg.SecretEnvelope = nil
		}
		return msg.Envelope
	}
}

func (g *GossipService) peersByOriginOrgPolicy(peer Peer) RoutingFilter {
	peersOrg := g.GetOrgOfPeer(peer.PKIid)
	if len(peersOrg) == 0 {
		return SelectNonePolicy
	}
	if bytes.Equal(g.SelfOrg, peersOrg) {
		return SelectAllPolicy
	}
	return func(peer Peer) bool {
		memberOrg := g.GetOrgOfPeer(peer.PKIid)
		if len(memberOrg) == 0 {
			return false
		}
		isFromMyOrg := bytes.Equal(g.SelfOrg, memberOrg)
		return isFromMyOrg || bytes.Equal(memberOrg, peersOrg)
	}
}

func (g *GossipService) newDiscoveryAdapter() *DiscoveryAdapter {
	return &DiscoveryAdapter{
		Comm:     g.Comm,
		Stopping: int32(0),
		GossipFunc: func(msg *SignedGossipMessage) {
			if g.Conf.PropagateIterations == 0 {
				return
			}
			g.Emitter.Add(&EmittedGossipMessage{
				SignedGossipMessage: msg,
				Filter: func(_ PKIidType) bool {
					return true
				},
			})
		},
		ForwardFunc: func(message *ReceivedMessage) {
			if g.Conf.PropagateIterations == 0 {
				return
			}
			g.Emitter.Add(&EmittedGossipMessage{
				SignedGossipMessage: message.SignedGossipMessage,
				Filter:              message.ConnInfo.ID.IsNotSameFilter,
			})
		},
		IncChan:          make(chan *ReceivedMessage),
		PresumedDeads:    g.PresumedDead,
		DisclosurePolicy: g.disclosurePolicy,
	}
}

func (cs *certStore) validateIdMsg(msg *SignedGossipMessage) error {
	idMsg := msg.GetPeerIdentity()
	if idMsg == nil {
		return errors.Errorf("Identity empty:%+v", msg)
	}
	pkiID := idMsg.PkiId
	cert := idMsg.Cert
	calculatedPKIID := cs.mcs.GetPKIidOfCert(PeerIdType(cert))
	claimedPKIID := PKIidType(pkiID)
	if !bytes.Equal(calculatedPKIID, claimedPKIID) {
		return errors.Errorf("Calculated pkiID doesn't match id: calculated:%v, claimedPKI-ID:%v", calculatedPKIID, claimedPKIID)
	}
	verifier := func(peerId []byte, signature, message []byte) error {
		return cs.mcs.Verify(PeerIdType(peerId), signature, message)
	}
	if err := msg.Verify(cert, verifier); err != nil {
		return errors.Wrap(err, "Failed verifying message")
	}
	return cs.mcs.ValidateIdType(PeerIdType(idMsg.Cert))
}

//---------- Public Methods -----------

func DefaultGossipInstance(portPrefix int, id int, maxMsgCount int, mcs *NaiveMessageCryptoService, boot ...int) Gossip {
	port := id + portPrefix
	conf := &GossipConfig{
		BindPort:                   port,
		BootstrapPeers:             BootPeers(portPrefix, boot...),
		ID:                         fmt.Sprintf("p%d", id),
		MaxBlockCountToStore:       maxMsgCount,
		MaxPropagationBurstLatency: time.Duration(500) * time.Millisecond,
		MaxPropagationBurstSize:    20,
		PropagateIterations:        1,
		PropagatePeerNum:           3,
		PullInterval:               time.Duration(4) * time.Second,
		PullPeerNum:                5,
		InternalEndpoint:           fmt.Sprintf("localhost:%d", port),
		ExternalEndpoint:           fmt.Sprintf("1.2.3.4:%d", port),
		PublishCertPeriod:          time.Duration(4) * time.Second,
		PublishStateInfoInterval:   time.Duration(1) * time.Second,
		RequestStateInfoInterval:   time.Duration(1) * time.Second,
	}
	selfID := PeerIdType(conf.InternalEndpoint)
	return NewGossipServiceWithServer(conf, &OrgCryptoService{}, mcs, selfID, nil, nil)
}

func NewGossipServiceWithServer(conf *GossipConfig, secAdvisor *OrgCryptoService, mcs *NaiveMessageCryptoService,
	id PeerIdType, pem []byte, key []byte) Gossip {
	var err error
	g := &GossipService{
		SelfOrg:              secAdvisor.OrgByPeerId(id),
		SecAdvisor:           secAdvisor,
		SelfId:               id,
		PresumedDead:         make(chan PKIidType, presumedDeadChanSize),
		Discovery:            nil,
		Mcs:                  mcs,
		Conf:                 conf,
		ChannelDeMultiplexer: NewChannelDemultiplexer(),
		ToDieChan:            make(chan struct{}, 1),
		StopFlag:             int32(0),
		StopSignal:           &sync.WaitGroup{},
		IncludeIdPeriod:      time.Now().Add(conf.PublishCertPeriod),
	}
	policy := NewGossipMessageComparator(0)
	g.StateInfoMsgStore = NewMessageStoreExpirable(policy, Noop, g.Conf.PublishStateInfoInterval*100, nil, nil, Noop)
	g.IdMapper = NewIdMapper(mcs, id, func(pkiID PKIidType, _ PeerIdType) {
		g.Comm.CloseConn(&RemotePeer{PKIID: pkiID})
		g.CertPuller.Remove(string(pkiID))
	})
	g.Comm, err = NewCommInstanceWithServer(conf.BindPort, g.IdMapper, id, pem, key, secAdvisor)
	if err != nil {
		return nil
	}
	g.ChanState = &ChannelState{
		Stopping:      int32(0),
		Channels:      make(map[string]*GossipChannel),
		GossipService: g,
	}
	g.Emitter = &BatchingEmitter{
		cb:         g.sendGossipBatch,
		delay:      conf.MaxPropagationBurstLatency,
		iterations: conf.PropagateIterations,
		burstSize:  conf.MaxPropagationBurstSize,
		lock:       &sync.Mutex{},
		buff:       make([]*batchedMessage, 0),
		stopFlag:   int32(0),
	}
	if conf.PropagateIterations != 0 {
		go g.Emitter.periodicEmit()
	}
	g.DiscAdapter = g.newDiscoveryAdapter()
	g.DisSecAdap = &DiscoverySecurityAdapter{
		IdMapper:        g.IdMapper,
		Mcs:             g.Mcs,
		IncludeIdPeriod: g.IncludeIdPeriod,
		Id:              g.SelfId,
	}
	g.Discovery = NewDiscoveryService(g.selfNetworkMember(), g.DiscAdapter, g.DisSecAdap, g.disclosurePolicy)
	g.CertPuller = g.createCertStorePuller()
	g.CertStore = newCertStore(g.CertPuller, g.IdMapper, id, mcs)
	g.StopSignal.Add(2)
	go g.start()
	go g.connect2BootstrapPeers()
	return g
}

func BootPeers(portPrefix int, ids ...int) []string {
	var peers []string
	for _, id := range ids {
		peers = append(peers, fmt.Sprintf("localhost:%d", id+portPrefix))
	}
	return peers
}

//---------- Private Methods -----------

func newCertStore(puller *PullMediator, idMapper *IdMapper, selfId PeerIdType, mcs *NaiveMessageCryptoService) *certStore {
	selfPKIID := idMapper.GetPKIidOfCert(selfId)
	certStore := &certStore{
		mcs:      mcs,
		puller:   puller,
		idMapper: idMapper,
		selfId:   selfId,
	}
	certStore.idMapper.Put(selfPKIID, selfId)
	pi := &PeerIdentity{
		Cert:     certStore.selfId,
		Metadata: nil,
		PkiId:    certStore.idMapper.GetPKIidOfCert(certStore.selfId),
	}
	m := &GossipMessage{
		Channel: nil,
		Nonce:   0,
		Tag:     GossipMessage_EMPTY,
		Content: &GossipMessage_PeerIdentity{
			PeerIdentity: pi,
		},
	}
	signer := func(msg []byte) ([]byte, error) {
		return certStore.idMapper.Sign(msg)
	}
	selfIDMsg := &SignedGossipMessage{
		GossipMessage: m,
	}
	selfIDMsg.Sign(signer)
	puller.Add(selfIDMsg)
	puller.RegisterMsgHook(RequestMsgType, func(_ []string, msgs []*SignedGossipMessage, _ *ReceivedMessage) {
		for _, msg := range msgs {
			pkiID := PKIidType(msg.GetPeerIdentity().PkiId)
			cert := PeerIdType(msg.GetPeerIdentity().Cert)
			certStore.idMapper.Put(pkiID, cert)
		}
	})
	return certStore
}

func partitionMessages(pred MessageAcceptor, a []*EmittedGossipMessage) ([]*EmittedGossipMessage, []*EmittedGossipMessage) {
	var s1 []*EmittedGossipMessage
	var s2 []*EmittedGossipMessage
	for _, m := range a {
		if pred(m) {
			s1 = append(s1, m)
		} else {
			s2 = append(s2, m)
		}
	}
	return s1, s2
}

func selectOnlyDiscoveryMessages(m interface{}) bool {
	msg, isGossipMsg := m.(*ReceivedMessage)
	if !isGossipMsg {
		return false
	}
	alive := msg.SignedGossipMessage.GetAliveMsg()
	memRes := msg.SignedGossipMessage.GetMemRes()
	memReq := msg.SignedGossipMessage.GetMemReq()
	selected := alive != nil || memReq != nil || memRes != nil
	return selected
}
