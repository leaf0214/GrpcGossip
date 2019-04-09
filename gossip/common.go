package gossip

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io"
	"math/big"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

//---------- Const -----------

const (
	MessageNoAction    InvalidationResult = iota
	MessageInvalidates
	MessageInvalidated
)

const (
	subscriptionBuffSize = 50
)

//---------- Var -----------

var (
	viperLock       sync.RWMutex
	ExpirationTimes = map[string]time.Time{}
)

//---------- Struct -----------

type NaiveMessageCryptoService struct {
	sync.RWMutex
	AllowedPkiIDS map[string]struct{}
	RevokedPkiIDS map[string]struct{}
}

type PubSub struct {
	sync.RWMutex
	subscriptions map[string]*Set
}

type Subscription struct {
	top string
	ttl time.Duration
	c   chan interface{}
}

type Set struct {
	items map[interface{}]struct{}
	lock  *sync.RWMutex
}

//---------- Define -----------

type OrgIdType []byte

type MessageAcceptor func(interface{}) bool

type PeerSuspector func(PeerIdType) bool

type PeerIdType []byte

type ChainID []byte

type PKIidType []byte

type InvalidationResult int

type MessageReplacingPolicy func(interface{}, interface{}) InvalidationResult

type Equals func(interface{}, interface{}) bool

//---------- Struct Methods -----------

func (cs *NaiveMessageCryptoService) OrgByPeerId(PeerIdType) OrgIdType {
	return nil
}

func (cs *NaiveMessageCryptoService) GetPKIidOfCert(peerId PeerIdType) PKIidType {
	return PKIidType(peerId)
}

func (cs *NaiveMessageCryptoService) Sign(msg []byte) ([]byte, error) {
	sig := make([]byte, len(msg))
	copy(sig, msg)
	return sig, nil
}

func (cs *NaiveMessageCryptoService) Verify(peerId PeerIdType, signature, message []byte) error {
	equal := bytes.Equal(signature, message)
	if !equal {
		return fmt.Errorf("wrong signature:%v, %v", signature, message)
	}
	return nil
}

func (cs *NaiveMessageCryptoService) VerifyBlock(chainID ChainID, seqNum uint64, signedBlock []byte) error {
	return nil
}

func (cs *NaiveMessageCryptoService) VerifyByChannel(_ ChainID, id PeerIdType, _, _ []byte) error {
	if cs.AllowedPkiIDS == nil {
		return nil
	}
	if _, allowed := cs.AllowedPkiIDS[string(id)]; allowed {
		return nil
	}
	return errors.New("forbidden")
}

func (cs *NaiveMessageCryptoService) ValidateIdType(peerId PeerIdType) error {
	cs.RLock()
	defer cs.RUnlock()
	if cs.RevokedPkiIDS == nil {
		return nil
	}
	if _, revoked := cs.RevokedPkiIDS[string(cs.GetPKIidOfCert(peerId))]; revoked {
		return errors.New("revoked")
	}
	return nil
}

func (cs *NaiveMessageCryptoService) Expiration(peerId PeerIdType) (time.Time, error) {
	if exp, exists := ExpirationTimes[string(peerId)]; exists {
		return exp, nil
	}
	return time.Now().Add(time.Hour), nil
}

func (cs *NaiveMessageCryptoService) Revoke(pkiID PKIidType) {
	cs.Lock()
	defer cs.Unlock()
	if cs.RevokedPkiIDS == nil {
		cs.RevokedPkiIDS = map[string]struct{}{}
	}
	cs.RevokedPkiIDS[string(pkiID)] = struct{}{}
}

func (ps *PubSub) Publish(topic string, item interface{}) error {
	ps.RLock()
	defer ps.RUnlock()
	s, subscribed := ps.subscriptions[topic]
	if !subscribed {
		return errors.New("no subscribers")
	}
	for _, sub := range s.ToArray() {
		c := sub.(*Subscription).c
		if len(c) == subscriptionBuffSize {
			continue
		}
		c <- item
	}
	return nil
}

func (ps *PubSub) Subscribe(topic string, ttl time.Duration) *Subscription {
	sub := &Subscription{
		top: topic,
		ttl: ttl,
		c:   make(chan interface{}, subscriptionBuffSize),
	}
	ps.Lock()
	s, exists := ps.subscriptions[topic]
	if !exists {
		s = NewSet()
		ps.subscriptions[topic] = s
	}
	ps.Unlock()
	s.Add(sub)
	time.AfterFunc(ttl, func() {
		ps.Lock()
		defer ps.Unlock()
		ps.subscriptions[sub.top].Remove(sub)
		if ps.subscriptions[sub.top].Size() != 0 {
			return
		}
		delete(ps.subscriptions, sub.top)
	})
	return sub
}

func (s *Subscription) Listen() (interface{}, error) {
	select {
	case <-time.After(s.ttl):
		return nil, errors.New("timed out")
	case item := <-s.c:
		return item, nil
	}
}

func (p PKIidType) IsNotSameFilter(that PKIidType) bool {
	return !bytes.Equal(p, that)
}

func (s *Set) Add(item interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items[item] = struct{}{}
}

func (s *Set) Exists(item interface{}) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, exists := s.items[item]
	return exists
}

func (s *Set) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.items)
}

func (s *Set) ToArray() []interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	a := make([]interface{}, len(s.items))
	i := 0
	for item := range s.items {
		a[i] = item
		i++
	}
	return a
}

func (s *Set) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items = make(map[interface{}]struct{})
}

func (s *Set) Remove(item interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.items, item)
}

//---------- Public Methods -----------

func NewPubSub() *PubSub {
	return &PubSub{subscriptions: make(map[string]*Set)}
}

func NewSet() *Set {
	return &Set{lock: &sync.RWMutex{}, items: make(map[interface{}]struct{})}
}

func SetVal(key string, val interface{}) {
	viperLock.Lock()
	defer viperLock.Unlock()
	viper.Set(key, val)
}

func SetDigestWaitTime(time time.Duration) {
	viper.Set("peer.gossip.digestWaitTime", time)
}

func SetRequestWaitTime(time time.Duration) {
	viper.Set("peer.gossip.requestWaitTime", time)
}

func SetResponseWaitTime(time time.Duration) {
	viper.Set("peer.gossip.responseWaitTime", time)
}

func SetAliveExpirationCheckInterval(interval time.Duration) {
	AliveExpirationCheckInterval = interval
}

func SetAliveExpirationTimeout(timeout time.Duration) {
	SetVal("peer.gossip.aliveExpirationTimeout", timeout)
	AliveExpirationCheckInterval = time.Duration(timeout / 10)
}

func SetAliveTimeInterval(interval time.Duration) {
	SetVal("peer.gossip.aliveTimeInterval", interval)
}

func SetReconnectInterval(interval time.Duration) {
	SetVal("peer.gossip.reconnectInterval", interval)
}

func SetMaxConnAttempts(attempts int) {
	MaxConnectionAttempts = attempts
}

func GetIntOrDefault(key string, defVal int) int {
	viperLock.RLock()
	defer viperLock.RUnlock()
	if val := viper.GetInt(key); val != 0 {
		return val
	}
	return defVal
}

func GetDurationOrDefault(key string, defVal time.Duration) time.Duration {
	viperLock.RLock()
	defer viperLock.RUnlock()
	if val := viper.GetDuration(key); val != 0 {
		return val
	}
	return defVal
}

func GetRandomIndices(indiceCount, highestIndex int) []int {
	if highestIndex+1 < indiceCount {
		return nil
	}
	indices := make([]int, 0)
	if highestIndex+1 == indiceCount {
		for i := 0; i < indiceCount; i++ {
			indices = append(indices, i)
		}
		return indices
	}
	for len(indices) < indiceCount {
		n := RandomInt(highestIndex + 1)
		if IndexInSlice(indices, n, numbericEqual) != -1 {
			continue
		}
		indices = append(indices, n)
	}
	return indices
}

func GenerateMAC(pkiID PKIidType, channelID ChainID) []byte {
	preImage := append([]byte(pkiID), []byte(channelID)...)
	return ComputeSHA256(preImage)
}

func ComputeSHA256(data []byte) (hash []byte) {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func IndexInSlice(array interface{}, o interface{}, equals Equals) int {
	arr := reflect.ValueOf(array)
	for i := 0; i < arr.Len(); i++ {
		if equals(arr.Index(i).Interface(), o) {
			return i
		}
	}
	return -1
}

func RandomInt(n int) int {
	m := int(RandomUInt64()) % n
	if m < 0 {
		return n + m
	}
	return m
}

func RandomUInt64() uint64 {
	b := make([]byte, 8)
	_, err := io.ReadFull(cryptorand.Reader, b)
	if err == nil {
		n := new(big.Int)
		return n.SetBytes(b).Uint64()
	}
	rand.Seed(rand.Int63())
	return uint64(rand.Int63())
}

func BytesToStrings(bytes [][]byte) []string {
	strings := make([]string, len(bytes))
	for i, b := range bytes {
		strings[i] = string(b)
	}
	return strings
}

func StringsToBytes(strings []string) [][]byte {
	b := make([][]byte, len(strings))
	for i, str := range strings {
		b[i] = []byte(str)
	}
	return b
}

//---------- Private Methods -----------

func numbericEqual(a interface{}, b interface{}) bool {
	return a.(int) == b.(int)
}
