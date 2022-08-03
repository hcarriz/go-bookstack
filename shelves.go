package bookstack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Shelf struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   int       `json:"created_by,omitempty"`
	UpdatedBy   int       `json:"updated_by,omitempty"`
	OwnedBy     int       `json:"owned_by,omitempty"`
}

type ShelfDetailed struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedBy   CreatedBy `json:"created_by,omitempty"`
	UpdatedBy   UpdatedBy `json:"updated_by,omitempty"`
	OwnedBy     OwnedBy   `json:"owned_by,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Tags        []Tag     `json:"tags,omitempty"`
	Cover       Cover     `json:"cover,omitempty"`
	Books       []Book    `json:"books,omitempty"`
}

type ShelfParams struct {
	Name        string      `json:"name,omitempty"`
	Description string      `json:"description,omitempty"`
	Books       []int       `json:"books,omitempty"`
	Tags        []TagParams `json:"tags,omitempty"`
	Image       string      `json:"image,omitempty"`
}

func (bp ShelfParams) Form() (string, io.Reader, error) {

	if bp.Image != "" {

		body := bytes.NewBuffer(nil)
		writer := multipart.NewWriter(body)

		defer writer.Close()

		if bp.Name != "" {

			if err := writer.WriteField("name", bp.Name); err != nil {
				return "", nil, err
			}

		}

		for _, x := range bp.Books {

			if err := writer.WriteField("books", strconv.Itoa(x)); err != nil {
				return "", nil, err
			}

		}

		// TODO: Tags

		if bp.Description != "" {

			if err := writer.WriteField("description", bp.Description); err != nil {
				return "", nil, err
			}

		}

		img, err := writer.CreateFormFile("image", filepath.Base(bp.Image))
		if err != nil {
			return "", nil, err
		}

		f, err := os.Open(bp.Image)
		if err != nil {
			return "", nil, err
		}

		defer f.Close()

		if _, err := io.Copy(img, f); err != nil {
			return "", nil, err
		}

		if err := writer.Close(); err != nil {
			return "", nil, err
		}

		return writer.FormDataContentType(), bytes.NewReader(body.Bytes()), nil

	}

	r, err := json.Marshal(bp)
	if err != nil {
		return "", nil, err
	}

	return appJSON, bytes.NewReader(r), nil

}

// ListShelves will return the shelves that match the given params.
func (b *Bookstack) ListShelves(ctx context.Context, params *QueryParams) ([]Shelf, error) {

	resp, err := b.request(ctx, http.MethodGet, params.String("/shelves"), blank{})
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]Shelf](resp)
}

// GetShelf will return a single shelf that matches id.
func (b *Bookstack) GetShelf(ctx context.Context, id int) (ShelfDetailed, error) {

	resp, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/shelves/%d", id), blank{})
	if err != nil {
		return ShelfDetailed{}, err
	}

	return ParseSingle[ShelfDetailed](resp)

}

// CreateShelf will create a shelf according to the given params.
func (b *Bookstack) CreateShelf(ctx context.Context, params ShelfParams) (Shelf, error) {

	resp, err := b.request(ctx, http.MethodPost, "/shelves", params)
	if err != nil {
		return Shelf{}, err
	}

	return ParseSingle[Shelf](resp)
}

// UpdateShelf will update a shelf with the given params.
func (b *Bookstack) UpdateShelf(ctx context.Context, id int, params ShelfParams) (Shelf, error) {

	resp, err := b.request(ctx, http.MethodPut, fmt.Sprintf("/shelves/%d", id), params)
	if err != nil {
		return Shelf{}, err
	}

	return ParseSingle[Shelf](resp)
}

// DeleteShelf will delete a shelf with the given id.
func (b *Bookstack) DeleteShelf(ctx context.Context, id int) (bool, error) {

	if _, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/shelves/%d", id), blank{}); err != nil {
		return false, err
	}

	return true, nil
}
