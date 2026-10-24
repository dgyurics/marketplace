<div align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.png?raw=true" alt="marketplace" width="300">
</div>

A self-hosted e-commerce framework designed for local businesses and community commerce. Minimal external dependencies and maximum self-reliance. See [https://marketly.sh](https://marketly.sh)

## Features

| Feature | Self-Hosted | Notes |
|---------|-------------|-------|
| Core backend (products, orders, users) | ✅ | Go + PostgreSQL |
| Web interface | ✅ | Vue 3 |
| Deployment | ✅ | Docker Compose orchestration |
| Admin dashboard | ✅ | Product + Order management |
| Geographic shipping | ✅ | Configure coverage and exclusions |
| Image processing | ✅ | imgproxy + rembg (AI) background removal |
| In-app notifications | ✅ | Status updates on orders & offers |
| Payment processing | ✅ | Pay on delivery, Stripe optional |
| Email delivery | ❌ | External SMTP (3rd party) |

## Planned Enhancements

* Cache-Control + ETags
* Optional image enhance + field suggestions
* No email required for login & registration
* Documentation for production setup & deployment
* Product variants (size, color, material, etc.)
* Geographic access control via Nginx and GeoIP2
* Product full-text search
* Simplify configuration & deployment to the max

## Local Development

See [Getting Started](deploy/local/readme.md) for setup instructions.

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
