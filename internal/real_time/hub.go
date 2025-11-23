package real_time

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeliveryHub struct {
	log         *zap.Logger
	mtx         sync.RWMutex
	connections map[uuid.UUID]chan any
}

func NewDeliveryHub(log *zap.Logger) *DeliveryHub {
	return &DeliveryHub{
		log:         log,
		connections: make(map[uuid.UUID]chan any),
	}
}

func (h *DeliveryHub) Register(devID uuid.UUID, ch chan any) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.connections[devID] = ch
	h.log.Info("device registered in realtime hub", zap.String("device_id", devID.String()))
}

func (h *DeliveryHub) Unregister(devID uuid.UUID, ch chan any) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	curr, ok := h.connections[devID]
	if ok && curr == ch {
		delete(h.connections, devID)
		h.log.Info("device unregistered from realtime hub", zap.String("device_id", devID.String()))

	}
}

func (h *DeliveryHub) Push(_ context.Context, deviceID uuid.UUID, payload any) error {
	h.mtx.RLock()
	ch, ok := h.connections[deviceID]
	h.mtx.RUnlock()

	if !ok {
		return fmt.Errorf("device is offline")
	}

	select {
	case ch <- payload:
		return nil
	default:
		return fmt.Errorf("realtime channel is full")
	}
}
