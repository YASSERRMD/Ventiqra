package realtime

import (
	"testing"
	"time"
)

func TestHubPublishIsSafeWithNoSubscribers(t *testing.T) {
	h := NewHub()
	defer h.Shutdown()
	// Must not block or panic.
	for i := 0; i < 100; i++ {
		h.Publish(Message{Type: "tick", CompanyID: "co-1", Payload: i})
	}
	if got := h.SubscriberCount("co-1"); got != 0 {
		t.Errorf("SubscriberCount = %d, want 0", got)
	}
}

func TestHubRegisterThenUnsubscribe(t *testing.T) {
	h := NewHub()
	defer h.Shutdown()

	c := &client{companyID: "co-1", send: make(chan []byte, 1)}
	h.joinLeave <- joinLeave{c: c, leave: false}
	waitFor(t, func() bool { return h.SubscriberCount("co-1") == 1 })

	h.joinLeave <- joinLeave{c: c, leave: true}
	waitFor(t, func() bool { return h.SubscriberCount("co-1") == 0 })
}

func TestHubDeliversToRegisteredClient(t *testing.T) {
	h := NewHub()
	defer h.Shutdown()

	c := &client{companyID: "co-1", send: make(chan []byte, 4)}
	h.joinLeave <- joinLeave{c: c, leave: false}
	waitFor(t, func() bool { return h.SubscriberCount("co-1") == 1 })

	h.Publish(Message{Type: "tick", CompanyID: "co-1", Payload: map[string]int{"day": 7}})

	select {
	case data := <-c.send:
		if !contains(string(data), `"type":"tick"`) {
			t.Errorf("unexpected payload: %s", data)
		}
		if !contains(string(data), `"day":7`) {
			t.Errorf("payload missing day: %s", data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for delivery")
	}
}

func TestHubIsolatesCompanies(t *testing.T) {
	h := NewHub()
	defer h.Shutdown()

	a := &client{companyID: "co-a", send: make(chan []byte, 4)}
	b := &client{companyID: "co-b", send: make(chan []byte, 4)}
	h.joinLeave <- joinLeave{c: a, leave: false}
	h.joinLeave <- joinLeave{c: b, leave: false}
	waitFor(t, func() bool { return h.SubscriberCount("co-a") == 1 && h.SubscriberCount("co-b") == 1 })

	h.Publish(Message{Type: "tick", CompanyID: "co-a"})

	select {
	case <-a.send:
	case <-time.After(time.Second):
		t.Fatal("co-a did not receive")
	}
	select {
	case data := <-b.send:
		t.Fatalf("co-b should not receive co-a message, got %s", data)
	case <-time.After(100 * time.Millisecond):
		// expected: no delivery
	}
}

func TestHubDropsSlowClient(t *testing.T) {
	h := NewHub()
	defer h.Shutdown()

	// A client with a tiny send buffer (1) that nobody drains.
	c := &client{companyID: "co-1", send: make(chan []byte, 1)}
	h.joinLeave <- joinLeave{c: c, leave: false}
	waitFor(t, func() bool { return h.SubscriberCount("co-1") == 1 })

	for i := 0; i < 100; i++ {
		h.Publish(Message{Type: "tick", CompanyID: "co-1"})
	}
	// The client should be evicted.
	waitFor(t, func() bool { return h.SubscriberCount("co-1") == 0 })
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition never became true")
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
