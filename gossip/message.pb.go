package gossip

import "github.com/golang/protobuf/proto"
import "fmt"
import "math"

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc" // finish
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type PullMsgType int32

const (
	PullMsgType_UNDEFINED    PullMsgType = 0
	PullMsgType_BLOCK_MSG    PullMsgType = 1
	PullMsgType_IDENTITY_MSG PullMsgType = 2
)

var PullMsgType_name = map[int32]string{
	0: "UNDEFINED",
	1: "BLOCK_MSG",
	2: "IDENTITY_MSG",
}
var PullMsgType_value = map[string]int32{
	"UNDEFINED":    0,
	"BLOCK_MSG":    1,
	"IDENTITY_MSG": 2,
}

func (x PullMsgType) String() string {
	return proto.EnumName(PullMsgType_name, int32(x))
}
func (PullMsgType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{0}
}

type GossipMessage_Tag int32

const (
	GossipMessage_UNDEFINED    GossipMessage_Tag = 0
	GossipMessage_EMPTY        GossipMessage_Tag = 1
	GossipMessage_ORG_ONLY     GossipMessage_Tag = 2
	GossipMessage_CHAN_ONLY    GossipMessage_Tag = 3
	GossipMessage_CHAN_AND_ORG GossipMessage_Tag = 4
	GossipMessage_CHAN_OR_ORG  GossipMessage_Tag = 5
)

var GossipMessage_Tag_name = map[int32]string{
	0: "UNDEFINED",
	1: "EMPTY",
	2: "ORG_ONLY",
	3: "CHAN_ONLY",
	4: "CHAN_AND_ORG",
	5: "CHAN_OR_ORG",
}
var GossipMessage_Tag_value = map[string]int32{
	"UNDEFINED":    0,
	"EMPTY":        1,
	"ORG_ONLY":     2,
	"CHAN_ONLY":    3,
	"CHAN_AND_ORG": 4,
	"CHAN_OR_ORG":  5,
}

func (x GossipMessage_Tag) String() string {
	return proto.EnumName(GossipMessage_Tag_name, int32(x))
}
func (GossipMessage_Tag) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{3, 0}
}

// Envelope contains a marshalled
// GossipMessage and a signature over it.
// It may also contain a SecretEnvelope
// which is a marshalled Secret
type Envelope struct {
	Payload              []byte          `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	Signature            []byte          `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	SecretEnvelope       *SecretEnvelope `protobuf:"bytes,3,opt,name=secret_envelope,json=secretEnvelope,proto3" json:"secret_envelope,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (e *Envelope) Reset()         { *e = Envelope{} }
func (e *Envelope) String() string { return proto.CompactTextString(e) }
func (*Envelope) ProtoMessage()    {}
func (*Envelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{0}
}
func (e *Envelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Envelope.Unmarshal(e, b)
}
func (e *Envelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Envelope.Marshal(b, e, deterministic)
}
func (e *Envelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Envelope.Merge(e, src)
}
func (e *Envelope) XXX_Size() int {
	return xxx_messageInfo_Envelope.Size(e)
}
func (e *Envelope) XXX_DiscardUnknown() {
	xxx_messageInfo_Envelope.DiscardUnknown(e)
}

var xxx_messageInfo_Envelope proto.InternalMessageInfo

func (e *Envelope) GetPayload() []byte {
	if e != nil {
		return e.Payload
	}
	return nil
}

func (e *Envelope) GetSignature() []byte {
	if e != nil {
		return e.Signature
	}
	return nil
}

func (e *Envelope) GetSecretEnvelope() *SecretEnvelope {
	if e != nil {
		return e.SecretEnvelope
	}
	return nil
}

// SecretEnvelope is a marshalled Secret
// and a signature over it.
// The signature should be validated by the peer
// that signed the Envelope the SecretEnvelope
// came with
type SecretEnvelope struct {
	Payload              []byte   `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	Signature            []byte   `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (se *SecretEnvelope) Reset()         { *se = SecretEnvelope{} }
func (se *SecretEnvelope) String() string { return proto.CompactTextString(se) }
func (*SecretEnvelope) ProtoMessage()     {}
func (*SecretEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{1}
}
func (se *SecretEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SecretEnvelope.Unmarshal(se, b)
}
func (se *SecretEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SecretEnvelope.Marshal(b, se, deterministic)
}
func (se *SecretEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SecretEnvelope.Merge(se, src)
}
func (se *SecretEnvelope) XXX_Size() int {
	return xxx_messageInfo_SecretEnvelope.Size(se)
}
func (se *SecretEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_SecretEnvelope.DiscardUnknown(se)
}

var xxx_messageInfo_SecretEnvelope proto.InternalMessageInfo

func (se *SecretEnvelope) GetPayload() []byte {
	if se != nil {
		return se.Payload
	}
	return nil
}

func (se *SecretEnvelope) GetSignature() []byte {
	if se != nil {
		return se.Signature
	}
	return nil
}

// Secret is an entity that might be omitted
// from an Envelope when the remote peer that is receiving
// the Envelope shouldn't know the secret's content.
type Secret struct {
	// Types that are valid to be assigned to Content:
	//	*Secret_InternalEndpoint
	Content              isSecret_Content `protobuf_oneof:"content"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Secret) Reset()         { *m = Secret{} }
func (m *Secret) String() string { return proto.CompactTextString(m) }
func (*Secret) ProtoMessage()    {}
func (*Secret) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{2}
}
func (m *Secret) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Secret.Unmarshal(m, b)
}
func (m *Secret) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Secret.Marshal(b, m, deterministic)
}
func (dst *Secret) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Secret.Merge(dst, src)
}
func (m *Secret) XXX_Size() int {
	return xxx_messageInfo_Secret.Size(m)
}
func (m *Secret) XXX_DiscardUnknown() {
	xxx_messageInfo_Secret.DiscardUnknown(m)
}

var xxx_messageInfo_Secret proto.InternalMessageInfo

type isSecret_Content interface {
	isSecret_Content()
}

type Secret_InternalEndpoint struct {
	InternalEndpoint string `protobuf:"bytes,1,opt,name=internalEndpoint,proto3,oneof"`
}

func (*Secret_InternalEndpoint) isSecret_Content() {}

func (m *Secret) GetContent() isSecret_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *Secret) GetInternalEndpoint() string {
	if x, ok := m.GetContent().(*Secret_InternalEndpoint); ok {
		return x.InternalEndpoint
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Secret) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Secret_OneofMarshaler, _Secret_OneofUnmarshaler, _Secret_OneofSizer, []interface{}{
		(*Secret_InternalEndpoint)(nil),
	}
}

func _Secret_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Secret)
	// content
	switch x := m.Content.(type) {
	case *Secret_InternalEndpoint:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.InternalEndpoint)
	case nil:
	default:
		return fmt.Errorf("Secret.Content has unexpected type %T", x)
	}
	return nil
}

func _Secret_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Secret)
	switch tag {
	case 1: // content.internalEndpoint
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.Content = &Secret_InternalEndpoint{x}
		return true, err
	default:
		return false, nil
	}
}

func _Secret_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Secret)
	// content
	switch x := m.Content.(type) {
	case *Secret_InternalEndpoint:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.InternalEndpoint)))
		n += len(x.InternalEndpoint)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// GossipMessage defines the message sent in a gossip network
type GossipMessage struct {
	// used mainly for testing, but will might be used in the future
	// for ensuring message delivery by acking
	Nonce uint64 `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	// The channel of the message.
	// Some GossipMessages may set this to nil, because
	// they are cross-channels but some may not
	Channel []byte `protobuf:"bytes,2,opt,name=channel,proto3" json:"channel,omitempty"`
	// determines to which peers it is allowed
	// to forward the message
	Tag GossipMessage_Tag `protobuf:"varint,3,opt,name=tag,proto3,enum=gossip.GossipMessage_Tag" json:"tag,omitempty"`
	// Types that are valid to be assigned to Content:
	//	*GossipMessage_AliveMsg
	//	*GossipMessage_MemReq
	//	*GossipMessage_MemRes
	//	*GossipMessage_DataMsg
	//	*GossipMessage_Hello
	//	*GossipMessage_DataDig
	//	*GossipMessage_DataReq
	//	*GossipMessage_DataUpdate
	//	*GossipMessage_Empty
	//	*GossipMessage_Conn
	//	*GossipMessage_StateInfo
	//	*GossipMessage_StateSnapshot
	//	*GossipMessage_StateInfoPullReq
	//	*GossipMessage_StateRequest
	//	*GossipMessage_StateResponse
	//	*GossipMessage_LeadershipMsg
	//	*GossipMessage_PeerIdentity
	//	*GossipMessage_Ack
	//	*GossipMessage_PrivateReq
	//	*GossipMessage_PrivateRes
	//	*GossipMessage_PrivateData
	Content              isGossipMessage_Content `protobuf_oneof:"content"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *GossipMessage) Reset()         { *m = GossipMessage{} }
func (m *GossipMessage) String() string { return proto.CompactTextString(m) }
func (*GossipMessage) ProtoMessage()    {}
func (*GossipMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{3}
}
func (m *GossipMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GossipMessage.Unmarshal(m, b)
}
func (m *GossipMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GossipMessage.Marshal(b, m, deterministic)
}
func (m *GossipMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GossipMessage.Merge(m, src)
}
func (m *GossipMessage) XXX_Size() int {
	return xxx_messageInfo_GossipMessage.Size(m)
}
func (m *GossipMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_GossipMessage.DiscardUnknown(m)
}

var xxx_messageInfo_GossipMessage proto.InternalMessageInfo

func (m *GossipMessage) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *GossipMessage) GetChannel() []byte {
	if m != nil {
		return m.Channel
	}
	return nil
}

func (m *GossipMessage) GetTag() GossipMessage_Tag {
	if m != nil {
		return m.Tag
	}
	return GossipMessage_UNDEFINED
}

type isGossipMessage_Content interface {
	isGossipMessage_Content()
}

type GossipMessage_AliveMsg struct {
	AliveMsg *AliveMessage `protobuf:"bytes,5,opt,name=alive_msg,json=aliveMsg,proto3,oneof"`
}

type GossipMessage_MemReq struct {
	MemReq *MembershipRequest `protobuf:"bytes,6,opt,name=mem_req,json=memReq,proto3,oneof"`
}

type GossipMessage_MemRes struct {
	MemRes *MembershipResponse `protobuf:"bytes,7,opt,name=mem_res,json=memRes,proto3,oneof"`
}

type GossipMessage_DataMsg struct {
	DataMsg *DataMessage `protobuf:"bytes,8,opt,name=data_msg,json=dataMsg,proto3,oneof"`
}

type GossipMessage_Hello struct {
	Hello *GossipHello `protobuf:"bytes,9,opt,name=hello,proto3,oneof"`
}

type GossipMessage_DataDig struct {
	DataDig *DataDigest `protobuf:"bytes,10,opt,name=data_dig,json=dataDig,proto3,oneof"`
}

type GossipMessage_DataReq struct {
	DataReq *DataRequest `protobuf:"bytes,11,opt,name=data_req,json=dataReq,proto3,oneof"`
}

type GossipMessage_DataUpdate struct {
	DataUpdate *DataUpdate `protobuf:"bytes,12,opt,name=data_update,json=dataUpdate,proto3,oneof"`
}

type GossipMessage_Empty struct {
	Empty *Empty `protobuf:"bytes,13,opt,name=empty,proto3,oneof"`
}

type GossipMessage_Conn struct {
	Conn *ConnEstablish `protobuf:"bytes,14,opt,name=conn,proto3,oneof"`
}

type GossipMessage_StateInfo struct {
	StateInfo *StateInfo `protobuf:"bytes,15,opt,name=state_info,json=stateInfo,proto3,oneof"`
}

type GossipMessage_StateSnapshot struct {
	StateSnapshot *StateInfoSnapshot `protobuf:"bytes,16,opt,name=state_snapshot,json=stateSnapshot,proto3,oneof"`
}

type GossipMessage_StateInfoPullReq struct {
	StateInfoPullReq *StateInfoPullRequest `protobuf:"bytes,17,opt,name=state_info_pull_req,json=stateInfoPullReq,proto3,oneof"`
}

type GossipMessage_StateRequest struct {
	StateRequest *RemoteStateRequest `protobuf:"bytes,18,opt,name=state_request,json=stateRequest,proto3,oneof"`
}

type GossipMessage_StateResponse struct {
	StateResponse *RemoteStateResponse `protobuf:"bytes,19,opt,name=state_response,json=stateResponse,proto3,oneof"`
}

type GossipMessage_LeadershipMsg struct {
	LeadershipMsg *LeadershipMessage `protobuf:"bytes,20,opt,name=leadership_msg,json=leadershipMsg,proto3,oneof"`
}

type GossipMessage_PeerIdentity struct {
	PeerIdentity *PeerIdentity `protobuf:"bytes,21,opt,name=peer_identity,json=peerIdentity,proto3,oneof"`
}

type GossipMessage_Ack struct {
	Ack *Acknowledgement `protobuf:"bytes,22,opt,name=ack,proto3,oneof"`
}

type GossipMessage_PrivateReq struct {
	PrivateReq *RemotePvtDataRequest `protobuf:"bytes,23,opt,name=privateReq,proto3,oneof"`
}

type GossipMessage_PrivateRes struct {
	PrivateRes *RemotePvtDataResponse `protobuf:"bytes,24,opt,name=privateRes,proto3,oneof"`
}

type GossipMessage_PrivateData struct {
	PrivateData *PrivateDataMessage `protobuf:"bytes,25,opt,name=private_data,json=privateData,proto3,oneof"`
}

func (*GossipMessage_AliveMsg) isGossipMessage_Content() {}

func (*GossipMessage_MemReq) isGossipMessage_Content() {}

func (*GossipMessage_MemRes) isGossipMessage_Content() {}

func (*GossipMessage_DataMsg) isGossipMessage_Content() {}

func (*GossipMessage_Hello) isGossipMessage_Content() {}

func (*GossipMessage_DataDig) isGossipMessage_Content() {}

func (*GossipMessage_DataReq) isGossipMessage_Content() {}

func (*GossipMessage_DataUpdate) isGossipMessage_Content() {}

func (*GossipMessage_Empty) isGossipMessage_Content() {}

func (*GossipMessage_Conn) isGossipMessage_Content() {}

func (*GossipMessage_StateInfo) isGossipMessage_Content() {}

func (*GossipMessage_StateSnapshot) isGossipMessage_Content() {}

func (*GossipMessage_StateInfoPullReq) isGossipMessage_Content() {}

func (*GossipMessage_StateRequest) isGossipMessage_Content() {}

func (*GossipMessage_StateResponse) isGossipMessage_Content() {}

func (*GossipMessage_LeadershipMsg) isGossipMessage_Content() {}

func (*GossipMessage_PeerIdentity) isGossipMessage_Content() {}

func (*GossipMessage_Ack) isGossipMessage_Content() {}

func (*GossipMessage_PrivateReq) isGossipMessage_Content() {}

func (*GossipMessage_PrivateRes) isGossipMessage_Content() {}

func (*GossipMessage_PrivateData) isGossipMessage_Content() {}

func (m *GossipMessage) GetContent() isGossipMessage_Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *GossipMessage) GetAliveMsg() *AliveMessage {
	if x, ok := m.GetContent().(*GossipMessage_AliveMsg); ok {
		return x.AliveMsg
	}
	return nil
}

func (m *GossipMessage) GetMemReq() *MembershipRequest {
	if x, ok := m.GetContent().(*GossipMessage_MemReq); ok {
		return x.MemReq
	}
	return nil
}

func (m *GossipMessage) GetMemRes() *MembershipResponse {
	if x, ok := m.GetContent().(*GossipMessage_MemRes); ok {
		return x.MemRes
	}
	return nil
}

func (m *GossipMessage) GetDataMsg() *DataMessage {
	if x, ok := m.GetContent().(*GossipMessage_DataMsg); ok {
		return x.DataMsg
	}
	return nil
}

func (m *GossipMessage) GetHello() *GossipHello {
	if x, ok := m.GetContent().(*GossipMessage_Hello); ok {
		return x.Hello
	}
	return nil
}

func (m *GossipMessage) GetDataDig() *DataDigest {
	if x, ok := m.GetContent().(*GossipMessage_DataDig); ok {
		return x.DataDig
	}
	return nil
}

func (m *GossipMessage) GetDataReq() *DataRequest {
	if x, ok := m.GetContent().(*GossipMessage_DataReq); ok {
		return x.DataReq
	}
	return nil
}

func (m *GossipMessage) GetDataUpdate() *DataUpdate {
	if x, ok := m.GetContent().(*GossipMessage_DataUpdate); ok {
		return x.DataUpdate
	}
	return nil
}

func (m *GossipMessage) GetEmpty() *Empty {
	if x, ok := m.GetContent().(*GossipMessage_Empty); ok {
		return x.Empty
	}
	return nil
}

func (m *GossipMessage) GetConn() *ConnEstablish {
	if x, ok := m.GetContent().(*GossipMessage_Conn); ok {
		return x.Conn
	}
	return nil
}

func (m *GossipMessage) GetStateInfo() *StateInfo {
	if x, ok := m.GetContent().(*GossipMessage_StateInfo); ok {
		return x.StateInfo
	}
	return nil
}

func (m *GossipMessage) GetStateSnapshot() *StateInfoSnapshot {
	if x, ok := m.GetContent().(*GossipMessage_StateSnapshot); ok {
		return x.StateSnapshot
	}
	return nil
}

func (m *GossipMessage) GetStateInfoPullReq() *StateInfoPullRequest {
	if x, ok := m.GetContent().(*GossipMessage_StateInfoPullReq); ok {
		return x.StateInfoPullReq
	}
	return nil
}

func (m *GossipMessage) GetStateRequest() *RemoteStateRequest {
	if x, ok := m.GetContent().(*GossipMessage_StateRequest); ok {
		return x.StateRequest
	}
	return nil
}

func (m *GossipMessage) GetStateResponse() *RemoteStateResponse {
	if x, ok := m.GetContent().(*GossipMessage_StateResponse); ok {
		return x.StateResponse
	}
	return nil
}

func (m *GossipMessage) GetLeadershipMsg() *LeadershipMessage {
	if x, ok := m.GetContent().(*GossipMessage_LeadershipMsg); ok {
		return x.LeadershipMsg
	}
	return nil
}

func (m *GossipMessage) GetPeerIdentity() *PeerIdentity {
	if x, ok := m.GetContent().(*GossipMessage_PeerIdentity); ok {
		return x.PeerIdentity
	}
	return nil
}

func (m *GossipMessage) GetAck() *Acknowledgement {
	if x, ok := m.GetContent().(*GossipMessage_Ack); ok {
		return x.Ack
	}
	return nil
}

func (m *GossipMessage) GetPrivateReq() *RemotePvtDataRequest {
	if x, ok := m.GetContent().(*GossipMessage_PrivateReq); ok {
		return x.PrivateReq
	}
	return nil
}

func (m *GossipMessage) GetPrivateRes() *RemotePvtDataResponse {
	if x, ok := m.GetContent().(*GossipMessage_PrivateRes); ok {
		return x.PrivateRes
	}
	return nil
}

func (m *GossipMessage) GetPrivateData() *PrivateDataMessage {
	if x, ok := m.GetContent().(*GossipMessage_PrivateData); ok {
		return x.PrivateData
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*GossipMessage) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _GossipMessage_OneofMarshaler, _GossipMessage_OneofUnmarshaler, _GossipMessage_OneofSizer, []interface{}{
		(*GossipMessage_AliveMsg)(nil),
		(*GossipMessage_MemReq)(nil),
		(*GossipMessage_MemRes)(nil),
		(*GossipMessage_DataMsg)(nil),
		(*GossipMessage_Hello)(nil),
		(*GossipMessage_DataDig)(nil),
		(*GossipMessage_DataReq)(nil),
		(*GossipMessage_DataUpdate)(nil),
		(*GossipMessage_Empty)(nil),
		(*GossipMessage_Conn)(nil),
		(*GossipMessage_StateInfo)(nil),
		(*GossipMessage_StateSnapshot)(nil),
		(*GossipMessage_StateInfoPullReq)(nil),
		(*GossipMessage_StateRequest)(nil),
		(*GossipMessage_StateResponse)(nil),
		(*GossipMessage_LeadershipMsg)(nil),
		(*GossipMessage_PeerIdentity)(nil),
		(*GossipMessage_Ack)(nil),
		(*GossipMessage_PrivateReq)(nil),
		(*GossipMessage_PrivateRes)(nil),
		(*GossipMessage_PrivateData)(nil),
	}
}

func _GossipMessage_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*GossipMessage)
	// content
	switch x := m.Content.(type) {
	case *GossipMessage_AliveMsg:
		b.EncodeVarint(5<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.AliveMsg); err != nil {
			return err
		}
	case *GossipMessage_MemReq:
		b.EncodeVarint(6<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.MemReq); err != nil {
			return err
		}
	case *GossipMessage_MemRes:
		b.EncodeVarint(7<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.MemRes); err != nil {
			return err
		}
	case *GossipMessage_DataMsg:
		b.EncodeVarint(8<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DataMsg); err != nil {
			return err
		}
	case *GossipMessage_Hello:
		b.EncodeVarint(9<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Hello); err != nil {
			return err
		}
	case *GossipMessage_DataDig:
		b.EncodeVarint(10<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DataDig); err != nil {
			return err
		}
	case *GossipMessage_DataReq:
		b.EncodeVarint(11<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DataReq); err != nil {
			return err
		}
	case *GossipMessage_DataUpdate:
		b.EncodeVarint(12<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.DataUpdate); err != nil {
			return err
		}
	case *GossipMessage_Empty:
		b.EncodeVarint(13<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Empty); err != nil {
			return err
		}
	case *GossipMessage_Conn:
		b.EncodeVarint(14<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Conn); err != nil {
			return err
		}
	case *GossipMessage_StateInfo:
		b.EncodeVarint(15<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.StateInfo); err != nil {
			return err
		}
	case *GossipMessage_StateSnapshot:
		b.EncodeVarint(16<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.StateSnapshot); err != nil {
			return err
		}
	case *GossipMessage_StateInfoPullReq:
		b.EncodeVarint(17<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.StateInfoPullReq); err != nil {
			return err
		}
	case *GossipMessage_StateRequest:
		b.EncodeVarint(18<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.StateRequest); err != nil {
			return err
		}
	case *GossipMessage_StateResponse:
		b.EncodeVarint(19<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.StateResponse); err != nil {
			return err
		}
	case *GossipMessage_LeadershipMsg:
		b.EncodeVarint(20<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.LeadershipMsg); err != nil {
			return err
		}
	case *GossipMessage_PeerIdentity:
		b.EncodeVarint(21<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PeerIdentity); err != nil {
			return err
		}
	case *GossipMessage_Ack:
		b.EncodeVarint(22<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Ack); err != nil {
			return err
		}
	case *GossipMessage_PrivateReq:
		b.EncodeVarint(23<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PrivateReq); err != nil {
			return err
		}
	case *GossipMessage_PrivateRes:
		b.EncodeVarint(24<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PrivateRes); err != nil {
			return err
		}
	case *GossipMessage_PrivateData:
		b.EncodeVarint(25<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PrivateData); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("GossipMessage.Content has unexpected type %T", x)
	}
	return nil
}

func _GossipMessage_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*GossipMessage)
	switch tag {
	case 5: // content.alive_msg
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(AliveMessage)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_AliveMsg{msg}
		return true, err
	case 6: // content.mem_req
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MembershipRequest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_MemReq{msg}
		return true, err
	case 7: // content.mem_res
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(MembershipResponse)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_MemRes{msg}
		return true, err
	case 8: // content.data_msg
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DataMessage)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_DataMsg{msg}
		return true, err
	case 9: // content.hello
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(GossipHello)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_Hello{msg}
		return true, err
	case 10: // content.data_dig
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DataDigest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_DataDig{msg}
		return true, err
	case 11: // content.data_req
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DataRequest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_DataReq{msg}
		return true, err
	case 12: // content.data_update
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(DataUpdate)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_DataUpdate{msg}
		return true, err
	case 13: // content.empty
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Empty)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_Empty{msg}
		return true, err
	case 14: // content.conn
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(ConnEstablish)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_Conn{msg}
		return true, err
	case 15: // content.state_info
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(StateInfo)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_StateInfo{msg}
		return true, err
	case 16: // content.state_snapshot
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(StateInfoSnapshot)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_StateSnapshot{msg}
		return true, err
	case 17: // content.state_info_pull_req
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(StateInfoPullRequest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_StateInfoPullReq{msg}
		return true, err
	case 18: // content.state_request
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RemoteStateRequest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_StateRequest{msg}
		return true, err
	case 19: // content.state_response
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RemoteStateResponse)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_StateResponse{msg}
		return true, err
	case 20: // content.leadership_msg
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(LeadershipMessage)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_LeadershipMsg{msg}
		return true, err
	case 21: // content.peer_identity
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PeerIdentity)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_PeerIdentity{msg}
		return true, err
	case 22: // content.ack
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Acknowledgement)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_Ack{msg}
		return true, err
	case 23: // content.privateReq
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RemotePvtDataRequest)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_PrivateReq{msg}
		return true, err
	case 24: // content.privateRes
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(RemotePvtDataResponse)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_PrivateRes{msg}
		return true, err
	case 25: // content.private_data
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PrivateDataMessage)
		err := b.DecodeMessage(msg)
		m.Content = &GossipMessage_PrivateData{msg}
		return true, err
	default:
		return false, nil
	}
}

func _GossipMessage_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*GossipMessage)
	// content
	switch x := m.Content.(type) {
	case *GossipMessage_AliveMsg:
		s := proto.Size(x.AliveMsg)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_MemReq:
		s := proto.Size(x.MemReq)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_MemRes:
		s := proto.Size(x.MemRes)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_DataMsg:
		s := proto.Size(x.DataMsg)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_Hello:
		s := proto.Size(x.Hello)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_DataDig:
		s := proto.Size(x.DataDig)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_DataReq:
		s := proto.Size(x.DataReq)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_DataUpdate:
		s := proto.Size(x.DataUpdate)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_Empty:
		s := proto.Size(x.Empty)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_Conn:
		s := proto.Size(x.Conn)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_StateInfo:
		s := proto.Size(x.StateInfo)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_StateSnapshot:
		s := proto.Size(x.StateSnapshot)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_StateInfoPullReq:
		s := proto.Size(x.StateInfoPullReq)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_StateRequest:
		s := proto.Size(x.StateRequest)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_StateResponse:
		s := proto.Size(x.StateResponse)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_LeadershipMsg:
		s := proto.Size(x.LeadershipMsg)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_PeerIdentity:
		s := proto.Size(x.PeerIdentity)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_Ack:
		s := proto.Size(x.Ack)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_PrivateReq:
		s := proto.Size(x.PrivateReq)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_PrivateRes:
		s := proto.Size(x.PrivateRes)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *GossipMessage_PrivateData:
		s := proto.Size(x.PrivateData)
		n += 2 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// StateInfo is used for a peer to relay its state information
// to other peers
type StateInfo struct {
	Timestamp *PeerTime `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	PkiId     []byte    `protobuf:"bytes,3,opt,name=pki_id,json=pkiId,proto3" json:"pki_id,omitempty"`
	// channel_MAC is an authentication code that proves
	// that the peer that sent this message knows
	// the name of the channel.
	Channel_MAC          []byte      `protobuf:"bytes,4,opt,name=channel_MAC,json=channelMAC,proto3" json:"channel_MAC,omitempty"`
	Properties           *Properties `protobuf:"bytes,5,opt,name=properties,proto3" json:"properties,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *StateInfo) Reset()         { *m = StateInfo{} }
func (m *StateInfo) String() string { return proto.CompactTextString(m) }
func (*StateInfo) ProtoMessage()    {}
func (*StateInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{4}
}
func (m *StateInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateInfo.Unmarshal(m, b)
}
func (m *StateInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateInfo.Marshal(b, m, deterministic)
}
func (dst *StateInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateInfo.Merge(dst, src)
}
func (m *StateInfo) XXX_Size() int {
	return xxx_messageInfo_StateInfo.Size(m)
}
func (m *StateInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_StateInfo.DiscardUnknown(m)
}

var xxx_messageInfo_StateInfo proto.InternalMessageInfo

func (m *StateInfo) GetTimestamp() *PeerTime {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *StateInfo) GetPkiId() []byte {
	if m != nil {
		return m.PkiId
	}
	return nil
}

func (m *StateInfo) GetChannel_MAC() []byte {
	if m != nil {
		return m.Channel_MAC
	}
	return nil
}

func (m *StateInfo) GetProperties() *Properties {
	if m != nil {
		return m.Properties
	}
	return nil
}

type Properties struct {
	LedgerHeight         uint64       `protobuf:"varint,1,opt,name=ledger_height,json=ledgerHeight,proto3" json:"ledger_height,omitempty"`
	LeftChannel          bool         `protobuf:"varint,2,opt,name=left_channel,json=leftChannel,proto3" json:"left_channel,omitempty"`
	Chaincodes           []*Chaincode `protobuf:"bytes,3,rep,name=chaincodes,proto3" json:"chaincodes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *Properties) Reset()         { *m = Properties{} }
func (m *Properties) String() string { return proto.CompactTextString(m) }
func (*Properties) ProtoMessage()    {}
func (*Properties) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{5}
}
func (m *Properties) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Properties.Unmarshal(m, b)
}
func (m *Properties) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Properties.Marshal(b, m, deterministic)
}
func (dst *Properties) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Properties.Merge(dst, src)
}
func (m *Properties) XXX_Size() int {
	return xxx_messageInfo_Properties.Size(m)
}
func (m *Properties) XXX_DiscardUnknown() {
	xxx_messageInfo_Properties.DiscardUnknown(m)
}

var xxx_messageInfo_Properties proto.InternalMessageInfo

func (m *Properties) GetLedgerHeight() uint64 {
	if m != nil {
		return m.LedgerHeight
	}
	return 0
}

func (m *Properties) GetLeftChannel() bool {
	if m != nil {
		return m.LeftChannel
	}
	return false
}

func (m *Properties) GetChaincodes() []*Chaincode {
	if m != nil {
		return m.Chaincodes
	}
	return nil
}

// StateInfoSnapshot is an aggregation of StateInfo messages
type StateInfoSnapshot struct {
	Elements             []*Envelope `protobuf:"bytes,1,rep,name=elements,proto3" json:"elements,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (sis *StateInfoSnapshot) Reset()         { *sis = StateInfoSnapshot{} }
func (sis *StateInfoSnapshot) String() string { return proto.CompactTextString(sis) }
func (*StateInfoSnapshot) ProtoMessage()      {}
func (*StateInfoSnapshot) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{6}
}
func (sis *StateInfoSnapshot) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateInfoSnapshot.Unmarshal(sis, b)
}
func (sis *StateInfoSnapshot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateInfoSnapshot.Marshal(b, sis, deterministic)
}
func (sis *StateInfoSnapshot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateInfoSnapshot.Merge(sis, src)
}
func (sis *StateInfoSnapshot) XXX_Size() int {
	return xxx_messageInfo_StateInfoSnapshot.Size(sis)
}
func (sis *StateInfoSnapshot) XXX_DiscardUnknown() {
	xxx_messageInfo_StateInfoSnapshot.DiscardUnknown(sis)
}

var xxx_messageInfo_StateInfoSnapshot proto.InternalMessageInfo

func (sis *StateInfoSnapshot) GetElements() []*Envelope {
	if sis != nil {
		return sis.Elements
	}
	return nil
}

// StateInfoPullRequest is used to fetch a StateInfoSnapshot
// from a remote peer
type StateInfoPullRequest struct {
	// channel_MAC is an authentication code that proves
	// that the peer that sent this message knows
	// the name of the channel.
	Channel_MAC          []byte   `protobuf:"bytes,1,opt,name=channel_MAC,json=channelMAC,proto3" json:"channel_MAC,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StateInfoPullRequest) Reset()         { *m = StateInfoPullRequest{} }
func (m *StateInfoPullRequest) String() string { return proto.CompactTextString(m) }
func (*StateInfoPullRequest) ProtoMessage()    {}
func (*StateInfoPullRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{7}
}
func (m *StateInfoPullRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StateInfoPullRequest.Unmarshal(m, b)
}
func (m *StateInfoPullRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StateInfoPullRequest.Marshal(b, m, deterministic)
}
func (dst *StateInfoPullRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateInfoPullRequest.Merge(dst, src)
}
func (m *StateInfoPullRequest) XXX_Size() int {
	return xxx_messageInfo_StateInfoPullRequest.Size(m)
}
func (m *StateInfoPullRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StateInfoPullRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StateInfoPullRequest proto.InternalMessageInfo

func (m *StateInfoPullRequest) GetChannel_MAC() []byte {
	if m != nil {
		return m.Channel_MAC
	}
	return nil
}

// ConnEstablish is the message used for the gossip handshake
// Whenever a peer connects to another peer, it handshakes
// with it by sending this message that proves its identity
type ConnEstablish struct {
	PkiId                []byte   `protobuf:"bytes,1,opt,name=pki_id,json=pkiId,proto3" json:"pki_id,omitempty"`
	Identity             []byte   `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity,omitempty"`
	TlsCertHash          []byte   `protobuf:"bytes,3,opt,name=tls_cert_hash,json=tlsCertHash,proto3" json:"tls_cert_hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConnEstablish) Reset()         { *m = ConnEstablish{} }
func (m *ConnEstablish) String() string { return proto.CompactTextString(m) }
func (*ConnEstablish) ProtoMessage()    {}
func (*ConnEstablish) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{8}
}
func (m *ConnEstablish) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConnEstablish.Unmarshal(m, b)
}
func (m *ConnEstablish) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConnEstablish.Marshal(b, m, deterministic)
}
func (dst *ConnEstablish) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConnEstablish.Merge(dst, src)
}
func (m *ConnEstablish) XXX_Size() int {
	return xxx_messageInfo_ConnEstablish.Size(m)
}
func (m *ConnEstablish) XXX_DiscardUnknown() {
	xxx_messageInfo_ConnEstablish.DiscardUnknown(m)
}

var xxx_messageInfo_ConnEstablish proto.InternalMessageInfo

func (m *ConnEstablish) GetPkiId() []byte {
	if m != nil {
		return m.PkiId
	}
	return nil
}

func (m *ConnEstablish) GetIdentity() []byte {
	if m != nil {
		return m.Identity
	}
	return nil
}

func (m *ConnEstablish) GetTlsCertHash() []byte {
	if m != nil {
		return m.TlsCertHash
	}
	return nil
}

// PeerIdentity defines the identity of the peer
// Used to make other peers learn of the identity
// of a certain peer
type PeerIdentity struct {
	PkiId                []byte   `protobuf:"bytes,1,opt,name=pki_id,json=pkiId,proto3" json:"pki_id,omitempty"`
	Cert                 []byte   `protobuf:"bytes,2,opt,name=cert,proto3" json:"cert,omitempty"`
	Metadata             []byte   `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerIdentity) Reset()         { *m = PeerIdentity{} }
func (m *PeerIdentity) String() string { return proto.CompactTextString(m) }
func (*PeerIdentity) ProtoMessage()    {}
func (*PeerIdentity) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{9}
}
func (m *PeerIdentity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerIdentity.Unmarshal(m, b)
}
func (m *PeerIdentity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerIdentity.Marshal(b, m, deterministic)
}
func (dst *PeerIdentity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerIdentity.Merge(dst, src)
}
func (m *PeerIdentity) XXX_Size() int {
	return xxx_messageInfo_PeerIdentity.Size(m)
}
func (m *PeerIdentity) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerIdentity.DiscardUnknown(m)
}

var xxx_messageInfo_PeerIdentity proto.InternalMessageInfo

func (m *PeerIdentity) GetPkiId() []byte {
	if m != nil {
		return m.PkiId
	}
	return nil
}

func (m *PeerIdentity) GetCert() []byte {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *PeerIdentity) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

// DataRequest is a message used for a peer to request
// certain data blocks from a remote peer
type DataRequest struct {
	Nonce                uint64      `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Digests              [][]byte    `protobuf:"bytes,2,rep,name=digests,proto3" json:"digests,omitempty"`
	MsgType              PullMsgType `protobuf:"varint,3,opt,name=msg_type,json=msgType,proto3,enum=gossip.PullMsgType" json:"msg_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (dr *DataRequest) Reset()         { *dr = DataRequest{} }
func (dr *DataRequest) String() string { return proto.CompactTextString(dr) }
func (*DataRequest) ProtoMessage()     {}
func (*DataRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{10}
}
func (dr *DataRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataRequest.Unmarshal(dr, b)
}
func (dr *DataRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataRequest.Marshal(b, dr, deterministic)
}
func (dr *DataRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataRequest.Merge(dr, src)
}
func (dr *DataRequest) XXX_Size() int {
	return xxx_messageInfo_DataRequest.Size(dr)
}
func (dr *DataRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DataRequest.DiscardUnknown(dr)
}

var xxx_messageInfo_DataRequest proto.InternalMessageInfo

func (dr *DataRequest) GetNonce() uint64 {
	if dr != nil {
		return dr.Nonce
	}
	return 0
}

func (dr *DataRequest) GetDigests() [][]byte {
	if dr != nil {
		return dr.Digests
	}
	return nil
}

func (dr *DataRequest) GetMsgType() PullMsgType {
	if dr != nil {
		return dr.MsgType
	}
	return PullMsgType_UNDEFINED
}

// GossipHello is the message that is used for the peer to initiate
// a pull round with another peer
type GossipHello struct {
	Nonce                uint64      `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Metadata             []byte      `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`
	MsgType              PullMsgType `protobuf:"varint,3,opt,name=msg_type,json=msgType,proto3,enum=gossip.PullMsgType" json:"msg_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GossipHello) Reset()         { *m = GossipHello{} }
func (m *GossipHello) String() string { return proto.CompactTextString(m) }
func (*GossipHello) ProtoMessage()    {}
func (*GossipHello) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{11}
}
func (m *GossipHello) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GossipHello.Unmarshal(m, b)
}
func (m *GossipHello) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GossipHello.Marshal(b, m, deterministic)
}
func (dst *GossipHello) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GossipHello.Merge(dst, src)
}
func (m *GossipHello) XXX_Size() int {
	return xxx_messageInfo_GossipHello.Size(m)
}
func (m *GossipHello) XXX_DiscardUnknown() {
	xxx_messageInfo_GossipHello.DiscardUnknown(m)
}

var xxx_messageInfo_GossipHello proto.InternalMessageInfo

func (m *GossipHello) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *GossipHello) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *GossipHello) GetMsgType() PullMsgType {
	if m != nil {
		return m.MsgType
	}
	return PullMsgType_UNDEFINED
}

// DataUpdate is the final message in the pull phase
// sent from the receiver to the initiator
type DataUpdate struct {
	Nonce                uint64      `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Data                 []*Envelope `protobuf:"bytes,2,rep,name=data,proto3" json:"data,omitempty"`
	MsgType              PullMsgType `protobuf:"varint,3,opt,name=msg_type,json=msgType,proto3,enum=gossip.PullMsgType" json:"msg_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (du *DataUpdate) Reset()         { *du = DataUpdate{} }
func (du *DataUpdate) String() string { return proto.CompactTextString(du) }
func (*DataUpdate) ProtoMessage()     {}
func (*DataUpdate) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{12}
}
func (du *DataUpdate) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataUpdate.Unmarshal(du, b)
}
func (du *DataUpdate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataUpdate.Marshal(b, du, deterministic)
}
func (du *DataUpdate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataUpdate.Merge(du, src)
}
func (du *DataUpdate) XXX_Size() int {
	return xxx_messageInfo_DataUpdate.Size(du)
}
func (du *DataUpdate) XXX_DiscardUnknown() {
	xxx_messageInfo_DataUpdate.DiscardUnknown(du)
}

var xxx_messageInfo_DataUpdate proto.InternalMessageInfo

func (du *DataUpdate) GetNonce() uint64 {
	if du != nil {
		return du.Nonce
	}
	return 0
}

func (du *DataUpdate) GetData() []*Envelope {
	if du != nil {
		return du.Data
	}
	return nil
}

func (du *DataUpdate) GetMsgType() PullMsgType {
	if du != nil {
		return du.MsgType
	}
	return PullMsgType_UNDEFINED
}

// DataDigest is the message sent from the receiver peer
// to the initator peer and contains the data items it has
type DataDigest struct {
	Nonce                uint64      `protobuf:"varint,1,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Digests              [][]byte    `protobuf:"bytes,2,rep,name=digests,proto3" json:"digests,omitempty"`
	MsgType              PullMsgType `protobuf:"varint,3,opt,name=msg_type,json=msgType,proto3,enum=gossip.PullMsgType" json:"msg_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (dd *DataDigest) Reset()         { *dd = DataDigest{} }
func (dd *DataDigest) String() string { return proto.CompactTextString(dd) }
func (*DataDigest) ProtoMessage()     {}
func (*DataDigest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{13}
}
func (dd *DataDigest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataDigest.Unmarshal(dd, b)
}
func (dd *DataDigest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataDigest.Marshal(b, dd, deterministic)
}
func (dd *DataDigest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataDigest.Merge(dd, src)
}
func (dd *DataDigest) XXX_Size() int {
	return xxx_messageInfo_DataDigest.Size(dd)
}
func (dd *DataDigest) XXX_DiscardUnknown() {
	xxx_messageInfo_DataDigest.DiscardUnknown(dd)
}

var xxx_messageInfo_DataDigest proto.InternalMessageInfo

func (dd *DataDigest) GetNonce() uint64 {
	if dd != nil {
		return dd.Nonce
	}
	return 0
}

func (dd *DataDigest) GetDigests() [][]byte {
	if dd != nil {
		return dd.Digests
	}
	return nil
}

func (dd *DataDigest) GetMsgType() PullMsgType {
	if dd != nil {
		return dd.MsgType
	}
	return PullMsgType_UNDEFINED
}

// DataMessage is the message that contains a block
type DataMessage struct {
	Payload              *Payload `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DataMessage) Reset()         { *m = DataMessage{} }
func (m *DataMessage) String() string { return proto.CompactTextString(m) }
func (*DataMessage) ProtoMessage()    {}
func (*DataMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{14}
}
func (m *DataMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DataMessage.Unmarshal(m, b)
}
func (m *DataMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DataMessage.Marshal(b, m, deterministic)
}
func (dst *DataMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DataMessage.Merge(dst, src)
}
func (m *DataMessage) XXX_Size() int {
	return xxx_messageInfo_DataMessage.Size(m)
}
func (m *DataMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_DataMessage.DiscardUnknown(m)
}

var xxx_messageInfo_DataMessage proto.InternalMessageInfo

func (m *DataMessage) GetPayload() *Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

// PrivateDataMessage message which includes private
// data information to distributed once transaction
// has been endorsed
type PrivateDataMessage struct {
	Payload              *PrivatePayload `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PrivateDataMessage) Reset()         { *m = PrivateDataMessage{} }
func (m *PrivateDataMessage) String() string { return proto.CompactTextString(m) }
func (*PrivateDataMessage) ProtoMessage()    {}
func (*PrivateDataMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{15}
}
func (m *PrivateDataMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrivateDataMessage.Unmarshal(m, b)
}
func (m *PrivateDataMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrivateDataMessage.Marshal(b, m, deterministic)
}
func (dst *PrivateDataMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrivateDataMessage.Merge(dst, src)
}
func (m *PrivateDataMessage) XXX_Size() int {
	return xxx_messageInfo_PrivateDataMessage.Size(m)
}
func (m *PrivateDataMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_PrivateDataMessage.DiscardUnknown(m)
}

var xxx_messageInfo_PrivateDataMessage proto.InternalMessageInfo

func (m *PrivateDataMessage) GetPayload() *PrivatePayload {
	if m != nil {
		return m.Payload
	}
	return nil
}

// Payload contains a block
type Payload struct {
	SeqNum               uint64   `protobuf:"varint,1,opt,name=seq_num,json=seqNum,proto3" json:"seq_num,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	PrivateData          [][]byte `protobuf:"bytes,3,rep,name=private_data,json=privateData,proto3" json:"private_data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (p *Payload) Reset()         { *p = Payload{} }
func (p *Payload) String() string { return proto.CompactTextString(p) }
func (*Payload) ProtoMessage()    {}
func (*Payload) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{16}
}
func (p *Payload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Payload.Unmarshal(p, b)
}
func (p *Payload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Payload.Marshal(b, p, deterministic)
}
func (p *Payload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Payload.Merge(p, src)
}
func (p *Payload) XXX_Size() int {
	return xxx_messageInfo_Payload.Size(p)
}
func (p *Payload) XXX_DiscardUnknown() {
	xxx_messageInfo_Payload.DiscardUnknown(p)
}

var xxx_messageInfo_Payload proto.InternalMessageInfo

func (p *Payload) GetSeqNum() uint64 {
	if p != nil {
		return p.SeqNum
	}
	return 0
}

func (p *Payload) GetData() []byte {
	if p != nil {
		return p.Data
	}
	return nil
}

func (p *Payload) GetPrivateData() [][]byte {
	if p != nil {
		return p.PrivateData
	}
	return nil
}

// PrivatePayload payload to encapsulate private
// data with collection name to enable routing
// based on collection partitioning
type PrivatePayload struct {
	CollectionName       string   `protobuf:"bytes,1,opt,name=collection_name,json=collectionName,proto3" json:"collection_name,omitempty"`
	Namespace            string   `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	TxId                 string   `protobuf:"bytes,3,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	PrivateRwset         []byte   `protobuf:"bytes,4,opt,name=private_rwset,json=privateRwset,proto3" json:"private_rwset,omitempty"`
	PrivateSimHeight     uint64   `protobuf:"varint,5,opt,name=private_sim_height,json=privateSimHeight,proto3" json:"private_sim_height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PrivatePayload) Reset()         { *m = PrivatePayload{} }
func (m *PrivatePayload) String() string { return proto.CompactTextString(m) }
func (*PrivatePayload) ProtoMessage()    {}
func (*PrivatePayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{17}
}
func (m *PrivatePayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PrivatePayload.Unmarshal(m, b)
}
func (m *PrivatePayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PrivatePayload.Marshal(b, m, deterministic)
}
func (dst *PrivatePayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PrivatePayload.Merge(dst, src)
}
func (m *PrivatePayload) XXX_Size() int {
	return xxx_messageInfo_PrivatePayload.Size(m)
}
func (m *PrivatePayload) XXX_DiscardUnknown() {
	xxx_messageInfo_PrivatePayload.DiscardUnknown(m)
}

var xxx_messageInfo_PrivatePayload proto.InternalMessageInfo

func (m *PrivatePayload) GetCollectionName() string {
	if m != nil {
		return m.CollectionName
	}
	return ""
}

func (m *PrivatePayload) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *PrivatePayload) GetTxId() string {
	if m != nil {
		return m.TxId
	}
	return ""
}

func (m *PrivatePayload) GetPrivateRwset() []byte {
	if m != nil {
		return m.PrivateRwset
	}
	return nil
}

func (m *PrivatePayload) GetPrivateSimHeight() uint64 {
	if m != nil {
		return m.PrivateSimHeight
	}
	return 0
}

// AliveMessage is sent to inform remote peers
// of a peer's existence and activity
type AliveMessage struct {
	Membership           *Member   `protobuf:"bytes,1,opt,name=membership,proto3" json:"membership,omitempty"`
	Timestamp            *PeerTime `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Identity             []byte    `protobuf:"bytes,4,opt,name=identity,proto3" json:"identity,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *AliveMessage) Reset()         { *m = AliveMessage{} }
func (m *AliveMessage) String() string { return proto.CompactTextString(m) }
func (*AliveMessage) ProtoMessage()    {}
func (*AliveMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{18}
}
func (m *AliveMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AliveMessage.Unmarshal(m, b)
}
func (m *AliveMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AliveMessage.Marshal(b, m, deterministic)
}
func (dst *AliveMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AliveMessage.Merge(dst, src)
}
func (m *AliveMessage) XXX_Size() int {
	return xxx_messageInfo_AliveMessage.Size(m)
}
func (m *AliveMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_AliveMessage.DiscardUnknown(m)
}

var xxx_messageInfo_AliveMessage proto.InternalMessageInfo

func (m *AliveMessage) GetMembership() *Member {
	if m != nil {
		return m.Membership
	}
	return nil
}

func (m *AliveMessage) GetTimestamp() *PeerTime {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *AliveMessage) GetIdentity() []byte {
	if m != nil {
		return m.Identity
	}
	return nil
}

// Leadership Message is sent during leader election to inform
// remote peers about intent of peer to proclaim itself as leader
type LeadershipMessage struct {
	PkiId                []byte    `protobuf:"bytes,1,opt,name=pki_id,json=pkiId,proto3" json:"pki_id,omitempty"`
	Timestamp            *PeerTime `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	IsDeclaration        bool      `protobuf:"varint,3,opt,name=is_declaration,json=isDeclaration,proto3" json:"is_declaration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *LeadershipMessage) Reset()         { *m = LeadershipMessage{} }
func (m *LeadershipMessage) String() string { return proto.CompactTextString(m) }
func (*LeadershipMessage) ProtoMessage()    {}
func (*LeadershipMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{19}
}
func (m *LeadershipMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LeadershipMessage.Unmarshal(m, b)
}
func (m *LeadershipMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LeadershipMessage.Marshal(b, m, deterministic)
}
func (dst *LeadershipMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LeadershipMessage.Merge(dst, src)
}
func (m *LeadershipMessage) XXX_Size() int {
	return xxx_messageInfo_LeadershipMessage.Size(m)
}
func (m *LeadershipMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_LeadershipMessage.DiscardUnknown(m)
}

var xxx_messageInfo_LeadershipMessage proto.InternalMessageInfo

func (m *LeadershipMessage) GetPkiId() []byte {
	if m != nil {
		return m.PkiId
	}
	return nil
}

func (m *LeadershipMessage) GetTimestamp() *PeerTime {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *LeadershipMessage) GetIsDeclaration() bool {
	if m != nil {
		return m.IsDeclaration
	}
	return false
}

// PeerTime defines the logical time of a peer's life
type PeerTime struct {
	IncNum               uint64   `protobuf:"varint,1,opt,name=inc_num,json=incNum,proto3" json:"inc_num,omitempty"`
	SeqNum               uint64   `protobuf:"varint,2,opt,name=seq_num,json=seqNum,proto3" json:"seq_num,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerTime) Reset()         { *m = PeerTime{} }
func (m *PeerTime) String() string { return proto.CompactTextString(m) }
func (*PeerTime) ProtoMessage()    {}
func (*PeerTime) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{20}
}
func (m *PeerTime) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerTime.Unmarshal(m, b)
}
func (m *PeerTime) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerTime.Marshal(b, m, deterministic)
}
func (dst *PeerTime) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerTime.Merge(dst, src)
}
func (m *PeerTime) XXX_Size() int {
	return xxx_messageInfo_PeerTime.Size(m)
}
func (m *PeerTime) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerTime.DiscardUnknown(m)
}

var xxx_messageInfo_PeerTime proto.InternalMessageInfo

func (m *PeerTime) GetIncNum() uint64 {
	if m != nil {
		return m.IncNum
	}
	return 0
}

func (m *PeerTime) GetSeqNum() uint64 {
	if m != nil {
		return m.SeqNum
	}
	return 0
}

// MembershipRequest is used to ask membership information
// from a remote peer
type MembershipRequest struct {
	SelfInformation      *Envelope `protobuf:"bytes,1,opt,name=self_information,json=selfInformation,proto3" json:"self_information,omitempty"`
	Known                [][]byte  `protobuf:"bytes,2,rep,name=known,proto3" json:"known,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *MembershipRequest) Reset()         { *m = MembershipRequest{} }
func (m *MembershipRequest) String() string { return proto.CompactTextString(m) }
func (*MembershipRequest) ProtoMessage()    {}
func (*MembershipRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{21}
}
func (m *MembershipRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MembershipRequest.Unmarshal(m, b)
}
func (m *MembershipRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MembershipRequest.Marshal(b, m, deterministic)
}
func (dst *MembershipRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MembershipRequest.Merge(dst, src)
}
func (m *MembershipRequest) XXX_Size() int {
	return xxx_messageInfo_MembershipRequest.Size(m)
}
func (m *MembershipRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_MembershipRequest.DiscardUnknown(m)
}

var xxx_messageInfo_MembershipRequest proto.InternalMessageInfo

func (m *MembershipRequest) GetSelfInformation() *Envelope {
	if m != nil {
		return m.SelfInformation
	}
	return nil
}

func (m *MembershipRequest) GetKnown() [][]byte {
	if m != nil {
		return m.Known
	}
	return nil
}

// MembershipResponse is used for replying to MembershipRequests
type MembershipResponse struct {
	Alive                []*Envelope `protobuf:"bytes,1,rep,name=alive,proto3" json:"alive,omitempty"`
	Dead                 []*Envelope `protobuf:"bytes,2,rep,name=dead,proto3" json:"dead,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (mr *MembershipResponse) Reset()         { *mr = MembershipResponse{} }
func (mr *MembershipResponse) String() string { return proto.CompactTextString(mr) }
func (*MembershipResponse) ProtoMessage()     {}
func (*MembershipResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{22}
}
func (mr *MembershipResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MembershipResponse.Unmarshal(mr, b)
}
func (mr *MembershipResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MembershipResponse.Marshal(b, mr, deterministic)
}
func (mr *MembershipResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MembershipResponse.Merge(mr, src)
}
func (mr *MembershipResponse) XXX_Size() int {
	return xxx_messageInfo_MembershipResponse.Size(mr)
}
func (mr *MembershipResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MembershipResponse.DiscardUnknown(mr)
}

var xxx_messageInfo_MembershipResponse proto.InternalMessageInfo

func (mr *MembershipResponse) GetAlive() []*Envelope {
	if mr != nil {
		return mr.Alive
	}
	return nil
}

func (mr *MembershipResponse) GetDead() []*Envelope {
	if mr != nil {
		return mr.Dead
	}
	return nil
}

// Member holds membership-related information
// about a peer
type Member struct {
	Endpoint             string   `protobuf:"bytes,1,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Metadata             []byte   `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`
	PkiId                []byte   `protobuf:"bytes,3,opt,name=pki_id,json=pkiId,proto3" json:"pki_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Member) Reset()         { *m = Member{} }
func (m *Member) String() string { return proto.CompactTextString(m) }
func (*Member) ProtoMessage()    {}
func (*Member) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{23}
}
func (m *Member) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Member.Unmarshal(m, b)
}
func (m *Member) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Member.Marshal(b, m, deterministic)
}
func (dst *Member) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Member.Merge(dst, src)
}
func (m *Member) XXX_Size() int {
	return xxx_messageInfo_Member.Size(m)
}
func (m *Member) XXX_DiscardUnknown() {
	xxx_messageInfo_Member.DiscardUnknown(m)
}

var xxx_messageInfo_Member proto.InternalMessageInfo

func (m *Member) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *Member) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *Member) GetPkiId() []byte {
	if m != nil {
		return m.PkiId
	}
	return nil
}

// Empty is used for pinging and in tests
type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{24}
}
func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (dst *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(dst, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

// RemoteStateRequest is used to ask a set of blocks
// from a remote peer
type RemoteStateRequest struct {
	StartSeqNum          uint64   `protobuf:"varint,1,opt,name=start_seq_num,json=startSeqNum,proto3" json:"start_seq_num,omitempty"`
	EndSeqNum            uint64   `protobuf:"varint,2,opt,name=end_seq_num,json=endSeqNum,proto3" json:"end_seq_num,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoteStateRequest) Reset()         { *m = RemoteStateRequest{} }
func (m *RemoteStateRequest) String() string { return proto.CompactTextString(m) }
func (*RemoteStateRequest) ProtoMessage()    {}
func (*RemoteStateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{25}
}
func (m *RemoteStateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoteStateRequest.Unmarshal(m, b)
}
func (m *RemoteStateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoteStateRequest.Marshal(b, m, deterministic)
}
func (dst *RemoteStateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoteStateRequest.Merge(dst, src)
}
func (m *RemoteStateRequest) XXX_Size() int {
	return xxx_messageInfo_RemoteStateRequest.Size(m)
}
func (m *RemoteStateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoteStateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RemoteStateRequest proto.InternalMessageInfo

func (m *RemoteStateRequest) GetStartSeqNum() uint64 {
	if m != nil {
		return m.StartSeqNum
	}
	return 0
}

func (m *RemoteStateRequest) GetEndSeqNum() uint64 {
	if m != nil {
		return m.EndSeqNum
	}
	return 0
}

// RemoteStateResponse is used to send a set of blocks
// to a remote peer
type RemoteStateResponse struct {
	Payloads             []*Payload `protobuf:"bytes,1,rep,name=payloads,proto3" json:"payloads,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *RemoteStateResponse) Reset()         { *m = RemoteStateResponse{} }
func (m *RemoteStateResponse) String() string { return proto.CompactTextString(m) }
func (*RemoteStateResponse) ProtoMessage()    {}
func (*RemoteStateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{26}
}
func (m *RemoteStateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoteStateResponse.Unmarshal(m, b)
}
func (m *RemoteStateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoteStateResponse.Marshal(b, m, deterministic)
}
func (dst *RemoteStateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoteStateResponse.Merge(dst, src)
}
func (m *RemoteStateResponse) XXX_Size() int {
	return xxx_messageInfo_RemoteStateResponse.Size(m)
}
func (m *RemoteStateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoteStateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RemoteStateResponse proto.InternalMessageInfo

func (m *RemoteStateResponse) GetPayloads() []*Payload {
	if m != nil {
		return m.Payloads
	}
	return nil
}

// RemotePrivateDataRequest message used to request
// missing private rwset
type RemotePvtDataRequest struct {
	Digests              []*PvtDataDigest `protobuf:"bytes,1,rep,name=digests,proto3" json:"digests,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *RemotePvtDataRequest) Reset()         { *m = RemotePvtDataRequest{} }
func (m *RemotePvtDataRequest) String() string { return proto.CompactTextString(m) }
func (*RemotePvtDataRequest) ProtoMessage()    {}
func (*RemotePvtDataRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{27}
}
func (m *RemotePvtDataRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemotePvtDataRequest.Unmarshal(m, b)
}
func (m *RemotePvtDataRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemotePvtDataRequest.Marshal(b, m, deterministic)
}
func (dst *RemotePvtDataRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemotePvtDataRequest.Merge(dst, src)
}
func (m *RemotePvtDataRequest) XXX_Size() int {
	return xxx_messageInfo_RemotePvtDataRequest.Size(m)
}
func (m *RemotePvtDataRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RemotePvtDataRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RemotePvtDataRequest proto.InternalMessageInfo

func (m *RemotePvtDataRequest) GetDigests() []*PvtDataDigest {
	if m != nil {
		return m.Digests
	}
	return nil
}

// PvtDataDigest defines a digest of private data
type PvtDataDigest struct {
	TxId                 string   `protobuf:"bytes,1,opt,name=tx_id,json=txId,proto3" json:"tx_id,omitempty"`
	Namespace            string   `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Collection           string   `protobuf:"bytes,3,opt,name=collection,proto3" json:"collection,omitempty"`
	BlockSeq             uint64   `protobuf:"varint,4,opt,name=block_seq,json=blockSeq,proto3" json:"block_seq,omitempty"`
	SeqInBlock           uint64   `protobuf:"varint,5,opt,name=seq_in_block,json=seqInBlock,proto3" json:"seq_in_block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (dig *PvtDataDigest) Reset()         { *dig = PvtDataDigest{} }
func (dig *PvtDataDigest) String() string { return proto.CompactTextString(dig) }
func (*PvtDataDigest) ProtoMessage()      {}
func (*PvtDataDigest) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{28}
}
func (dig *PvtDataDigest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PvtDataDigest.Unmarshal(dig, b)
}
func (dig *PvtDataDigest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PvtDataDigest.Marshal(b, dig, deterministic)
}
func (dig *PvtDataDigest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PvtDataDigest.Merge(dig, src)
}
func (dig *PvtDataDigest) XXX_Size() int {
	return xxx_messageInfo_PvtDataDigest.Size(dig)
}
func (dig *PvtDataDigest) XXX_DiscardUnknown() {
	xxx_messageInfo_PvtDataDigest.DiscardUnknown(dig)
}

var xxx_messageInfo_PvtDataDigest proto.InternalMessageInfo

func (dig *PvtDataDigest) GetTxId() string {
	if dig != nil {
		return dig.TxId
	}
	return ""
}

func (dig *PvtDataDigest) GetNamespace() string {
	if dig != nil {
		return dig.Namespace
	}
	return ""
}

func (dig *PvtDataDigest) GetCollection() string {
	if dig != nil {
		return dig.Collection
	}
	return ""
}

func (dig *PvtDataDigest) GetBlockSeq() uint64 {
	if dig != nil {
		return dig.BlockSeq
	}
	return 0
}

func (dig *PvtDataDigest) GetSeqInBlock() uint64 {
	if dig != nil {
		return dig.SeqInBlock
	}
	return 0
}

// RemotePrivateData message to response on private
// data replication request
type RemotePvtDataResponse struct {
	Elements             []*PvtDataElement `protobuf:"bytes,1,rep,name=elements,proto3" json:"elements,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (res *RemotePvtDataResponse) Reset()         { *res = RemotePvtDataResponse{} }
func (res *RemotePvtDataResponse) String() string { return proto.CompactTextString(res) }
func (*RemotePvtDataResponse) ProtoMessage()      {}
func (*RemotePvtDataResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{29}
}
func (res *RemotePvtDataResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemotePvtDataResponse.Unmarshal(res, b)
}
func (res *RemotePvtDataResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemotePvtDataResponse.Marshal(b, res, deterministic)
}
func (res *RemotePvtDataResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemotePvtDataResponse.Merge(res, src)
}
func (res *RemotePvtDataResponse) XXX_Size() int {
	return xxx_messageInfo_RemotePvtDataResponse.Size(res)
}
func (res *RemotePvtDataResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RemotePvtDataResponse.DiscardUnknown(res)
}

var xxx_messageInfo_RemotePvtDataResponse proto.InternalMessageInfo

func (res *RemotePvtDataResponse) GetElements() []*PvtDataElement {
	if res != nil {
		return res.Elements
	}
	return nil
}

type PvtDataElement struct {
	Digest *PvtDataDigest `protobuf:"bytes,1,opt,name=digest,proto3" json:"digest,omitempty"`
	// the payload is a marshaled kvrwset.KVRWSet
	Payload              [][]byte `protobuf:"bytes,2,rep,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PvtDataElement) Reset()         { *m = PvtDataElement{} }
func (m *PvtDataElement) String() string { return proto.CompactTextString(m) }
func (*PvtDataElement) ProtoMessage()    {}
func (*PvtDataElement) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{30}
}
func (m *PvtDataElement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PvtDataElement.Unmarshal(m, b)
}
func (m *PvtDataElement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PvtDataElement.Marshal(b, m, deterministic)
}
func (dst *PvtDataElement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PvtDataElement.Merge(dst, src)
}
func (m *PvtDataElement) XXX_Size() int {
	return xxx_messageInfo_PvtDataElement.Size(m)
}
func (m *PvtDataElement) XXX_DiscardUnknown() {
	xxx_messageInfo_PvtDataElement.DiscardUnknown(m)
}

var xxx_messageInfo_PvtDataElement proto.InternalMessageInfo

func (m *PvtDataElement) GetDigest() *PvtDataDigest {
	if m != nil {
		return m.Digest
	}
	return nil
}

func (m *PvtDataElement) GetPayload() [][]byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

// PvtPayload augments private rwset data and tx index
// inside the block
type PvtDataPayload struct {
	TxSeqInBlock uint64 `protobuf:"varint,1,opt,name=tx_seq_in_block,json=txSeqInBlock,proto3" json:"tx_seq_in_block,omitempty"`
	// Encodes marhslaed bytes of rwset.TxPvtReadWriteSet
	// defined in rwset.proto
	Payload              []byte   `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PvtDataPayload) Reset()         { *m = PvtDataPayload{} }
func (m *PvtDataPayload) String() string { return proto.CompactTextString(m) }
func (*PvtDataPayload) ProtoMessage()    {}
func (*PvtDataPayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{31}
}
func (m *PvtDataPayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PvtDataPayload.Unmarshal(m, b)
}
func (m *PvtDataPayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PvtDataPayload.Marshal(b, m, deterministic)
}
func (dst *PvtDataPayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PvtDataPayload.Merge(dst, src)
}
func (m *PvtDataPayload) XXX_Size() int {
	return xxx_messageInfo_PvtDataPayload.Size(m)
}
func (m *PvtDataPayload) XXX_DiscardUnknown() {
	xxx_messageInfo_PvtDataPayload.DiscardUnknown(m)
}

var xxx_messageInfo_PvtDataPayload proto.InternalMessageInfo

func (m *PvtDataPayload) GetTxSeqInBlock() uint64 {
	if m != nil {
		return m.TxSeqInBlock
	}
	return 0
}

func (m *PvtDataPayload) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type Acknowledgement struct {
	Error                string   `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Acknowledgement) Reset()         { *m = Acknowledgement{} }
func (m *Acknowledgement) String() string { return proto.CompactTextString(m) }
func (*Acknowledgement) ProtoMessage()    {}
func (*Acknowledgement) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{32}
}
func (m *Acknowledgement) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Acknowledgement.Unmarshal(m, b)
}
func (m *Acknowledgement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Acknowledgement.Marshal(b, m, deterministic)
}
func (dst *Acknowledgement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Acknowledgement.Merge(dst, src)
}
func (m *Acknowledgement) XXX_Size() int {
	return xxx_messageInfo_Acknowledgement.Size(m)
}
func (m *Acknowledgement) XXX_DiscardUnknown() {
	xxx_messageInfo_Acknowledgement.DiscardUnknown(m)
}

var xxx_messageInfo_Acknowledgement proto.InternalMessageInfo

func (m *Acknowledgement) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

// Chaincode represents a Chaincode that is installed
// on a peer
type Chaincode struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Version              string   `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Metadata             []byte   `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Chaincode) Reset()         { *m = Chaincode{} }
func (m *Chaincode) String() string { return proto.CompactTextString(m) }
func (*Chaincode) ProtoMessage()    {}
func (*Chaincode) Descriptor() ([]byte, []int) {
	return fileDescriptor_message_7c42328ef5ef9997, []int{33}
}
func (m *Chaincode) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Chaincode.Unmarshal(m, b)
}
func (m *Chaincode) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Chaincode.Marshal(b, m, deterministic)
}
func (dst *Chaincode) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Chaincode.Merge(dst, src)
}
func (m *Chaincode) XXX_Size() int {
	return xxx_messageInfo_Chaincode.Size(m)
}
func (m *Chaincode) XXX_DiscardUnknown() {
	xxx_messageInfo_Chaincode.DiscardUnknown(m)
}

var xxx_messageInfo_Chaincode proto.InternalMessageInfo

func (m *Chaincode) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Chaincode) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *Chaincode) GetMetadata() []byte {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func init() {
	proto.RegisterType((*Envelope)(nil), "gossip.Envelope")
	proto.RegisterType((*SecretEnvelope)(nil), "gossip.SecretEnvelope")
	proto.RegisterType((*Secret)(nil), "gossip.Secret")
	proto.RegisterType((*GossipMessage)(nil), "gossip.GossipMessage")
	proto.RegisterType((*StateInfo)(nil), "gossip.StateInfo")
	proto.RegisterType((*Properties)(nil), "gossip.Properties")
	proto.RegisterType((*StateInfoSnapshot)(nil), "gossip.StateInfoSnapshot")
	proto.RegisterType((*StateInfoPullRequest)(nil), "gossip.StateInfoPullRequest")
	proto.RegisterType((*ConnEstablish)(nil), "gossip.ConnEstablish")
	proto.RegisterType((*PeerIdentity)(nil), "gossip.PeerIdentity")
	proto.RegisterType((*DataRequest)(nil), "gossip.DataRequest")
	proto.RegisterType((*GossipHello)(nil), "gossip.GossipHello")
	proto.RegisterType((*DataUpdate)(nil), "gossip.DataUpdate")
	proto.RegisterType((*DataDigest)(nil), "gossip.DataDigest")
	proto.RegisterType((*DataMessage)(nil), "gossip.DataMessage")
	proto.RegisterType((*PrivateDataMessage)(nil), "gossip.PrivateDataMessage")
	proto.RegisterType((*Payload)(nil), "gossip.Payload")
	proto.RegisterType((*PrivatePayload)(nil), "gossip.PrivatePayload")
	proto.RegisterType((*AliveMessage)(nil), "gossip.AliveMessage")
	proto.RegisterType((*LeadershipMessage)(nil), "gossip.LeadershipMessage")
	proto.RegisterType((*PeerTime)(nil), "gossip.PeerTime")
	proto.RegisterType((*MembershipRequest)(nil), "gossip.MembershipRequest")
	proto.RegisterType((*MembershipResponse)(nil), "gossip.MembershipResponse")
	proto.RegisterType((*Member)(nil), "gossip.Member")
	proto.RegisterType((*Empty)(nil), "gossip.Empty")
	proto.RegisterType((*RemoteStateRequest)(nil), "gossip.RemoteStateRequest")
	proto.RegisterType((*RemoteStateResponse)(nil), "gossip.RemoteStateResponse")
	proto.RegisterType((*RemotePvtDataRequest)(nil), "gossip.RemotePvtDataRequest")
	proto.RegisterType((*PvtDataDigest)(nil), "gossip.PvtDataDigest")
	proto.RegisterType((*RemotePvtDataResponse)(nil), "gossip.RemotePvtDataResponse")
	proto.RegisterType((*PvtDataElement)(nil), "gossip.PvtDataElement")
	proto.RegisterType((*PvtDataPayload)(nil), "gossip.PvtDataPayload")
	proto.RegisterType((*Acknowledgement)(nil), "gossip.Acknowledgement")
	proto.RegisterType((*Chaincode)(nil), "gossip.Chaincode")
	proto.RegisterEnum("gossip.PullMsgType", PullMsgType_name, PullMsgType_value)
	proto.RegisterEnum("gossip.GossipMessage_Tag", GossipMessage_Tag_name, GossipMessage_Tag_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GossipClient is the client API for Gossip service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GossipClient interface {
	// GossipStream is the gRPC stream used for sending and receiving messages
	GossipStream(ctx context.Context, opts ...grpc.CallOption) (Gossip_GossipStreamClient, error)
	// Ping is used to probe a remote peer's aliveness
	Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error)
}

type gossipClient struct {
	cc *grpc.ClientConn
}

func NewGossipClient(cc *grpc.ClientConn) GossipClient {
	return &gossipClient{cc}
}

func (c *gossipClient) GossipStream(ctx context.Context, opts ...grpc.CallOption) (Gossip_GossipStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Gossip_serviceDesc.Streams[0], "/gossip.Gossip/GossipStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &gossipGossipStreamClient{stream}
	return x, nil
}

type Gossip_GossipStreamClient interface {
	Send(*Envelope) error
	Recv() (*Envelope, error)
	grpc.ClientStream
}

type gossipGossipStreamClient struct {
	grpc.ClientStream
}

func (x *gossipGossipStreamClient) Send(m *Envelope) error {
	return x.ClientStream.SendMsg(m)
}

func (x *gossipGossipStreamClient) Recv() (*Envelope, error) {
	m := new(Envelope)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *gossipClient) Ping(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/gossip.Gossip/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GossipServer is the server API for Gossip service.
type GossipServer interface {
	// GossipStream is the gRPC stream used for sending and receiving messages
	GossipStream(Gossip_GossipStreamServer) error
	// Ping is used to probe a remote peer's aliveness
	Ping(context.Context, *Empty) (*Empty, error)
}

func RegisterGossipServer(s *grpc.Server, srv GossipServer) {
	s.RegisterService(&_Gossip_serviceDesc, srv)
}

func _Gossip_GossipStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GossipServer).GossipStream(&gossipGossipStreamServer{stream})
}

type Gossip_GossipStreamServer interface {
	Send(*Envelope) error
	Recv() (*Envelope, error)
	grpc.ServerStream
}

type gossipGossipStreamServer struct {
	grpc.ServerStream
}

func (x *gossipGossipStreamServer) Send(m *Envelope) error {
	return x.ServerStream.SendMsg(m)
}

func (x *gossipGossipStreamServer) Recv() (*Envelope, error) {
	m := new(Envelope)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Gossip_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GossipServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gossip.Gossip/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GossipServer).Ping(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gossip_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gossip.Gossip",
	HandlerType: (*GossipServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Gossip_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GossipStream",
			Handler:       _Gossip_GossipStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "gossip/message.proto",
}

func init() { proto.RegisterFile("gossip/message.proto", fileDescriptor_message_7c42328ef5ef9997) }

var fileDescriptor_message_7c42328ef5ef9997 = []byte{
	// 1874 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x58, 0x5b, 0x6f, 0xe3, 0xc6,
	0x15, 0x16, 0x6d, 0x5d, 0x8f, 0x2e, 0x96, 0xc7, 0xde, 0x5d, 0xc6, 0x49, 0x13, 0x87, 0xed, 0x26,
	0xdb, 0x7a, 0x23, 0x6f, 0x9d, 0x16, 0x0d, 0x90, 0xb6, 0x0b, 0x5b, 0x52, 0x2c, 0x21, 0x2b, 0xad,
	0x4b, 0x7b, 0xd1, 0xba, 0x2f, 0xc4, 0x98, 0x1c, 0x53, 0xac, 0xc9, 0x21, 0xcd, 0x19, 0x3b, 0xf6,
	0x63, 0xd1, 0x87, 0x00, 0x7d, 0xe9, 0x6f, 0xe8, 0x53, 0xff, 0x66, 0x31, 0x33, 0xbc, 0x4a, 0xf6,
	0x02, 0x1b, 0x20, 0x6f, 0x3c, 0xf7, 0x99, 0x33, 0x67, 0xbe, 0x73, 0x86, 0xb0, 0xed, 0x86, 0x8c,
	0x79, 0xd1, 0x7e, 0x40, 0x18, 0xc3, 0x2e, 0x19, 0x44, 0x71, 0xc8, 0x43, 0x54, 0x57, 0xdc, 0x9d,
	0x67, 0x76, 0x18, 0x04, 0x21, 0xdd, 0xb7, 0x43, 0xdf, 0x27, 0x36, 0xf7, 0x42, 0xaa, 0x14, 0x8c,
	0x7f, 0x69, 0xd0, 0x1c, 0xd3, 0x5b, 0xe2, 0x87, 0x11, 0x41, 0x3a, 0x34, 0x22, 0x7c, 0xef, 0x87,
	0xd8, 0xd1, 0xb5, 0x5d, 0xed, 0x45, 0xc7, 0x4c, 0x49, 0xf4, 0x09, 0xb4, 0x98, 0xe7, 0x52, 0xcc,
	0x6f, 0x62, 0xa2, 0xaf, 0x49, 0x59, 0xce, 0x40, 0xaf, 0x61, 0x83, 0x11, 0x3b, 0x26, 0xdc, 0x22,
	0x89, 0x2b, 0x7d, 0x7d, 0x57, 0x7b, 0xd1, 0x3e, 0x78, 0x3a, 0x50, 0xf1, 0x07, 0xa7, 0x52, 0x9c,
	0x06, 0x32, 0x7b, 0xac, 0x44, 0x1b, 0x13, 0xe8, 0x95, 0x35, 0x7e, 0xea, 0x52, 0x8c, 0x43, 0xa8,
	0x2b, 0x4f, 0xe8, 0x25, 0xf4, 0x3d, 0xca, 0x49, 0x4c, 0xb1, 0x3f, 0xa6, 0x4e, 0x14, 0x7a, 0x94,
	0x4b, 0x57, 0xad, 0x49, 0xc5, 0x5c, 0x91, 0x1c, 0xb5, 0xa0, 0x61, 0x87, 0x94, 0x13, 0xca, 0x8d,
	0x1f, 0xdb, 0xd0, 0x3d, 0x96, 0xcb, 0x9e, 0xa9, 0x5c, 0xa2, 0x6d, 0xa8, 0xd1, 0x90, 0xda, 0x44,
	0xda, 0x57, 0x4d, 0x45, 0x88, 0x25, 0xda, 0x0b, 0x4c, 0x29, 0xf1, 0x93, 0x65, 0xa4, 0x24, 0xda,
	0x83, 0x75, 0x8e, 0x5d, 0x99, 0x83, 0xde, 0xc1, 0x47, 0x69, 0x0e, 0x4a, 0x3e, 0x07, 0x67, 0xd8,
	0x35, 0x85, 0x16, 0xfa, 0x1a, 0x5a, 0xd8, 0xf7, 0x6e, 0x89, 0x15, 0x30, 0x57, 0xaf, 0xc9, 0xb4,
	0x6d, 0xa7, 0x26, 0x87, 0x42, 0x90, 0x58, 0x4c, 0x2a, 0x66, 0x53, 0x2a, 0xce, 0x98, 0x8b, 0x7e,
	0x07, 0x8d, 0x80, 0x04, 0x56, 0x4c, 0xae, 0xf5, 0xba, 0x34, 0xc9, 0xa2, 0xcc, 0x48, 0x70, 0x41,
	0x62, 0xb6, 0xf0, 0x22, 0x93, 0x5c, 0xdf, 0x10, 0xc6, 0x27, 0x15, 0xb3, 0x1e, 0x90, 0xc0, 0x24,
	0xd7, 0xe8, 0xf7, 0xa9, 0x15, 0xd3, 0x1b, 0xd2, 0x6a, 0xe7, 0x21, 0x2b, 0x16, 0x85, 0x94, 0x91,
	0xcc, 0x8c, 0xa1, 0x57, 0xd0, 0x74, 0x30, 0xc7, 0x72, 0x81, 0x4d, 0x69, 0xb7, 0x95, 0xda, 0x8d,
	0x30, 0xc7, 0xf9, 0xfa, 0x1a, 0x42, 0x4d, 0x2c, 0x6f, 0x0f, 0x6a, 0x0b, 0xe2, 0xfb, 0xa1, 0xde,
	0x2a, 0xab, 0xab, 0x14, 0x4c, 0x84, 0x68, 0x52, 0x31, 0x95, 0x0e, 0xda, 0x4f, 0xdc, 0x3b, 0x9e,
	0xab, 0x83, 0xd4, 0x47, 0x45, 0xf7, 0x23, 0xcf, 0x55, 0xbb, 0x90, 0xde, 0x47, 0x9e, 0x9b, 0xad,
	0x47, 0xec, 0xbe, 0xbd, 0xba, 0x9e, 0x7c, 0xdf, 0xd2, 0x42, 0x6d, 0xbc, 0x2d, 0x2d, 0x6e, 0x22,
	0x07, 0x73, 0xa2, 0x77, 0x56, 0xa3, 0xbc, 0x93, 0x92, 0x49, 0xc5, 0x04, 0x27, 0xa3, 0xd0, 0x73,
	0xa8, 0x91, 0x20, 0xe2, 0xf7, 0x7a, 0x57, 0x1a, 0x74, 0x53, 0x83, 0xb1, 0x60, 0x8a, 0x0d, 0x48,
	0x29, 0xda, 0x83, 0xaa, 0x1d, 0x52, 0xaa, 0xf7, 0xa4, 0xd6, 0x93, 0x54, 0x6b, 0x18, 0x52, 0x3a,
	0x66, 0x1c, 0x5f, 0xf8, 0x1e, 0x5b, 0x4c, 0x2a, 0xa6, 0x54, 0x42, 0x07, 0x00, 0x8c, 0x63, 0x4e,
	0x2c, 0x8f, 0x5e, 0x86, 0xfa, 0x86, 0x34, 0xd9, 0xcc, 0xae, 0x89, 0x90, 0x4c, 0xe9, 0xa5, 0xc8,
	0x4e, 0x8b, 0xa5, 0x04, 0x3a, 0x82, 0x9e, 0xb2, 0x61, 0x14, 0x47, 0x6c, 0x11, 0x72, 0xbd, 0x5f,
	0x3e, 0xf4, 0xcc, 0xee, 0x34, 0x51, 0x98, 0x54, 0xcc, 0xae, 0x34, 0x49, 0x19, 0x68, 0x06, 0x5b,
	0x79, 0x5c, 0x2b, 0xba, 0xf1, 0x7d, 0x99, 0xbf, 0x4d, 0xe9, 0xe8, 0x93, 0x15, 0x47, 0x27, 0x37,
	0xbe, 0x9f, 0x27, 0xb2, 0xcf, 0x96, 0xf8, 0xe8, 0x10, 0x94, 0x7f, 0xe1, 0x44, 0x28, 0xe9, 0xa8,
	0x5c, 0x50, 0x26, 0x09, 0x42, 0x4e, 0xa4, 0xbb, 0xdc, 0x4d, 0x87, 0x15, 0x68, 0x34, 0x4a, 0x77,
	0x15, 0x27, 0x25, 0xa7, 0x6f, 0x49, 0x1f, 0x1f, 0x3f, 0xe8, 0x23, 0xab, 0xca, 0x2e, 0x2b, 0x32,
	0x44, 0x6e, 0x7c, 0x82, 0x1d, 0x55, 0xbc, 0xb2, 0x44, 0xb7, 0xcb, 0xb9, 0x79, 0x93, 0x49, 0xf3,
	0x42, 0xed, 0xe6, 0x26, 0xa2, 0x5c, 0xbf, 0x85, 0x6e, 0x44, 0x48, 0x6c, 0x79, 0x0e, 0xa1, 0xdc,
	0xe3, 0xf7, 0xfa, 0x93, 0xf2, 0x35, 0x3c, 0x21, 0x24, 0x9e, 0x26, 0x32, 0xb1, 0x8d, 0xa8, 0x40,
	0x8b, 0xcb, 0x8e, 0xed, 0x2b, 0xfd, 0xa9, 0x34, 0x79, 0x96, 0xdd, 0x5c, 0xfb, 0x8a, 0x86, 0x3f,
	0xf8, 0xc4, 0x71, 0x49, 0x40, 0xa8, 0xd8, 0xbc, 0xd0, 0x42, 0x7f, 0x06, 0x88, 0x62, 0xef, 0x56,
	0x65, 0x41, 0x7f, 0x56, 0x4e, 0xbe, 0xda, 0xef, 0xc9, 0x2d, 0x2f, 0x57, 0x71, 0xc1, 0x02, 0xbd,
	0x2e, 0xd8, 0x33, 0x5d, 0x97, 0xf6, 0xbf, 0x78, 0xc4, 0x3e, 0xcb, 0x58, 0xc1, 0x04, 0xbd, 0x86,
	0x4e, 0x42, 0x59, 0xa2, 0xd0, 0xf5, 0x8f, 0xca, 0xc7, 0x76, 0xa2, 0x64, 0xe5, 0x6b, 0xdd, 0x8e,
	0x72, 0xae, 0x61, 0xc1, 0xfa, 0x19, 0x76, 0x51, 0x17, 0x5a, 0xef, 0xe6, 0xa3, 0xf1, 0x77, 0xd3,
	0xf9, 0x78, 0xd4, 0xaf, 0xa0, 0x16, 0xd4, 0xc6, 0xb3, 0x93, 0xb3, 0xf3, 0xbe, 0x86, 0x3a, 0xd0,
	0x7c, 0x6b, 0x1e, 0x5b, 0x6f, 0xe7, 0x6f, 0xce, 0xfb, 0x6b, 0x42, 0x6f, 0x38, 0x39, 0x9c, 0x2b,
	0x72, 0x1d, 0xf5, 0xa1, 0x23, 0xc9, 0xc3, 0xf9, 0xc8, 0x7a, 0x6b, 0x1e, 0xf7, 0xab, 0x68, 0x03,
	0xda, 0x4a, 0xc1, 0x94, 0x8c, 0x5a, 0x11, 0x89, 0xff, 0xa7, 0x41, 0x2b, 0xab, 0x48, 0x34, 0x80,
	0x16, 0xf7, 0x02, 0xc2, 0x38, 0x0e, 0x22, 0x89, 0xb8, 0xed, 0x83, 0x7e, 0xf1, 0x84, 0xce, 0xbc,
	0x80, 0x98, 0xb9, 0x0a, 0x7a, 0x02, 0xf5, 0xe8, 0xca, 0xb3, 0x3c, 0x47, 0x02, 0x71, 0xc7, 0xac,
	0x45, 0x57, 0xde, 0xd4, 0x41, 0x9f, 0x41, 0x3b, 0xc1, 0x69, 0x6b, 0x76, 0x38, 0xd4, 0xab, 0x52,
	0x06, 0x09, 0x6b, 0x76, 0x38, 0x14, 0x37, 0x34, 0x8a, 0xc3, 0x88, 0xc4, 0xdc, 0x23, 0x2c, 0x41,
	0x64, 0x94, 0x27, 0x28, 0x95, 0x98, 0x05, 0x2d, 0xe3, 0x47, 0x0d, 0x20, 0x17, 0xa1, 0x5f, 0x42,
	0x57, 0x1e, 0x7d, 0x6c, 0x2d, 0x88, 0xe7, 0x2e, 0x78, 0xd2, 0x38, 0x3a, 0x8a, 0x39, 0x91, 0x3c,
	0xf4, 0x39, 0x74, 0x7c, 0x72, 0xc9, 0xad, 0x62, 0x13, 0x69, 0x9a, 0x6d, 0xc1, 0x1b, 0x26, 0x8d,
	0xe4, 0xb7, 0x20, 0x16, 0xe6, 0x51, 0x3b, 0x74, 0x08, 0xd3, 0xd7, 0x77, 0xd7, 0x8b, 0x60, 0x31,
	0x4c, 0x25, 0x66, 0x41, 0xc9, 0x38, 0x84, 0xcd, 0x15, 0x34, 0x40, 0x2f, 0xa1, 0x49, 0x7c, 0x59,
	0x88, 0x4c, 0xd7, 0xa4, 0x97, 0x2c, 0x73, 0x59, 0x4f, 0xce, 0x34, 0x8c, 0x3f, 0xc0, 0xf6, 0x43,
	0x38, 0xb0, 0x9c, 0x39, 0x6d, 0x39, 0x73, 0xc6, 0x25, 0x74, 0x4b, 0xa0, 0x57, 0x38, 0x02, 0xad,
	0x78, 0x04, 0x3b, 0xd0, 0xcc, 0xae, 0x9a, 0x6a, 0x9d, 0x19, 0x8d, 0x0c, 0xe8, 0x72, 0x9f, 0x59,
	0x36, 0x89, 0xb9, 0xb5, 0xc0, 0x6c, 0x91, 0x1c, 0x5e, 0x9b, 0xfb, 0x6c, 0x48, 0x62, 0x3e, 0xc1,
	0x6c, 0x61, 0xbc, 0x83, 0x4e, 0xf1, 0x4a, 0x3e, 0x16, 0x06, 0x41, 0x55, 0xb8, 0x49, 0x42, 0xc8,
	0x6f, 0x11, 0x3a, 0x20, 0x1c, 0xcb, 0xda, 0x57, 0x9e, 0x33, 0xda, 0x08, 0xa0, 0x5d, 0xb8, 0x79,
	0x8f, 0x77, 0x7d, 0x47, 0x76, 0x24, 0xa6, 0xaf, 0xed, 0xae, 0x8b, 0xae, 0x9f, 0x90, 0x68, 0x00,
	0xcd, 0x80, 0xb9, 0x16, 0xbf, 0x4f, 0xc6, 0x9f, 0x5e, 0xde, 0x96, 0x44, 0x16, 0x67, 0xcc, 0x3d,
	0xbb, 0x8f, 0x88, 0xd9, 0x08, 0xd4, 0x87, 0x11, 0x42, 0xbb, 0xd0, 0x0f, 0x1f, 0x09, 0x57, 0x5c,
	0xef, 0x5a, 0x79, 0xbd, 0x1f, 0x1c, 0xf0, 0x0e, 0x20, 0x6f, 0x75, 0x8f, 0xc4, 0xfb, 0x15, 0x54,
	0x93, 0x58, 0x0f, 0x57, 0x49, 0xf5, 0x27, 0x45, 0xf6, 0x55, 0x64, 0xd5, 0xca, 0x7f, 0xf6, 0xc4,
	0x7e, 0xa3, 0xce, 0x31, 0x9d, 0xde, 0x7e, 0x5d, 0x1e, 0x25, 0xdb, 0x07, 0x1b, 0x99, 0xb5, 0x62,
	0x67, 0xb3, 0xa5, 0xf1, 0x1d, 0xa0, 0x55, 0x04, 0x44, 0xaf, 0x96, 0x1d, 0x3c, 0x5d, 0x82, 0xcb,
	0x15, 0x3f, 0xe7, 0xd0, 0x48, 0x78, 0xe8, 0x19, 0x34, 0x18, 0xb9, 0xb6, 0xe8, 0x4d, 0x90, 0x6c,
	0xb7, 0xce, 0xc8, 0xf5, 0xfc, 0x26, 0x10, 0xd5, 0x59, 0x38, 0x55, 0x95, 0xd7, 0xcf, 0x97, 0xd0,
	0x79, 0x5d, 0x26, 0xa2, 0x84, 0xbf, 0xff, 0x59, 0x83, 0x5e, 0x39, 0x2c, 0xfa, 0x12, 0x36, 0xf2,
	0xb9, 0xde, 0xa2, 0x38, 0x50, 0x99, 0x6d, 0x99, 0xbd, 0x9c, 0x3d, 0xc7, 0x01, 0x11, 0xa3, 0xb3,
	0x90, 0xb2, 0x08, 0xdb, 0x6a, 0x74, 0x6e, 0x99, 0x39, 0x03, 0x6d, 0x41, 0x8d, 0xdf, 0xa5, 0x70,
	0xd9, 0x32, 0xab, 0xfc, 0x6e, 0xea, 0x08, 0x24, 0x4b, 0x57, 0x14, 0xff, 0xc0, 0x08, 0x4f, 0xf0,
	0x32, 0x5d, 0xa6, 0x29, 0x78, 0xe8, 0x25, 0xa0, 0x54, 0x89, 0x79, 0x41, 0x8a, 0x79, 0x35, 0xb9,
	0xdd, 0x7e, 0x22, 0x39, 0xf5, 0x82, 0x04, 0xf7, 0xe6, 0x80, 0x0a, 0xcb, 0xb5, 0x43, 0x7a, 0xe9,
	0xb9, 0x2c, 0x19, 0x63, 0x3f, 0x1b, 0xa8, 0x87, 0xca, 0x60, 0x98, 0x69, 0x0c, 0xa5, 0xc2, 0x09,
	0xb6, 0xaf, 0xb0, 0x4b, 0xcc, 0x4d, 0x7b, 0x49, 0xc0, 0x8c, 0x7f, 0x6b, 0xd0, 0x29, 0x0e, 0xca,
	0x68, 0x00, 0x10, 0x64, 0xf3, 0x6c, 0x72, 0x64, 0xbd, 0xf2, 0xa4, 0x6b, 0x16, 0x34, 0x3e, 0xb8,
	0xb1, 0x14, 0xe1, 0xab, 0x5a, 0x86, 0x2f, 0xe3, 0x9f, 0x1a, 0x6c, 0xae, 0x4c, 0x1c, 0x8f, 0x01,
	0xd4, 0x87, 0x06, 0x7e, 0x0e, 0x3d, 0x8f, 0x59, 0x0e, 0xb1, 0x7d, 0x1c, 0x63, 0x91, 0x02, 0x79,
	0x54, 0x4d, 0xb3, 0xeb, 0xb1, 0x51, 0xce, 0x34, 0xfe, 0x08, 0xcd, 0xd4, 0x5a, 0x94, 0x9f, 0x47,
	0xed, 0x62, 0xf9, 0x79, 0xd4, 0x16, 0xe5, 0x57, 0xa8, 0xcb, 0xb5, 0x62, 0x5d, 0x1a, 0x97, 0xb0,
	0xb9, 0xf2, 0x86, 0x40, 0xdf, 0x42, 0x9f, 0x11, 0xff, 0x52, 0x0e, 0x8f, 0x71, 0xa0, 0x62, 0x6b,
	0xe5, 0x05, 0x67, 0x10, 0xb1, 0x21, 0x34, 0xa7, 0xb9, 0xa2, 0xb8, 0xef, 0x62, 0x18, 0xa2, 0xc9,
	0xbd, 0x56, 0x84, 0x71, 0x01, 0x68, 0xf5, 0xd5, 0x81, 0xbe, 0x80, 0x9a, 0x7c, 0xe4, 0x3c, 0xda,
	0xa6, 0x94, 0x58, 0xe2, 0x14, 0xc1, 0xce, 0x7b, 0x70, 0x8a, 0x60, 0xc7, 0xf8, 0x2b, 0xd4, 0x55,
	0x0c, 0x71, 0x66, 0xa4, 0xf4, 0x0a, 0x34, 0x33, 0xfa, 0xbd, 0x18, 0xfb, 0xf0, 0x10, 0x61, 0x34,
	0xa0, 0x26, 0x1f, 0x01, 0xc6, 0xdf, 0x00, 0xad, 0x8e, 0xba, 0xa2, 0x89, 0x31, 0x8e, 0x63, 0x6e,
	0x95, 0xaf, 0x7e, 0x5b, 0x32, 0x4f, 0xd5, 0xfd, 0xff, 0x14, 0xda, 0x84, 0x3a, 0x56, 0xf9, 0x10,
	0x5a, 0x84, 0x3a, 0x4a, 0x6e, 0x1c, 0xc1, 0xd6, 0x03, 0x03, 0x30, 0xda, 0x83, 0x66, 0x82, 0x32,
	0x69, 0x2b, 0x5f, 0x81, 0xb3, 0x4c, 0xc1, 0x38, 0x86, 0xed, 0x87, 0x86, 0x4a, 0xb4, 0x9f, 0x63,
	0xad, 0xf2, 0x91, 0x3d, 0x5a, 0x12, 0x45, 0x85, 0xd4, 0x19, 0x04, 0x1b, 0xff, 0xd5, 0xa0, 0x5b,
	0x12, 0xe5, 0x68, 0xa1, 0x15, 0xd0, 0xe2, 0xfd, 0x00, 0xf3, 0x29, 0x40, 0x7e, 0x7b, 0x13, 0x94,
	0x29, 0x70, 0xd0, 0xc7, 0xd0, 0xba, 0xf0, 0x43, 0xfb, 0x4a, 0xe4, 0x44, 0x5e, 0xac, 0xaa, 0xd9,
	0x94, 0x8c, 0x53, 0x72, 0x8d, 0x76, 0xa1, 0x23, 0x52, 0xe5, 0x51, 0x4b, 0xb2, 0x12, 0x74, 0x01,
	0x46, 0xae, 0xa7, 0xf4, 0x48, 0x70, 0x8c, 0xef, 0xe1, 0xc9, 0x83, 0x13, 0x30, 0x3a, 0x58, 0x99,
	0x7e, 0x9e, 0x2e, 0x6d, 0x77, 0xac, 0xc4, 0x85, 0x19, 0xe8, 0x1c, 0x7a, 0x65, 0x19, 0xfa, 0x0a,
	0xea, 0x2a, 0x1b, 0x49, 0xe1, 0x3f, 0x92, 0xb2, 0x44, 0xa9, 0xf8, 0x03, 0x23, 0x69, 0x67, 0x69,
	0x73, 0xf8, 0x4b, 0xe6, 0x3a, 0x05, 0xf0, 0xe7, 0xb0, 0xc1, 0xef, 0xac, 0xd2, 0xf6, 0x92, 0x81,
	0x91, 0xdf, 0x9d, 0x66, 0x1b, 0x2c, 0xbb, 0x2c, 0xfe, 0x13, 0x31, 0xbe, 0x84, 0x8d, 0xa5, 0x07,
	0x87, 0xb8, 0x74, 0x24, 0x8e, 0xc3, 0x38, 0x39, 0x1f, 0x45, 0x18, 0xef, 0xa0, 0x95, 0x8d, 0x8d,
	0xa2, 0x03, 0x15, 0x9a, 0x85, 0xfc, 0x16, 0x31, 0x6e, 0x49, 0xcc, 0xc4, 0x01, 0xa9, 0xf3, 0x4b,
	0xc9, 0xf7, 0x4d, 0x4e, 0xbf, 0xf9, 0x13, 0xb4, 0x0b, 0x9d, 0x78, 0xf9, 0x71, 0xd0, 0x85, 0xd6,
	0xd1, 0x9b, 0xb7, 0xc3, 0xef, 0xad, 0xd9, 0xe9, 0x71, 0x5f, 0x13, 0x6f, 0x80, 0xe9, 0x68, 0x3c,
	0x3f, 0x9b, 0x9e, 0x9d, 0x4b, 0xce, 0xda, 0xc1, 0x3f, 0xa0, 0xae, 0x26, 0x21, 0xf4, 0x0d, 0x74,
	0xd4, 0xd7, 0x29, 0x8f, 0x09, 0x0e, 0xd0, 0xca, 0xc5, 0xde, 0x59, 0xe1, 0x18, 0x95, 0x17, 0xda,
	0x2b, 0x0d, 0x7d, 0x01, 0xd5, 0x13, 0x8f, 0xba, 0xa8, 0xfc, 0x48, 0xdf, 0x29, 0x93, 0x46, 0xe5,
	0xe8, 0xab, 0xbf, 0xef, 0xb9, 0x1e, 0x5f, 0xdc, 0x5c, 0x88, 0x4e, 0xb3, 0xbf, 0xb8, 0x8f, 0x48,
	0xac, 0xa6, 0xf2, 0xfd, 0x4b, 0x7c, 0x11, 0x7b, 0xf6, 0xbe, 0xfc, 0x2f, 0xc6, 0xf6, 0x95, 0xd9,
	0x45, 0x5d, 0x92, 0x5f, 0xff, 0x3f, 0x00, 0x00, 0xff, 0xff, 0xdd, 0x1d, 0xb3, 0x7e, 0x5f, 0x13,
	0x00, 0x00,
}
