package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetConfigValue_ServerFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"server.port", "8080", func(fc *FileConfig) bool { return fc.Server.Port == "8080" }, false},
		{"server.bind", "0.0.0.0", func(fc *FileConfig) bool { return fc.Server.Bind == "0.0.0.0" }, false},
		{"server.token", "secret", func(fc *FileConfig) bool { return fc.Server.Token == "secret" }, false},
		{"server.stateDir", "/tmp/state", func(fc *FileConfig) bool { return fc.Server.StateDir == "/tmp/state" }, false},
		{"server.unknown", "value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_BrowserAndInstanceDefaultsFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"browser.version", "144.0.7559.133", func(fc *FileConfig) bool { return fc.Browser.ChromeVersion == "144.0.7559.133" }, false},
		{"browser.binary", "/tmp/chrome", func(fc *FileConfig) bool { return fc.Browser.ChromeBinary == "/tmp/chrome" }, false},
		{"instanceDefaults.mode", "headed", func(fc *FileConfig) bool { return fc.InstanceDefaults.Mode == "headed" }, false},
		{"instanceDefaults.maxTabs", "50", func(fc *FileConfig) bool { return *fc.InstanceDefaults.MaxTabs == 50 }, false},
		{"instanceDefaults.stealthLevel", "full", func(fc *FileConfig) bool { return fc.InstanceDefaults.StealthLevel == "full" }, false},
		{"instanceDefaults.tabEvictionPolicy", "close_lru", func(fc *FileConfig) bool { return fc.InstanceDefaults.TabEvictionPolicy == "close_lru" }, false},
		{"instanceDefaults.blockAds", "yes", func(fc *FileConfig) bool { return *fc.InstanceDefaults.BlockAds == true }, false},
		{"profiles.baseDir", "/tmp/profiles", func(fc *FileConfig) bool { return fc.Profiles.BaseDir == "/tmp/profiles" }, false},
		{"instanceDefaults.noRestore", "maybe", nil, true},
		{"instanceDefaults.maxTabs", "many", nil, true},
		{"instanceDefaults.unknown", "value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_SecurityFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"security.allowEvaluate", "true", func(fc *FileConfig) bool { return *fc.Security.AllowEvaluate == true }, false},
		{"security.allowMacro", "1", func(fc *FileConfig) bool { return *fc.Security.AllowMacro == true }, false},
		{"security.allowScreencast", "false", func(fc *FileConfig) bool { return *fc.Security.AllowScreencast == false }, false},
		{"security.allowDownload", "on", func(fc *FileConfig) bool { return *fc.Security.AllowDownload == true }, false},
		{"security.allowUpload", "off", func(fc *FileConfig) bool { return *fc.Security.AllowUpload == false }, false},
		{"security.allowEvaluate", "maybe", nil, true},
		{"security.unknown", "true", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_MultiInstanceFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"multiInstance.strategy", "explicit", func(fc *FileConfig) bool { return fc.MultiInstance.Strategy == "explicit" }, false},
		{"multiInstance.allocationPolicy", "round_robin", func(fc *FileConfig) bool { return fc.MultiInstance.AllocationPolicy == "round_robin" }, false},
		{"multiInstance.instancePortStart", "9900", func(fc *FileConfig) bool { return *fc.MultiInstance.InstancePortStart == 9900 }, false},
		{"multiInstance.unknown", "value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_AttachFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"attach.enabled", "true", func(fc *FileConfig) bool { return fc.Attach.Enabled != nil && *fc.Attach.Enabled }, false},
		{"attach.allowHosts", "localhost, chrome.internal", func(fc *FileConfig) bool {
			return len(fc.Attach.AllowHosts) == 2 && fc.Attach.AllowHosts[1] == "chrome.internal"
		}, false},
		{"attach.allowSchemes", "ws,wss", func(fc *FileConfig) bool {
			return len(fc.Attach.AllowSchemes) == 2 && fc.Attach.AllowSchemes[0] == "ws"
		}, false},
		{"attach.enabled", "maybe", nil, true},
		{"attach.unknown", "value", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_TimeoutsFields(t *testing.T) {
	tests := []struct {
		path    string
		value   string
		check   func(*FileConfig) bool
		wantErr bool
	}{
		{"timeouts.actionSec", "60", func(fc *FileConfig) bool { return fc.Timeouts.ActionSec == 60 }, false},
		{"timeouts.navigateSec", "120", func(fc *FileConfig) bool { return fc.Timeouts.NavigateSec == 120 }, false},
		{"timeouts.shutdownSec", "30", func(fc *FileConfig) bool { return fc.Timeouts.ShutdownSec == 30 }, false},
		{"timeouts.waitNavMs", "2000", func(fc *FileConfig) bool { return fc.Timeouts.WaitNavMs == 2000 }, false},
		{"timeouts.actionSec", "fast", nil, true},
		{"timeouts.unknown", "10", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"="+tt.value, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.check(fc) {
				t.Errorf("SetConfigValue() did not set value correctly")
			}
		})
	}
}

func TestSetConfigValue_InvalidPaths(t *testing.T) {
	tests := []string{
		"port",          // missing section
		"",              // empty
		"unknown.field", // unknown section
		"server",        // missing field
		"a.b.c",         // too many parts (we only split on first .)
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			fc := &FileConfig{}
			err := SetConfigValue(fc, path, "value")
			if err == nil {
				t.Errorf("SetConfigValue(%q) should have failed", path)
			}
		})
	}
}

func TestPatchConfigJSON(t *testing.T) {
	fc := &FileConfig{
		Server: ServerConfig{
			Port: "9867",
			Bind: "127.0.0.1",
		},
		InstanceDefaults: InstanceDefaultsConfig{
			StealthLevel: "light",
		},
	}

	// Patch to change port and add token
	patch := `{"server": {"port": "8080", "token": "secret"}}`
	if err := PatchConfigJSON(fc, patch); err != nil {
		t.Fatalf("PatchConfigJSON() error = %v", err)
	}

	if fc.Server.Port != "8080" {
		t.Errorf("port = %v, want 8080", fc.Server.Port)
	}
	if fc.Server.Token != "secret" {
		t.Errorf("token = %v, want secret", fc.Server.Token)
	}
	// Bind should be preserved
	if fc.Server.Bind != "127.0.0.1" {
		t.Errorf("bind = %v, want 127.0.0.1 (should be preserved)", fc.Server.Bind)
	}
	// InstanceDefaults.StealthLevel should be preserved
	if fc.InstanceDefaults.StealthLevel != "light" {
		t.Errorf("stealthLevel = %v, want light (should be preserved)", fc.InstanceDefaults.StealthLevel)
	}
}

func TestPatchConfigJSON_NestedMerge(t *testing.T) {
	fc := &FileConfig{
		InstanceDefaults: InstanceDefaultsConfig{
			StealthLevel:      "light",
			TabEvictionPolicy: "reject",
		},
	}

	// Patch instanceDefaults section, should merge not replace
	patch := `{"instanceDefaults": {"stealthLevel": "full"}}`
	if err := PatchConfigJSON(fc, patch); err != nil {
		t.Fatalf("PatchConfigJSON() error = %v", err)
	}

	if fc.InstanceDefaults.StealthLevel != "full" {
		t.Errorf("stealthLevel = %v, want full", fc.InstanceDefaults.StealthLevel)
	}
	// tabEvictionPolicy should be preserved
	if fc.InstanceDefaults.TabEvictionPolicy != "reject" {
		t.Errorf("tabEvictionPolicy = %v, want reject (should be preserved)", fc.InstanceDefaults.TabEvictionPolicy)
	}
}

func TestPatchConfigJSON_InvalidJSON(t *testing.T) {
	fc := &FileConfig{}
	err := PatchConfigJSON(fc, "not json")
	if err == nil {
		t.Error("PatchConfigJSON() should fail on invalid JSON")
	}
}

func TestLoadAndSaveFileConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	_ = os.Setenv("PINCHTAB_CONFIG", configPath)
	defer func() { _ = os.Unsetenv("PINCHTAB_CONFIG") }()

	// Load (should return empty config for non-existent file)
	fc, path, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("LoadFileConfig() error = %v", err)
	}
	if path != configPath {
		t.Errorf("path = %v, want %v", path, configPath)
	}

	// Modify
	fc.Server.Port = "8080"
	fc.InstanceDefaults.StealthLevel = "full"

	// Save
	if err := SaveFileConfig(fc, path); err != nil {
		t.Fatalf("SaveFileConfig() error = %v", err)
	}

	// Load again
	fc2, _, err := LoadFileConfig()
	if err != nil {
		t.Fatalf("LoadFileConfig() second time error = %v", err)
	}

	if fc2.Server.Port != "8080" {
		t.Errorf("loaded port = %v, want 8080", fc2.Server.Port)
	}
	if fc2.InstanceDefaults.StealthLevel != "full" {
		t.Errorf("loaded stealthLevel = %v, want full", fc2.InstanceDefaults.StealthLevel)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{"true", true, false},
		{"True", true, false},
		{"TRUE", true, false},
		{"1", true, false},
		{"yes", true, false},
		{"on", true, false},
		{"false", false, false},
		{"False", false, false},
		{"0", false, false},
		{"no", false, false},
		{"off", false, false},
		{"maybe", false, true},
		{"", false, true},
		{"2", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBool(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBool(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
