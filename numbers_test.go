package godrbdutils

import "testing"

func TestEmpty(t *testing.T) {
	got, err := GetNumber(10, 20, []int{})
	if err != nil {
		t.Fatalf("Did not expect err: %v", err)
	}
	if got != 10 {
		t.Fatalf("Expected: %d, but got: %d", 10, got)
	}
}

func TestLowest(t *testing.T) {
	got, err := GetNumber(10, 20, []int{11})
	if err != nil {
		t.Fatalf("Did not expect err: %v", err)
	}
	exp := 12
	if got != exp {
		t.Fatalf("Expected: %d, but got: %d", exp, got)
	}
}

func TestLowestSomeUsed(t *testing.T) {
	got, err := GetNumber(10, 20, []int{19})
	// 19 is the lowest, so 18 should be used
	if err != nil {
		t.Fatalf("Did not expect err: %v", err)
	}
	exp := 20
	if got != exp {
		t.Fatalf("Expected: %d, but got: %d", exp, got)
	}
}

func TestHighestSomeUsed(t *testing.T) {
	got, err := GetNumber(10, 20, []int{10, 11, 12, 19})
	if err != nil {
		t.Fatalf("Did not expect err: %v", err)
	}
	if got != 20 {
		t.Fatalf("Expected: %d, but got: %d", 20, got)
	}
}

func TestMiddle(t *testing.T) {
	got, err := GetNumber(10, 20, []int{10, 11, 12, 20})
	if err != nil {
		t.Fatalf("Did not expect err: %v", err)
	}
	if got != 13 {
		t.Fatalf("Expected: %d, but got: %d", 13, got)
	}
}

func TestNoneInRange(t *testing.T) {
	got, err := GetNumber(10, 15, []int{10, 11, 12, 13, 14, 15})
	if err == nil {
		t.Fatalf("Expected an err, but got: %d", got)
	}
}

func TestMinMaxWeird(t *testing.T) {
	got, err := GetNumber(15, 10, []int{})
	if err == nil {
		t.Fatalf("Expected an err, but got: %d", got)
	}
}
