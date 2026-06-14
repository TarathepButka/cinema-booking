package showtime

import "testing"

func TestGenerateSeats(t *testing.T) {
	seats := GenerateSeats([]Zone{
		{Name: "FRONT", Rows: []string{"A"}, SeatsPerRow: 2, Price: 180},
		{Name: "BACK", Rows: []string{"B", "C"}, SeatsPerRow: 1, Price: 240},
	})

	if len(seats) != 4 {
		t.Fatalf("GenerateSeats() returned %d seats, want 4", len(seats))
	}

	wantLabels := []string{"A1", "A2", "B1", "C1"}
	for i, want := range wantLabels {
		seat := seats[i]
		if seat.SeatLabel != want {
			t.Errorf("seat %d label = %q, want %q", i, seat.SeatLabel, want)
		}
		if seat.Status != SeatAvailable {
			t.Errorf("seat %s status = %q, want %q", seat.SeatLabel, seat.Status, SeatAvailable)
		}
	}

	if seats[0].Zone != "FRONT" || seats[0].Price != 180 {
		t.Fatalf("front seat metadata was not preserved: %#v", seats[0])
	}
	if seats[2].Zone != "BACK" || seats[2].Price != 240 {
		t.Fatalf("back seat metadata was not preserved: %#v", seats[2])
	}
}
