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
	"time"
)

type Book struct {
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

type BookDetailed struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Slug        string    `json:"slug,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"created_by,omitempty"`
	UpdatedBy struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"updated_by,omitempty"`
	OwnedBy struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"owned_by,omitempty"`
	Tags []struct {
		ID    int    `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
		Order int    `json:"order,omitempty"`
	} `json:"tags,omitempty"`
	Cover struct {
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
	} `json:"cover,omitempty"`
}

type BookParams struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
}

func (bp BookParams) Form() (string, io.Reader, error) {

	if bp.Image != "" {

		body := bytes.NewBuffer(nil)
		writer := multipart.NewWriter(body)

		defer writer.Close()

		if bp.Name != "" {

			if err := writer.WriteField("name", bp.Name); err != nil {
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

	return "application/json", bytes.NewReader(r), nil

}

// ListBooks will return the books that match the given params.
func (b *Bookstack) ListBooks(ctx context.Context, params *QueryParams) ([]Book, error) {

	resp, err := b.request(ctx, http.MethodGet, params.String("/books"), nil)
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]Book](resp)
}

// GetBook will return a single book that matches id.
func (b *Bookstack) GetBook(ctx context.Context, id int) (BookDetailed, error) {

	resp, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/books/%d", id), nil)
	if err != nil {
		return BookDetailed{}, err
	}

	return ParseSingle[BookDetailed](resp)

}

// CreateBook will create a book according to the given params.
func (b *Bookstack) CreateBook(ctx context.Context, params BookParams) (Book, error) {

	resp, err := b.form(ctx, http.MethodPost, "/books", params)
	if err != nil {
		return Book{}, err
	}

	return ParseSingle[Book](resp)
}

// UpdateBook will update a book with the given params.
func (b *Bookstack) UpdateBook(ctx context.Context, id int, params BookParams) (Book, error) {

	resp, err := b.form(ctx, http.MethodPut, fmt.Sprintf("/books/%d", id), params)
	if err != nil {
		return Book{}, err
	}

	return ParseSingle[Book](resp)
}

// DeleteBook will delete a book with the given id.
func (b *Bookstack) DeleteBook(ctx context.Context, id int) (bool, error) {

	if _, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/books/%d", id), nil); err != nil {
		return false, err
	}

	return true, nil
}
