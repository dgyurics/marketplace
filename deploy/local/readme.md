# Local Development

## Prerequisites

- [Go](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)
- [Docker](https://docs.docker.com/get-docker/)
- [Stripe CLI](https://docs.stripe.com/stripe-cli)

## Setup

All commands run from the project root.

1. Copy env
```bash
cp deploy/local/example.env deploy/local/.env
```

2. Generate RSA keys
```bash
make keys
cp private.pem deploy/local/
cp public.pem deploy/local/
```

3. Start Postgres and supporting services
```bash
docker compose -f deploy/local/docker-compose.yaml up -d
```

4. Start backend and frontend
```bash
make -j2 dev-backend dev-frontend
```

| Service | URL |
|---------|-----|
| App | http://localhost |
| Mailpit | http://localhost:8025 |

Admin login: `admin@marketplace.com` / `admin`

## Stripe (Optional)

To enable payment processing locally:

1. Create a [Stripe account](https://stripe.com)
2. Set the following in `deploy/local/.env`:
   - `STRIPE_SECRET_KEY` — Stripe Dashboard → Developers → API Keys
   - `STRIPE_WEBHOOK_SIGNING_SECRET` — Stripe Dashboard → Developers → Webhooks (create a test endpoint)
   - `VITE_STRIPE_PUBLISHABLE_KEY` — Stripe Dashboard → Developers → API Keys
3. Start all services:
```bash
make -j3 dev-backend dev-frontend stripe-listen
```

The Stripe CLI forwards webhook events to `http://localhost:8000/payment/events`. Dashboard webhook configuration is only needed for production.