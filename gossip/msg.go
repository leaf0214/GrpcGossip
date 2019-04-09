package gossip

import (
	"sync"

	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"time"
)

//---------- Struct -----------

type ReceivedMessage struct {
	*SignedGossipMessage
	Conn     *Connection
	ConnInfo *ConnectionInfo
}

type SignedGossipMessage struct {
	*Envelope
	*GossipMessage
}

type MessageStore struct {
	policy            MessageReplacingPolicy
	lock              sync.RWMutex
	messages          []*msg
	invTrigger        invalidationTrigger
	msgTTL            time.Duration
	expiredCount      int
	externalLock      func()
	externalUnlock    func()
	expireMsgCallback func(msg interface{})
	doneCh            chan struct{}
	stopOnce          sync.Once
}

type ConnectionInfo struct {
	ID PKIidType
	Id PeerIdType
}

type msg struct {
	data    interface{}
	created time.Time
	expired bool
}

type msgComparator struct {
	dataBlockStorageSize int
}

//---------- Define -----------

type Signer func([]byte) ([]byte, error)

type verifier func(peerId []byte, signature, message []byte) error

type invalidationTrigger func(interface{})

func Noop(_ interface{}) {
}

//---------- Struct Methods -----------

func (s *MessageStore) Add(message interface{}) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	n := len(s.messages)
	for i := 0; i < n; i++ {
		m := s.messages[i]
		switch s.policy(message, m.data) {
		case MessageInvalidated:
			return false
		case MessageInvalidates:
			s.invTrigger(m.data)
			s.messages = append(s.messages[:i], s.messages[i+1:]...)
			n--
			i--
		}
	}
	s.messages = append(s.messages, &msg{data: message, created: time.Now()})
	return true
}

func (s *MessageStore) CheckValid(message interface{}) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, m := range s.messages {
		if s.policy(message, m.data) == MessageInvalidated {
			return false
		}
	}
	return true
}

func (s *MessageStore) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.messages) - s.expiredCount
}

func (s *MessageStore) Get() []interface{} {
	res := make([]interface{}, 0)
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, msg := range s.messages {
		if !msg.expired {
			res = append(res, msg.data)
		}
	}
	return res
}

func (s *MessageStore) Purge(shouldBePurged func(interface{}) bool) {
	shouldMsgBePurged := func(m *msg) bool {
		return shouldBePurged(m.data)
	}
	if !s.isPurgeNeeded(shouldMsgBePurged) {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	n := len(s.messages)
	for i := 0; i < n; i++ {
		if !shouldMsgBePurged(s.messages[i]) {
			continue
		}
		s.invTrigger(s.messages[i].data)
		s.messages = append(s.messages[:i], s.messages[i+1:]...)
		n--
		i--
	}
}

func (s *MessageStore) Stop() {
	stopFunc := func() {
		close(s.doneCh)
	}
	s.stopOnce.Do(stopFunc)
}

func (s *MessageStore) expirationRoutine() {
	for {
		select {
		case <-s.doneCh:
			return
		case <-time.After(s.msgTTL / 100):
			hasMessageExpired := func(m *msg) bool {
				if !m.expired && time.Since(m.created) > s.msgTTL {
					return true
				}
				if time.Since(m.created) > (s.msgTTL * 2) {
					return true
				}
				return false
			}
			if s.isPurgeNeeded(hasMessageExpired) {
				s.expireMessages()
			}
		}
	}
}

func (s *MessageStore) expireMessages() {
	s.externalLock()
	s.lock.Lock()
	defer s.lock.Unlock()
	defer s.externalUnlock()
	n := len(s.messages)
	for i := 0; i < n; i++ {
		m := s.messages[i]
		if !m.expired {
			if time.Since(m.created) > s.msgTTL {
				m.expired = true
				s.expireMsgCallback(m.data)
				s.expiredCount++
			}
			continue
		}
		if time.Since(m.created) > (s.msgTTL * 2) {
			s.messages = append(s.messages[:i], s.messages[i+1:]...)
			n--
			i--
			s.expiredCount--
		}
	}
}

func (s *MessageStore) isPurgeNeeded(shouldBePurged func(*msg) bool) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, m := range s.messages {
		if shouldBePurged(m) {
			return true
		}
	}
	return false
}

func (m *ReceivedMessage) Respond(msg *GossipMessage) {
	sMsg, err := msg.NoopSign()
	if err != nil {
		return
	}
	m.Conn.Send(sMsg, func(error) {}, BlockingSend)
}

func (m *ReceivedMessage) Ack(err error) {
	ackMsg := &GossipMessage{
		Nonce: m.SignedGossipMessage.Nonce,
		Content: &GossipMessage_Ack{
			Ack: &Acknowledgement{},
		},
	}
	if err != nil {
		ackMsg.GetAck().Error = err.Error()
	}
	m.Respond(ackMsg)
}

func (m *SignedGossipMessage) Sign(signer Signer) (*Envelope, error) {
	var secretEnvelope *SecretEnvelope
	if m.Envelope != nil {
		secretEnvelope = m.Envelope.SecretEnvelope
	}
	m.Envelope = nil
	payload, err := proto.Marshal(m.GossipMessage)
	if err != nil {
		return nil, err
	}
	sig, err := signer(payload)
	if err != nil {
		return nil, err
	}
	e := &Envelope{
		Payload:        payload,
		Signature:      sig,
		SecretEnvelope: secretEnvelope,
	}
	m.Envelope = e
	return e, nil
}

func (m *SignedGossipMessage) Verify(peerId []byte, verify verifier) error {
	if m.Envelope == nil {
		return errors.New("missing envelope")
	}
	if len(m.Envelope.Payload) == 0 {
		return errors.New("empty payload")
	}
	if len(m.Envelope.Signature) == 0 {
		return errors.New("empty signature")
	}
	payloadSigVerificationErr := verify(peerId, m.Envelope.Signature, m.Envelope.Payload)
	if payloadSigVerificationErr != nil {
		return payloadSigVerificationErr
	}
	if m.Envelope.SecretEnvelope != nil {
		payload := m.Envelope.SecretEnvelope.Payload
		sig := m.Envelope.SecretEnvelope.Signature
		if len(payload) == 0 {
			return errors.New("empty payload")
		}
		if len(sig) == 0 {
			return errors.New("empty signature")
		}
		return verify(peerId, sig, payload)
	}
	return nil
}

func (mc *msgComparator) getMsgReplacingPolicy() MessageReplacingPolicy {
	return func(this interface{}, that interface{}) InvalidationResult {
		return mc.invalidationPolicy(this, that)
	}
}

func (mc *msgComparator) invalidationPolicy(this interface{}, that interface{}) InvalidationResult {
	thisMsg := this.(*SignedGossipMessage)
	thatMsg := that.(*SignedGossipMessage)
	if thisMsg.GetAliveMsg() != nil && thatMsg.GetAliveMsg() != nil {
		return aliveInvalidationPolicy(thisMsg.GetAliveMsg(), thatMsg.GetAliveMsg())
	}
	if thisMsg.GetDataMsg() != nil && thatMsg.GetDataMsg() != nil {
		return mc.dataInvalidationPolicy(thisMsg.GetDataMsg(), thatMsg.GetDataMsg())
	}
	if thisMsg.GetStateInfo() != nil && thatMsg.GetStateInfo() != nil {
		return mc.stateInvalidationPolicy(thisMsg.GetStateInfo(), thatMsg.GetStateInfo())
	}
	if thisMsg.GetPeerIdentity() != nil && thatMsg.GetPeerIdentity() != nil {
		return mc.idInvalidationPolicy(thisMsg.GetPeerIdentity(), thatMsg.GetPeerIdentity())
	}
	if thisMsg.GetLeadershipMsg() != nil && thatMsg.GetLeadershipMsg() != nil {
		return leaderInvalidationPolicy(thisMsg.GetLeadershipMsg(), thatMsg.GetLeadershipMsg())
	}
	return MessageNoAction
}

func (mc *msgComparator) stateInvalidationPolicy(thisStateMsg *StateInfo, thatStateMsg *StateInfo) InvalidationResult {
	if !bytes.Equal(thisStateMsg.PkiId, thatStateMsg.PkiId) {
		return MessageNoAction
	}
	return compareTimestamps(thisStateMsg.Timestamp, thatStateMsg.Timestamp)
}

func (mc *msgComparator) idInvalidationPolicy(thisIdMsg *PeerIdentity, thatIdMsg *PeerIdentity) InvalidationResult {
	if bytes.Equal(thisIdMsg.PkiId, thatIdMsg.PkiId) {
		return MessageInvalidated
	}
	return MessageNoAction
}

func (mc *msgComparator) dataInvalidationPolicy(thisDataMsg *DataMessage, thatDataMsg *DataMessage) InvalidationResult {
	if thisDataMsg.Payload.SeqNum == thatDataMsg.Payload.SeqNum {
		return MessageInvalidated
	}
	diff := abs(thisDataMsg.Payload.SeqNum, thatDataMsg.Payload.SeqNum)
	if diff <= uint64(mc.dataBlockStorageSize) {
		return MessageNoAction
	}
	if thisDataMsg.Payload.SeqNum > thatDataMsg.Payload.SeqNum {
		return MessageInvalidates
	}
	return MessageInvalidated
}

func (m *GossipMessage) IsPullMsg() bool {
	return m.GetDataReq() != nil || m.GetDataUpdate() != nil || m.GetHello() != nil || m.GetDataDig() != nil
}

func (m *GossipMessage) IsChannelRestricted() bool {
	return m.Tag == GossipMessage_CHAN_AND_ORG || m.Tag == GossipMessage_CHAN_ONLY || m.Tag == GossipMessage_CHAN_OR_ORG
}

func (m *GossipMessage) IsOrgRestricted() bool {
	return m.Tag == GossipMessage_CHAN_AND_ORG || m.Tag == GossipMessage_ORG_ONLY
}

func (m *GossipMessage) IsPrivateDataMsg() bool {
	return m.GetPrivateReq() != nil || m.GetPrivateRes() != nil || m.GetPrivateData() != nil
}

func (m *GossipMessage) IsTagLegal() error {
	if m.Tag == GossipMessage_UNDEFINED {
		return fmt.Errorf("undefined tag")
	}
	if m.GetDataMsg() != nil {
		if m.Tag != GossipMessage_CHAN_AND_ORG {
			return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_CHAN_AND_ORG)])
		}
		return nil
	}
	if m.GetAliveMsg() != nil || m.GetMemReq() != nil || m.GetMemRes() != nil {
		if m.Tag != GossipMessage_EMPTY {
			return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_EMPTY)])
		}
		return nil
	}
	if m.GetPeerIdentity() != nil {
		if m.Tag != GossipMessage_ORG_ONLY {
			return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_ORG_ONLY)])
		}
		return nil
	}
	if m.IsPullMsg() {
		switch m.GetPullMsgType() {
		case PullMsgType_BLOCK_MSG:
			if m.Tag != GossipMessage_CHAN_AND_ORG {
				return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_CHAN_AND_ORG)])
			}
			return nil
		case PullMsgType_IDENTITY_MSG:
			if m.Tag != GossipMessage_EMPTY {
				return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_EMPTY)])
			}
			return nil
		default:
			return fmt.Errorf("invalid PullMsgType:%s", PullMsgType_name[int32(m.GetPullMsgType())])
		}
	}
	if m.GetStateInfo() != nil || m.GetStateInfoPullReq() != nil || m.GetStateSnapshot() != nil || m.GetStateRequest() != nil || m.GetStateResponse() != nil {
		if m.Tag != GossipMessage_CHAN_OR_ORG {
			return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_CHAN_OR_ORG)])
		}
		return nil
	}
	if m.GetLeadershipMsg() != nil {
		if m.Tag != GossipMessage_CHAN_AND_ORG {
			return fmt.Errorf("tag should be %s", GossipMessage_Tag_name[int32(GossipMessage_CHAN_AND_ORG)])
		}
		return nil
	}
	return fmt.Errorf("unknown message type:%v", m)
}

func (m *GossipMessage) GetPullMsgType() PullMsgType {
	if helloMsg := m.GetHello(); helloMsg != nil {
		return helloMsg.MsgType
	}
	if digMsg := m.GetDataDig(); digMsg != nil {
		return digMsg.MsgType
	}
	if reqMsg := m.GetDataReq(); reqMsg != nil {
		return reqMsg.MsgType
	}
	if resMsg := m.GetDataUpdate(); resMsg != nil {
		return resMsg.MsgType
	}
	return PullMsgType_UNDEFINED
}

func (m *GossipMessage) NoopSign() (*SignedGossipMessage, error) {
	signer := func([]byte) ([]byte, error) {
		return nil, nil
	}
	sMsg := &SignedGossipMessage{
		GossipMessage: m,
	}
	_, err := sMsg.Sign(signer)
	return sMsg, err
}

func (e *Envelope) ToGossipMessage() (*SignedGossipMessage, error) {
	msg := &GossipMessage{}
	if err := proto.Unmarshal(e.Payload, msg); err != nil {
		return nil, fmt.Errorf("failed unmarshaling GossipMessage from envelope:%v", err)
	}
	return &SignedGossipMessage{
		GossipMessage: msg,
		Envelope:      e,
	}, nil
}

func (e *Envelope) SignSecret(signer Signer, secret *Secret) error {
	payload, err := proto.Marshal(secret)
	if err != nil {
		return err
	}
	sig, err := signer(payload)
	if err != nil {
		return err
	}
	e.SecretEnvelope = &SecretEnvelope{
		Payload:   payload,
		Signature: sig,
	}
	return nil
}

func (se *SecretEnvelope) InternalEndpoint() string {
	secret := &Secret{}
	if err := proto.Unmarshal(se.Payload, secret); err != nil {
		return ""
	}
	return secret.GetInternalEndpoint()
}

//---------- Public Methods -----------

func NewMessageStore(policy MessageReplacingPolicy, trigger invalidationTrigger) *MessageStore {
	return newMsgStore(policy, trigger)
}

func NewMessageStoreExpirable(policy MessageReplacingPolicy, trigger invalidationTrigger, msgTTL time.Duration, externalLock func(), externalUnlock func(), externalExpire func(interface{})) *MessageStore {
	store := newMsgStore(policy, trigger)
	store.msgTTL = msgTTL
	if externalLock != nil {
		store.externalLock = externalLock
	}
	if externalUnlock != nil {
		store.externalUnlock = externalUnlock
	}
	if externalExpire != nil {
		store.expireMsgCallback = externalExpire
	}
	go store.expirationRoutine()
	return store
}

func NewGossipMessageComparator(dataBlockStorageSize int) MessageReplacingPolicy {
	return (&msgComparator{dataBlockStorageSize: dataBlockStorageSize}).getMsgReplacingPolicy()
}

//---------- Private Methods -----------

func newMsgStore(policy MessageReplacingPolicy, trigger invalidationTrigger) *MessageStore {
	return &MessageStore{
		policy:            policy,
		messages:          make([]*msg, 0),
		invTrigger:        trigger,
		externalLock:      func() {},
		externalUnlock:    func() {},
		expireMsgCallback: func(interface{}) {},
		expiredCount:      0,
		doneCh:            make(chan struct{}),
	}
}

func aliveInvalidationPolicy(thisMsg *AliveMessage, thatMsg *AliveMessage) InvalidationResult {
	if !bytes.Equal(thisMsg.Membership.PkiId, thatMsg.Membership.PkiId) {
		return MessageNoAction
	}
	return compareTimestamps(thisMsg.Timestamp, thatMsg.Timestamp)
}

func leaderInvalidationPolicy(thisMsg *LeadershipMessage, thatMsg *LeadershipMessage) InvalidationResult {
	if !bytes.Equal(thisMsg.PkiId, thatMsg.PkiId) {
		return MessageNoAction
	}
	return compareTimestamps(thisMsg.Timestamp, thatMsg.Timestamp)
}

func compareTimestamps(thisTS *PeerTime, thatTS *PeerTime) InvalidationResult {
	if thisTS.IncNum == thatTS.IncNum {
		if thisTS.SeqNum > thatTS.SeqNum {
			return MessageInvalidates
		}
		return MessageInvalidated
	}
	if thisTS.IncNum < thatTS.IncNum {
		return MessageInvalidated
	}
	return MessageInvalidates
}

func abs(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return b - a
}
