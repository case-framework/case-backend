package study

type PaginationInfos struct {
	TotalCount  int64 `json:"totalCount"`
	CurrentPage int64 `json:"currentPage"`
	TotalPages  int64 `json:"totalPages"`
	PageSize    int64 `json:"pageSize"`
}
