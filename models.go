package bookstack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
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

type TagReq struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Blank struct{}

func (b Blank) Form() (string, io.Reader, error) {
	return "", nil, nil
}
