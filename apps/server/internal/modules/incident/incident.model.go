package incident

import "time"

type Model struct {
	ID           string     `json:"id"`
	StatusPageID string     `json:"status_page_id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Style        string     `json:"style"`
	Active       bool       `json:"active"`
	ResolvedAt   *time.Time `json:"resolved_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
