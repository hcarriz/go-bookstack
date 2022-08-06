package bookstack

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PreviewHTML struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}
type Search struct {
	ID          int         `json:"id,omitempty"`
	BookID      int         `json:"book_id,omitempty"`
	Slug        string      `json:"slug,omitempty"`
	Name        string      `json:"name,omitempty"`
	CreatedAt   time.Time   `json:"created_at,omitempty"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
	Type        ContentType `json:"type,omitempty"`
	URL         string      `json:"url,omitempty"`
	PreviewHTML PreviewHTML `json:"preview_html,omitempty"`
	Tags        []Tag       `json:"tags,omitempty"`
	ChapterID   int         `json:"chapter_id,omitempty"`
	Draft       bool        `json:"draft,omitempty"`
	Template    bool        `json:"template,omitempty"`
}

const SearchDateFormat = "2006-01-02"

type SearchParams struct {
	// Time Filters
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`

	// User Filters
	UpdatedBy *string `json:"updated_by,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
	OwnedBy   *string `json:"owned_by,omitempty"`

	// Content Filters
	InName *string `json:"in_name,omitempty"`
	InBody *string `json:"in_body,omitempty"`

	// Option Filters
	IsRestricted  bool          `json:"is_restricted,omitempty"`
	ViewedByMe    bool          `json:"viewed_by_me,omitempty"`
	NotViewedByMe bool          `json:"not_viewed_by_me,omitempty"`
	Type          []ContentType `json:"type,omitempty"`

	// Query
	Query string
	Page  *int
	Count *int
}

func (s SearchParams) String(q string) string {

	l := url.Values{}

	query := []string{
		s.Query,
	}

	if s.Page != nil {
		l.Add("page", fmt.Sprint(s.Page))
	}

	if s.Count != nil {
		l.Add("count", fmt.Sprint(s.Count))
	}

	if s.UpdatedAfter != nil {
		query = append(query, fmt.Sprintf("{updated_after:%s}", s.UpdatedAfter.Format(SearchDateFormat)))

	}

	if s.UpdatedBefore != nil {
		query = append(query, fmt.Sprintf("{updated_before:%s}", s.UpdatedBefore.Format(SearchDateFormat)))
	}

	if s.CreatedAfter != nil {
		query = append(query, fmt.Sprintf("{updated_after:%s}", s.CreatedAfter.Format(SearchDateFormat)))

	}

	if s.CreatedBefore != nil {
		query = append(query, fmt.Sprintf("{updated_before:%s}", s.CreatedBefore.Format(SearchDateFormat)))
	}

	if s.UpdatedBy != nil {
		by := "me"

		if *s.UpdatedBy != "" {
			by = *s.UpdatedBy
		}

		query = append(query, fmt.Sprintf("{updated_by:%s}", by))
	}

	if s.CreatedBy != nil {
		by := "me"

		if *s.CreatedBy != "" {
			by = *s.CreatedBy
		}

		query = append(query, fmt.Sprintf("{created_by:%s}", by))
	}

	if s.OwnedBy != nil {
		by := "me"

		if *s.OwnedBy != "" {
			by = *s.OwnedBy
		}

		query = append(query, fmt.Sprintf("{owned_by:%s}", by))
	}

	if s.InName != nil {
		query = append(query, fmt.Sprintf("{in_name:%s}", *s.InName))
	}

	if s.InBody != nil {
		query = append(query, fmt.Sprintf("{in_body:%s}", *s.InBody))
	}

	if s.ViewedByMe {
		query = append(query, "{viewed_by_me}")
	}

	if s.NotViewedByMe {
		query = append(query, "{not_viewed_by_me}")
	}

	if s.IsRestricted {
		query = append(query, "{is_restricted}")
	}

	if len(s.Type) > 0 {

		l := []string{}

		for _, t := range s.Type {
			l = append(l, string(t))
		}

		query = append(query, fmt.Sprintf("{type:%s}", strings.Join(l, "|")))
	}

	l.Add("query", strings.Join(query, " "))

	return fmt.Sprintf("%s?%s", q, l.Encode())
}

func (b *Bookstack) Search(ctx context.Context, query SearchParams) ([]Search, error) {

	raw, err := b.request(ctx, http.MethodGet, query.String("/search"), blank{})
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]Search](raw)

}
