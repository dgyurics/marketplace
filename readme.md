<p align="center">
  <img src="https://github.com/dgyurics/marketplace/blob/main/logo.webp?raw=true" alt="marketplace">
</p>

A self-hosted e-commerce framework designed for local businesses and community commerce. Minimal external dependencies and maximum self-reliance. See [https://selfco.io](https://selfco.io)

## Features

| Feature | Self-Hosted | Notes |
|---------|-------------|-------|
| Core backend (products, orders, users) | ✅ | Go + PostgreSQL |
| Modern web interface | ✅ | Vue 3 SPA |
| Back-office tools | ✅ | Product and order management |
| Container orchestration | ✅ | Full Docker deployment |
| Image optimization & serving | ✅ | imgproxy service |
| AI background removal | ✅ | rembg service |
| Payment processing | ❌ | Stripe (3rd party) |
| Email delivery | ❌ | External SMTP (3rd party) |

## Planned Enhancements

* API caching and rate limiting
* Self-hosted mail server via docker-mailserver
* Database backup and recovery utilities
* Partial and full refund capabilities
* Multi-currency and localization support
* Mobile-responsive design (phone/tablet support)
* Product full-text search with filtering
* Offline payment options for verified customers
* Address validation via libpostal
* Geographic access control via Nginx and GeoIP2
* Discount codes and promotional campaigns
* Chat integration for verified customers
* Product variants (size, color, material, etc.)

## Local Development

Coming soon

## Production Deployment

Coming soon

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
