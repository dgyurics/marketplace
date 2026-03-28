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

* Mobile device support
* Documentation for setup and configuration
* Offline payment for trusted users
* Database backup and recovery scripts
* Product variants (size, color, material, etc.)
* Replace auth and account recovery with email-free solution
* Replace email notifications with in-app messaging/notifications
* Geographic access control via Nginx and GeoIP2
* Product full-text search

## Local Development

Coming soon

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
