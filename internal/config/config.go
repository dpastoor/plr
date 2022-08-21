package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
)

type Scenarios struct {
	Users    []User    `json:"users"`
	Sessions []Session `json:"sessions"`
	Url      string    `json:"url"`
}

// User defines a new User
type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Session defines a new selenium session
type Session struct {
	User            string   `json:"user,omitempty"`
	RemoteCmdBase64 string   `json:"remote_cmd_base64,omitempty"`
	Name            *string  `json:"name,omitempty"`
	Id              *string  `json:"id,omitempty"`
	Headless        *bool    `json:"headless,omitempty"`
	New             *bool    `json:"new,omitempty"`
	Delay           *float64 `json:"delay,omitempty"`
}

func (cfg Scenarios) Validate() error {
	for _, user := range cfg.Users {
		if user.Name == "" || user.Password == "" {
			return errors.New("must set user name and password for all users")
		}
	}
	users := lo.Map(cfg.Users, func(user User, _ int) string {
		return user.Name
	})
	sessionUsers := lo.Map(cfg.Sessions, func(session Session, _ int) string {
		return session.User
	})
	_, notUsers := lo.Difference(users, sessionUsers)
	if len(notUsers) > 0 {
		return fmt.Errorf("all session user(s) must be defined in user section, currently missing: %v", strings.Join(lo.Uniq(notUsers), ", "))
	}

	for _, session := range cfg.Sessions {
		if session.New != nil && !*session.New {
			if session.Name == nil || (session.Name != nil && *session.Name == "") {
				return errors.New("any non-new session must also have a name")
			}
		}
		if session.RemoteCmdBase64 == "" {
			return errors.New("must set remote_cmd_base64 for all sessions")
		}
	}
	return nil
}

func read(path string) (Scenarios, error) {
	var config Scenarios
	file, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	err = config.Validate()
	if err != nil {
		return config, err
	}
	return config, nil
}
