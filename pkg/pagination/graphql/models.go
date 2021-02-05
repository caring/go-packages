
// PageInfo represents params from a page response
type PageInfo struct {
	StartCursor     string `json:"start_cursor"`
	EndCursor       string `json:"end_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
	HasPreviousPage bool   `json:"has_previous_page"`
}