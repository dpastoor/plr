package config_test

import (
	"testing"

	"github.com/dpastoor/plr/internal/config"
	"github.com/metrumresearchgroup/wrapt"
)

func TestNewConfig(tt *testing.T) {
	tests := []struct {
		name    string
		session config.Session
	}{
		{
			name: "simple",
			session: config.Session{
				User: "user1",
			},
		},
		{
			name: "with optional strings",
			session: config.Session{
				User: "user1",
				Name: config.StringPtr("test-session"),
				Id:   config.StringPtr("abc-1234"),
			},
		},
		{
			name: "truthy set",
			session: config.Session{
				User:     "user1",
				New:      config.BoolPtr(true),
				Headless: config.BoolPtr(true),
				Delay:    config.Float64Ptr(1.5),
			},
		},
		{
			name: "falsey set",
			session: config.Session{
				User:     "user1",
				New:      config.BoolPtr(false),
				Headless: config.BoolPtr(false),
				Delay:    config.Float64Ptr(0.0),
			},
		},
	}
	config, err := config.Read("testdata/opts.json")
	if err != nil {
		tt.Fatalf("failed to read config: %v", err)
	}
	for i, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			t.R.Equal(test.session, config.Sessions[i])
		})
	}
}

func TestMissingUser(tt *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "missing user2",
			path: "testdata/missing-user.json",
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			_, err := config.Read(test.path)
			t.R.Error(err)
			t.A.ErrorContains(err, "user2")
		})
	}
}
