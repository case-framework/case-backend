package study

func getTotalPages(totalCount int64, limit int64) int64 {
	if limit == 0 {
		return 0
	}
	return (totalCount + limit - 1) / limit
}
