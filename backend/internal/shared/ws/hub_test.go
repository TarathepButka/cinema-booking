package ws

import (
	"testing"
	"time"
)

func TestHubBroadcastsOnlyToMatchingShowtime(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	target := &Client{ShowtimeID: "showtime-1", Send: make(chan []byte, 1), Hub: hub}
	other := &Client{ShowtimeID: "showtime-2", Send: make(chan []byte, 1), Hub: hub}
	hub.register <- target
	hub.register <- other

	deadline := time.Now().Add(time.Second)
	for hub.ClientCount("showtime-1") != 1 || hub.ClientCount("showtime-2") != 1 {
		if time.Now().After(deadline) {
			t.Fatal("clients were not registered before timeout")
		}
		time.Sleep(time.Millisecond)
	}

	want := []byte(`{"type":"SEAT_UPDATE","seatLabel":"A1","status":"LOCKED"}`)
	hub.Broadcast("showtime-1", want)

	select {
	case got := <-target.Send:
		if string(got) != string(want) {
			t.Fatalf("broadcast payload = %s, want %s", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("matching showtime client did not receive broadcast")
	}

	select {
	case got := <-other.Send:
		t.Fatalf("other showtime unexpectedly received broadcast: %s", got)
	case <-time.After(20 * time.Millisecond):
	}
}
