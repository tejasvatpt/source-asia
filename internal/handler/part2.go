package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tejasvatpt/source-asia/internal/catalog"
	"github.com/tejasvatpt/source-asia/internal/models"
)

const (
	defaultProductLimit = 20
	maxProductLimit     = 100
	maxURLsPerRequest   = 20
	maxURLLength        = 2048
)

type Part2Handler struct {
	store *catalog.Store
}

func NewPart2Handler(s *catalog.Store) *Part2Handler {
	return &Part2Handler{store: s}
}

func (h *Part2Handler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createProduct(w, r)
	case http.MethodGet:
		h.listProducts(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET or POST")
	}
}

func (h *Part2Handler) HandleProductByID(w http.ResponseWriter, r *http.Request) {
	id, action, ok := parseProductPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}

	if action == "" && r.Method == http.MethodGet {
		h.getProduct(w, id)
		return
	}

	if action == "media" && r.Method == http.MethodPost {
		h.addMedia(w, r, id)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET for product detail or POST for media")
}

func (h *Part2Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var body models.CreateProductRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON: "+err.Error())
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.SKU = strings.TrimSpace(body.SKU)

	if err := validateProductInput(body); err != nil {
		writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	product, err := h.store.Create(body)
	if errors.Is(err, catalog.ErrDuplicateSKU) {
		writeError(w, http.StatusConflict, "duplicate_sku", "sku already exists")
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

func (h *Part2Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := readPagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}

	products, total := h.store.List(limit, offset)

	writeJSON(w, http.StatusOK, models.ProductListResponse{
		Limit:    limit,
		Offset:   offset,
		Total:    total,
		Products: products,
	})
}

func (h *Part2Handler) getProduct(w http.ResponseWriter, id string) {
	product, err := h.store.GetByID(id)
	if errors.Is(err, catalog.ErrProductNotFound) {
		writeError(w, http.StatusNotFound, "product_not_found", "product not found")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func (h *Part2Handler) addMedia(w http.ResponseWriter, r *http.Request, id string) {
	var body models.AddMediaRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON: "+err.Error())
		return
	}

	if len(body.ImageURLs) == 0 && len(body.VideoURLs) == 0 {
		writeError(w, http.StatusBadRequest, "missing_media", "at least one image_urls or video_urls value is required")
		return
	}

	if err := validateMedia(body.ImageURLs, body.VideoURLs); err != nil {
		writeError(w, http.StatusBadRequest, "validation_failed", err.Error())
		return
	}

	product, err := h.store.AddMedia(id, body)
	if errors.Is(err, catalog.ErrProductNotFound) {
		writeError(w, http.StatusNotFound, "product_not_found", "product not found")
		return
	}

	writeJSON(w, http.StatusOK, product)
}

func readPagination(r *http.Request) (int, int, error) {
	limit := defaultProductLimit
	offset := 0

	// Defaults keep the endpoint easy to call, while the max limit protects the
	// service from returning too much data in one response.
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			return 0, 0, errors.New("limit must be a positive number")
		}
		if parsed > maxProductLimit {
			return 0, 0, errors.New("limit cannot be greater than 100")
		}
		limit = parsed
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		parsed, err := strconv.Atoi(rawOffset)
		if err != nil || parsed < 0 {
			return 0, 0, errors.New("offset must be zero or a positive number")
		}
		offset = parsed
	}

	return limit, offset, nil
}

func validateProductInput(body models.CreateProductRequest) error {
	if body.Name == "" {
		return errors.New("name is required and must be non-empty")
	}

	if body.SKU == "" {
		return errors.New("sku is required and must be non-empty")
	}

	return validateMedia(body.ImageURLs, body.VideoURLs)
}

func validateMedia(imageURLs, videoURLs []string) error {
	if len(imageURLs) > maxURLsPerRequest {
		return errors.New("image_urls cannot contain more than 20 URLs")
	}

	if len(videoURLs) > maxURLsPerRequest {
		return errors.New("video_urls cannot contain more than 20 URLs")
	}

	for i, value := range imageURLs {
		imageURLs[i] = strings.TrimSpace(value)
		if err := validateURL(value); err != nil {
			return errors.New("invalid image url: " + err.Error())
		}
	}

	for i, value := range videoURLs {
		videoURLs[i] = strings.TrimSpace(value)
		if err := validateURL(value); err != nil {
			return errors.New("invalid video url: " + err.Error())
		}
	}

	return nil
}

func validateURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("url must be non-empty")
	}

	if len(value) > maxURLLength {
		return errors.New("url cannot be longer than 2048 characters")
	}

	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return errors.New("url must be valid")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("url must start with http:// or https://")
	}

	if parsed.Host == "" {
		return errors.New("url must include a host")
	}

	return nil
}

func parseProductPath(path string) (string, string, bool) {
	trimmed := strings.TrimPrefix(path, "/products/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")

	// Supported paths:
	// /products/{id}
	// /products/{id}/media
	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", true
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] == "media" {
		return parts[0], parts[1], true
	}

	return "", "", false
}
