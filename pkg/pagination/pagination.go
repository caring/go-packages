package pagination

import (
	"encoding/base64"
	"errors"
	"fmt"
)

// Pager represents params from list request
type Pager struct {
	// Store the decoded cursor within the service, encode/decode it as it passes through the API
	DecCursor         string
	Limit             int64
	ForwardPagination bool
}

// NewPager creates Pager object from proto struct
func NewPager(pr PaginationRequest) (*Pager, error) {
	after := pr.GetAfter()
	before := pr.GetBefore()

	if len(after) > 0 && len(before) > 0 {
		return nil, errors.New("invalid pagination request. you may only use one of after or before")
	}

	first := pr.GetFirst()
	last := pr.GetLast()

	if first < 0 || last < 0 {
		return nil, errors.New("invalid pagination request. first and last must be a positive number")
	}

	var (
		cursor  string
		limit   int64
		forward bool
		err     error
	)
	// There are 3 cases to account for. Forward params given, backward params given
	// and no params given. No params given defaults to forward pagination with no cursor
	if len(after) > 0 {
		cursor = after
		limit = first
		forward = true
	} else if len(before) > 0 {
		cursor = before
		limit = last
		forward = false
	} else {
		cursor = after
		limit = first
		forward = true
	}
	cursor, err = DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}
	return &Pager{
		DecCursor:         cursor,
		Limit:             limit,
		ForwardPagination: forward,
	}, nil
}

// Page is a struct representation of data related to pagination
// ToProto converts a DB layer struct to a protobuf struct
func (p *Page) ToProto() *PageInfo {
	return &PageInfo{
		StartCursor:     EncodeCursor(p.StartCursor),
		EndCursor:       EncodeCursor(p.EndCursor),
		HasNextPage:     p.HasNextPage,
		HasPreviousPage: p.HasPreviousPage,
	}
}

// NewPageInfo creates PageInfo object
func NewPage(hasNextPage bool, hasPrevPage bool, firstCursor string, lastCursor string) *Page {
	return &Page{
		StartCursor:     firstCursor,
		EndCursor:       lastCursor,
		HasNextPage:     hasNextPage,
		HasPreviousPage: hasPrevPage,
	}
}

// DecodeCursor decodes base64 cursor
func DecodeCursor(c string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(c)
	if err != nil {
		return "", errors.New("Decode error: " + err.Error() + " for base64 cursor " + fmt.Sprint(c))
	}
	return string(decoded), nil
}

// EncodeCursor base64 encodes the given string
func EncodeCursor(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}
