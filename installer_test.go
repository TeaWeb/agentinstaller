package agentinstaller

import "testing"

func TestInstall(t *testing.T) {
	installer := NewInstaller()
	installer.Master = "http://127.0.0.1:7777"
	installer.Id = "WaqRBOHMDOsN4ead"
	installer.Key = "VmoUQc8WcRSMh6XPFU45MUOg1ps8qHB8"
	installer.Dir = "/opt"
	isInstalled, err := installer.Start()
	if isInstalled {
		t.Log("isInstalled:", isInstalled)
	}
	if err != nil {
		t.Fatal(err)
	}
}
