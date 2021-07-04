package main

import (
	"os"
	"testing"
)

func TestIsModuleAwareUnset(t *testing.T) {
	if !IsModuleAware() {
		t.Error("IsModuleAware unset failed, expected true, got false")
	}
}

func TestIsModuleAwareAuto(t *testing.T) {
	os.Setenv("GO111MODULE", "auto")

	if !IsModuleAware() {
		t.Error("IsModuleAware auto failed, expected true, got false")
	}
}

func TestIsModuleAwareOn(t *testing.T) {
	os.Setenv("GO111MODULE", "on")

	if !IsModuleAware() {
		t.Error("IsModuleAware on failed, expected true, got false")
	}
}

func TestIsModuleAwareOff(t *testing.T) {
	os.Setenv("GO111MODULE", "off")
	if IsModuleAware() {
		t.Errorf("IsModuleAware off failed, expected false, got true")
	}
}
