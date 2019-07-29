package store

import (
	"context"
	"sync/atomic"
	"unsafe"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

// EventStoreProxy is a mechanism for providing an EventStore with a replaceable
// underlying implementation. It uses an atomic so that calls are not impeded by
// mutex overhead.
type EventStoreProxy struct {
	impl unsafe.Pointer

	// guard against the store being garbage collected,
	// which would cause a crash if there were no more references
	// to it and the impl was dereferenced.
	gcGuard EventStore
}

func NewEventStoreProxy(s EventStore) *EventStoreProxy {
	return &EventStoreProxy{
		impl:    unsafe.Pointer(&s),
		gcGuard: s,
	}
}

func (e *EventStoreProxy) Impl() EventStore {
	return *((*EventStore)(atomic.LoadPointer(&e.impl)))
}

func (e *EventStoreProxy) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	return e.Impl().DeleteEventByEntityCheck(ctx, entity, check)
}

func (e *EventStoreProxy) GetEvents(ctx context.Context, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return e.Impl().GetEvents(ctx, pred)
}

func (e *EventStoreProxy) GetEventsByEntity(ctx context.Context, entity string, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return e.Impl().GetEventsByEntity(ctx, entity, pred)
}

func (e *EventStoreProxy) GetEventByEntityCheck(ctx context.Context, entity, check string) (*types.Event, error) {
	return e.Impl().GetEventByEntityCheck(ctx, entity, check)
}

func (e *EventStoreProxy) UpdateEvent(ctx context.Context, event *types.Event) (old, new *types.Event, err error) {
	return e.Impl().UpdateEvent(ctx, event)
}

type closer interface {
	Close() error
}

func (e *EventStoreProxy) UpdateEventStore(to EventStore) {
	oldptr := atomic.SwapPointer(&e.impl, unsafe.Pointer(&to))
	old := *((*EventStore)(oldptr))
	if s, ok := old.(closer); ok {
		// Attempt to close the old store
		defer s.Close()
	}
	e.gcGuard = to
}