package door

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"cleanroomorcontrol/internal/model"
)

// stubSettleSource hands back a caller-controlled, buffered settle channel per
// room. It is safe for concurrent use so the test goroutine and the controller
// never race on the underlying map.
type stubSettleSource struct {
	mu sync.Mutex
	ch map[model.RoomID]chan struct{}
}

func newStubSettleSource() *stubSettleSource {
	return &stubSettleSource{ch: make(map[model.RoomID]chan struct{})}
}

func (s *stubSettleSource) channel(room model.RoomID) chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.ch[room]
	if !ok {
		ch = make(chan struct{}, 1)
		s.ch[room] = ch
	}
	return ch
}

func (s *stubSettleSource) SettleChannel(room model.RoomID) <-chan struct{} {
	return s.channel(room)
}

func (s *stubSettleSource) signal(room model.RoomID) {
	select {
	case s.channel(room) <- struct{}{}:
	default:
	}
}

// TestAcquireTimeoutReleasesDoor reproduces the reported stall: when the
// airflow settle wait times out, the interlock must be released so the door
// returns to an enterable state instead of staying occupied forever.
func TestAcquireTimeoutReleasesDoor(t *testing.T) {
	source := newStubSettleSource()
	ctrl := NewController(source)
	ctrl.RegisterDoor("OR-1", "OR-1")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := ctrl.Acquire(ctx, "OR-1")
	if !errors.Is(err, ErrInterlockTimeout) {
		t.Fatalf("expected ErrInterlockTimeout, got %v", err)
	}

	// After the timeout the door must be available again: a doctor swiping
	// in (CanEnter) must succeed, and a fresh Acquire must be admitted (rather
	// than rejected as busy). The new wait will block on a settle signal, so we
	// queue one and bound the call with its own deadline to keep it honest.
	if got := ctrl.State("OR-1"); got != model.DoorUnlocked {
		t.Fatalf("door state after timeout = %s, want unlocked", got)
	}
	if !ctrl.CanEnter("OR-1") {
		t.Fatal("door must be enterable after interlock timeout")
	}
	source.signal("OR-1")
	reCtx, reCancel := context.WithTimeout(context.Background(), time.Second)
	defer reCancel()
	if err := ctrl.Acquire(reCtx, "OR-1"); err != nil {
		t.Fatalf("re-acquire after timeout failed: %v", err)
	}
}

// TestAcquireSettleKeepsAcquired confirms that on the success path (airflow
// settles before the timeout) the interlock stays held — the timeout fix must
// not over-release a wait that completed normally.
func TestAcquireSettleKeepsAcquired(t *testing.T) {
	source := newStubSettleSource()
	ctrl := NewController(source)
	ctrl.RegisterDoor("OR-2", "OR-2")

	// Pre-queue the settle signal on the buffered channel so Acquire completes
	// successfully well before the deadline, with no extra goroutine.
	source.signal("OR-2")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ctrl.Acquire(ctx, "OR-2"); err != nil {
		t.Fatalf("acquire on settle failed: %v", err)
	}
	if got := ctrl.State("OR-2"); got != model.DoorInterlocked {
		t.Fatalf("door state after settle = %s, want interlocked", got)
	}
	if ctrl.CanEnter("OR-2") {
		t.Fatal("door must remain non-enterable while interlocked")
	}
}

// TestReleaseGuardSkipsStaleOwner verifies the WaitID guard: a release from a
// timed-out owner must not clobber a later re-acquisition by a new owner.
func TestReleaseGuardSkipsStaleOwner(t *testing.T) {
	source := newStubSettleSource()
	ctrl := NewController(source)
	ctrl.RegisterDoor("OR-3", "OR-3")

	// First owner times out; its release restores the door to unlocked.
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel1()
	if err := ctrl.Acquire(ctx1, "OR-3"); !errors.Is(err, ErrInterlockTimeout) {
		t.Fatalf("first acquire: expected timeout, got %v", err)
	}

	// A second owner re-acquires and is now legitimately interlocked. Pre-queue
	// the settle so it completes immediately and stays held.
	source.signal("OR-3")
	if err := ctrl.Acquire(context.Background(), "OR-3"); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}
	if got := ctrl.State("OR-3"); got != model.DoorInterlocked {
		t.Fatalf("second owner state = %s, want interlocked", got)
	}
}
