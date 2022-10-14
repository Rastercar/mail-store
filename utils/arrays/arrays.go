package arrays

func ToChunks(arr []string, chunkSize int) [][]string {
	if len(arr) == 0 {
		return [][]string{}
	}

	divided := make([][]string, (len(arr)+chunkSize-1)/chunkSize)

	prev := 0

	i := 0

	till := len(arr) - chunkSize

	for prev < till {
		next := prev + chunkSize
		divided[i] = arr[prev:next]
		prev = next
		i++
	}

	divided[i] = arr[prev:]

	return divided
}

func RemoveDuplicates(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, item := range s {
		if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}

	return list
}
