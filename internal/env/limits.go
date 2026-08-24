package env

import (
	"sync"

	"cleanroomorcontrol/internal/model"
)

type LimitsStore struct {
	mu     sync.RWMutex
	limits map[model.RoomID]model.EnvLimits
}

func NewLimitsStore() *LimitsStore {
	return &LimitsStore{limits: make(map[model.RoomID]model.EnvLimits)}
}

func (s *LimitsStore) Set(room model.RoomID, limits model.EnvLimits) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits[room] = limits
}

func (s *LimitsStore) Get(room model.RoomID) (model.EnvLimits, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limits, ok := s.limits[room]
	return limits, ok
}
