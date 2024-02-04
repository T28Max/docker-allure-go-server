package token

type UserInfo struct {
	Pass  string
	Roles []string
}

func (ua *UserInfo) GetRoles() []string {
	return ua.Roles
}

func (ua *UserInfo) GetPass() string {
	return ua.Pass
}
