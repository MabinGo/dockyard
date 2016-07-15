package db

import (
	"fmt"

	"github.com/containerops/dockyard/auth/dao"
)

// Auth implements Authenticator interface to authenticate user against DB.
type DBAuthn struct{}

func NewDBAuthn() (*DBAuthn, error) {
	return &DBAuthn{}, nil
}

// Authenticate calls dao to authenticate user.
func (d *DBAuthn) Authenticate(userName, password string) (*dao.User, error) {
	user := &dao.User{Name: userName}
	if exist, err := user.Get(); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("Not found user:%s", userName)
	} else if exist && user.Status == dao.INACTIVE {
		return nil, fmt.Errorf("User:%s is inactive", userName)
	}

	if user.Password == dao.GeneratePwdBySalt(password, user.Salt) {
		return user, nil
	} else {
		return nil, fmt.Errorf("Password incorrect")
	}
}
