package auth

import (
	"crypto/sha256"
)

type AuthUser struct {
	Username  string
	Flags     map[string]bool
	Passwords map[string]bool
}

func DefaultUser() *AuthUser {
	return &AuthUser{
		Username: "default",
		Flags: map[string]bool{
			"nopass": true,
		},
		Passwords: map[string]bool{},
	}
}

type AuthManager struct {
	// All types of users stored inside manager
	Users map[string]*AuthUser
}

func NewAuthManager() *AuthManager {
	return &AuthManager{
		Users: map[string]*AuthUser{
			"default": DefaultUser(),
		},
	}
}

func (m *AuthManager) GetUserFlags(username string) []string {
	user := m.Users[username]
	if user == nil {
		return []string{}
	}
	flags := []string{}
	for flag := range user.Flags {
		flags = append(flags, flag)
	}
	return flags
}

func (m *AuthManager) GetUserPswds(username string) []string {
	user := m.Users[username]	
	if user == nil {
		return []string{}
	}
	passwords := []string{}
	for password := range user.Passwords {
		passwords = append(passwords, password)
	}
	return passwords
}

func (m *AuthManager) SetUserPassword(username string, password string) {
	passHash := sha256.Sum256([]byte(password))
	m.Users[username].Passwords[string(passHash[:])] = true
	m.Users[username].Flags["nopass"] = false
}