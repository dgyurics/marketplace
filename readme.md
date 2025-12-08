<p align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.webp?raw=true" alt="marketplace">
</p>

A self-hosted e-commerce framework designed for local businesses and community commerce. Minimal external dependencies and maximum self-reliance. See [https://selfco.io](https://selfco.io)

## Features

| Feature | Self-Hosted | Notes |
|---------|-------------|-------|
| Core backend (products, orders, users) | ✅ | Go + PostgreSQL |
| Web interface | ✅ | Vue 3 |
| Admin dashboard | ✅ | Product + Order management |
| Deployment | ✅ | Docker Compose orchestration |
| Image processing | ✅ | imgproxy + rembg (AI) background removal |
| Payment processing | ❌ | Stripe (3rd party) |
| Email delivery | ❌ | External SMTP (3rd party) |

## Planned Enhancements

* Geographic shipping controls
* Customer support messaging system
* Offline payment for trusted users
* Self-hosted email server via boky/postfix
* Database backup and recovery scripts
* Mobile device support
* Product full-text search
* Chat integration for trusted users
* Product variants (size, color, material, etc.)
* Geographic access control via Nginx and GeoIP2

## Local Development

Coming soon

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
