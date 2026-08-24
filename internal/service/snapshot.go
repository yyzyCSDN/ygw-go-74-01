package service

import (
	"encoding/json"
	"os"
	"path/filepath"

	"cleanroomorcontrol/internal/differential"
)

type FileSnapshotStore struct {
	dir string
}

func NewFileSnapshotStore(dir string) *FileSnapshotStore {
	return &FileSnapshotStore{dir: dir}
}

func (s *FileSnapshotStore) path() string {
	return filepath.Join(s.dir, "pressure-snapshot.json")
}

func (s *FileSnapshotStore) LoadLatestPressure() (differential.PressureSnapshot, error) {
	data, err := os.ReadFile(s.path())
	if err != nil {
		if os.IsNotExist(err) {
			return differential.PressureSnapshot{}, differential.ErrSnapshotMissing
		}
		return differential.PressureSnapshot{}, err
	}
	var snapshot differential.PressureSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return differential.PressureSnapshot{}, err
	}
	return snapshot, nil
}

func (s *FileSnapshotStore) SavePressure(snapshot differential.PressureSnapshot) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0o644)
}
