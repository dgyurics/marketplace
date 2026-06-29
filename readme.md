<p align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.webp?raw=true" alt="marketplace">
</p>

A self-hosted e-commerce framework designed for local businesses and community commerce. Minimal external dependencies and maximum self-reliance. See [https://selfco.io](https://selfco.io)

## Features

| Feature | Self-Hosted | Notes |
|---------|-------------|-------|
| Core backend (products, orders, users) | ✅ | Go + PostgreSQL |
| Web interface | ✅ | Vue 3 |
| Deployment | ✅ | Docker Compose orchestration |
| Admin dashboard | ✅ | Product + Order management |
| Geographic shipping | ✅ | Configure coverage and exclusions |
| Image processing | ✅ | imgproxy + rembg (AI) background removal |
| Email delivery | ❌ | External SMTP (3rd party) |
| Payment processing | ❌ | Stripe (3rd party) |

## Planned Enhancements

* Documentation for production setup and configuration
* Demo mode for the admin dashboard with anonymized data
* Pay on delivery for approved buyers
* Product variants (size, color, material, etc.)
* Support email-free auth and account recovery
* Support email-free operation with in-app notifications
* Geographic access control via Nginx and GeoIP2
* Product full-text search
* Simplify deployment and configuration to the max

## Local Development

See [Getting Started](deploy/local/readme.md) for setup instructions.

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
