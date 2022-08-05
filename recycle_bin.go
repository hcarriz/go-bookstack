package bookstack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RecycleBinItem struct {
	ID            int             `json:"id,omitempty"`
	DeletedBy     int             `json:"deleted_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at,omitempty"`
	DeletableType DeletableType   `json:"deletable_type,omitempty"`
	DeletableID   int             `json:"deletable_id,omitempty"`
	Deletable     json.RawMessage `json:"deletable,omitempty"`
}

type RecycledBook struct {
	Book
	PagesCount    int `json:"pages_count,omitempty"`
	ChaptersCount int `json:"chapters_count,omitempty"`
}

type RecycledPage struct {
	ID            int       `json:"id,omitempty"`
	BookID        int       `json:"book_id,omitempty"`
	ChapterID     int       `json:"chapter_id,omitempty"`
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
	OwnedBy       int       `json:"owned_by,omitempty"`
	Editor        string    `json:"editor,omitempty"`
	BookSlug      string    `json:"book_slug,omitempty"`
	Parent        Parent    `json:"parent,omitempty"`
}

type RecycledChapter struct {
	Chapter
	BookSlug   string `json:"book_slug,omitempty"`
	PagesCount int    `json:"pages_count,omitempty"`
	Parent     Parent `json:"parent,omitempty"`
}

type Parent struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   int       `json:"created_by,omitempty"`
	UpdatedBy   int       `json:"updated_by,omitempty"`
	OwnedBy     int       `json:"owned_by,omitempty"`
	Type        string    `json:"type,omitempty"`
}

func (i RecycleBinItem) Book() (*RecycledBook, bool) {

	result, err := ParseSingle[RecycledBook]([]byte(i.Deletable))
	if err != nil {
		return nil, false
	}

	return &result, true
}

func (i RecycleBinItem) Chapter() (*RecycledChapter, bool) {

	result, err := ParseSingle[RecycledChapter]([]byte(i.Deletable))
	if err != nil {
		return nil, false
	}

	return &result, true
}

func (i RecycleBinItem) Shelf() (*Shelf, bool) {

	result, err := ParseSingle[Shelf]([]byte(i.Deletable))
	if err != nil {
		return nil, false
	}

	return &result, true
}

func (i RecycleBinItem) Page() (*RecycledPage, bool) {

	result, err := ParseSingle[RecycledPage]([]byte(i.Deletable))
	if err != nil {
		return nil, false
	}

	return &result, true
}

type DeletableType string

const (
	DeletedBook    DeletableType = "book"
	DeletedChapter DeletableType = "chapter"
	DeletedShelf   DeletableType = "bookshelf"
	DeletedPage    DeletableType = "page"
)

// ListRecycleBinItems will list the items in the recycle bin.
func (b *Bookstack) ListRecycleBinItems(ctx context.Context) ([]RecycleBinItem, error) {

	raw, err := b.request(ctx, http.MethodGet, "/recycle-bin", blank{})
	if err != nil {
		return nil, err
	}

	list, err := ParseMultiple[[]RecycleBinItem](raw)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// RestoreRecycleBinItem will restore an item from the recycle bin.
func (b *Bookstack) RestoreRecyleBinItem(ctx context.Context, id int) (int, error) {

	raw, err := b.request(ctx, http.MethodPut, fmt.Sprintf("/recycle-bin/%d", id), blank{})
	if err != nil {
		return 0, err
	}

	resp := struct {
		Count int `json:"restore_count"`
	}{}

	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// DeleteRecycleBinItem will delete an item from the recycle bin.
func (b *Bookstack) DeleteRecycleBinItem(ctx context.Context, id int) (int, error) {

	raw, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/recycle-bin/%d", id), blank{})
	if err != nil {
		return 0, err
	}

	resp := struct {
		Count int `json:"delete_count"`
	}{}

	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}
