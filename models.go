package bookstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

type Response struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Total int             `json:"total,omitempty"`
	Err   struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

func (r Response) Error() error {
	switch {
	case r.Err.Code != 0 || r.Err.Message != "":
		return fmt.Errorf("%d %s", r.Err.Code, r.Err.Message)
	default:
		return nil
	}
}

type QueryParams struct {
	Count          int
	Offset         int
	SortField      string
	SortDescending bool
	FilterField    string
	FilterValue    string
}

func (q *QueryParams) String(l string) string {

	if q == nil {
		return l
	}

	u := url.Values{}

	if q.Count != 0 {
		u.Add("count", strconv.Itoa(q.Count))
	}

	if q.Offset != 0 {
		u.Add("offset", strconv.Itoa(q.Offset))
	}

	if q.SortField != "" {

		v := "+"
		if q.SortDescending {
			v = "-"
		}

		u.Add("sort", fmt.Sprintf("%s%s", v, q.SortField))
	}

	if q.FilterField != "" && q.FilterValue != "" {
		u.Add(fmt.Sprintf("filter[%s]", q.FilterField), q.FilterValue)
	}

	return fmt.Sprintf("%s?%s", l, u.Encode())

}

type Tag struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Order int    `json:"order,omitempty"`
}

type TagParams struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type CreatedBy struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type UpdatedBy struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type OwnedBy struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Cover struct {
	ID         int       `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	URL        string    `json:"url,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	CreatedBy  int       `json:"created_by,omitempty"`
	UpdatedBy  int       `json:"updated_by,omitempty"`
	Path       string    `json:"path,omitempty"`
	Type       string    `json:"type,omitempty"`
	UploadedTo int       `json:"uploaded_to,omitempty"`
}

type blank struct{}

func (b blank) Form() (string, io.Reader, error) {
	return "", nil, nil
}
