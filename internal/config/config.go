package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
)

type Config struct {
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
	User     string   `json:"user"`
	Name     *string  `json:"name"`
	Id       *string  `json:"id"`
	Headless *bool    `json:"headless"`
	New      *bool    `json:"new"`
	Delay    *float64 `json:"delay"`
}

func (cfg Config) Validate() error {
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
	return nil
}

func read(path string) (Config, error) {
	var config Config
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
