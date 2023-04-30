package ciam

type Role uint8

func (r Role) IsRegisteredUser() bool {
	return r == roleRegisteredUser
}

const (
	roleAnonymUser Role = iota
	roleRegisteredUser
)
