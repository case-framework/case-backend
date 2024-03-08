package utils

func ContainsString(slice []string, searchTerm string) bool {
	for _, s := range slice {
		if searchTerm == s {
			return true
		}
	}
	return false
}
