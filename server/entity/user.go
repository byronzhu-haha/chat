package entity

type UserState byte

const (
	UserStateOffline UserState = iota
	UserStateOnline
)

type User struct {
	id    string
	name  string
	pwd   string
	state UserState
}

func NewUser(id, name, pwd string) *User {
	return &User{
		id:   id,
		name: name,
		pwd:  pwd,
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

func (u *User) State() UserState {
	return u.state
}

func (u *User) SetState(state UserState) {
	u.state = state
}
