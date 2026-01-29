package rcon

import (
	"errors"
	"reflect"
	"testing"
)

type MockConsole struct {
	WriteFunc func(cmd string) (int, error)
	ReadFunc  func() (string, int, error)
}

func (m *MockConsole) Write(cmd string) (int, error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(cmd)
	}
	return 0, nil
}

func (m *MockConsole) Read() (string, int, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc()
	}
	return "", 0, nil
}

func TestReload(t *testing.T) {
	tests := []struct {
		name    string
		mock    Console
		wantErr bool
	}{
		{
			name: "success",
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					if cmd != "reload" {
						t.Errorf("unexpected command: %s", cmd)
					}
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "Reload complete.", 1, nil
				},
			},
			wantErr: false,
		},
		{
			name: "write error",
			mock: &MockConsole{
				WriteFunc: func(_ string) (int, error) {
					return 0, errors.New("write error")
				},
				ReadFunc: nil,
			},
			wantErr: true,
		},
		{
			name: "read error",
			mock: &MockConsole{
				WriteFunc: func(_ string) (int, error) {
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "", 0, errors.New("read error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Reload(tt.mock); (err != nil) != tt.wantErr {
				t.Errorf("Reload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWhitelistSwitch(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		mock    Console
		wantErr bool
	}{
		{
			name:    "enable whitelist",
			enabled: true,
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					if cmd != "whitelist on" {
						t.Errorf("unexpected command: %s", cmd)
					}
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "Whitelist is now on", 1, nil
				},
			},
			wantErr: false,
		},
		{
			name:    "disable whitelist",
			enabled: false,
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					if cmd != "whitelist off" {
						t.Errorf("unexpected command: %s", cmd)
					}
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "Whitelist is now off", 1, nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WhitelistSwitch(tt.mock, tt.enabled); (err != nil) != tt.wantErr {
				t.Errorf("WhitelistSwitch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestListWhitelist(t *testing.T) {
	tests := []struct {
		name    string
		mock    Console
		want    []string
		wantErr bool
	}{
		{
			name: "list users",
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					if cmd != "whitelist list" {
						t.Errorf("unexpected command: %s", cmd)
					}
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "There are 2 whitelisted players: hoge, fuga", 1, nil
				},
			},
			want:    []string{"hoge", "fuga"},
			wantErr: false,
		},
		{
			name: "no users",
			mock: &MockConsole{
				WriteFunc: func(_ string) (int, error) {
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "There are no whitelisted players", 1, nil
				},
			},
			want:    []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ListWhitelist(tt.mock)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListWhitelist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListWhitelist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOp(t *testing.T) {
	tests := []struct {
		name    string
		users   []string
		mock    Console
		wantErr bool
	}{
		{
			name:  "success",
			users: []string{"user1"},
			mock: &MockConsole{
				WriteFunc: func(cmd string) (int, error) {
					if cmd != "op user1" {
						t.Errorf("unexpected command: %s", cmd)
					}
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "Made user1 a server operator", 1, nil
				},
			},
			wantErr: false,
		},
		{
			name:  "player does not exist",
			users: []string{"user1"},
			mock: &MockConsole{
				WriteFunc: func(_ string) (int, error) {
					return 1, nil
				},
				ReadFunc: func() (string, int, error) {
					return "That player does not exist", 1, nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Op(tt.mock, tt.users); (err != nil) != tt.wantErr {
				t.Errorf("Op() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
