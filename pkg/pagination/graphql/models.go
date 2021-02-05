// Pager represents params from list request
type Pager struct {
	DecCursor         string `json:"decCursor"`
	Limit             int64  `json:"limi"`
	ForwardPagination bool   `json:"forwardPagination"`
}

// PageInfo represents params from a page response
type PageInfo struct {
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
}