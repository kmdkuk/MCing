package utils

import "maps"

// MergeMap merges two maps.
func MergeMap(m1, m2 map[string]string) map[string]string {
	m := make(map[string]string)
	maps.Copy(m, m1)
	maps.Copy(m, m2)
	if len(m) == 0 {
		return nil
	}
	return m
}
