package settings

import "testing"

func TestInitLoadsExampleConfig(t *testing.T) {
	orig := GlobalConfig
	t.Cleanup(func() {
		GlobalConfig = orig
	})
	t.Chdir("..")

	if err := Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	if GlobalConfig == nil {
		t.Fatal("GlobalConfig is nil")
	}
	if GlobalConfig.App.Name == "" {
		t.Fatalf("app name is empty: %#v", GlobalConfig.App)
	}
}
