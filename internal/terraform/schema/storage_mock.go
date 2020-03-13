package schema

import (
	"log"

	"github.com/fsnotify/fsnotify"
	tfjson "github.com/hashicorp/terraform-json"
)

func MockStorage(ps *tfjson.ProviderSchemas) *Storage {
	s := NewStorage()
	if ps == nil {
		ps = &tfjson.ProviderSchemas{}
	}
	s.ps = ps
	s.sync = true
	s.w = &MockWatcher{}
	return s
}

type MockWatcher struct{}

func (w *MockWatcher) AddWorkspace(string) error {
	return nil
}

func (w *MockWatcher) Close() error {
	return nil
}

func (w *MockWatcher) Events() chan fsnotify.Event {
	return nil
}

func (w *MockWatcher) Errors() chan error {
	return nil
}

func (w *MockWatcher) OnPluginChange(func(*watchedWorkspace) error) {}

func (w *MockWatcher) SetLogger(*log.Logger) {}
