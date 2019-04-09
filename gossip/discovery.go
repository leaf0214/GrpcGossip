package gossip

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	cmn "xianhetian.com/tendermint/tmlibs/common"
	flow "xianhetian.com/tendermint/tmlibs/flowrate"
	"xianhetian.com/tendermint/types"
)

//---------- Const -----------

const (
	msgExpirationFactor  = 20
	defaultHelloInterval = time.Duration(5) * time.Second
)

//---------- Var -----------

var MaxConnectionAttempts = 120

var AliveExpirationCheckInterval time.Duration

//---------- Struct -----------

type Peer struct {
	Endpoint         string
	InternalEndpoint string
	PKIid            PKIidType
	Metadata         []byte
	Properties       *Properties
	*Envelope

	Conn        net.Conn
	Pong        chan bool
	Timeout     chan bool
	Send        chan struct{}
	Channels    map[string]*types.Channel
	BufReader   *bufio.Reader
	BufWriter   *bufio.Writer
	SendMonitor *flow.Monitor
	RecvMonitor *flow.Monitor
	PongTimer   *time.Timer
	PingTimer   *cmn.RepeatTimer
	FlushTimer  *cmn.ThrottleTimer
	NodeInfo    types.NodeInfo
}

type GossipDiscovery struct {
	incTime          uint64
	seqNum           uint64
	self             Peer
	deadLastTS       map[string]*timestamp
	aliveLastTS      map[string]*timestamp
	id2Member        map[string]*Peer
	aliveMembership  *MembershipStore
	deadMembership   *MembershipStore
	selfAliveMessage *SignedGossipMessage

	msgStore *aliveMsgStore

	comm  *DiscoveryAdapter
	crypt *DiscoverySecurityAdapter
	lock  *sync.RWMutex

	toDieChan        chan struct{}
	toDieFlag        int32
	port             int
	disclosurePolicy DisclosurePolicy
	pubsub           *PubSub

	aliveTimeInterval            time.Duration
	aliveExpirationTimeout       time.Duration
	aliveExpirationCheckInterval time.Duration
	reconnectInterval            time.Duration
}

type DiscoveryAdapter struct {
	Stopping         int32
	Comm             *Communicate
	PresumedDeads    chan PKIidType
	IncChan          chan *ReceivedMessage
	GossipFunc       func(message *SignedGossipMessage)
	ForwardFunc      func(message *ReceivedMessage)
	DisclosurePolicy DisclosurePolicy
}

type DiscoverySecurityAdapter struct {
	Id              PeerIdType
	IncludeIdPeriod time.Time
	IdMapper        *IdMapper
	Mcs             *NaiveMessageCryptoService
}

type MembershipStore struct {
	m map[string]*SignedGossipMessage
	sync.RWMutex
}

type PeerId struct {
	ID      PKIidType
	SelfOrg bool
}

type aliveMsgStore struct {
	*MessageStore
}

type timestamp struct {
	incTime  time.Time
	seqNum   uint64
	lastSeen time.Time
}

//---------- Define -----------

type Peers []Peer

type DisclosurePolicy func(*Peer) (Sieve, EnvelopeFilter)

type EnvelopeFilter func(*SignedGossipMessage) *Envelope

type Sieve func(*SignedGossipMessage) bool

type identifier func() (*PeerId, error)

//---------- Struct Methods -----------

func (d *GossipDiscovery) Lookup(PKIID PKIidType) *Peer {
	if bytes.Equal(PKIID, d.self.PKIid) {
		return &d.self
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	nm := d.id2Member[string(PKIID)]
	return nm
}

func (d *GossipDiscovery) Connect(peer Peer, id identifier) {
	for _, endpoint := range []string{peer.InternalEndpoint, peer.Endpoint} {
		if endpoint == fmt.Sprintf("127.0.0.1:%d", d.port) || endpoint == fmt.Sprintf("localhost:%d", d.port) ||
			endpoint == d.self.InternalEndpoint || endpoint == d.self.Endpoint {
			return
		}
	}
	go func() {
		for i := 0; i < MaxConnectionAttempts && !d.toDie(); i++ {
			id, err := id()
			if err != nil {
				if d.toDie() {
					return
				}
				time.Sleep(d.reconnectInterval)
				continue
			}
			peer := &Peer{
				InternalEndpoint: peer.InternalEndpoint,
				Endpoint:         peer.Endpoint,
				PKIid:            id.ID,
			}
			m, err := d.createMembershipRequest(id.SelfOrg)
			if err != nil {
				continue
			}
			req, err := m.NoopSign()
			if err != nil {
				continue
			}
			req.Nonce = RandomUInt64()
			req, err = req.NoopSign()
			if err != nil {
				continue
			}
			go d.sendUntilAcked(peer, req)
			return
		}
	}()
}

func (d *GossipDiscovery) GetMembership() []Peer {
	if d.toDie() {
		return []Peer{}
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	var response []Peer
	for _, m := range d.aliveMembership.ToSlice() {
		peer := m.GetAliveMsg()
		response = append(response, Peer{
			PKIid:            peer.Membership.PkiId,
			Endpoint:         peer.Membership.Endpoint,
			Metadata:         peer.Membership.Metadata,
			InternalEndpoint: d.id2Member[string(m.GetAliveMsg().Membership.PkiId)].InternalEndpoint,
			Envelope:         m.Envelope,
		})
	}
	return response
}

func (d *GossipDiscovery) Self() Peer {
	var env *Envelope
	msg, _ := d.aliveMsgAndInternalEndpoint()
	sMsg, err := msg.NoopSign()
	if err == nil {
		env = sMsg.Envelope
	}
	mem := msg.GetAliveMsg().Membership
	return Peer{
		Endpoint: mem.Endpoint,
		Metadata: mem.Metadata,
		PKIid:    mem.PkiId,
		Envelope: env,
	}
}

func (d *GossipDiscovery) InitiateSync(peerNum int) {
	if d.toDie() {
		return
	}
	var peers2SendTo []*Peer
	m, err := d.createMembershipRequest(true)
	if err != nil {
		return
	}
	memReq, err := m.NoopSign()
	if err != nil {
		return
	}
	d.lock.RLock()
	n := d.aliveMembership.Size()
	k := peerNum
	if k > n {
		k = n
	}
	aliveMembersAsSlice := d.aliveMembership.ToSlice()
	for _, i := range GetRandomIndices(k, n-1) {
		pulledPeer := aliveMembersAsSlice[i].GetAliveMsg().Membership
		var internalEndpoint string
		if aliveMembersAsSlice[i].Envelope.SecretEnvelope != nil {
			internalEndpoint = aliveMembersAsSlice[i].Envelope.SecretEnvelope.InternalEndpoint()
		}
		netMember := &Peer{
			Endpoint:         pulledPeer.Endpoint,
			Metadata:         pulledPeer.Metadata,
			PKIid:            pulledPeer.PkiId,
			InternalEndpoint: internalEndpoint,
		}
		peers2SendTo = append(peers2SendTo, netMember)
	}
	d.lock.RUnlock()
	for _, netMember := range peers2SendTo {
		d.comm.SendToPeer(netMember, memReq)
	}
}

func (d *GossipDiscovery) Stop() {
	atomic.StoreInt32(&d.toDieFlag, int32(1))
	d.msgStore.Stop()
	d.toDieChan <- struct{}{}
}

func (d *GossipDiscovery) isAlive(pkiID PKIidType) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	_, alive := d.aliveLastTS[string(pkiID)]
	return alive
}

func (d *GossipDiscovery) sendUntilAcked(peer *Peer, message *SignedGossipMessage) {
	nonce := message.Nonce
	for i := 0; i < MaxConnectionAttempts && !d.toDie(); i++ {
		sub := d.pubsub.Subscribe(fmt.Sprintf("%d", nonce), time.Second*5)
		d.comm.SendToPeer(peer, message)
		if _, timeoutErr := sub.Listen(); timeoutErr == nil {
			return
		}
		time.Sleep(d.reconnectInterval)
	}
}

func (d *GossipDiscovery) handlePresumedDeadPeers() {
	for !d.toDie() {
		select {
		case deadPeer := <-d.comm.PresumedDeads:
			if d.isAlive(deadPeer) {
				d.expireDeadMembers([]PKIidType{deadPeer})
			}
		case s := <-d.toDieChan:
			d.toDieChan <- s
			return
		}
	}
}

func (d *GossipDiscovery) handleMessages() {
	in := d.comm.Receive()
	for !d.toDie() {
		select {
		case s := <-d.toDieChan:
			d.toDieChan <- s
			return
		case msg := <-in:
			d.handleMsgFromComm(msg)
		}
	}
}

func (d *GossipDiscovery) handleAliveMessage(m *SignedGossipMessage) {
	if !d.crypt.ValidateAliveMsg(m) {
		return
	}
	pkiID := m.GetAliveMsg().Membership.PkiId
	if bytes.Equal(pkiID, d.self.PKIid) {
		return
	}
	ts := m.GetAliveMsg().Timestamp
	d.lock.RLock()
	_, known := d.id2Member[string(pkiID)]
	d.lock.RUnlock()
	if !known {
		d.learnNewMembers([]*SignedGossipMessage{m}, []*SignedGossipMessage{})
		return
	}
	d.lock.RLock()
	_, isAlive := d.aliveLastTS[string(pkiID)]
	lastDeadTS, isDead := d.deadLastTS[string(pkiID)]
	d.lock.RUnlock()
	if !isAlive && !isDead {
		return
	}
	if isAlive && isDead {
		return
	}
	if isDead {
		if !before(lastDeadTS, ts) {
			return
		}
		d.resurrectMember(m, *ts)
	}
	d.lock.RLock()
	lastAliveTS, isAlive := d.aliveLastTS[string(pkiID)]
	d.lock.RUnlock()
	if isAlive && before(lastAliveTS, ts) {
		d.learnExistingMembers([]*SignedGossipMessage{m})
	}
}

func (d *GossipDiscovery) handleMsgFromComm(msg *ReceivedMessage) {
	if msg == nil {
		return
	}
	m := msg.SignedGossipMessage
	if m.GetAliveMsg() == nil && m.GetMemRes() == nil && m.GetMemReq() == nil {
		return
	}
	if memReq := m.GetMemReq(); memReq != nil {
		selfInfoGossipMsg, err := memReq.SelfInformation.ToGossipMessage()
		if err != nil {
			return
		}
		if d.msgStore.CheckValid(selfInfoGossipMsg) {
			d.handleAliveMessage(selfInfoGossipMsg)
		}
		var internalEndpoint string
		if m.Envelope.SecretEnvelope != nil {
			internalEndpoint = m.Envelope.SecretEnvelope.InternalEndpoint()
		}
		go d.sendMemResponse(selfInfoGossipMsg.GetAliveMsg().Membership, internalEndpoint, m.Nonce)
		return
	}
	if m.GetAliveMsg() != nil {
		if !d.msgStore.Add(m) {
			return
		}
		d.handleAliveMessage(m)
		if d.comm.toDie() {
			return
		}
		d.comm.ForwardFunc(msg)
		return
	}
	if memResp := m.GetMemRes(); memResp != nil {
		d.pubsub.Publish(fmt.Sprintf("%d", m.Nonce), m.Nonce)
		for _, env := range memResp.Alive {
			am, err := env.ToGossipMessage()
			if err != nil {
				return
			}
			if am.GetAliveMsg() == nil {
				return
			}
			if d.msgStore.CheckValid(am) {
				d.handleAliveMessage(am)
			}
		}
		for _, env := range memResp.Dead {
			dm, err := env.ToGossipMessage()
			if err != nil {
				return
			}
			if !d.crypt.ValidateAliveMsg(dm) {
				continue
			}
			if !d.msgStore.CheckValid(dm) {
				return
			}
			var newDeadMembers []*SignedGossipMessage
			d.lock.RLock()
			if _, known := d.id2Member[string(dm.GetAliveMsg().Membership.PkiId)]; !known {
				newDeadMembers = append(newDeadMembers, dm)
			}
			d.lock.RUnlock()
			d.learnNewMembers([]*SignedGossipMessage{}, newDeadMembers)
		}
	}
}

func (d *GossipDiscovery) sendMemResponse(targetMember *Member, internalEndpoint string, nonce uint64) {
	targetPeer := &Peer{
		Endpoint:         targetMember.Endpoint,
		Metadata:         targetMember.Metadata,
		PKIid:            targetMember.PkiId,
		InternalEndpoint: internalEndpoint,
	}
	var aliveMsg *SignedGossipMessage
	var err error
	d.lock.RLock()
	aliveMsg = d.selfAliveMessage
	d.lock.RUnlock()
	if aliveMsg == nil {
		aliveMsg, err = d.createSignedAliveMessage(true)
		if err != nil {
			return
		}
	}
	memResp := d.createMembershipResponse(aliveMsg, targetPeer)
	if memResp == nil {
		d.comm.CloseConn(targetPeer)
		return
	}
	msg, err := (&GossipMessage{
		Tag:   GossipMessage_EMPTY,
		Nonce: nonce,
		Content: &GossipMessage_MemRes{
			MemRes: memResp,
		},
	}).NoopSign()
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	d.comm.SendToPeer(targetPeer, msg)
}

func (d *GossipDiscovery) createMembershipResponse(aliveMsg *SignedGossipMessage, targetMember *Peer) *MembershipResponse {
	shouldBeDisclosed, omitConcealedFields := d.disclosurePolicy(targetMember)
	if !shouldBeDisclosed(aliveMsg) {
		return nil
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	var deadPeers []*Envelope
	for _, dm := range d.deadMembership.ToSlice() {
		if !shouldBeDisclosed(dm) {
			continue
		}
		deadPeers = append(deadPeers, omitConcealedFields(dm))
	}
	var aliveSnapshot []*Envelope
	for _, am := range d.aliveMembership.ToSlice() {
		if !shouldBeDisclosed(am) {
			continue
		}
		aliveSnapshot = append(aliveSnapshot, omitConcealedFields(am))
	}
	return &MembershipResponse{
		Alive: append(aliveSnapshot, omitConcealedFields(aliveMsg)),
		Dead:  deadPeers,
	}
}

func (d *GossipDiscovery) resurrectMember(am *SignedGossipMessage, t PeerTime) {
	d.lock.Lock()
	defer d.lock.Unlock()
	peer := am.GetAliveMsg().Membership
	pkiID := peer.PkiId
	d.aliveLastTS[string(pkiID)] = &timestamp{
		lastSeen: time.Now(),
		seqNum:   t.SeqNum,
		incTime:  time.Unix(int64(0), int64(t.IncNum)),
	}
	var internalEndpoint string
	if prevNetMem := d.id2Member[string(pkiID)]; prevNetMem != nil {
		internalEndpoint = prevNetMem.InternalEndpoint
	}
	if am.Envelope.SecretEnvelope != nil {
		internalEndpoint = am.Envelope.SecretEnvelope.InternalEndpoint()
	}
	d.id2Member[string(pkiID)] = &Peer{
		Endpoint:         peer.Endpoint,
		Metadata:         peer.Metadata,
		PKIid:            peer.PkiId,
		InternalEndpoint: internalEndpoint,
	}
	delete(d.deadLastTS, string(pkiID))
	d.deadMembership.Remove(PKIidType(pkiID))
	d.aliveMembership.Put(PKIidType(pkiID), &SignedGossipMessage{GossipMessage: am.GossipMessage, Envelope: am.Envelope})
}

func (d *GossipDiscovery) periodicalReconnectToDead() {
	for !d.toDie() {
		wg := &sync.WaitGroup{}
		for _, peer := range d.copyLastSeen(d.deadLastTS) {
			wg.Add(1)
			go func(peer Peer) {
				defer wg.Done()
				if !d.comm.Ping(&peer) {
					return
				}
				m, err := d.createMembershipRequest(true)
				if err != nil {
					return
				}
				req, err := m.NoopSign()
				if err != nil {
					return
				}
				d.comm.SendToPeer(&peer, req)
			}(peer)
		}
		wg.Wait()
		time.Sleep(d.reconnectInterval)
	}
}

func (d *GossipDiscovery) createMembershipRequest(includeInternalEndpoint bool) (*GossipMessage, error) {
	am, err := d.createSignedAliveMessage(includeInternalEndpoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req := &MembershipRequest{
		SelfInformation: am.Envelope,
		Known:           [][]byte{},
	}
	return &GossipMessage{
		Tag:   GossipMessage_EMPTY,
		Nonce: uint64(0),
		Content: &GossipMessage_MemReq{
			MemReq: req,
		},
	}, nil
}

func (d *GossipDiscovery) copyLastSeen(lastSeenMap map[string]*timestamp) []Peer {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var res []Peer
	for pkiIDStr := range lastSeenMap {
		res = append(res, *(d.id2Member[pkiIDStr]))
	}
	return res
}

func (d *GossipDiscovery) periodicalCheckAlive() {
	for !d.toDie() {
		time.Sleep(d.aliveExpirationCheckInterval)
		dead := d.getDeadMembers()
		if len(dead) > 0 {
			d.expireDeadMembers(dead)
		}
	}
}

func (d *GossipDiscovery) expireDeadMembers(dead []PKIidType) {
	var deadMembers2Expire []*Peer
	d.lock.Lock()
	for _, pkiID := range dead {
		if _, isAlive := d.aliveLastTS[string(pkiID)]; !isAlive {
			continue
		}
		deadMembers2Expire = append(deadMembers2Expire, d.id2Member[string(pkiID)])
		lastTS, hasLastTS := d.aliveLastTS[string(pkiID)]
		if hasLastTS {
			d.deadLastTS[string(pkiID)] = lastTS
			delete(d.aliveLastTS, string(pkiID))
		}
		if am := d.aliveMembership.GetMsgByID(pkiID); am != nil {
			d.deadMembership.Put(pkiID, am)
			d.aliveMembership.Remove(pkiID)
		}
	}
	d.lock.Unlock()
	for _, member2Expire := range deadMembers2Expire {
		d.comm.CloseConn(member2Expire)
	}
}

func (d *GossipDiscovery) getDeadMembers() []PKIidType {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var dead []PKIidType
	for id, last := range d.aliveLastTS {
		elapsedNonAliveTime := time.Since(last.lastSeen)
		if elapsedNonAliveTime > d.aliveExpirationTimeout {
			dead = append(dead, PKIidType(id))
		}
	}
	return dead
}

func (d *GossipDiscovery) periodicalSendAlive() {
	for !d.toDie() {
		time.Sleep(d.aliveTimeInterval)
		msg, err := d.createSignedAliveMessage(true)
		if err != nil {
			return
		}
		d.lock.Lock()
		d.selfAliveMessage = msg
		d.lock.Unlock()
		d.comm.Gossip(msg)
	}
}

func (d *GossipDiscovery) aliveMsgAndInternalEndpoint() (*GossipMessage, string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.seqNum++
	msg := &GossipMessage{
		Tag: GossipMessage_EMPTY,
		Content: &GossipMessage_AliveMsg{
			AliveMsg: &AliveMessage{
				Membership: &Member{
					Endpoint: d.self.Endpoint,
					Metadata: d.self.Metadata,
					PkiId:    d.self.PKIid,
				},
				Timestamp: &PeerTime{
					IncNum: uint64(d.incTime),
					SeqNum: d.seqNum,
				},
			},
		},
	}
	return msg, d.self.InternalEndpoint
}

func (d *GossipDiscovery) createSignedAliveMessage(includeInternalEndpoint bool) (*SignedGossipMessage, error) {
	msg, internalEndpoint := d.aliveMsgAndInternalEndpoint()
	envp := d.crypt.SignMessage(msg, internalEndpoint)
	if envp == nil {
		return nil, errors.New("Failed signing message")
	}
	signedMsg := &SignedGossipMessage{
		GossipMessage: msg,
		Envelope:      envp,
	}
	if !includeInternalEndpoint {
		signedMsg.Envelope.SecretEnvelope = nil
	}
	return signedMsg, nil
}

func (d *GossipDiscovery) learnExistingMembers(aliveArr []*SignedGossipMessage) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, m := range aliveArr {
		am := m.GetAliveMsg()
		if am == nil {
			return
		}
		var internalEndpoint string
		if prevNetMem := d.id2Member[string(am.Membership.PkiId)]; prevNetMem != nil {
			internalEndpoint = prevNetMem.InternalEndpoint
		}
		if m.Envelope.SecretEnvelope != nil {
			internalEndpoint = m.Envelope.SecretEnvelope.InternalEndpoint()
		}
		peer := d.id2Member[string(am.Membership.PkiId)]
		peer.Endpoint = am.Membership.Endpoint
		peer.Metadata = am.Membership.Metadata
		peer.InternalEndpoint = internalEndpoint
		if _, isKnownAsDead := d.deadLastTS[string(am.Membership.PkiId)]; isKnownAsDead {
			continue
		}
		if _, isKnownAsAlive := d.aliveLastTS[string(am.Membership.PkiId)]; !isKnownAsAlive {
			continue
		}
		alive := d.aliveLastTS[string(am.Membership.PkiId)]
		alive.incTime = time.Unix(int64(0), int64(am.Timestamp.IncNum))
		alive.lastSeen = time.Now()
		alive.seqNum = am.Timestamp.SeqNum
		var sgm *SignedGossipMessage
		if sgm = d.aliveMembership.GetMsgByID(m.GetAliveMsg().Membership.PkiId); sgm == nil {
			msg := &SignedGossipMessage{GossipMessage: m.GossipMessage, Envelope: sgm.Envelope}
			d.aliveMembership.Put(m.GetAliveMsg().Membership.PkiId, msg)
			continue
		}
		sgm.GossipMessage = m.GossipMessage
		sgm.Envelope = m.Envelope
	}
}

func (d *GossipDiscovery) learnNewMembers(aliveMembers []*SignedGossipMessage, deadMembers []*SignedGossipMessage) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, am := range aliveMembers {
		if bytes.Equal(am.GetAliveMsg().Membership.PkiId, d.self.PKIid) {
			continue
		}
		d.aliveLastTS[string(am.GetAliveMsg().Membership.PkiId)] = &timestamp{
			incTime:  time.Unix(int64(0), int64(am.GetAliveMsg().Timestamp.IncNum)),
			lastSeen: time.Now(),
			seqNum:   am.GetAliveMsg().Timestamp.SeqNum,
		}
		d.aliveMembership.Put(am.GetAliveMsg().Membership.PkiId, &SignedGossipMessage{GossipMessage: am.GossipMessage, Envelope: am.Envelope})
	}
	for _, dm := range deadMembers {
		if bytes.Equal(dm.GetAliveMsg().Membership.PkiId, d.self.PKIid) {
			continue
		}
		d.deadLastTS[string(dm.GetAliveMsg().Membership.PkiId)] = &timestamp{
			incTime:  time.Unix(int64(0), int64(dm.GetAliveMsg().Timestamp.IncNum)),
			lastSeen: time.Now(),
			seqNum:   dm.GetAliveMsg().Timestamp.SeqNum,
		}
		d.deadMembership.Put(dm.GetAliveMsg().Membership.PkiId, &SignedGossipMessage{GossipMessage: dm.GossipMessage, Envelope: dm.Envelope})
	}
	for _, a := range [][]*SignedGossipMessage{aliveMembers, deadMembers} {
		for _, m := range a {
			peer := m.GetAliveMsg()
			if peer == nil {
				return
			}
			var internalEndpoint string
			if m.Envelope.SecretEnvelope != nil {
				internalEndpoint = m.Envelope.SecretEnvelope.InternalEndpoint()
			}
			if prevNetMem := d.id2Member[string(peer.Membership.PkiId)]; prevNetMem != nil {
				internalEndpoint = prevNetMem.InternalEndpoint
			}
			d.id2Member[string(peer.Membership.PkiId)] = &Peer{
				Endpoint:         peer.Membership.Endpoint,
				Metadata:         peer.Membership.Metadata,
				PKIid:            peer.Membership.PkiId,
				InternalEndpoint: internalEndpoint,
			}
		}
	}
}

func (d *GossipDiscovery) toDie() bool {
	toDie := atomic.LoadInt32(&d.toDieFlag) == int32(1)
	return toDie
}

func (da *DiscoveryAdapter) Gossip(msg *SignedGossipMessage) {
	if da.toDie() {
		return
	}
	da.GossipFunc(msg)
}

func (da *DiscoveryAdapter) Ping(peer *Peer) bool {
	err := da.Comm.Probe(&RemotePeer{Endpoint: peer.PreferredEndpoint(), PKIID: peer.PKIid})
	return err == nil
}

func (da *DiscoveryAdapter) SendToPeer(peer *Peer, msg *SignedGossipMessage) {
	if da.toDie() {
		return
	}
	if memReq := msg.GetMemReq(); memReq != nil && len(peer.PKIid) != 0 {
		selfMsg, err := memReq.SelfInformation.ToGossipMessage()
		_, omitConcealedFields := da.DisclosurePolicy(peer)
		selfMsg.Envelope = omitConcealedFields(selfMsg)
		oldKnown := memReq.Known
		memReq = &MembershipRequest{
			SelfInformation: selfMsg.Envelope,
			Known:           oldKnown,
		}
		msg.Content = &GossipMessage_MemReq{
			MemReq: memReq,
		}
		msg, err = msg.NoopSign()
		if err != nil {
			return
		}
	}
	da.Comm.Send(msg, &RemotePeer{PKIID: peer.PKIid, Endpoint: peer.PreferredEndpoint()})
}

func (da *DiscoveryAdapter) Receive() <-chan *ReceivedMessage {
	return da.IncChan
}

func (da *DiscoveryAdapter) CloseConn(peer *Peer) {
	da.Comm.CloseConn(&RemotePeer{PKIID: peer.PKIid})
}

func (da *DiscoveryAdapter) toDie() bool {
	return atomic.LoadInt32(&da.Stopping) == int32(1)
}

func (sa *DiscoverySecurityAdapter) SignMessage(m *GossipMessage, internalEndpoint string) *Envelope {
	signer := func(msg []byte) ([]byte, error) {
		return sa.Mcs.Sign(msg)
	}
	if m.GetAliveMsg() != nil && time.Now().Before(sa.IncludeIdPeriod) {
		m.GetAliveMsg().Identity = sa.Id
	}
	sMsg := &SignedGossipMessage{
		GossipMessage: m,
	}
	e, err := sMsg.Sign(signer)
	if err != nil {
		return nil
	}
	if internalEndpoint == "" {
		return e
	}
	e.SignSecret(signer, &Secret{
		Content: &Secret_InternalEndpoint{
			InternalEndpoint: internalEndpoint,
		},
	})
	return e
}

func (sa *DiscoverySecurityAdapter) ValidateAliveMsg(m *SignedGossipMessage) bool {
	am := m.GetAliveMsg()
	if am == nil || am.Membership == nil || am.Membership.PkiId == nil || m.Envelope == nil || m.Envelope.Payload == nil || m.Envelope.Signature == nil {
		return false
	}
	var idTypes PeerIdType
	if am.Identity != nil {
		idTypes = PeerIdType(am.Identity)
		claimedPKIID := am.Membership.PkiId
		if err := sa.IdMapper.Put(claimedPKIID, idTypes); err != nil {
			return false
		}
	} else {
		idTypes, _ = sa.IdMapper.Get(am.Membership.PkiId)
	}
	if idTypes == nil {
		return false
	}
	am = m.GetAliveMsg()
	verifier := func(peerId []byte, signature, message []byte) error {
		return sa.Mcs.Verify(PeerIdType(peerId), signature, message)
	}
	if err := m.Verify(idTypes, verifier); err != nil {
		return false
	}
	return true
}

func (s *aliveMsgStore) Add(msg interface{}) bool {
	if msg.(*SignedGossipMessage).GetAliveMsg() == nil {
		panic(fmt.Sprint("Msg ", msg, " is not AliveMsg"))
	}
	return s.MessageStore.Add(msg)
}

func (s *aliveMsgStore) CheckValid(msg interface{}) bool {
	if msg.(*SignedGossipMessage).GetAliveMsg() == nil {
		panic(fmt.Sprint("Msg ", msg, " is not AliveMsg"))
	}
	return s.MessageStore.CheckValid(msg)
}

func (p Peer) PreferredEndpoint() string {
	if p.InternalEndpoint != "" {
		return p.InternalEndpoint
	}
	return p.Endpoint
}

func (m *MembershipStore) GetMsgByID(pkiID PKIidType) *SignedGossipMessage {
	m.RLock()
	defer m.RUnlock()
	if msg, exists := m.m[string(pkiID)]; exists {
		return msg
	}
	return nil
}

func (m *MembershipStore) Size() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

func (m *MembershipStore) Put(pkiID PKIidType, msg *SignedGossipMessage) {
	m.Lock()
	defer m.Unlock()
	m.m[string(pkiID)] = msg
}

func (m *MembershipStore) Remove(pkiID PKIidType) {
	m.Lock()
	defer m.Unlock()
	delete(m.m, string(pkiID))
}

func (m *MembershipStore) ToSlice() []*SignedGossipMessage {
	m.RLock()
	defer m.RUnlock()
	peers := make([]*SignedGossipMessage, len(m.m))
	i := 0
	for _, peer := range m.m {
		peers[i] = peer
		i++
	}
	return peers
}

//---------- Public Methods -----------

func NewDiscoveryService(self Peer, comm *DiscoveryAdapter, crypt *DiscoverySecurityAdapter, disPol DisclosurePolicy) *GossipDiscovery {
	d := &GossipDiscovery{
		self:                         self,
		incTime:                      uint64(time.Now().UnixNano()),
		seqNum:                       uint64(0),
		deadLastTS:                   make(map[string]*timestamp),
		aliveLastTS:                  make(map[string]*timestamp),
		id2Member:                    make(map[string]*Peer),
		aliveMembership:              NewMembershipStore(),
		deadMembership:               NewMembershipStore(),
		crypt:                        crypt,
		comm:                         comm,
		lock:                         &sync.RWMutex{},
		toDieChan:                    make(chan struct{}, 1),
		toDieFlag:                    int32(0),
		disclosurePolicy:             disPol,
		pubsub:                       NewPubSub(),
		aliveTimeInterval:            getAliveTimeInterval(),
		aliveExpirationTimeout:       getAliveExpirationTimeout(),
		aliveExpirationCheckInterval: getAliveExpirationCheckInterval(),
		reconnectInterval:            getReconnectInterval(),
	}
	endpoint := d.self.InternalEndpoint
	internalEndpointSplit := strings.Split(endpoint, ":")
	myPort, _ := strconv.ParseInt(internalEndpointSplit[1], 10, 64)
	d.port = int(myPort)
	policy := NewGossipMessageComparator(0)
	aliveMsgTTL := d.aliveExpirationTimeout * msgExpirationFactor
	externalLock := func() { d.lock.Lock() }
	externalUnlock := func() { d.lock.Unlock() }
	callback := func(m interface{}) {
		msg := m.(*SignedGossipMessage)
		if msg.GetAliveMsg() == nil {
			return
		}
		id := msg.GetAliveMsg().Membership.PkiId
		d.aliveMembership.Remove(id)
		d.deadMembership.Remove(id)
		delete(d.id2Member, string(id))
		delete(d.deadLastTS, string(id))
		delete(d.aliveLastTS, string(id))
	}
	d.msgStore = &aliveMsgStore{
		MessageStore: NewMessageStoreExpirable(policy, func(interface{}) {}, aliveMsgTTL, externalLock, externalUnlock, callback),
	}
	go d.periodicalSendAlive()
	go d.periodicalCheckAlive()
	go d.handleMessages()
	go d.periodicalReconnectToDead()
	go d.handlePresumedDeadPeers()
	return d
}

func NewMembershipStore() *MembershipStore {
	return &MembershipStore{m: make(map[string]*SignedGossipMessage)}
}

//---------- Private Methods -----------

func before(a *timestamp, b *PeerTime) bool {
	return (uint64(a.incTime.UnixNano()) == b.IncNum && a.seqNum < b.SeqNum) || uint64(a.incTime.UnixNano()) < b.IncNum
}

func getAliveTimeInterval() time.Duration {
	return GetDurationOrDefault("peer.gossip.aliveTimeInterval", defaultHelloInterval)
}

func getAliveExpirationTimeout() time.Duration {
	return GetDurationOrDefault("peer.gossip.aliveExpirationTimeout", 5*getAliveTimeInterval())
}

func getAliveExpirationCheckInterval() time.Duration {
	if AliveExpirationCheckInterval != 0 {
		return AliveExpirationCheckInterval
	}
	return time.Duration(getAliveExpirationTimeout() / 10)
}

func getReconnectInterval() time.Duration {
	return GetDurationOrDefault("peer.gossip.reconnectInterval", getAliveExpirationTimeout())
}
