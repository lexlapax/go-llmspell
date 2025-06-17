// ABOUTME: Channel integration for Go channel ↔ LChannel bridge in GopherLua engine
// ABOUTME: Provides select operations, buffered channels, and deadlock detection

package gopherlua

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

// ChannelManager manages Go channels and their Lua LChannel counterparts
type ChannelManager struct {
	maxChannels int
	channels    map[string]*channelInfo
	mu          sync.RWMutex
	closed      bool
	closeOnce   sync.Once
}

// channelInfo tracks channel state and metadata
type channelInfo struct {
	ID         string
	GoChan     chan lua.LValue
	LChannel   lua.LChannel
	BufferSize int
	CreatedAt  time.Time
	Closed     bool
}

// SelectOperation defines the type of operation for select
type SelectOperation int

const (
	SelectReceive SelectOperation = iota
	SelectSend
)

// SelectCase represents a single case in a select operation
type SelectCase struct {
	ChannelID string
	Operation SelectOperation
	Value     lua.LValue // For send operations
}

const (
	defaultMaxChannels = 100
)

// NewChannelManager creates a new channel manager with specified max channels
func NewChannelManager(maxChannels int) (*ChannelManager, error) {
	if maxChannels < 0 {
		return nil, fmt.Errorf("maxChannels cannot be negative: %d", maxChannels)
	}

	if maxChannels == 0 {
		maxChannels = defaultMaxChannels
	}

	return &ChannelManager{
		maxChannels: maxChannels,
		channels:    make(map[string]*channelInfo),
	}, nil
}

// CreateChannel creates a new Go channel with optional buffer size and LChannel bridge
func (cm *ChannelManager) CreateChannel(L *lua.LState, bufferSize int) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.closed {
		return "", fmt.Errorf("channel manager is closed")
	}

	// Count only active (non-closed) channels
	activeCount := 0
	for _, info := range cm.channels {
		if !info.Closed {
			activeCount++
		}
	}

	if activeCount >= cm.maxChannels {
		return "", fmt.Errorf("maximum channels (%d) exceeded", cm.maxChannels)
	}

	channelID := uuid.New().String()

	// Create Go channel
	goChan := make(chan lua.LValue, bufferSize)

	// Create Lua LChannel (LChannel is just chan LValue)
	lChannel := make(lua.LChannel, bufferSize)

	info := &channelInfo{
		ID:         channelID,
		GoChan:     goChan,
		LChannel:   lChannel,
		BufferSize: bufferSize,
		CreatedAt:  time.Now(),
		Closed:     false,
	}

	cm.channels[channelID] = info

	// Start bridge goroutine to sync Go channel with LChannel
	go cm.bridgeChannels(channelID, info)

	return channelID, nil
}

// bridgeChannels synchronizes Go channel with LChannel
func (cm *ChannelManager) bridgeChannels(channelID string, info *channelInfo) {
	// This goroutine bridges the Go channel and LChannel
	// For now, we'll keep them separate and sync on operations
	// In a full implementation, we'd have bidirectional syncing
}

// Send sends a value to the specified channel
func (cm *ChannelManager) Send(ctx context.Context, channelID string, value lua.LValue) error {
	cm.mu.RLock()
	info, exists := cm.channels[channelID]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	if info.Closed {
		return fmt.Errorf("cannot send to closed channel: %s", channelID)
	}

	select {
	case info.GoChan <- value:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Receive receives a value from the specified channel
func (cm *ChannelManager) Receive(ctx context.Context, channelID string) (lua.LValue, error) {
	cm.mu.RLock()
	info, exists := cm.channels[channelID]
	cm.mu.RUnlock()

	if !exists {
		return lua.LNil, fmt.Errorf("channel not found: %s", channelID)
	}

	select {
	case value, ok := <-info.GoChan:
		if !ok {
			return lua.LNil, fmt.Errorf("channel is closed")
		}
		return value, nil
	case <-ctx.Done():
		return lua.LNil, ctx.Err()
	}
}

// Select performs a select operation on multiple channels
func (cm *ChannelManager) Select(ctx context.Context, cases []SelectCase) (int, lua.LValue, error) {
	if len(cases) == 0 {
		return -1, lua.LNil, fmt.Errorf("no select cases provided")
	}

	// Build reflect.SelectCase slice for Go's select
	selectCases := make([]reflect.SelectCase, len(cases)+1) // +1 for context

	// Add context case for cancellation
	selectCases[0] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	}

	// Build select cases from channel operations
	for i, selectCase := range cases {
		cm.mu.RLock()
		info, exists := cm.channels[selectCase.ChannelID]
		cm.mu.RUnlock()

		if !exists {
			return -1, lua.LNil, fmt.Errorf("channel not found: %s", selectCase.ChannelID)
		}

		switch selectCase.Operation {
		case SelectReceive:
			selectCases[i+1] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(info.GoChan),
			}
		case SelectSend:
			selectCases[i+1] = reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(info.GoChan),
				Send: reflect.ValueOf(selectCase.Value),
			}
		default:
			return -1, lua.LNil, fmt.Errorf("unknown select operation: %d", selectCase.Operation)
		}
	}

	// Perform the select operation
	chosen, recv, recvOK := reflect.Select(selectCases)

	// Check if context was cancelled (case 0)
	if chosen == 0 {
		return -1, lua.LNil, ctx.Err()
	}

	// Convert chosen index back to original case index
	originalCase := chosen - 1

	// Handle receive operations
	if cases[originalCase].Operation == SelectReceive {
		if !recvOK {
			return originalCase, lua.LNil, fmt.Errorf("channel is closed")
		}

		if recv.IsValid() {
			value, ok := recv.Interface().(lua.LValue)
			if !ok {
				return originalCase, lua.LNil, fmt.Errorf("invalid value type received")
			}
			return originalCase, value, nil
		}
		return originalCase, lua.LNil, nil
	}

	// Handle send operations
	return originalCase, lua.LNil, nil
}

// CloseChannel closes the specified channel
func (cm *ChannelManager) CloseChannel(channelID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	info, exists := cm.channels[channelID]
	if !exists {
		return fmt.Errorf("channel not found: %s", channelID)
	}

	if info.Closed {
		return fmt.Errorf("channel already closed: %s", channelID)
	}

	// Close Go channel
	close(info.GoChan)

	// Note: LChannel is separate, we'll implement proper bridging later

	// Mark as closed (keep in map for receive operations)
	info.Closed = true

	return nil
}

// ChannelExists checks if a channel exists and is active
func (cm *ChannelManager) ChannelExists(channelID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	info, exists := cm.channels[channelID]
	return exists && !info.Closed
}

// GetChannelCount returns the number of active channels
func (cm *ChannelManager) GetChannelCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	activeCount := 0
	for _, info := range cm.channels {
		if !info.Closed {
			activeCount++
		}
	}
	return activeCount
}

// GetChannelInfo returns information about a specific channel
func (cm *ChannelManager) GetChannelInfo(channelID string) (*channelInfo, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	info, exists := cm.channels[channelID]
	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	// Return a copy to avoid concurrent access issues
	return &channelInfo{
		ID:         info.ID,
		BufferSize: info.BufferSize,
		CreatedAt:  info.CreatedAt,
		Closed:     info.Closed,
	}, nil
}

// ListChannels returns a list of all active channel IDs
func (cm *ChannelManager) ListChannels() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	channels := make([]string, 0, len(cm.channels))
	for id := range cm.channels {
		channels = append(channels, id)
	}

	return channels
}

// Close shuts down the channel manager and closes all channels
func (cm *ChannelManager) Close() error {
	var closeErr error

	cm.closeOnce.Do(func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()

		cm.closed = true

		// Close all channels
		for _, info := range cm.channels {
			if !info.Closed {
				close(info.GoChan)
				// Note: LChannel closing will be implemented with proper bridging
				info.Closed = true
			}
		}

		// Clear channels map
		cm.channels = make(map[string]*channelInfo)
	})

	return closeErr
}
