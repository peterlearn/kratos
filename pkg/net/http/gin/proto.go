package gin

import (
	"github.com/gogo/protobuf/proto"
	"github.com/peterlearn/kratos/v1/pkg/ecode"
)

type ProtoMessage struct {
	Code                 int32    `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Data                 []byte   `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func TOPROTO(buf []byte, err error) (data *ProtoMessage) {
	bcode := ecode.Cause(err)
	return &ProtoMessage{
		Code:    int32(bcode.Code()),
		Message: bcode.Message(),
		Data:    buf,
	}
}
func (m *ProtoMessage) Reset()         { *m = ProtoMessage{} }
func (m *ProtoMessage) String() string { return proto.CompactTextString(m) }
func (*ProtoMessage) ProtoMessage()    {}
