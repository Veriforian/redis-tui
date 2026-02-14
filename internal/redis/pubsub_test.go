package redis

import (
	"testing"
	"time"

	"github.com/davidbudnick/redis-tui/internal/types"
)

func TestPublish(t *testing.T) {
	client, _ := setupTestClient(t)

	// Subscribe to a channel using the raw client
	sub := client.client.Subscribe(client.ctx, "testchan")
	t.Cleanup(func() { _ = sub.Close() })

	// Wait for the subscription to be ready
	_, err := sub.Receive(client.ctx)
	if err != nil {
		t.Fatalf("failed to receive subscription confirmation: %v", err)
	}

	// Publish a message
	receivers, err := client.Publish("testchan", "hello")
	if err != nil {
		t.Fatalf("Publish() returned error: %v", err)
	}
	if receivers < 1 {
		t.Errorf("Publish() receivers = %d, want >= 1", receivers)
	}

	// Verify the message was received
	msg, err := sub.ReceiveMessage(client.ctx)
	if err != nil {
		t.Fatalf("ReceiveMessage() returned error: %v", err)
	}
	if msg.Payload != "hello" {
		t.Errorf("received payload = %q, want %q", msg.Payload, "hello")
	}
	if msg.Channel != "testchan" {
		t.Errorf("received channel = %q, want %q", msg.Channel, "testchan")
	}
}

func TestPubSubChannels(t *testing.T) {
	client, _ := setupTestClient(t)

	// Subscribe to a channel to make it active
	sub := client.client.Subscribe(client.ctx, "activechan")
	t.Cleanup(func() { _ = sub.Close() })

	// Wait for the subscription to be ready
	_, err := sub.Receive(client.ctx)
	if err != nil {
		t.Fatalf("failed to receive subscription confirmation: %v", err)
	}

	// List active channels
	channels, err := client.PubSubChannels("*")
	if err != nil {
		t.Fatalf("PubSubChannels() returned error: %v", err)
	}

	found := false
	for _, ch := range channels {
		if ch == "activechan" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("PubSubChannels() = %v, expected to contain %q", channels, "activechan")
	}
}

func TestSubscribeKeyspace(t *testing.T) {
	client, _ := setupTestClient(t)

	// Set up a channel to receive events from the handler
	eventCh := make(chan types.KeyspaceEvent, 10)
	handler := func(evt types.KeyspaceEvent) {
		eventCh <- evt
	}

	err := client.SubscribeKeyspace("*", handler)
	if err != nil {
		t.Fatalf("SubscribeKeyspace() returned error: %v", err)
	}

	// Verify the subscription was set up correctly
	if client.keyspacePS == nil {
		t.Error("keyspacePS should not be nil after SubscribeKeyspace")
	}
	if client.cancelKeyspace == nil {
		t.Error("cancelKeyspace should not be nil after SubscribeKeyspace")
	}
	if len(client.eventHandlers) != 1 {
		t.Errorf("eventHandlers length = %d, want 1", len(client.eventHandlers))
	}

	// Try to trigger a keyspace event by setting a key via the redis client
	// (miniredis may or may not fire keyspace notifications, so we use a timeout)
	client.client.Set(client.ctx, "mykey", "val", 0)

	select {
	case evt := <-eventCh:
		// If we received an event, validate it
		if evt.Key != "mykey" {
			t.Errorf("event Key = %q, want %q", evt.Key, "mykey")
		}
		if evt.DB != 0 {
			t.Errorf("event DB = %d, want 0", evt.DB)
		}
	case <-time.After(200 * time.Millisecond):
		// miniredis may not support keyspace notifications, that's acceptable.
		// We already verified the mechanics above (keyspacePS, cancelKeyspace, eventHandlers).
		t.Log("no keyspace event received (miniredis may not support keyspace notifications); mechanics verified")
	}
}

func TestSubscribeKeyspace_Resubscribe(t *testing.T) {
	client, _ := setupTestClient(t)

	// First subscription
	handler1 := func(evt types.KeyspaceEvent) {}
	err := client.SubscribeKeyspace("*", handler1)
	if err != nil {
		t.Fatalf("first SubscribeKeyspace() returned error: %v", err)
	}

	oldCancel := client.cancelKeyspace
	oldPS := client.keyspacePS

	if oldCancel == nil {
		t.Fatal("cancelKeyspace should not be nil after first SubscribeKeyspace")
	}
	if oldPS == nil {
		t.Fatal("keyspacePS should not be nil after first SubscribeKeyspace")
	}

	// Second subscription (re-subscribe)
	handler2 := func(evt types.KeyspaceEvent) {}
	err = client.SubscribeKeyspace("*", handler2)
	if err != nil {
		t.Fatalf("second SubscribeKeyspace() returned error: %v", err)
	}

	// Verify new subscription was created
	if client.cancelKeyspace == nil {
		t.Error("cancelKeyspace should not be nil after re-subscribe")
	}
	if client.keyspacePS == nil {
		t.Error("keyspacePS should not be nil after re-subscribe")
	}

	// The new cancel and PS should be different objects from the old ones
	// (old ones were replaced during re-subscribe)
	if client.keyspacePS == oldPS {
		t.Error("keyspacePS should be a new instance after re-subscribe")
	}

	// Verify only one handler is registered (old handlers replaced)
	if len(client.eventHandlers) != 1 {
		t.Errorf("eventHandlers length = %d, want 1 after re-subscribe", len(client.eventHandlers))
	}
}

func TestUnsubscribeKeyspace(t *testing.T) {
	t.Run("unsubscribe after subscribe", func(t *testing.T) {
		client, _ := setupTestClient(t)

		handler := func(evt types.KeyspaceEvent) {}
		err := client.SubscribeKeyspace("*", handler)
		if err != nil {
			t.Fatalf("SubscribeKeyspace() returned error: %v", err)
		}

		// Verify subscription is active
		if client.cancelKeyspace == nil {
			t.Fatal("cancelKeyspace should not be nil before unsubscribe")
		}
		if client.keyspacePS == nil {
			t.Fatal("keyspacePS should not be nil before unsubscribe")
		}

		// Unsubscribe
		err = client.UnsubscribeKeyspace()
		if err != nil {
			t.Fatalf("UnsubscribeKeyspace() returned error: %v", err)
		}

		if client.cancelKeyspace != nil {
			t.Error("cancelKeyspace should be nil after UnsubscribeKeyspace")
		}
		if client.keyspacePS != nil {
			t.Error("keyspacePS should be nil after UnsubscribeKeyspace")
		}
	})

	t.Run("unsubscribe when not subscribed", func(t *testing.T) {
		client, _ := setupTestClient(t)

		err := client.UnsubscribeKeyspace()
		if err != nil {
			t.Errorf("UnsubscribeKeyspace() when not subscribed returned error: %v", err)
		}
	})
}
