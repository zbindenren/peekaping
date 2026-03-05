package incident

type CreateDto struct {
	StatusPageID string `json:"status_page_id" validate:"required"`
	Title        string `json:"title" validate:"required"`
	Content      string `json:"content"`
	Style        string `json:"style" validate:"required,oneof=info warning danger"`
}

type UpdateDto struct {
	Title   *string `json:"title,omitempty"`
	Content *string `json:"content,omitempty"`
	Style   *string `json:"style,omitempty" validate:"omitempty,oneof=info warning danger"`
	Active  *bool   `json:"active,omitempty"`
}
