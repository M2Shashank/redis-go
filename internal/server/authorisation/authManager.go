package auth

type authUser struct {
	Username string
}

type AuthManager struct {
	// All types of users stored inside manager
	Users map[string]*authUser
}

func NewAuthManager() *AuthManager {
	manager := &AuthManager{
		Users: make(map[string]*authUser),
	}
	
	manager.Users["default"] = &authUser{
		Username: "default",
	}
	return manager
}