package ssh

import (
	"testing"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

func TestJumpHostConfigBuilding(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	params := api.SSHConnectParams{
		Host:     "target.internal",
		Port:     22,
		Username: "appuser",
		Auth: api.SSHAuthParams{
			Type:     "password",
			Password: "target-password",
		},
		JumpHost: &api.SSHConnectParams{
			Host:     "bastion.example.com",
			Port:     2222,
			Username: "jules",
			Auth: api.SSHAuthParams{
				Type:     "password",
				Password: "secret-password",
			},
		},
	}

	config, err := mgr.buildClientConfig(params)
	if err != nil {
		t.Fatalf("failed to build target config: %v", err)
	}
	if config.User != "appuser" {
		t.Errorf("expected user appuser, got %s", config.User)
	}

	jumpConfig, err := mgr.buildClientConfig(*params.JumpHost)
	if err != nil {
		t.Fatalf("failed to build jump config: %v", err)
	}
	if jumpConfig.User != "jules" {
		t.Errorf("expected user jules, got %s", jumpConfig.User)
	}
}

func TestProxyJumpParsingErrorHandling(t *testing.T) {
	mgr := NewManager(func(method string, params interface{}) {})

	// Ensure empty proxy jump returns error
	_, err := mgr.connectViaProxyJump("", "127.0.0.1:22", nil)
	if err == nil {
		t.Fatal("expected error on empty proxy jump")
	}

	// For a real connection test, we'd need a mock SSH server.
	// We just ensure it starts dialing and fails with a dial error
	// rather than a parsing crash.
	_, err = mgr.connectViaProxyJump("user@invalid.host.local:2222", "127.0.0.1:22", nil)
	if err == nil {
		t.Fatal("expected dial error to invalid host")
	}
}
