package iapetus

import (
	"testing"
)

type mockBackend struct {
	name   string
	status string
	called *bool
}

func (m *mockBackend) RunTask(task *Task) error {
	if m.called != nil {
		*m.called = true
	}
	return nil
}
func (m *mockBackend) ValidateTask(task *Task) error { return nil }
func (m *mockBackend) GetName() string               { return m.name }
func (m *mockBackend) GetStatus() string             { return m.status }

func TestRegisterAndGetBackend(t *testing.T) {
	mb := &mockBackend{name: "mock", status: "available"}
	RegisterBackend("mock", mb)
	b := GetBackend("mock")
	if b == nil {
		t.Fatal("expected backend to be registered and retrievable")
	}
	if b.GetName() != "mock" {
		t.Errorf("expected GetName to return 'mock', got %q", b.GetName())
	}
	if b.GetStatus() != "available" {
		t.Errorf("expected GetStatus to return 'available', got %q", b.GetStatus())
	}
}

func TestGetBackend_NotFound(t *testing.T) {
	b := GetBackend("doesnotexist")
	if b != nil {
		t.Errorf("expected nil for unregistered backend, got %v", b)
	}
}
