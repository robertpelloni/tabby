package ssh

import (
	"testing"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
)

func TestFingerprint(t *testing.T) {
	p1 := api.SSHConnectParams{
		Host:     "localhost",
		Port:     22,
		Username: "root",
	}
	p2 := api.SSHConnectParams{
		Host:     "localhost",
		Port:     22,
		Username: "root",
	}
	p3 := api.SSHConnectParams{
		Host:     "localhost",
		Port:     2222,
		Username: "root",
	}

	f1 := getFingerprint(p1)
	f2 := getFingerprint(p2)
	f3 := getFingerprint(p3)

	if f1 != f2 {
		t.Errorf("expected fingerprints to match for identical params")
	}
	if f1 == f3 {
		t.Errorf("expected fingerprints to differ for different ports")
	}
}
