package gossip

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
	"xianhetian.com/tendermint/p2p/grpc/service"
)

//---------- Const -----------

const (
	handshakeTimeout = time.Second * time.Duration(10)
	defConnTimeout   = time.Second * time.Duration(2)
	defRecvBuffSize  = 20
	defSendBuffSize  = 20
)

const (
	BlockingSend    = blockingBehavior(true)
	NonBlockingSend = blockingBehavior(false)
)

//---------- Interface -----------

type stream interface {
	Send(envelope *Envelope) error
	Recv() (*Envelope, error)
	grpc.Stream
}

//---------- Struct -----------

type Communicate struct {
	sa            *OrgCryptoService
	tlsCerts      *TLSCertificates
	pubSub        *PubSub
	peerId        PeerIdType
	idMapper      *IdMapper
	connStore     *connectionStore
	PKIID         []byte
	deadEndpoints chan PKIidType
	msgPublisher  *ChannelDeMultiplexer
	lock          *sync.Mutex
	listener      net.Listener
	gSrv          *grpc.Server
	exitChan      chan struct{}
	stopWG        sync.WaitGroup
	subscriptions []chan *ReceivedMessage
	stopping      int32
}

type Connection struct {
	cancel       context.CancelFunc
	outBuff      chan *msgSending
	pkiID        PKIidType        // 远程端点的pkiID
	handler      handler          // 在接收消息时调用的方法
	conn         *grpc.ClientConn // gRPC与远程端点的连接
	clientStream Gossip_GossipStreamClient
	serverStream Gossip_GossipStreamServer
	stopFlag     int32
	stopChan     chan struct{}
	sync.RWMutex
}

type RemotePeer struct {
	Endpoint string
	PKIID    PKIidType
}

type TLSCertificates struct {
	TLSServerCert atomic.Value
	TLSClientCert atomic.Value
}

type sendResult struct {
	error
	RemotePeer
}

type connectionStore struct {
	isClosing        bool
	connFactory      *Communicate
	sync.RWMutex
	pki2Conn         map[string]*Connection
	destinationLocks map[string]*sync.Mutex
}

type msgSending struct {
	envelope *Envelope
	onErr    func(error)
}

type ackSendOperation struct {
	sendFunc   func(peer *RemotePeer, msg *SignedGossipMessage)
	waitForAck func(*RemotePeer) error
}

//---------- Define -----------

type AggregatedSendResults []sendResult

type blockingBehavior bool

type handler func(*SignedGossipMessage)

//---------- Struct Methods -----------

func (c *Communicate) GossipStream(stream Gossip_GossipStreamServer) error {
	if c.isStopping() {
		return fmt.Errorf("shutting down")
	}
	connInfo, err := c.authenticateRemotePeer(stream, false)
	if err != nil {
		return err
	}
	conn := c.connStore.onConnected(stream, connInfo)
	h := func(m *SignedGossipMessage) {
		c.msgPublisher.DeMultiplex(&ReceivedMessage{
			Conn:                conn,
			SignedGossipMessage: m,
			ConnInfo:            connInfo,
		})
	}
	conn.handler = interceptAcks(h, connInfo.ID, c.pubSub)
	defer func() {
		c.connStore.closeByPKIid(connInfo.ID)
		conn.close()
	}()
	return conn.serviceConnection()
}

func (c *Communicate) Ping(context.Context, *Empty) (*Empty, error) {
	return &Empty{}, nil
}

func (c *Communicate) Send(msg *SignedGossipMessage, peers ...*RemotePeer) {
	if c.isStopping() || len(peers) == 0 {
		return
	}
	for _, remotePeer := range peers {
		go func(peer *RemotePeer, msg *SignedGossipMessage) { c.sendToEndpoint(peer, msg, NonBlockingSend) }(remotePeer, msg)
	}
}

// 向远程Peer发送消息，等待MinAck的确认
func (c *Communicate) SendWithAck(msg *SignedGossipMessage, timeout time.Duration, minAck int, peers ...*RemotePeer) AggregatedSendResults {
	if len(peers) == 0 {
		return nil
	}
	var err error
	msg.Nonce = RandomUInt64()
	msg, err = msg.NoopSign()
	if c.isStopping() || err != nil {
		if err == nil {
			err = errors.New("comm is stopping")
		}
		var results []sendResult
		for _, p := range peers {
			results = append(results, sendResult{
				error:      err,
				RemotePeer: *p,
			})
		}
		return results
	}
	sndFunc := func(peer *RemotePeer, msg *SignedGossipMessage) {
		c.sendToEndpoint(peer, msg, BlockingSend)
	}
	subscriptions := make(map[string]func() error)
	for _, p := range peers {
		topic := fmt.Sprintf("%d %s", msg.Nonce, hex.EncodeToString(p.PKIID))
		sub := c.pubSub.Subscribe(topic, timeout)
		subscriptions[string(p.PKIID)] = func() error {
			msg, err := sub.Listen()
			if err != nil {
				return err
			}
			if msg, isAck := msg.(*Acknowledgement); !isAck {
				return fmt.Errorf("received a message of type %s, expected *proto.Acknowledgement", reflect.TypeOf(msg))
			} else if msg.Error != "" {
				return errors.New(msg.Error)
			}
			return nil
		}
	}
	waitForAck := func(p *RemotePeer) error {
		return subscriptions[string(p.PKIID)]()
	}
	ackOperation := &ackSendOperation{sendFunc: sndFunc, waitForAck: waitForAck}
	return ackOperation.send(msg, minAck, peers...)
}

// 探测远程节点是否响应
func (c *Communicate) Probe(remotePeer *RemotePeer) error {
	if c.isStopping() {
		return fmt.Errorf("stopping")
	}
	cp := service.NewClient()
	cp.Config.UseTLS = false
	_, err := cp.Conn(remotePeer.Endpoint)
	if err != nil {
		return err
	}
	cl := NewGossipClient(&cp.ClientConn)
	defer cp.Close()
	ctx, cancel := context.WithTimeout(context.Background(), defConnTimeout)
	defer cancel()
	_, err = cl.Ping(ctx, &Empty{})
	return err
}

// 返回远程Peer是否验证成功
func (c *Communicate) Handshake(remotePeer *RemotePeer) (PeerIdType, error) {
	cp := service.NewClient()
	cp.Config.UseTLS = false
	_, err := cp.Conn(remotePeer.Endpoint)
	if err != nil {
		return nil, err
	}
	defer cp.Close()
	cl := NewGossipClient(&cp.ClientConn)
	ctx, cancel := context.WithTimeout(context.Background(), defConnTimeout)
	defer cancel()
	if _, err := cl.Ping(ctx, &Empty{}); err != nil {
		return nil, err
	}
	ctx, cancel = context.WithTimeout(context.Background(), handshakeTimeout)
	defer cancel()
	stream, err := cl.GossipStream(ctx)
	if err != nil {
		return nil, err
	}
	connInfo, err := c.authenticateRemotePeer(stream, true)
	if err != nil {
		return nil, err
	}
	if len(remotePeer.PKIID) > 0 && !bytes.Equal(connInfo.ID, remotePeer.PKIID) {
		return nil, fmt.Errorf("PKI-ID of remote peer doesn't match expected PKI-ID")
	}
	return connInfo.Id, nil
}

func (c *Communicate) Receive(acceptor MessageAcceptor) <-chan *ReceivedMessage {
	genericChan := c.msgPublisher.AddChannel(acceptor)
	specificChan := make(chan *ReceivedMessage, 10)
	if c.isStopping() {
		return specificChan
	}
	c.lock.Lock()
	c.subscriptions = append(c.subscriptions, specificChan)
	c.lock.Unlock()
	c.stopWG.Add(1)
	go func() {
		defer c.stopWG.Done()
		for {
			select {
			case msg := <-genericChan:
				if msg == nil {
					return
				}
				select {
				case specificChan <- msg.(*ReceivedMessage):
				case <-c.exitChan:
					return
				}
			case <-c.exitChan:
				return
			}
		}
	}()
	return specificChan
}

func (c *Communicate) PresumedDead() <-chan PKIidType {
	return c.deadEndpoints
}

func (c *Communicate) CloseConn(peer *RemotePeer) {
	c.connStore.Lock()
	defer c.connStore.Unlock()
	if conn, exists := c.connStore.pki2Conn[string(peer.PKIID)]; exists {
		conn.close()
		delete(c.connStore.pki2Conn, string(conn.pkiID))
	}
}

func (c *Communicate) Stop() {
	if !atomic.CompareAndSwapInt32(&c.stopping, 0, int32(1)) {
		return
	}
	if c.gSrv != nil {
		c.gSrv.Stop()
	}
	if c.listener != nil {
		c.listener.Close()
	}
	c.connStore.Lock()
	c.connStore.isClosing = true
	var connections2Close []*Connection
	for _, conn := range c.connStore.pki2Conn {
		connections2Close = append(connections2Close, conn)
	}
	c.connStore.Unlock()
	wg := sync.WaitGroup{}
	for _, conn := range connections2Close {
		wg.Add(1)
		go func(conn *Connection) {
			c.connStore.closeByPKIid(conn.pkiID)
			wg.Done()
		}(conn)
	}
	wg.Wait()
	c.msgPublisher.Close()
	close(c.exitChan)
	c.stopWG.Wait()
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, ch := range c.subscriptions {
		close(ch)
	}
}

func (c *Communicate) createConnection(endpoint string, expectedPKIID PKIidType) (*Connection, error) {
	var err error
	var stream Gossip_GossipStreamClient
	var pkiID PKIidType
	var connInfo *ConnectionInfo
	if c.isStopping() {
		return nil, errors.New("Stopping")
	}
	cp := service.NewClient()
	cp.Config.UseTLS = false
	_, err = cp.Conn(endpoint)
	if err != nil {
		return nil, err
	}
	cl := NewGossipClient(&cp.ClientConn)
	ctx, cancel := context.WithTimeout(context.Background(), defConnTimeout)
	defer cancel()
	if _, err = cl.Ping(ctx, &Empty{}); err != nil {
		cp.ClientConn.Close()
		return nil, errors.WithStack(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	if stream, err = cl.GossipStream(ctx); err == nil {
		connInfo, err = c.authenticateRemotePeer(stream, true)
		if err != nil {
			cp.ClientConn.Close()
			cancel()
			return nil, errors.WithStack(err)
		}
		pkiID = connInfo.ID
		if expectedPKIID != nil && !bytes.Equal(pkiID, expectedPKIID) {
			actualOrg := c.sa.OrgByPeerId(connInfo.Id)
			ids, _ := c.idMapper.Get(expectedPKIID)
			oldOrg := c.sa.OrgByPeerId(ids)
			if !bytes.Equal(actualOrg, oldOrg) {
				cp.ClientConn.Close()
				cancel()
				return nil, errors.New("authentication failure")
			}
		}
		conn := newConnection(&cp.ClientConn, stream, nil)
		conn.pkiID = pkiID
		conn.cancel = cancel
		h := func(m *SignedGossipMessage) {
			c.msgPublisher.DeMultiplex(&ReceivedMessage{
				Conn:                conn,
				SignedGossipMessage: m,
				ConnInfo:            connInfo,
			})
		}
		conn.handler = interceptAcks(h, connInfo.ID, c.pubSub)
		return conn, nil
	}
	cp.ClientConn.Close()
	cancel()
	return nil, errors.WithStack(err)
}

func (c *Communicate) createConnectionMsg(pkiID PKIidType, certHash []byte, cert PeerIdType, signer Signer) (*SignedGossipMessage, error) {
	m := &GossipMessage{
		Tag:   GossipMessage_EMPTY,
		Nonce: 0,
		Content: &GossipMessage_Conn{
			Conn: &ConnEstablish{
				TlsCertHash: certHash,
				Identity:    cert,
				PkiId:       pkiID,
			},
		},
	}
	sMsg := &SignedGossipMessage{
		GossipMessage: m,
	}
	_, err := sMsg.Sign(signer)
	return sMsg, errors.WithStack(err)
}

func (c *Communicate) sendToEndpoint(peer *RemotePeer, msg *SignedGossipMessage, shouldBlock blockingBehavior) {
	if c.isStopping() {
		return
	}
	var err error
	conn, err := c.connStore.getConnection(peer)
	if err == nil {
		disConnectOnErr := func(err error) {
			c.disconnect(peer.PKIID)
		}
		conn.Send(msg, disConnectOnErr, shouldBlock)
		return
	}
	c.disconnect(peer.PKIID)
}

func (c *Communicate) isStopping() bool {
	return atomic.LoadInt32(&c.stopping) == int32(1)
}

func (c *Communicate) authenticateRemotePeer(stream stream, initiator bool) (*ConnectionInfo, error) {
	var remoteAddress string
	p, ok := peer.FromContext(stream.Context())
	if ok {
		if address := p.Addr; address != nil {
			remoteAddress = address.String()
		}
	}
	var err error
	var cMsg *SignedGossipMessage
	useTLS := c.tlsCerts != nil
	var selfCertHash []byte
	if useTLS {
		certReference := c.tlsCerts.TLSServerCert
		if initiator {
			certReference = c.tlsCerts.TLSClientCert
		}
		rawCert := certReference.Load().(*tls.Certificate).Certificate[0]
		if len(rawCert) != 0 {
			selfCertHash = ComputeSHA256(rawCert)
		}
	}
	signer := func(msg []byte) ([]byte, error) {
		return c.idMapper.Sign(msg)
	}
	cMsg, err = c.createConnectionMsg(c.PKIID, selfCertHash, c.peerId, signer)
	if err != nil {
		return nil, err
	}
	stream.Send(cMsg.Envelope)
	m, err := readWithTimeout(stream, GetDurationOrDefault("peer.gossip.connTimeout", defConnTimeout), remoteAddress)
	if err != nil {
		return nil, err
	}
	receivedMsg := m.GetConn()
	if receivedMsg == nil {
		return nil, fmt.Errorf("wrong type, receivedMsg is nil")
	}
	if receivedMsg.PkiId == nil {
		return nil, fmt.Errorf("no PKI-ID")
	}
	if err = c.idMapper.Put(receivedMsg.PkiId, receivedMsg.Identity); err != nil {
		return nil, err
	}
	connInfo := &ConnectionInfo{
		ID: receivedMsg.PkiId,
		Id: receivedMsg.Identity,
	}
	verifier := func(peerId []byte, signature, message []byte) error {
		pkiID := c.idMapper.GetPKIidOfCert(PeerIdType(peerId))
		return c.idMapper.Verify(pkiID, signature, message)
	}
	if err = m.Verify(receivedMsg.Identity, verifier); err != nil {
		return nil, err
	}
	return connInfo, nil
}

func (c *Communicate) disconnect(pkiID PKIidType) {
	if c.isStopping() {
		return
	}
	c.deadEndpoints <- pkiID
	c.connStore.closeByPKIid(pkiID)
}

func (conn *Connection) Send(msg *SignedGossipMessage, onErr func(error), shouldBlock blockingBehavior) {
	if conn.toDie() {
		return
	}
	m := &msgSending{
		envelope: msg.Envelope,
		onErr:    onErr,
	}
	if len(conn.outBuff) == cap(conn.outBuff) && !shouldBlock {
		return
	}
	conn.outBuff <- m
}

func (conn *Connection) close() {
	if conn.toDie() {
		return
	}
	amIFirst := atomic.CompareAndSwapInt32(&conn.stopFlag, int32(0), int32(1))
	if !amIFirst {
		return
	}
	conn.stopChan <- struct{}{}
	for len(conn.outBuff) > 0 {
		<-conn.outBuff
	}
	conn.Lock()
	defer conn.Unlock()
	if conn.clientStream != nil {
		conn.clientStream.CloseSend()
	}
	if conn.conn != nil {
		conn.conn.Close()
	}
	if conn.cancel != nil {
		conn.cancel()
	}
}

func (conn *Connection) toDie() bool {
	return atomic.LoadInt32(&(conn.stopFlag)) == int32(1)
}

func (conn *Connection) serviceConnection() error {
	errChan := make(chan error, 1)
	msgChan := make(chan *SignedGossipMessage, GetIntOrDefault("peer.gossip.recvBuffSize", defRecvBuffSize))
	quit := make(chan struct{})
	go conn.readFromStream(errChan, quit, msgChan)
	go conn.writeToStream()
	for !conn.toDie() {
		select {
		case stop := <-conn.stopChan:
			conn.stopChan <- stop
			return nil
		case err := <-errChan:
			return err
		case msg := <-msgChan:
			conn.handler(msg)
		}
	}
	return nil
}

func (conn *Connection) getStream() stream {
	conn.Lock()
	defer conn.Unlock()
	if conn.clientStream != nil {
		return conn.clientStream
	}
	if conn.serverStream != nil {
		return conn.serverStream
	}
	return nil
}

func (conn *Connection) writeToStream() {
	for !conn.toDie() {
		stream := conn.getStream()
		if stream == nil {
			return
		}
		select {
		case m := <-conn.outBuff:
			if err := stream.Send(m.envelope); err != nil {
				go m.onErr(err)
				return
			}
		case stop := <-conn.stopChan:
			conn.stopChan <- stop
			return
		}
	}
}

func (conn *Connection) readFromStream(errChan chan error, quit chan struct{}, msgChan chan *SignedGossipMessage) {
	for !conn.toDie() {
		stream := conn.getStream()
		if stream == nil {
			errChan <- fmt.Errorf("stream is nil")
			return
		}
		envelope, err := stream.Recv()
		if conn.toDie() {
			return
		}
		if err != nil {
			errChan <- err
			return
		}
		msg, err := envelope.ToGossipMessage()
		if err != nil {
			errChan <- err
		}
		select {
		case msgChan <- msg:
		case <-quit:
			return
		}
	}
}

func (cs *connectionStore) getConnection(peer *RemotePeer) (*Connection, error) {
	cs.RLock()
	isClosing := cs.isClosing
	cs.RUnlock()
	if isClosing {
		return nil, fmt.Errorf("shutting down")
	}
	pkiID := peer.PKIID
	endpoint := peer.Endpoint
	cs.Lock()
	destinationLock, hasConnected := cs.destinationLocks[string(pkiID)]
	if !hasConnected {
		destinationLock = &sync.Mutex{}
		cs.destinationLocks[string(pkiID)] = destinationLock
	}
	cs.Unlock()
	destinationLock.Lock()
	cs.RLock()
	conn, exists := cs.pki2Conn[string(pkiID)]
	if exists {
		cs.RUnlock()
		destinationLock.Unlock()
		return conn, nil
	}
	cs.RUnlock()
	createdConnection, err := cs.connFactory.createConnection(endpoint, pkiID)
	if err != nil {
		return nil, err
	}
	destinationLock.Unlock()
	cs.RLock()
	isClosing = cs.isClosing
	cs.RUnlock()
	if isClosing {
		return nil, fmt.Errorf("ConnStore is closing")
	}
	cs.Lock()
	delete(cs.destinationLocks, string(pkiID))
	defer cs.Unlock()
	conn, exists = cs.pki2Conn[string(pkiID)]
	if exists {
		if createdConnection != nil {
			createdConnection.close()
		}
		return conn, nil
	}
	if err != nil {
		return nil, err
	}
	conn = createdConnection
	cs.pki2Conn[string(createdConnection.pkiID)] = conn
	go conn.serviceConnection()
	return conn, nil
}

func (cs *connectionStore) onConnected(serverStream Gossip_GossipStreamServer, connInfo *ConnectionInfo) *Connection {
	cs.Lock()
	defer cs.Unlock()
	if c, exists := cs.pki2Conn[string(connInfo.ID)]; exists {
		c.close()
	}
	conn := newConnection(nil, nil, serverStream)
	conn.pkiID = connInfo.ID
	cs.pki2Conn[string(connInfo.ID)] = conn
	return conn
}

func (cs *connectionStore) closeByPKIid(pkiID PKIidType) {
	cs.Lock()
	defer cs.Unlock()
	if conn, exists := cs.pki2Conn[string(pkiID)]; exists {
		conn.close()
		delete(cs.pki2Conn, string(pkiID))
	}
}

func (aso *ackSendOperation) send(msg *SignedGossipMessage, minAckNum int, peers ...*RemotePeer) []sendResult {
	successAcks := 0
	var results []sendResult
	acks := make(chan sendResult, len(peers))
	for _, p := range peers {
		go func(p *RemotePeer) {
			aso.sendFunc(p, msg)
			err := aso.waitForAck(p)
			acks <- sendResult{RemotePeer: *p, error: err}
		}(p)
	}
	for {
		ack := <-acks
		results = append(results, sendResult{
			error:      ack.error,
			RemotePeer: ack.RemotePeer,
		})
		if ack.error == nil {
			successAcks++
		}
		if successAcks == minAckNum || len(results) == len(peers) {
			break
		}
	}
	return results
}

// 成功确认的次数
func (ar AggregatedSendResults) AckCount() int {
	var c int
	for _, ack := range ar {
		if ack.error == nil {
			c++
		}
	}
	return c
}

func (ar AggregatedSendResults) String() string {
	errMap := map[string]int{}
	for _, ack := range ar {
		if ack.error == nil {
			continue
		}
		errMap[ack.Error()]++
	}
	ackCount := ar.AckCount()
	output := map[string]interface{}{}
	if ackCount > 0 {
		output["successes"] = ackCount
	}
	if ackCount < len(ar) {
		output["failures"] = errMap
	}
	b, _ := json.Marshal(output)
	return string(b)
}

func (sr sendResult) Error() string {
	if sr.error != nil {
		return sr.error.Error()
	}
	return ""
}

//---------- Public Methods -----------

func NewCommInstanceWithServer(port int, idMapper *IdMapper, peerId PeerIdType,
	pem []byte, key []byte, sa *OrgCryptoService) (*Communicate, error) {
	s := service.NewServer()
	s.Config.UseTLS = false
	var ll string
	var certs *TLSCertificates
	if port > 0 {
		pem, key, ll, certs = createGRPCLayer(port)
	}
	if len(pem) > 0 && len(key) > 0 {
		s.Config.UseTLS = false
		s.Config.Addr = ll
	}
	s.Listen()
	commInst := &Communicate{
		sa:            sa,
		pubSub:        NewPubSub(),
		PKIID:         idMapper.GetPKIidOfCert(peerId),
		idMapper:      idMapper,
		peerId:        peerId,
		listener:      s.Listener,
		gSrv:          &s.Server,
		msgPublisher:  NewChannelDemultiplexer(),
		lock:          &sync.Mutex{},
		deadEndpoints: make(chan PKIidType, 100),
		stopping:      int32(0),
		exitChan:      make(chan struct{}),
		subscriptions: make([]chan *ReceivedMessage, 0),
		tlsCerts:      certs,
	}
	commInst.connStore = &connectionStore{
		connFactory:      commInst,
		isClosing:        false,
		pki2Conn:         make(map[string]*Connection),
		destinationLocks: make(map[string]*sync.Mutex),
	}
	if port > 0 {
		commInst.stopWG.Add(1)
		RegisterGossipServer(&s.Server, commInst)
		go func() {
			defer commInst.stopWG.Done()
			err := s.Serve(s.Listener)
			if err != nil {
				panic(err)
			}
		}()
	}
	return commInst, nil
}

//---------- Private Methods -----------

func newConnection(c *grpc.ClientConn, cs Gossip_GossipStreamClient, ss Gossip_GossipStreamServer) *Connection {
	connection := &Connection{
		outBuff:      make(chan *msgSending, GetIntOrDefault("peer.gossip.sendBuffSize", defSendBuffSize)),
		conn:         c,
		clientStream: cs,
		serverStream: ss,
		stopFlag:     int32(0),
		stopChan:     make(chan struct{}, 1),
	}
	return connection
}

func interceptAcks(nextHandler handler, remotePeerID PKIidType, pubSub *PubSub) func(*SignedGossipMessage) {
	return func(m *SignedGossipMessage) {
		if m.GetAck() != nil {
			topic := fmt.Sprintf("%d %s", m.Nonce, hex.EncodeToString(remotePeerID))
			pubSub.Publish(topic, m.GetAck())
			return
		}
		nextHandler(m)
	}
}

func readWithTimeout(stream interface{}, timeout time.Duration, address string) (*SignedGossipMessage, error) {
	incChan := make(chan *SignedGossipMessage, 1)
	errChan := make(chan error, 1)
	go func() {
		if srvStr, isServerStr := stream.(Gossip_GossipStreamServer); isServerStr {
			if m, err := srvStr.Recv(); err == nil {
				msg, err := m.ToGossipMessage()
				if err != nil {
					errChan <- err
					return
				}
				incChan <- msg
			}
		} else if clStr, isClientStr := stream.(Gossip_GossipStreamClient); isClientStr {
			if m, err := clStr.Recv(); err == nil {
				msg, err := m.ToGossipMessage()
				if err != nil {
					errChan <- err
					return
				}
				incChan <- msg
			}
		}
	}()
	select {
	case <-time.NewTicker(timeout).C:
		return nil, errors.Errorf("Timed out waiting for connection message from %s", address)
	case m := <-incChan:
		return m, nil
	case err := <-errChan:
		return nil, errors.WithStack(err)
	}
}

func createGRPCLayer(port int) ([]byte, []byte, string, *TLSCertificates) {
	s := service.NewServer()
	c := service.NewClient()
	serverPem, serverKey := s.Config.Certificate, s.Config.Key
	clientPem, clientKey := c.Config.Certificate, c.Config.Key
	serverCert, _ := tls.X509KeyPair(serverPem, serverKey)
	clientCert, _ := tls.X509KeyPair(clientPem, clientKey)
	certs := &TLSCertificates{}
	certs.TLSServerCert.Store(&serverCert)
	certs.TLSClientCert.Store(&clientCert)
	return serverPem, serverKey, fmt.Sprintf("%s:%d", "", port), certs
}
