package version

import (
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	// Default value should be "dev"
	info := Info()
	if info != Version {
		t.Errorf("Info() = %q, want %q", info, Version)
	}
}

func TestFull(t *testing.T) {
	full := Full()
	if !strings.Contains(full, Version) {
		t.Errorf("Full() = %q, should contain Version %q", full, Version)
	}
	if !strings.Contains(full, "commit:") {
		t.Errorf("Full() = %q, should contain 'commit:'", full)
	}
	if !strings.Contains(full, "built:") {
		t.Errorf("Full() = %q, should contain 'built:'", full)
	}
}
