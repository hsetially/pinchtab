package config

import (
	"strings"
	"testing"
)

func TestValidateFileConfig_Valid(t *testing.T) {
	maxTabs := 20
	fc := &FileConfig{
		Server: ServerConfig{
			Port: "9867",
			Bind: "127.0.0.1",
		},
		InstanceDefaults: InstanceDefaultsConfig{
			Mode:              "headless",
			MaxTabs:           &maxTabs,
			StealthLevel:      "light",
			TabEvictionPolicy: "reject",
		},
		MultiInstance: MultiInstanceConfig{
			Strategy:         "simple",
			AllocationPolicy: "fcfs",
		},
		Timeouts: TimeoutsConfig{
			ActionSec:   30,
			NavigateSec: 60,
		},
	}

	errs := ValidateFileConfig(fc)
	if len(errs) > 0 {
		t.Errorf("expected no errors for valid config, got: %v", errs)
	}
}

func TestValidateFileConfig_InvalidPort(t *testing.T) {
	tests := []struct {
		port    string
		wantErr bool
	}{
		{"9867", false},
		{"1", false},
		{"65535", false},
		{"0", true},
		{"65536", true},
		{"-1", true},
		{"abc", true},
		{"", false}, // empty is ok (uses default)
	}

	for _, tt := range tests {
		fc := &FileConfig{
			Server: ServerConfig{Port: tt.port},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("port=%q: got error=%v, want error=%v (errs: %v)", tt.port, hasErr, tt.wantErr, errs)
		}
	}
}

func TestValidateFileConfig_InvalidStealthLevel(t *testing.T) {
	tests := []struct {
		level   string
		wantErr bool
	}{
		{"light", false},
		{"medium", false},
		{"full", false},
		{"", false}, // empty is ok
		{"none", true},
		{"max", true},
		{"LIGHT", true}, // case sensitive
	}

	for _, tt := range tests {
		fc := &FileConfig{
			InstanceDefaults: InstanceDefaultsConfig{StealthLevel: tt.level},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("stealthLevel=%q: got error=%v, want error=%v", tt.level, hasErr, tt.wantErr)
		}
	}
}

func TestValidateFileConfig_InvalidEvictionPolicy(t *testing.T) {
	tests := []struct {
		policy  string
		wantErr bool
	}{
		{"reject", false},
		{"close_oldest", false},
		{"close_lru", false},
		{"", false},
		{"drop", true},
		{"lru", true},
	}

	for _, tt := range tests {
		fc := &FileConfig{
			InstanceDefaults: InstanceDefaultsConfig{TabEvictionPolicy: tt.policy},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("tabEvictionPolicy=%q: got error=%v, want error=%v", tt.policy, hasErr, tt.wantErr)
		}
	}
}

func TestValidateFileConfig_InvalidStrategy(t *testing.T) {
	tests := []struct {
		strategy string
		wantErr  bool
	}{
		{"simple", false},
		{"explicit", false},
		{"simple-autorestart", false},
		{"", false},
		{"auto", true},
		{"default", true},
	}

	for _, tt := range tests {
		fc := &FileConfig{
			MultiInstance: MultiInstanceConfig{Strategy: tt.strategy},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("strategy=%q: got error=%v, want error=%v", tt.strategy, hasErr, tt.wantErr)
		}
	}
}

func TestValidateFileConfig_InvalidAllocationPolicy(t *testing.T) {
	tests := []struct {
		policy  string
		wantErr bool
	}{
		{"fcfs", false},
		{"round_robin", false},
		{"random", false},
		{"", false},
		{"fifo", true},
		{"roundrobin", true}, // underscore required
	}

	for _, tt := range tests {
		fc := &FileConfig{
			MultiInstance: MultiInstanceConfig{AllocationPolicy: tt.policy},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("allocationPolicy=%q: got error=%v, want error=%v", tt.policy, hasErr, tt.wantErr)
		}
	}
}

func TestValidateFileConfig_InvalidAttachScheme(t *testing.T) {
	tests := []struct {
		schemes []string
		wantErr bool
	}{
		{[]string{"ws"}, false},
		{[]string{"wss"}, false},
		{[]string{"ws", "wss"}, false},
		{[]string{"http"}, true},
		{[]string{"ws", "https"}, true},
	}

	for _, tt := range tests {
		fc := &FileConfig{
			Attach: AttachConfig{AllowSchemes: tt.schemes},
		}
		errs := ValidateFileConfig(fc)
		hasErr := len(errs) > 0
		if hasErr != tt.wantErr {
			t.Errorf("allowSchemes=%v: got error=%v, want error=%v", tt.schemes, hasErr, tt.wantErr)
		}
	}
}

func TestValidateFileConfig_InvalidMaxTabs(t *testing.T) {
	zero := 0
	negative := -1
	positive := 10

	tests := []struct {
		name    string
		maxTabs *int
		wantErr bool
	}{
		{"nil", nil, false},
		{"positive", &positive, false},
		{"zero", &zero, true},
		{"negative", &negative, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FileConfig{
				InstanceDefaults: InstanceDefaultsConfig{MaxTabs: tt.maxTabs},
			}
			errs := ValidateFileConfig(fc)
			hasErr := len(errs) > 0
			if hasErr != tt.wantErr {
				t.Errorf("maxTabs=%v: got error=%v, want error=%v", tt.maxTabs, hasErr, tt.wantErr)
			}
		})
	}
}

func TestValidateFileConfig_InvalidTimeouts(t *testing.T) {
	fc := &FileConfig{
		Timeouts: TimeoutsConfig{
			ActionSec:   -1,
			NavigateSec: -1,
			ShutdownSec: -1,
			WaitNavMs:   -1,
		},
	}

	errs := ValidateFileConfig(fc)
	if len(errs) != 4 {
		t.Errorf("expected 4 timeout errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateFileConfig_InstancePortRange(t *testing.T) {
	start := 9900
	end := 9800 // invalid: start > end

	fc := &FileConfig{
		Server: ServerConfig{},
		MultiInstance: MultiInstanceConfig{
			InstancePortStart: &start,
			InstancePortEnd:   &end,
		},
	}

	errs := ValidateFileConfig(fc)
	if len(errs) != 1 {
		t.Errorf("expected 1 error for invalid port range, got %d: %v", len(errs), errs)
	}
	if len(errs) > 0 && !strings.Contains(errs[0].Error(), "start port") {
		t.Errorf("expected port range error, got: %v", errs[0])
	}
}

func TestValidateFileConfig_MultipleErrors(t *testing.T) {
	zero := 0
	fc := &FileConfig{
		Server: ServerConfig{
			Port: "99999", // invalid
		},
		InstanceDefaults: InstanceDefaultsConfig{
			MaxTabs:           &zero,           // invalid
			StealthLevel:      "superstealth",  // invalid
			TabEvictionPolicy: "delete_oldest", // invalid
		},
		MultiInstance: MultiInstanceConfig{
			Strategy:         "magical",  // invalid
			AllocationPolicy: "balanced", // invalid
		},
	}

	errs := ValidateFileConfig(fc)
	if len(errs) < 5 {
		t.Errorf("expected at least 5 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "server.port",
		Message: "port out of range",
	}
	expected := "server.port: port out of range"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestValidEnumValues(t *testing.T) {
	// Test that the valid values match the validation functions
	for _, level := range ValidStealthLevels() {
		if !isValidStealthLevel(level) {
			t.Errorf("ValidStealthLevels contains %q but isValidStealthLevel returns false", level)
		}
	}

	for _, policy := range ValidEvictionPolicies() {
		if !isValidEvictionPolicy(policy) {
			t.Errorf("ValidEvictionPolicies contains %q but isValidEvictionPolicy returns false", policy)
		}
	}

	for _, strategy := range ValidStrategies() {
		if !isValidStrategy(strategy) {
			t.Errorf("ValidStrategies contains %q but isValidStrategy returns false", strategy)
		}
	}

	for _, policy := range ValidAllocationPolicies() {
		if !isValidAllocationPolicy(policy) {
			t.Errorf("ValidAllocationPolicies contains %q but isValidAllocationPolicy returns false", policy)
		}
	}

	for _, scheme := range ValidAttachSchemes() {
		if !isValidAttachScheme(scheme) {
			t.Errorf("ValidAttachSchemes contains %q but isValidAttachScheme returns false", scheme)
		}
	}
}
