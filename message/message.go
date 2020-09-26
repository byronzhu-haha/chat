package message

import (
	"bytes"
	"encoding/gob"
	"github.com/byronzhu-haha/chat/server/entity"
)

const serverLogo = "server"

func marshal(v interface{}) ([]byte, error) {
	var buf = &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func unmarshal(data []byte, res interface{}) error {
	return gob.NewDecoder(bytes.NewBuffer(data)).Decode(res)
}

type MsgType byte

const (
	MsgTypeReq MsgType = iota
	MsgTypeResp
	MsgTypeChat
)

type Message struct {
	MsgType MsgType
	Head    []byte
	Body    []byte
}

func Pack(msgType MsgType, head, body []byte) ([]byte, error) {
	return marshal(&Message{
		MsgType: msgType,
		Head:    head,
		Body:    body,
	})
}

func Unpack(data []byte) (msg Message, err error) {
	err = unmarshal(data, &msg)
	return msg, err
}

func (m *Message) IsRequestMsg() bool {
	return m.MsgType == MsgTypeReq
}

func (m *Message) IsRespMsg() bool {
	return m.MsgType == MsgTypeResp
}

func (m *Message) IsChatMsg() bool {
	return m.MsgType == MsgTypeChat
}

type RequestHeader struct {
	SrcAddr  string
	DestAddr string
}

func PackRequestHeader(srcAddr string) ([]byte, error) {
	return marshal(&RequestHeader{
		SrcAddr:  srcAddr,
		DestAddr: serverLogo,
	})
}

func UnpackRequestHeader(data []byte) (head RequestHeader, err error) {
	err = unmarshal(data, &head)
	return head, err
}

type Code int32

const (
	CodeOk Code = iota
	CodeFailed
	CodeTimeout
	CodeInvalidOperate
)

type ResponseHeader struct {
	Seq      int
	Code     Code
	DestAddr string
}

func PackResponseHeader(destAddr string, seq int, code Code) ([]byte, error) {
	return marshal(&ResponseHeader{
		Seq:      seq,
		Code:     code,
		DestAddr: destAddr,
	})
}

func UnpackResponseHeader(data []byte) (head ResponseHeader, err error) {
	err = unmarshal(data, &head)
	return head, err
}

type ChatHeader struct {
	SrcAddr    string
	SrcUserID  string
	DestUserID string
}

func PackChatHeader(srcAddr, srcUserID, destUserID string) ([]byte, error) {
	return marshal(&ChatHeader{
		SrcAddr:    srcAddr,
		SrcUserID:  srcUserID,
		DestUserID: destUserID,
	})
}

func UnpackChatHeader(data []byte) (head ChatHeader, err error) {
	err = unmarshal(data, &head)
	return head, err
}

type OperateType byte

const (
	OperateTypeRegister OperateType = iota + 1 // 注册
	OperateTypeLogin                           // 登录
	OperateTypeLogout                          // 登出
	OperateTypeDelete                          // 注销
)

type ServerMetadata struct {
	Operate OperateType
	User    *entity.User
}

func PackMetadata(op OperateType, user *entity.User) ([]byte, error) {
	meta := &ServerMetadata{
		Operate: op,
		User:    user,
	}
	var buf = &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(meta)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func UnpackMetadata(buf []byte) (ServerMetadata, error) {
	var res ServerMetadata
	err := gob.NewDecoder(bytes.NewBuffer(buf)).Decode(&res)
	if err != nil {
		return res, err
	}
	return res, nil
}
