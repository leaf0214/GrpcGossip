package gossip

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

//---------- Var -----------

var (
	usageThreshold = time.Hour
)

//---------- Struct -----------

type IdMapper struct {
	onPurge    purgeTrigger
	mcs        *NaiveMessageCryptoService
	pkiID2Cert map[string]*storedId
	sync.RWMutex
	stopChan   chan struct{}
	sync.Once
	selfPKIID  string
}

type storedId struct {
	pkiID           PKIidType
	lastAccessTime  int64
	peerId          PeerIdType
	expirationTimer *time.Timer
}

//---------- Define -----------

type purgeTrigger func(PKIidType, PeerIdType)

//---------- Struct Methods -----------

func (is *IdMapper) Put(pkiID PKIidType, idType PeerIdType) error {
	if pkiID == nil {
		return errors.New("PKIID is nil")
	}
	if idType == nil {
		return errors.New("idType is nil")
	}
	expirationDate, err := is.mcs.Expiration(idType)
	if err != nil {
		return errors.Wrap(err, "failed classifying idType")
	}
	if err := is.mcs.ValidateIdType(idType); err != nil {
		return err
	}
	id := is.mcs.GetPKIidOfCert(idType)
	if !bytes.Equal(pkiID, id) {
		return errors.New("idType doesn't match the computed pkiID")
	}
	is.Lock()
	defer is.Unlock()
	if _, exists := is.pkiID2Cert[string(pkiID)]; exists {
		return nil
	}
	var expirationTimer *time.Timer
	if !expirationDate.IsZero() {
		if time.Now().After(expirationDate) {
			return errors.New("idType expired")
		}
		timeToLive := expirationDate.Add(time.Millisecond).Sub(time.Now())
		expirationTimer = time.AfterFunc(timeToLive, func() {
			is.delete(pkiID, idType)
		})
	}
	is.pkiID2Cert[string(id)] = &storedId{
		pkiID:           pkiID,
		lastAccessTime:  time.Now().UnixNano(),
		peerId:          idType,
		expirationTimer: expirationTimer,
	}
	return nil
}

func (is *IdMapper) Get(pkiID PKIidType) (PeerIdType, error) {
	is.RLock()
	defer is.RUnlock()
	storedId, exists := is.pkiID2Cert[string(pkiID)]
	if !exists {
		return nil, errors.New("PKIID wasn't found")
	}
	return storedId.fetchId(), nil
}

func (is *IdMapper) Sign(msg []byte) ([]byte, error) {
	return is.mcs.Sign(msg)
}

func (is *IdMapper) Verify(vkID, signature, message []byte) error {
	cert, err := is.Get(vkID)
	if err != nil {
		return err
	}
	return is.mcs.Verify(cert, signature, message)
}

func (is *IdMapper) GetPKIidOfCert(id PeerIdType) PKIidType {
	return is.mcs.GetPKIidOfCert(id)
}

func (is *IdMapper) SuspectPeers(isSuspected PeerSuspector) {
	for _, id := range is.validateIds(isSuspected) {
		if id.expirationTimer != nil {
			id.expirationTimer.Stop()
		}
		is.delete(id.pkiID, id.peerId)
	}
}

func (is *IdMapper) Stop() {
	is.Once.Do(func() {
		is.stopChan <- struct{}{}
	})
}

func (is *IdMapper) periodicalPurgeUnusedIds() {
	usageTh := time.Duration(atomic.LoadInt64((*int64)(&usageThreshold)))
	for {
		select {
		case <-is.stopChan:
			return
		case <-time.After(usageTh / 10):
			is.SuspectPeers(func(_ PeerIdType) bool {
				return false
			})
		}
	}
}

func (is *IdMapper) validateIds(isSuspected PeerSuspector) []*storedId {
	now := time.Now()
	usageTh := time.Duration(atomic.LoadInt64((*int64)(&usageThreshold)))
	is.RLock()
	defer is.RUnlock()
	var revokedIds []*storedId
	for pkiID, storedId := range is.pkiID2Cert {
		if pkiID != is.selfPKIID && storedId.fetchLastAccessTime().Add(usageTh).Before(now) {
			revokedIds = append(revokedIds, storedId)
			continue
		}
		if !isSuspected(storedId.peerId) {
			continue
		}
		if err := is.mcs.ValidateIdType(storedId.fetchId()); err != nil {
			revokedIds = append(revokedIds, storedId)
		}
	}
	return revokedIds
}

func (is *IdMapper) delete(pkiID PKIidType, id PeerIdType) {
	is.Lock()
	defer is.Unlock()
	is.onPurge(pkiID, id)
	delete(is.pkiID2Cert, string(pkiID))
}

func (si *storedId) fetchId() PeerIdType {
	atomic.StoreInt64(&si.lastAccessTime, time.Now().UnixNano())
	return si.peerId
}

func (si *storedId) fetchLastAccessTime() time.Time {
	return time.Unix(0, atomic.LoadInt64(&si.lastAccessTime))
}

//---------- Public Methods -----------

func NewIdMapper(mcs *NaiveMessageCryptoService, selfId PeerIdType, onPurge purgeTrigger) *IdMapper {
	selfPKIID := mcs.GetPKIidOfCert(selfId)
	idMapper := &IdMapper{
		onPurge:    onPurge,
		mcs:        mcs,
		pkiID2Cert: make(map[string]*storedId),
		stopChan:   make(chan struct{}),
		selfPKIID:  string(selfPKIID),
	}
	idMapper.Put(selfPKIID, selfId)
	go idMapper.periodicalPurgeUnusedIds()
	return idMapper
}
