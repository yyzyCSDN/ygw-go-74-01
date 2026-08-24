package main

import "testing"

func TestLoadConfigDefaults(t *testing.T) {
	cfg := loadConfigFromArgs(nil)
	if cfg.Addr != "127.0.0.1" || cfg.Port != 8090 || cfg.DataDir != "./data" || cfg.WebDir != "web" {
		t.Fatalf("defaults = %+v", cfg)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	cfg := loadConfigFromArgs([]string{"-addr", "0.0.0.0", "-port", "9090", "-dir", "/tmp/data", "-web-dir", "/tmp/web"})
	if cfg.Addr != "0.0.0.0" || cfg.Port != 9090 || cfg.DataDir != "/tmp/data" || cfg.WebDir != "/tmp/web" {
		t.Fatalf("overrides = %+v", cfg)
	}
}
