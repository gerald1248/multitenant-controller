package main

func unique(slice []string) []string {
	keys := make(map[string]struct{})
	list := []string{}
	for _, entry := range slice {
			if _, ok := keys[entry]; !ok {
					keys[entry] = struct{}{}
					list = append(list, entry)
			}
	}
	return list
}