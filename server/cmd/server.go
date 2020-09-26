package cmd

import (
	"context"
	"errors"
	"github.com/byronzhu-haha/chat/message"
	"github.com/byronzhu-haha/chat/server/conn"
	"github.com/byronzhu-haha/chat/server/entity"
	"github.com/byronzhu-haha/chat/server/repo"
	"github.com/byronzhu-haha/log"
	"os"
)

type ChatServer struct {
	init        bool
	connManager *conn.Manager
	userRepo    repo.Repo
	messages    chan message.Message
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		init:        true,
		connManager: conn.NewManager(),
		userRepo:    repo.NewUserManager(),
		messages:    make(chan message.Message, 1000),
	}
}

func (s *ChatServer) Run() {
	if !s.init {
		log.Errorf("server is not init")
		os.Exit(1)
	}
	err := s.connManager.Start()
	if err != nil {
		panic(err.Error())
	}
	done := make(chan struct{})
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go s.connManager.HandleMetadata(ctx, s.messages)
	go s.HandleMessage()
	select {
	case <-done:
		cancel()
	}
}

func (s *ChatServer) HandleMessage() {
	for m := range s.messages {
		head, _ := message.UnpackRequestHeader(m.Head)
		meta, err := message.UnpackMetadata(m.Body)
		if err != nil {
			log.Errorf("unpack meta failed, err: %v", err)
			continue
		}

		var (
			resp []byte
			code  = message.CodeOk
		)
		switch meta.Operate {
		case message.OperateTypeRegister:
			resp, err = s.Register(meta.User)
		case message.OperateTypeLogin:
			resp, err = s.Login(head.SrcAddr, meta.User)
		case message.OperateTypeLogout:
			resp, err = s.Logout(meta.User)
		case message.OperateTypeDelete:
			resp, err = s.Delete(meta.User)
		default:
			log.Infof("invalid operate, srcAddr: %s", head.SrcAddr)
			code = message.CodeInvalidOperate
		}
		if err != nil {
			log.Errorf("pack resp message failed, err: %+v", err)
			code = message.CodeFailed
		}
		respHead, _ := message.PackResponseHeader(head.SrcAddr, 0, code)
		msg, _ := message.Pack(message.MsgTypeResp, respHead, resp)
		s.connManager.SendMsg(head.SrcAddr, msg)
	}
}

func (s *ChatServer) Register(user *entity.User) (resp []byte, err error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	id := repo.GenerateOneID()
	err = s.userRepo.Save(entity.NewUser(id, user.Name(), user.Pwd()))
	if err == nil {
		resp = []byte(id)
	}
	return
}

func (s *ChatServer) Login(addr string, user *entity.User) (resp []byte, err error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	u, err := s.userRepo.Get(user.ID())
	if err != nil {
		return resp, err
	}
	if user.Pwd() != u.Pwd() {
		err = errors.New("passwd error")
		return resp, err
	}
	err = repo.SetUserIP(user.ID(), addr)
	if err != nil {
		return nil, err
	}
	u.SetState(entity.UserStateOnline)
	return
}

func (s *ChatServer) Logout(user *entity.User) (resp []byte, err error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	u, err := s.userRepo.Get(user.ID())
	if err != nil {
		return resp, err
	}
	u.SetState(entity.UserStateOffline)
	return
}

func (s *ChatServer) Delete(user *entity.User) (resp []byte, err error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}
	err = s.userRepo.Del(user.ID())
	return
}
