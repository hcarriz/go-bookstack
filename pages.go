package bookstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Page struct {
	ID            int       `json:"id,omitempty"`
	BookID        int       `json:"book_id,omitempty"`
	PageID        int       `json:"page_id,omitempty"`
	Name          string    `json:"name,omitempty"`
	Slug          string    `json:"slug,omitempty"`
	Priority      int       `json:"priority,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	CreatedBy     int       `json:"created_by,omitempty"`
	UpdatedBy     int       `json:"updated_by,omitempty"`
	Draft         bool      `json:"draft,omitempty"`
	RevisionCount int       `json:"revision_count,omitempty"`
	Template      bool      `json:"template,omitempty"`
}

type PageDetailed struct {
	ID            int       `json:"id,omitempty"`
	BookID        int       `json:"book_id,omitempty"`
	ChapterID     int       `json:"chapter_id,omitempty"`
	Name          string    `json:"name,omitempty"`
	Slug          string    `json:"slug,omitempty"`
	HTML          string    `json:"html,omitempty"`
	Priority      int       `json:"priority,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	CreatedBy     CreatedBy `json:"created_by,omitempty"`
	UpdatedBy     UpdatedBy `json:"updated_by,omitempty"`
	OwnedBy       OwnedBy   `json:"owned_by,omitempty"`
	Draft         bool      `json:"draft,omitempty"`
	Markdown      string    `json:"markdown,omitempty"`
	RevisionCount int       `json:"revision_count,omitempty"`
	Template      bool      `json:"template,omitempty"`
	Tags          []Tag     `json:"tags,omitempty"`
}

type PageParams struct {
	BookID    int         `json:"book_id,omitempty"`
	ChapterID int         `json:"chapterID,omitempty"`
	Name      string      `json:"name,omitempty"`
	HTML      string      `json:"html,omitempty"`
	Markdown  string      `json:"markdown,omitempty"`
	Tags      []TagParams `json:"tags,omitempty"`
}

func (bp PageParams) Form() (string, io.Reader, error) {

	r, err := json.Marshal(bp)
	if err != nil {
		return "", nil, err
	}

	return appJSON, bytes.NewReader(r), nil

}

// ListPages will return the pages that match the given params.
func (b *Bookstack) ListPages(ctx context.Context, params *QueryParams) ([]Page, error) {

	resp, err := b.request(ctx, http.MethodGet, params.String("/pages"), blank{})
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]Page](resp)
}

// GetPage will return a single page that matches id.
func (b *Bookstack) GetPage(ctx context.Context, id int) (PageDetailed, error) {

	resp, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/pages/%d", id), blank{})
	if err != nil {
		return PageDetailed{}, err
	}

	return ParseSingle[PageDetailed](resp)

}

// CreatePage will create a page according to the given params.
func (b *Bookstack) CreatePage(ctx context.Context, params PageParams) (Page, error) {

	resp, err := b.request(ctx, http.MethodPost, "/pages", params)
	if err != nil {
		return Page{}, err
	}

	return ParseSingle[Page](resp)
}

// UpdatePage will update a page with the given params.
func (b *Bookstack) UpdatePage(ctx context.Context, id int, params PageParams) (Page, error) {

	resp, err := b.request(ctx, http.MethodPut, fmt.Sprintf("/pages/%d", id), params)
	if err != nil {
		return Page{}, err
	}

	return ParseSingle[Page](resp)
}

// DeletePage will delete a page with the given id.
func (b *Bookstack) DeletePage(ctx context.Context, id int) (bool, error) {

	if _, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/pages/%d", id), blank{}); err != nil {
		return false, err
	}

	return true, nil
}

// ExportPageHTML will return a page in HTML format.
func (b *Bookstack) ExportPageHTML(ctx context.Context, id int) (io.Reader, error) {

	data, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/pages/%d/export/html", id), blank{})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil

}

// ExportPagePDF will return a page in PDF format.
func (b *Bookstack) ExportPagePDF(ctx context.Context, id int) (io.Reader, error) {

	data, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/pages/%d/export/pdf", id), blank{})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil

}

// ExportPageMarkdown will return a page in Markdown format.
func (b *Bookstack) ExportPageMarkdown(ctx context.Context, id int) (io.Reader, error) {

	data, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/pages/%d/export/markdown", id), blank{})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil

}

// ExportPagePlaintext will return a page in Plaintext format.
func (b *Bookstack) ExportPagePlaintext(ctx context.Context, id int) (io.Reader, error) {

	data, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/pages/%d/export/plaintext", id), blank{})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil

}
