package raw

func RawIntersection(aliceIds []int, bobIds []int) map[int]bool {
	aliceSet := make(map[int]bool)
	for _, id := range aliceIds {
		aliceSet[id] = true
	}

	resultSet := make(map[int]bool)
	for _, id := range bobIds {
		if aliceSet[id] {
			resultSet[id] = true
		}
	}

	return resultSet
}
