package utils

import (
	"testing"
)

func TestStringSliceIntersection(t *testing.T) {
	slice1 := []string{"one", "two", "three"}
	slice2 := []string{"two", "four", "three"}
	intersection := StringSliceIntersection(slice1, slice2)
	if intersection[0] != "two" && intersection[1] != "three" && len(intersection) != 2 {
		t.Errorf("Expected two,three; got %v", intersection)
	}
}

func TestStringSliceDifference(t *testing.T) {
	slice1 := []string{"one", "two", "three"}
	slice2 := []string{"two", "four", "three"}
	difference := StringSliceDifference(slice1, slice2)
	if difference[0] != "one" && len(difference) != 1 {
		t.Errorf("Expected one; got %v", difference)
	}
}

func TestStringsToInts(t *testing.T) {
	ints, err := StringsToInts([]string{"1", "2"})
	if ints[0] != 1 && ints[1] != 2 && len(ints) != 2 {
		t.Errorf("Expected 1,2; got %v", ints)
	}
	ints, err = StringsToInts([]string{"one", "two"})
	if ints != nil {
		t.Errorf("Expected error, got %v, err: %v", ints, err.Error())
	}
}
