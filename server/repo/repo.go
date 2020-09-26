package repo

import (
	"errors"
	"github.com/byronzhu-haha/chat/server/entity"
	"strconv"
	"sync"
	"sync/atomic"
)

type Repo interface {
	Save(u *entity.User) error
	Get(id string) (*entity.User, error)
	Del(id string) error
}

type UserIPRepo interface {
	SetUserIP(uid, ip string) error
	GetUserIP(uid string) (string, error)
}

type UserManager struct {
	users map[string]*entity.User
	mu    sync.RWMutex
}

func NewUserManager() Repo {
	return &UserManager{
		users: make(map[string]*entity.User),
	}
}

func (m *UserManager) Save(u *entity.User) error {
	if u == nil {
		return errors.New("user is nil")
	}
	m.mu.Lock()
	m.users[u.ID()] = u
	m.mu.Unlock()
	return nil
}

func (m *UserManager) Get(id string) (*entity.User, error) {
	m.mu.RLock()
	u, ok := m.users[id]
	m.mu.RUnlock()
	if !ok {
		return nil, errors.New("not found user")
	}
	return u, nil
}

func (m *UserManager) Del(id string) error {
	m.mu.Lock()
	delete(m.users, id)
	m.mu.Unlock()
	return nil
}

type UserIPManager struct {
	init bool
	ips  map[string]string
	mu   sync.RWMutex
}

var ipMgr UserIPRepo

func init() {
	ipMgr = &UserIPManager{
		ips: map[string]string{},
	}
}

func SetUserIP(uid, ip string) error {
	return ipMgr.SetUserIP(uid, ip)
}

func GetUserIP(uid string) (string, error) {
	return ipMgr.GetUserIP(uid)
}

func (m *UserIPManager) wrap(fn func()) {
	if !m.init {
		panic("ip manager is not init")
	}
	fn()
}

func (m *UserIPManager) SetUserIP(uid, ip string) (err error) {
	m.wrap(func() {
		if uid == "" || ip == "" {
			err = errors.New("uid or ip must not be nil")
			return
		}
		m.mu.Lock()
		m.ips[uid] = ip
		m.mu.Unlock()
	})
	return nil
}

func (m *UserIPManager) GetUserIP(uid string) (ip string, err error) {
	m.wrap(func() {
		if uid == "" {
			err = errors.New("uid must not be nil")
			return
		}
		m.mu.RLock()
		res, ok := m.ips[uid]
		m.mu.RUnlock()
		if !ok {
			err = errors.New("not found ip")
			return
		}
		ip = res
	})
	return ip, nil
}

var idGenerator uint64

func GenerateOneID() string {
	n := atomic.AddUint64(&idGenerator, 100000)
	return strconv.FormatUint(n, 10)
}
