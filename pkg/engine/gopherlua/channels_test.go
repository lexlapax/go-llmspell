// ABOUTME: Tests for channel integration functionality in GopherLua engine
// ABOUTME: Tests Go channel ↔ LChannel bridge, select operations, and channel management

package gopherlua

import (
	"context"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func TestChannelManager_NewChannelManager(t *testing.T) {
	tests := []struct {
		name        string
		maxChannels int
		wantErr     bool
	}{
		{
			name:        "valid manager with default max channels",
			maxChannels: 0, // Should use default
			wantErr:     false,
		},
		{
			name:        "valid manager with custom max channels",
			maxChannels: 10,
			wantErr:     false,
		},
		{
			name:        "invalid manager with negative max channels",
			maxChannels: -1,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewChannelManager(tt.maxChannels)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewChannelManager() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewChannelManager() unexpected error: %v", err)
				return
			}

			if manager == nil {
				t.Errorf("NewChannelManager() returned nil manager")
				return
			}

			// Verify manager state
			if manager.maxChannels <= 0 {
				t.Errorf("NewChannelManager() maxChannels should be positive, got %d", manager.maxChannels)
			}
		})
	}
}

func TestChannelManager_CreateChannel(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	tests := []struct {
		name       string
		bufferSize int
		wantErr    bool
	}{
		{
			name:       "unbuffered channel",
			bufferSize: 0,
			wantErr:    false,
		},
		{
			name:       "buffered channel",
			bufferSize: 5,
			wantErr:    false,
		},
		{
			name:       "large buffer channel",
			bufferSize: 100,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channelID, err := manager.CreateChannel(L, tt.bufferSize)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateChannel() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CreateChannel() unexpected error: %v", err)
				return
			}

			if channelID == "" {
				t.Errorf("CreateChannel() returned empty channel ID")
				return
			}

			// Verify channel exists
			if !manager.ChannelExists(channelID) {
				t.Errorf("CreateChannel() channel %s should exist", channelID)
			}
		})
	}
}

func TestChannelManager_SendReceive(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create a buffered channel with enough capacity
	channelID, err := manager.CreateChannel(L, 10)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test sending values
	testValues := []lua.LValue{
		lua.LString("hello"),
		lua.LNumber(42),
		lua.LBool(true),
		lua.LNil,
	}

	// Send values to channel
	for i, value := range testValues {
		err := manager.Send(ctx, channelID, value)
		if err != nil {
			t.Errorf("Send() value %d error: %v", i, err)
		}
	}

	// Receive values from channel
	for i, expectedValue := range testValues {
		receivedValue, err := manager.Receive(ctx, channelID)
		if err != nil {
			t.Errorf("Receive() value %d error: %v", i, err)
			continue
		}

		if receivedValue.Type() != expectedValue.Type() {
			t.Errorf("Receive() value %d type mismatch: expected %s, got %s",
				i, expectedValue.Type(), receivedValue.Type())
		}

		// For non-nil values, check the actual value
		if expectedValue != lua.LNil {
			if receivedValue.String() != expectedValue.String() {
				t.Errorf("Receive() value %d mismatch: expected %s, got %s",
					i, expectedValue.String(), receivedValue.String())
			}
		}
	}
}

func TestChannelManager_SelectOperation(t *testing.T) {
	manager, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create multiple channels
	ch1, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel 1: %v", err)
	}

	ch2, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel 2: %v", err)
	}

	ch3, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel 3: %v", err)
	}

	ctx := context.Background()

	// Send to channel 2
	err = manager.Send(ctx, ch2, lua.LString("from_ch2"))
	if err != nil {
		t.Fatalf("Failed to send to channel 2: %v", err)
	}

	// Create select cases
	selectCases := []SelectCase{
		{ChannelID: ch1, Operation: SelectReceive},
		{ChannelID: ch2, Operation: SelectReceive},
		{ChannelID: ch3, Operation: SelectReceive},
	}

	// Perform select operation
	selectedCase, value, err := manager.Select(ctx, selectCases)
	if err != nil {
		t.Fatalf("Select() error: %v", err)
	}

	// Should select channel 2 (index 1)
	if selectedCase != 1 {
		t.Errorf("Select() expected case 1, got %d", selectedCase)
	}

	if value.String() != "from_ch2" {
		t.Errorf("Select() expected 'from_ch2', got %s", value.String())
	}
}

func TestChannelManager_SelectWithTimeout(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create unbuffered channel (will block)
	channelID, err := manager.CreateChannel(L, 0)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create select cases that will timeout
	selectCases := []SelectCase{
		{ChannelID: channelID, Operation: SelectReceive},
	}

	// Perform select operation with timeout
	_, _, err = manager.Select(ctx, selectCases)
	if err == nil {
		t.Errorf("Select() expected timeout error, got nil")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Select() expected context.DeadlineExceeded, got %v", err)
	}
}

func TestChannelManager_ChannelClosing(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create channel
	channelID, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	ctx := context.Background()

	// Send a value
	err = manager.Send(ctx, channelID, lua.LString("test"))
	if err != nil {
		t.Fatalf("Failed to send to channel: %v", err)
	}

	// Close the channel
	err = manager.CloseChannel(channelID)
	if err != nil {
		t.Errorf("CloseChannel() error: %v", err)
	}

	// Should still be able to receive the sent value
	value, err := manager.Receive(ctx, channelID)
	if err != nil {
		t.Errorf("Receive() from closed channel error: %v", err)
	}

	if value.String() != "test" {
		t.Errorf("Receive() expected 'test', got %s", value.String())
	}

	// Further receives should indicate channel is closed
	_, err = manager.Receive(ctx, channelID)
	if err == nil {
		t.Errorf("Receive() from empty closed channel should return error")
	}
}

func TestChannelManager_DeadlockDetection(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create unbuffered channel
	channelID, err := manager.CreateChannel(L, 0)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to send to unbuffered channel with no receiver (should timeout/deadlock)
	err = manager.Send(ctx, channelID, lua.LString("deadlock_test"))
	if err == nil {
		t.Errorf("Send() to unbuffered channel should timeout/error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Send() expected timeout, got %v", err)
	}
}

func TestChannelManager_ChannelLimits(t *testing.T) {
	manager, err := NewChannelManager(2) // Limited to 2 channels
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create maximum number of channels
	ch1, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel 1: %v", err)
	}

	ch2, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Fatalf("Failed to create channel 2: %v", err)
	}

	// Should fail to create another channel
	_, err = manager.CreateChannel(L, 1)
	if err == nil {
		t.Errorf("CreateChannel() should fail when limit exceeded")
	}

	// After closing one channel, should be able to create another
	err = manager.CloseChannel(ch1)
	if err != nil {
		t.Errorf("CloseChannel() error: %v", err)
	}

	ch3, err := manager.CreateChannel(L, 1)
	if err != nil {
		t.Errorf("CreateChannel() should succeed after closing channel: %v", err)
	}

	if ch3 == "" {
		t.Errorf("CreateChannel() returned empty ID")
	}

	// Verify we still have ch2 and ch3
	if !manager.ChannelExists(ch2) {
		t.Errorf("Channel ch2 should still exist")
	}

	if !manager.ChannelExists(ch3) {
		t.Errorf("Channel ch3 should exist")
	}

	if manager.ChannelExists(ch1) {
		t.Errorf("Channel ch1 should be closed")
	}
}

func TestChannelManager_Close(t *testing.T) {
	manager, err := NewChannelManager(5)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}

	L := lua.NewState()
	defer L.Close()

	// Create some channels
	for i := 0; i < 3; i++ {
		_, err := manager.CreateChannel(L, 1)
		if err != nil {
			t.Errorf("Failed to create channel %d: %v", i, err)
		}
	}

	// Close manager should succeed
	err = manager.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}

	// Should not be able to create new channels after close
	_, err = manager.CreateChannel(L, 1)
	if err == nil {
		t.Errorf("CreateChannel() should fail after Close()")
	}
}

func TestChannelManager_ConcurrentOperations(t *testing.T) {
	manager, err := NewChannelManager(10)
	if err != nil {
		t.Fatalf("Failed to create channel manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	L := lua.NewState()
	defer L.Close()

	// Create a buffered channel
	channelID, err := manager.CreateChannel(L, 10)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	ctx := context.Background()
	done := make(chan bool, 2)

	// Goroutine 1: Send values
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			err := manager.Send(ctx, channelID, lua.LNumber(float64(i)))
			if err != nil {
				t.Errorf("Send() error in goroutine 1: %v", err)
				return
			}
		}
	}()

	// Goroutine 2: Receive values
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			value, err := manager.Receive(ctx, channelID)
			if err != nil {
				t.Errorf("Receive() error in goroutine 2: %v", err)
				return
			}
			if value.Type() != lua.LTNumber {
				t.Errorf("Receive() expected number, got %s", value.Type())
				return
			}
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}
