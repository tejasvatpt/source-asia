package models

type RequestBody struct {
	UserID  string      `json:"user_id"`
	Payload interface{} `json:"payload"`
}

type RequestResponse struct {
	Message  string `json:"message"`
	UserID   string `json:"user_id"`
	Accepted bool   `json:"accepted"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type UserStats struct {
	UserID           string `json:"user_id"`
	AcceptedInWindow int    `json:"accepted_in_window"`
	RejectedTotal    int    `json:"rejected_total"`
}

type StatsResponse struct {
	Users []UserStats `json:"users"`
}

type Product struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
	CreatedAt string   `json:"created_at"`
}

type CreateProductRequest struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

type AddMediaRequest struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

type ProductListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SKU          string `json:"sku"`
	ImageCount   int    `json:"image_count"`
	VideoCount   int    `json:"video_count"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type ProductListResponse struct {
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Total    int               `json:"total"`
	Products []ProductListItem `json:"products"`
}
