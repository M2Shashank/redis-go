package auth

import (
	"crypto/sha256"
	"encoding/hex"
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
	for flag, present := range user.Flags {
		if present {
			flags = append(flags, flag)
		}
	}
	return flags
}

func (m *AuthManager) GetUserPswds(username string) []string {
	user := m.Users[username]
	if user == nil {
		return []string{}
	}
	passwords := []string{}
	for password, valid := range user.Passwords {
		if valid {
			passwords = append(passwords, password)
		}
	}
	return passwords
}

func (m *AuthManager) SetUserPassword(username string, password string) {
	passHash := sha256.Sum256([]byte(password))
	hexHash := hex.EncodeToString(passHash[:])
	m.Users[username].Passwords[hexHash] = true
	m.Users[username].Flags["nopass"] = false
}

func (m *AuthManager) ValidateCredentials(username string, password string) bool {
	user := m.Users[username]
	if user == nil {
		return false
	}

	passHash := sha256.Sum256([]byte(password))
	hexHash := hex.EncodeToString(passHash[:])

	if user.Passwords[hexHash] {
		return true
	}

	return false
}