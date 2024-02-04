package token

type UserAccess struct {
	UserName string
	Roles    []string
}

func (ua *UserAccess) GetRoles() []string {
	return ua.Roles
}

func (ua *UserAccess) GetUsername() string {
	return ua.UserName
}
