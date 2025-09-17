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

* Admin dashboard and management tools
* Database backup and recovery utilities
* Mobile-responsive design (phone/tablet support)
* Self-hosted mail server via docker-mailserver
* API caching and rate limiting
* Partial and full refund capabilities
* Product full-text search with filtering
* Product variants (size, color, material, etc.)
* Multi-currency and localization support
* Address validation via libpostal
* Discount codes and promotional campaigns
* Geographic access control via Nginx and GeoIP2
* Chat integration for verified customers
* Offline payment options for verified customers

## Local Development

Coming soon

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
