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

type User struct {
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	ExternalAuthID string    `json:"external_auth_id"`
	Slug           string    `json:"slug"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
	ID             int       `json:"id"`
	ProfileURL     string    `json:"profile_url"`
	EditURL        string    `json:"edit_url"`
	AvatarURL      string    `json:"avatar_url"`
	Roles          []struct {
		ID          int    `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"roles"`
}

type UserParams struct {
	Name           string `json:"name,omitempty"`
	Email          string `json:"email,omitempty"`
	ExternalAuthID string `json:"external_auth_id,omitempty"`
	Password       string `json:"password,omitempty"`
	Language       string `json:"language,omitempty"`
	Roles          []int  `json:"roles,omitempty"`
	SendInvite     bool   `json:"send_invite,omitempty"`
}

func (u UserParams) Form() (string, io.Reader, error) {

	data, err := json.Marshal(u)
	if err != nil {
		return "", nil, err
	}

	return "application/json", bytes.NewReader(data), nil
}

type UserDeleteParams struct {
	MigrateOwnershipID int `json:"migrate_ownership_id,omitempty"`
}

func (ud UserDeleteParams) Form() (string, io.Reader, error) {
	data, err := json.Marshal(ud)
	if err != nil {
		return "", nil, err
	}

	return appJSON, bytes.NewReader(data), nil
}

// ListUsers will return the users the match the given params.
func (b *Bookstack) ListUsers(ctx context.Context, params *QueryParams) ([]User, error) {

	resp, err := b.request(ctx, http.MethodGet, params.String("/users"), blank{})
	if err != nil {
		return nil, err
	}

	return ParseMultiple[[]User](resp)
}

// GetUser will return the user assigned to the given id, or an error.
func (b *Bookstack) GetUser(ctx context.Context, id int) (User, error) {

	resp, err := b.request(ctx, http.MethodGet, fmt.Sprintf("/users/%d", id), blank{})
	if err != nil {
		return User{}, err
	}

	return ParseSingle[User](resp)
}

// CreateUser will create a user from the given params.
func (b *Bookstack) CreateUser(ctx context.Context, params UserParams) (User, error) {

	resp, err := b.request(ctx, http.MethodPost, "/users", params)
	if err != nil {
		return User{}, err
	}

	return ParseSingle[User](resp)
}

// UpdateUser will update the a user with the given params.
func (b *Bookstack) UpdateUser(ctx context.Context, id int, params UserParams) (User, error) {

	resp, err := b.request(ctx, http.MethodPut, fmt.Sprintf("/users/%d", id), params)
	if err != nil {
		return User{}, err
	}

	return ParseSingle[User](resp)
}

// Delete user will delete a user.
func (b *Bookstack) DeleteUser(ctx context.Context, id int, params *UserDeleteParams) (bool, error) {

	p := UserDeleteParams{}

	if params != nil {
		p = *params
	}

	if _, err := b.request(ctx, http.MethodDelete, fmt.Sprintf("/users/%d", id), p); err != nil {
		return false, err
	}

	return true, nil
}
