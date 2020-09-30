package user

import (
	"bytes"
	"encoding/gob"
	"sort"
)

type State byte

const (
	Offline State = iota
	Online
)

type User struct {
	id      string
	name    string
	pwd     string
	state   State
	friends map[string]string
}

type BriefUser struct {
	ID    string
	Name  string
	State State
}

func NewUser(id, name, pwd string, state State) *User {
	return &User{
		id:      id,
		name:    name,
		pwd:     pwd,
		state:   state,
		friends: make(map[string]string),
	}
}

func (u *User) ID() string {
	return u.id
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Pwd() string {
	return u.pwd
}

func (u *User) State() State {
	return u.state
}

func (u *User) SetState(state State) {
	u.state = state
}

func (u *User) AddFriend(userid, username string) {
	u.friends[userid] = username
}

func (u *User) DelFriend(userid string) {
	delete(u.friends, userid)
}

func (u *User) ListFriend() []BriefUser {
	res := make([]BriefUser, 0, len(u.friends))
	for id, name := range u.friends {
		res = append(res, BriefUser{
			ID:    id,
			Name:  name,
			State: Offline,
		})
	}
	return res
}

func (u *User) SortFriend(us []BriefUser) {
	sort.Slice(us, func(i, j int) bool {
		if us[i].State > us[j].State {
			return true
		}
		if us[i].State == us[j].State {
			return us[i].ID < us[j].ID
		}
		return false
	})
}

func (u *User) Marshal() (buf []byte, err error) {
	err = gob.NewEncoder(bytes.NewBuffer(buf)).Encode(u)
	return buf, err
}

func (u *User) Unmarshal(buf []byte) error {
	return gob.NewDecoder(bytes.NewBuffer(buf)).Decode(u)
}
