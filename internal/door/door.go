package door

import (
	"context"
	"strconv"
	"sync"

	"cleanroomorcontrol/internal/model"
)

type SettleSource interface {
	SettleChannel(room model.RoomID) <-chan struct{}
}

type Door struct {
	ID     string
	Room   model.RoomID
	State  model.DoorState
	WaitID string
}

type Controller struct {
	mu      sync.RWMutex
	doors   map[model.RoomID]*Door
	settle  SettleSource
	waitSeq int
}

func NewController(settle SettleSource) *Controller {
	return &Controller{
		doors:  make(map[model.RoomID]*Door),
		settle: settle,
	}
}

func (c *Controller) RegisterDoor(id string, room model.RoomID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.doors[room] = &Door{ID: id, Room: room, State: model.DoorUnlocked}
}

func (c *Controller) Acquire(ctx context.Context, room model.RoomID) error {
	c.mu.Lock()
	door, ok := c.doors[room]
	if !ok {
		c.mu.Unlock()
		return ErrDoorMissing
	}
	if door.State != model.DoorUnlocked {
		c.mu.Unlock()
		return ErrDoorBusy
	}
	door.State = model.DoorInterlocked
	door.WaitID = c.nextWaitID()
	c.mu.Unlock()
	return waitForSettle(ctx, c.settleChannel(room), func() {})
}

func (c *Controller) nextWaitID() string {
	c.waitSeq++
	return "W-" + strconv.Itoa(c.waitSeq)
}

func (c *Controller) settleChannel(room model.RoomID) <-chan struct{} {
	if c.settle == nil {
		return make(chan struct{})
	}
	return c.settle.SettleChannel(room)
}

func (c *Controller) Engage(room model.RoomID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	door, ok := c.doors[room]
	if !ok {
		return ErrDoorMissing
	}
	door.State = model.DoorInterlocked
	return nil
}

func (c *Controller) Release(room model.RoomID) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	door, ok := c.doors[room]
	if !ok {
		return ErrDoorMissing
	}
	door.State = model.DoorUnlocked
	door.WaitID = ""
	return nil
}

func (c *Controller) State(room model.RoomID) model.DoorState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	door, ok := c.doors[room]
	if !ok {
		return model.DoorLocked
	}
	return door.State
}

func (c *Controller) CanEnter(room model.RoomID) bool {
	return model.CanEnterDoor(c.State(room))
}

func (c *Controller) Snapshot() map[model.RoomID]model.DoorState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[model.RoomID]model.DoorState, len(c.doors))
	for room, door := range c.doors {
		out[room] = door.State
	}
	return out
}
