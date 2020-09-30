package cmd

import (
	"context"
	"errors"
	"github.com/byronzhu-haha/chat/entity/message"
	"github.com/byronzhu-haha/chat/entity/user"
	"github.com/byronzhu-haha/chat/server/conn"
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
			code = message.CodeOk
		)
		switch meta.Operate {
		case message.OperateTypeRegister:
			resp, err = s.Register(meta.Username, meta.Passwd)
		case message.OperateTypeLogin:
			resp, err = s.Login(head.SrcAddr, meta.Username, meta.Passwd)
		case message.OperateTypeLogout:
			resp, err = s.Logout(meta.Username)
		case message.OperateTypeDelete:
			resp, err = s.Delete(meta.Username)
		case message.OperateTypeSearchFriend:
			resp, err = s.SearchFriend(meta.Username, meta.Userid)
		case message.OperateTypeMakeFriend:
			resp, err = s.MakeFriend(meta.Userid, meta.DestUserID)
		case message.OperateTypeDeleteFriend:
			resp, err = s.DeleteFriend(meta.Userid, meta.DestUserID)
		case message.OperateTypeListFriend:
			resp, err = s.ListFriend(meta.Userid)
		default:
			log.Infof("invalid operate, srcAddr: %s", head.SrcAddr)
			code = message.CodeInvalidOperate
		}
		if err != nil {
			log.Errorf("pack resp message failed, err: %+v", err)
			code = message.CodeFailed
		}
		respHead, _ := message.PackResponseHeader(head.SrcAddr, meta.Operate, 0, code)
		msg, _ := message.Pack(message.MsgTypeResp, respHead, resp)
		s.connManager.SendMsg(head.SrcAddr, msg)
	}
}

func (s *ChatServer) Register(name, pwd string) (resp []byte, err error) {
	id := repo.GenerateOneID()
	err = s.userRepo.Save(user.NewUser(id, name, pwd, user.Offline))
	if err == nil {
		resp = []byte(id)
	}
	return
}

func (s *ChatServer) Login(addr string, userid, pwd string) (resp []byte, err error) {
	u, err := s.userRepo.Get(userid)
	if err != nil {
		return resp, err
	}
	if pwd != u.Pwd() {
		err = errors.New("passwd error")
		return resp, err
	}
	err = repo.SetUserIP(userid, addr)
	if err != nil {
		return resp, err
	}
	u.SetState(user.Online)
	return
}

func (s *ChatServer) Logout(userid string) (resp []byte, err error) {
	u, err := s.userRepo.Get(userid)
	if err != nil {
		return resp, err
	}
	u.SetState(user.Offline)
	return
}

func (s *ChatServer) Delete(userid string) (resp []byte, err error) {
	err = s.userRepo.Del(userid)
	return
}

func (s *ChatServer) SearchFriend(username, userid string) (resp []byte, err error) {
	var users = &message.UserList{}
	u, err := s.userRepo.Get(userid)
	if err == nil {
		*users = append(*users, u)
		return users.Marshal()
	}
	if username == "" {
		return resp, repo.ErrNotFoundUser
	}
	us, err := s.userRepo.List(username)
	if err != nil {
		return resp, err
	}
	for i := 0; i < len(us); i++ {
		*users = append(*users, us[i])
	}
	return users.Marshal()
}

func (s *ChatServer) MakeFriend(userid, friendID string) (resp []byte, err error) {
	err = s.userRepo.AddUserFriend(userid, friendID)
	if err != nil {
		return resp, err
	}

	return s.listFriend(userid)
}

func (s *ChatServer) DeleteFriend(userid, friendID string) (resp []byte, err error) {
	err = s.userRepo.DelUserFriend(userid, friendID)
	if err != nil {
		return resp, err
	}

	return s.listFriend(userid)
}

func (s *ChatServer) ListFriend(userid string) (resp []byte, err error) {
	return s.listFriend(userid)
}

func (s *ChatServer) listFriend(userid string) (resp []byte, err error) {
	var (
		us = s.userRepo.ListUserFriend(userid)
		fs = &message.UserList{}
	)

	for _, u := range us {
		*fs = append(*fs, user.NewUser(u.ID, u.Name, "", u.State))
	}

	return fs.Marshal()
}