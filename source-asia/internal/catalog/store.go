package catalog

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tejasvatpt/source-asia/internal/models"
)

var ErrDuplicateSKU = errors.New("duplicate sku")
var ErrProductNotFound = errors.New("product not found")

type Store struct {
	mu       sync.RWMutex
	nextID   int
	products map[string]*models.Product
	skuToID  map[string]string
	order    []string
}

func NewStore() *Store {
	return &Store{
		nextID:   1,
		products: make(map[string]*models.Product),
		skuToID:  make(map[string]string),
		order:    make([]string, 0),
	}
}

func (s *Store) Create(req models.CreateProductRequest) (models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.skuToID[req.SKU]; exists {
		return models.Product{}, ErrDuplicateSKU
	}

	id := fmt.Sprintf("%d", s.nextID)
	s.nextID++

	product := &models.Product{
		ID:        id,
		Name:      req.Name,
		SKU:       req.SKU,
		ImageURLs: copyStrings(req.ImageURLs),
		VideoURLs: copyStrings(req.VideoURLs),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	s.products[id] = product
	s.skuToID[req.SKU] = id
	s.order = append(s.order, id)

	return cloneProduct(product), nil
}

func (s *Store) List(limit, offset int) ([]models.ProductListItem, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.order)
	if offset > total {
		return []models.ProductListItem{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]models.ProductListItem, 0, end-offset)
	for _, id := range s.order[offset:end] {
		product := s.products[id]

		// The list endpoint only builds small summary objects. It does not copy
		// or return the full image/video URL slices for every product.
		item := models.ProductListItem{
			ID:         product.ID,
			Name:       product.Name,
			SKU:        product.SKU,
			ImageCount: len(product.ImageURLs),
			VideoCount: len(product.VideoURLs),
			CreatedAt:  product.CreatedAt,
		}

		if len(product.ImageURLs) > 0 {
			item.ThumbnailURL = product.ImageURLs[0]
		}

		items = append(items, item)
	}

	return items, total
}

func (s *Store) GetByID(id string) (models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	product, ok := s.products[id]
	if !ok {
		return models.Product{}, ErrProductNotFound
	}

	return cloneProduct(product), nil
}

func (s *Store) AddMedia(id string, req models.AddMediaRequest) (models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	product, ok := s.products[id]
	if !ok {
		return models.Product{}, ErrProductNotFound
	}

	product.ImageURLs = append(product.ImageURLs, req.ImageURLs...)
	product.VideoURLs = append(product.VideoURLs, req.VideoURLs...)

	return cloneProduct(product), nil
}

func cloneProduct(p *models.Product) models.Product {
	return models.Product{
		ID:        p.ID,
		Name:      p.Name,
		SKU:       p.SKU,
		ImageURLs: copyStrings(p.ImageURLs),
		VideoURLs: copyStrings(p.VideoURLs),
		CreatedAt: p.CreatedAt,
	}
}

func copyStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	// Return a copy so callers cannot change the store's internal slices.
	copied := make([]string, len(values))
	copy(copied, values)
	return copied
}
