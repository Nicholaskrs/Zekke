package helpers

func InArray[T comparable](src []T, val T) bool {
	for _, source := range src {
		if source == val {
			return true
		}
	}
	return false
}

// ArrayDiff used to get different value between two arrays
// It returns different value in array
func ArrayDiff[T comparable](arr1 []T, arr2 []T) []T {
	diff := make([]T, 0)
	for _, v1 := range arr1 {
		flagExists := false
		for _, v2 := range arr2 {
			if v1 == v2 {
				flagExists = true
				break
			}
		}
		if !flagExists {
			diff = append(diff, v1)
		}
	}

	for _, v1 := range arr2 {
		flagExists := false
		for _, v2 := range arr1 {
			if v1 == v2 {
				flagExists = true
				break
			}
		}
		if !flagExists {
			diff = append(diff, v1)
		}
	}

	return diff
}

// ArrayIntersectKey
func ArrayIntersectKey[T1 comparable, T2 comparable](map1, map2 map[T1]T2) map[T1]T2 {
	intersect := make(map[T1]T2)
	for key := range map1 {
		if _, exists := map2[key]; exists {
			intersect[key] = map1[key]
		}
	}
	return intersect
}
