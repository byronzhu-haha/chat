package repo

import (
	"errors"
	"github.com/byronzhu-haha/chat/entity/user"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
)

type Repo interface {
	Save(u *user.User) error
	Get(id string) (*user.User, error)
	Del(id string) error
	List(username string) ([]*user.User, error)
	DelUserFriend(userid, friendID string) error
	AddUserFriend(userid, friendID string) error
	ListUserFriend(userid string) []user.BriefUser
}

type UserIPRepo interface {
	SetUserIP(uid, ip string) error
	GetUserIP(uid string) (string, error)
}

var (
	ErrNotFoundUser = errors.New("not found user who want to search")
)

type UserManager struct {
	users map[string]*user.User
	mu    sync.RWMutex
}

func NewUserManager() Repo {
	return &UserManager{
		users: make(map[string]*user.User),
	}
}

func (m *UserManager) Save(u *user.User) error {
	if u == nil {
		return errors.New("user is nil")
	}
	m.mu.Lock()
	m.users[u.ID()] = u
	m.mu.Unlock()
	return nil
}

func (m *UserManager) Get(id string) (*user.User, error) {
	m.mu.RLock()
	u, ok := m.users[id]
	if !ok {
		m.mu.RUnlock()
		return nil, ErrNotFoundUser
	}
	m.mu.RUnlock()
	return u, nil
}

func (m *UserManager) List(username string) (res []*user.User, err error) {
	var (
		reg = regexp.MustCompile("*"+username+"*")
	)
	m.mu.RLock()
	for _, u := range m.users {
		if reg.MatchString(u.Name()) {
			res = append(res, u)
		}
	}
	m.mu.RUnlock()
	if len(res) == 0 {
		err = ErrNotFoundUser
	}
	return res, err
}

func (m *UserManager) Del(id string) error {
	m.mu.Lock()
	delete(m.users, id)
	m.mu.Unlock()
	return nil
}

func (m *UserManager) AddUserFriend(userid, friendID string) error {
	m.mu.Lock()
	u, ok := m.users[userid]
	if !ok {
		m.mu.Unlock()
		return ErrNotFoundUser
	}
	f, ok := m.users[friendID]
	if !ok {
		m.mu.Unlock()
		return ErrNotFoundUser
	}
	u.AddFriend(friendID, f.Name())
	m.mu.Unlock()
	return nil
}

func (m *UserManager) DelUserFriend(userid, friendID string) error {
	m.mu.Lock()
	u, ok := m.users[userid]
	if !ok {
		m.mu.Unlock()
		return ErrNotFoundUser
	}
	u.DelFriend(friendID)
	m.mu.Unlock()
	return nil
}

func (m *UserManager) ListUserFriend(userid string) []user.BriefUser {
	m.mu.RLock()
	u, ok := m.users[userid]
	if !ok {
		m.mu.RUnlock()
		return []user.BriefUser{}
	}
	res := u.ListFriend()
	for i, re := range res {
		f, ok := m.users[re.ID]
		if !ok {
			continue
		}
		res[i].State = f.State()
	}
	u.SortFriend(res)
	m.mu.RUnlock()
	return res
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
