package utils

import (
	"net/http"
	"strconv"
)

// Returns elements in 1 that are also in 2
func StringSliceIntersection(slice1 []string, slice2 []string) []string {
	intersection := []string{}
	for i := range slice1 {
		for j := range slice2 {
			if slice1[i] == slice2[j] {
				intersection = append(intersection, slice1[i])
			}
		}
	}
	return intersection
}

// Returns elements in 1 that are not in 2
func StringSliceDifference(slice1 []string, slice2 []string) []string {
	difference := []string{}
	for i := range slice1 {
		different := true
		for j := range slice2 {
			if slice1[i] == slice2[j] {
				different = false
				break
			}
		}
		if different {
			difference = append(difference, slice1[i])
		}
	}
	return difference
}

func StringsToInts(strs []string) ([]int, error) {
	if len(strs) == 1 && strs[0] == "" {
		return []int{}, nil
	}
	ints := make([]int, len(strs))
	for i := range strs {
		x, err := strconv.Atoi(strs[i])
		if err != nil {
			return nil, err
		}
		ints[i] = x
	}
	return ints, nil
}

func GetQueryArg(r *http.Request, key string) string {
	query := r.URL.Query()
	val, ok := query[key]
	if ok {
		return val[0]
	}
	return ""
}
