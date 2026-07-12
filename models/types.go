package models

// PaginatedProductResponse wraps a paginated list of products with page metadata.
type PaginatedProductResponse struct {
	Data       []Product `json:"data"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}
