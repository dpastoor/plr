package config

import (
	"encoding/json"
	"os"
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
