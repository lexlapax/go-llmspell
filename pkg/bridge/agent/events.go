// ABOUTME: Event system bridge v2.0.0 provides access to go-llms v0.3.5 event functionality
// ABOUTME: Wraps event bus, storage, filtering, serialization, aggregation, and replay capabilities

package agent

// EventBridge is now an alias for EventBridgeV2
// This maintains backward compatibility while using the new v2.0.0 implementation
type EventBridge = EventBridgeV2

// NewEventBridge creates a new event bridge using v2.0.0 implementation
func NewEventBridge() *EventBridge {
	return NewEventBridgeV2()
}
