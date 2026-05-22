# Source Asia Backend Assignment

Small Go HTTP server for both assignment parts.

```txt
Part 1 -> rate limited API
Part 2 -> product catalog with media URLs
```

Uses only Go standard library. Data is stored in memory.

## Run on Windows

If using GitHub:

```powershell
git clone <repo-link>
cd source-asia
```

If using zip, unzip it and open the folder:

```powershell
cd source-asia
```

Run:

```powershell
go run ./cmd/server
```

Server starts on:

```txt
http://localhost:8080
```

Check compile:

```powershell
go test ./...
```

## Part 1 - Rate Limited API

Endpoints:

```txt
POST /request
GET /stats
```

Rules:

```txt
5 accepted requests per user
rolling 60 second window
201 -> accepted
429 -> rate limit crossed
rejected count -> lifetime count while server is running
concurrency safe using mutexes
```

Send request:

```powershell
$body = @{
  user_id = "alice"
  payload = @{ action = "buy" }
} | ConvertTo-Json -Depth 5

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/request" `
  -ContentType "application/json" `
  -Body $body
```

Trigger rate limit:

```powershell
1..6 | ForEach-Object {
  $body = @{
    user_id = "bob"
    payload = "test"
  } | ConvertTo-Json -Depth 5

  Invoke-RestMethod -Method Post `
    -Uri "http://localhost:8080/request" `
    -ContentType "application/json" `
    -Body $body
}
```

First 5 requests pass. The 6th request returns:

```txt
429 -> rate_limit_exceeded
```

Check stats:

```powershell
Invoke-RestMethod -Method Get `
  -Uri "http://localhost:8080/stats" | ConvertTo-Json -Depth 5
```

Common errors:

```txt
bad JSON -> 400 invalid_json
missing user_id -> 400 missing_user_id
missing payload -> 400 missing_payload
limit crossed -> 429 rate_limit_exceeded
```

## Part 2 - Product Catalog

Endpoints:

```txt
POST /products
GET /products
GET /products/{id}
POST /products/{id}/media
```

Product rules:

```txt
name is required
sku is required and unique
media are URL strings only
URLs must start with http:// or https://
max URL length is 2048
max 20 image URLs and 20 video URLs per request
```

Create product:

```powershell
$body = @{
  name = "Product A"
  sku = "SKU-001"
  image_urls = @(
    "https://cdn.example.com/products/sku-001/img-1.jpg",
    "https://cdn.example.com/products/sku-001/img-2.jpg"
  )
  video_urls = @(
    "https://cdn.example.com/products/sku-001/demo.mp4"
  )
} | ConvertTo-Json -Depth 5

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/products" `
  -ContentType "application/json" `
  -Body $body
```

Expected:

```txt
201 -> product created
```

Run the same command again to check duplicate SKU:

```txt
409 -> duplicate_sku
```

List products:

```powershell
Invoke-RestMethod -Method Get `
  -Uri "http://localhost:8080/products?limit=20&offset=0" | ConvertTo-Json -Depth 5
```

The list response returns summary fields only:

```txt
id, name, sku, image_count, video_count, thumbnail_url, created_at
```

It does not return all image/video URLs. Full media URLs are returned only in the detail API.

Get product detail:

```powershell
Invoke-RestMethod -Method Get `
  -Uri "http://localhost:8080/products/1" | ConvertTo-Json -Depth 5
```

Add media:

```powershell
$body = @{
  image_urls = @("https://cdn.example.com/products/sku-001/img-3.jpg")
  video_urls = @("https://cdn.example.com/products/sku-001/review.mp4")
} | ConvertTo-Json -Depth 5

Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/products/1/media" `
  -ContentType "application/json" `
  -Body $body
```

Validation examples:

```txt
duplicate sku -> 409 duplicate_sku
invalid URL -> 400 validation_failed
empty media body -> 400 missing_media
unknown product -> 404 product_not_found
```

Unknown product check:

```powershell
curl.exe -i "http://localhost:8080/products/999"
```

## Storage

Part 1 stores recent request timestamps per user.

Part 2 uses:

```txt
products map[id]*Product
skuToID map[sku]id
order []id
```

This keeps create, duplicate SKU check, list, detail, and media append simple.

For the product list API, only summary fields are prepared. For the detail API, the full product with all media URLs is returned.

## Production Notes

Current limits:

```txt
restart clears data
single server instance only
no authentication
no real CDN upload
no database
```

In production I would use Redis for shared rate limiting, PostgreSQL for product/media data, and a CDN or object storage service for actual media files.

With PostgreSQL, products would be stored in a `products` table and media URLs would be stored in a separate `product_media` table. The `sku` column would have a unique index. The list API would fetch only product summary data, and the detail API would fetch media URLs for one selected product.

## Project Structure

```txt
source-asia/
  cmd/
    server/
      main.go
  internal/
    catalog/
      store.go
    handler/
      part1.go
      part2.go
    models/
      models.go
    ratelimit/
      limiter.go
  go.mod
  README.md
```

## AI Usage

AI was used to help understand the assignment requirements and shape the README text.
