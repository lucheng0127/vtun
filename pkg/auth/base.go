package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

type BaseAuthMgr struct {
	DB         string // User database file, csv format: <username>,<passwd base64 encode>
	AuthedUser map[string]string
	MLock      sync.Mutex
}

func (mgr *BaseAuthMgr) GetUsers(db string) ([][]string, error) {
	raw, err := os.ReadFile(db)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(raw))
	users, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (mgr *BaseAuthMgr) ListUser() ([]string, error) {
	users, err := mgr.GetUsers(mgr.DB)
	if err != nil {
		return nil, err
	}

	var u []string
	for _, uinfo := range users {
		if len(uinfo) < 2 {
			continue
		}

		u = append(u, uinfo[0])
	}

	return u, nil
}

func (mgr *BaseAuthMgr) AddUser(user, passwd string) error {
	users, err := mgr.GetUsers(mgr.DB)
	if err != nil {
		return err
	}

	for _, uinfo := range users {
		if len(uinfo) < 2 {
			continue
		}

		if user == uinfo[0] {
			return fmt.Errorf("user %s existed", user)
		}
	}

	// TODO: Add file lock if needed
	f, err := os.OpenFile(mgr.DB, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	passwdEnc := base64.StdEncoding.EncodeToString([]byte(passwd))
	line := fmt.Sprintf("%s,%s\n", user, passwdEnc)

	if _, err = f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func (mgr *BaseAuthMgr) ValidateUser(user, passwd string) error {
	if _, ok := mgr.AuthedUser[user]; ok {
		return fmt.Errorf("user %s already logged in", user)
	}

	users, err := mgr.GetUsers(mgr.DB)
	if err != nil {
		return err
	}

	for _, uinfo := range users {
		if len(uinfo) < 2 {
			continue
		}

		if uinfo[0] == user {
			if uinfo[1] == base64.StdEncoding.EncodeToString([]byte(passwd)) {
				mgr.MLock.Lock()
				mgr.AuthedUser[user] = "OK"
				mgr.MLock.Unlock()
				return nil
			}

			return fmt.Errorf("wrong passwd for user %s", user)
		}
	}

	return fmt.Errorf("user %s not exist", user)
}

func (mgr *BaseAuthMgr) LogoutUser(user string) {
	mgr.MLock.Lock()
	delete(mgr.AuthedUser, user)
	mgr.MLock.Unlock()
	log.Infof("user %s logged out", user)
}
