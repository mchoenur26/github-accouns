package models

import (
	"time"
)

// bentuk database
type Client struct {
	ID            int        `json:"id"`
	Name          string     `json:"name" validate:"required"`
	Slug          string     `json:"slug"`
	IsProject     string     `json:"is_project"`
	SelfCapture   string     `json:"self_capture"`
	ClientPrefix  string     `json:"client_prefix" validate:"required"`
	ClientLogoURL string     `json:"client_logo_url"`
	Address       *string    `json:"address"`
	PhoneNumber   *string    `json:"phone_number"`
	City          *string    `json:"city"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}
