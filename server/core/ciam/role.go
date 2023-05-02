package ciam

type Role uint8

func (r Role) IsRegisteredUser() bool {
	return r == roleRegisteredUser
}

const (
	roleAnonymUser Role = iota
	roleRegisteredUser
)

var (
	QuotasAnonymUser = Quotas{
		PromptLengthMax:   100,
		RequestsPerMinute: 1,
		RequestsPerDay:    5,
	}
	QuotasRegisteredUser = Quotas{
		PromptLengthMax:   300,
		RequestsPerMinute: 3,
		RequestsPerDay:    20,
	}
)
