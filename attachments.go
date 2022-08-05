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

type Attachment struct {
	ID         int       `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Extension  string    `json:"extension,omitempty"`
	UploadedTo int       `json:"uploaded_to,omitempty"`
	External   bool      `json:"external,omitempty"`
	Order      int       `json:"order,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	CreatedBy  int       `json:"created_by,omitempty"`
	UpdatedBy  int       `json:"updated_by,omitempty"`
}

type AttachmentDetailed struct {
	ID         int       `json:"id,omitempty"`
	Name       string    `json:"name,omitempty"`
	Extension  string    `json:"extension,omitempty"`
	UploadedTo int       `json:"uploaded_to,omitempty"`
	External   bool      `json:"external,omitempty"`
	Order      int       `json:"order,omitempty"`
	CreatedBy  CreatedBy `json:"created_by,omitempty"`
	UpdatedBy  UpdatedBy `json:"updated_by,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
	Links      Links     `json:"links,omitempty"`
	Content    string    `json:"content,omitempty"`
}
type Links struct {
	HTML     string `json:"html,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

type AttachmentParams struct {
	Name       string `json:"name,omitempty"`
	UploadedTo int    `json:"uploaded_to,omitempty"`
	File       string `json:"file,omitempty"`
	Link       string `json:"link,omitempty"`
}

func (a AttachmentParams) Form() (string, io.Reader, error) {

	if a.File != "" {

		body := bytes.NewBuffer(nil)
		writer := multipart.NewWriter(body)

		defer writer.Close()

		if a.Name != "" {

			if err := writer.WriteField("name", a.Name); err != nil {
				return "", nil, err
			}

		}

		if a.Link != "" {

			if err := writer.WriteField("link", a.Link); err != nil {
				return "", nil, err
			}

		}

		if a.UploadedTo != 0 {

			if err := writer.WriteField("uploaded_to", strconv.Itoa(a.UploadedTo)); err != nil {
				return "", nil, err
			}

		}

		img, err := writer.CreateFormFile("file", filepath.Base(a.File))
		if err != nil {
			return "", nil, err
		}

		f, err := os.Open(a.File)
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

	r, err := json.Marshal(a)
	if err != nil {
		return "", nil, err
	}

	return appJSON, bytes.NewReader(r), nil

}

// ListAttachments will return the attachments that match the given params.
func (b *Bookstack) ListAttachments(ctx context.Context, params *QueryParams) ([]Attachment, error) {

	resp, err := b.request(ctx, http.MethodGet, params.String("/attachments"), blank{})
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]Attachment](resp)
}

// GetAttachment will return a single attachment that matches id.
func (b *Bookstack) GetAttachment(ctx context.Context, id int) (AttachmentDetailed, error) {

	resp, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/attachments/%d", id), blank{})
	if err != nil {
		return AttachmentDetailed{}, err
	}

	return ParseSingle[AttachmentDetailed](resp)

}

// CreateAttachment will create a attachment according to the given params.
func (b *Bookstack) CreateAttachment(ctx context.Context, params AttachmentParams) (Attachment, error) {

	resp, err := b.request(ctx, http.MethodPost, "/attachments", params)
	if err != nil {
		return Attachment{}, err
	}

	return ParseSingle[Attachment](resp)
}

// UpdateAttachment will update a attachment with the given params.
func (b *Bookstack) UpdateAttachment(ctx context.Context, id int, params AttachmentParams) (Attachment, error) {

	resp, err := b.request(ctx, http.MethodPut, fmt.Sprintf("/attachments/%d", id), params)
	if err != nil {
		return Attachment{}, err
	}

	return ParseSingle[Attachment](resp)
}

// DeleteAttachment will delete a attachment with the given id.
func (b *Bookstack) DeleteAttachment(ctx context.Context, id int) (bool, error) {

	if _, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/attachments/%d", id), blank{}); err != nil {
		return false, err
	}

	return true, nil
}
