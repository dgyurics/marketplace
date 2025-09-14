<p align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.webp?raw=true" alt="marketplace">
</p>

A self-hosted e-commerce framework designed for local businesses and community commerce. Minimal external dependencies and maximum self-reliance. See [https://selfco.io](https://selfco.io)

## Features

| Feature | Self-Hosted | Notes |
|---------|-------------|-------|
| Core backend (products, orders, users) | ✅ | Go + PostgreSQL |
| Modern web interface | ✅ | Vue 3 SPA |
| Container orchestration | ✅ | Full Docker deployment |
| Image optimization & serving | ✅ | imgproxy service |
| AI background removal | ✅ | rembg service |
| Payment processing | ❌ | Stripe (3rd party) |
| Email delivery | ❌ | External SMTP (3rd party) |

## Planned Enhancements

* Containerized mail server via docker-mailserver
* Offline payment methods for trusted customers
* Comprehensive admin dashboard and management tools
* Database backup and recovery toolkit
* API caching and rate limiting for improved performance
* Partial and full refund support
* Full-text search using PostgreSQL trigrams
* Product variants and SKU management (size, color, etc.)
* Multi-currency and language support
* Customer pickup scheduling
* Discount codes and promotional campaigns

## Local Development

Coming soon

## Production Deployment

Coming soon