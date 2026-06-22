package roadmap

import "testing"

func TestDevelopProgressClampsAndShips(t *testing.T) {
	f := &Feature{Status: StatusDeveloping, Progress: 80}
	p, shipped := DevelopProgress(f, 15)
	if p != 95 || shipped {
		t.Errorf("step1: progress=%d shipped=%v", p, shipped)
	}
	p, shipped = DevelopProgress(f, 10)
	if p != 100 || !shipped {
		t.Errorf("step2: progress=%d shipped=%v", p, shipped)
	}
	if f.Status != StatusShipped {
		t.Errorf("status = %q, want shipped", f.Status)
	}
	// Further progress stays at 100 and doesn't re-ship.
	_, shipped = DevelopProgress(f, 50)
	if shipped {
		t.Error("should not ship again at 100")
	}
}

func TestDevelopProgressNilSafe(t *testing.T) {
	if p, shipped := DevelopProgress(nil, 10); p != 0 || shipped {
		t.Errorf("nil feature: progress=%d shipped=%v", p, shipped)
	}
}

func TestSortByPriorityDesc(t *testing.T) {
	features := []Feature{
		{Name: "a", Priority: 1},
		{Name: "b", Priority: 5},
		{Name: "c", Priority: 3},
	}
	order := SortByPriorityDesc(features)
	want := []string{"b", "c", "a"}
	for i, w := range want {
		if features[order[i]].Name != w {
			t.Errorf("pos %d = %q, want %q", i, features[order[i]].Name, w)
		}
	}
}

func TestTotalShippedValue(t *testing.T) {
	features := []Feature{
		{Status: StatusShipped, ValuePoints: 10},
		{Status: StatusDeveloping, ValuePoints: 20},
		{Status: StatusShipped, ValuePoints: 5},
	}
	if got := TotalShippedValue(features); got != 15 {
		t.Errorf("shipped value = %d, want 15", got)
	}
}

func TestIsValidStatus(t *testing.T) {
	for _, s := range []Status{StatusBacklog, StatusDeveloping, StatusShipped} {
		if !IsValidStatus(s) {
			t.Errorf("%q should be valid", s)
		}
	}
	if IsValidStatus("done") {
		t.Error("done should be invalid")
	}
}
