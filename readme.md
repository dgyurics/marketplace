<p align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.webp?raw=true" alt="marketplace">
</p>

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
| Email delivery | ❌ | External SMTP (3rd party) |
| Payment processing | ❌ | Stripe (3rd party) |

## Planned Enhancements

* Username-based auth with admin-assisted account recovery
* One-click buy with pay-on-delivery
* Pay-on-delivery option at checkout
* Remove gorilla/mux dependency
* Documentation for production setup and configuration
* Product variants (size, color, material, etc.)
* Geographic access control via Nginx and GeoIP2
* Product full-text search
* Simplify deployment and configuration to the max

## Local Development

See [Getting Started](deploy/local/readme.md) for setup instructions.

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
