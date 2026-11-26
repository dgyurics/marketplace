# Production Deployment

Marketplace runs behind your existing HTTPS reverse proxy. The included Nginx serves the app on a private HTTP port; it does not need a domain or TLS certificate. Your proxy owns the public domain, certificates, HTTPS redirects and certificate renewal.

## Requirements

- Docker Engine and Docker Compose on a Linux host with cgroups enabled
- Minimum 2 vCPU and 6 GB RAM; 4 vCPU and 8 GB RAM recommended
- A reverse proxy that terminates HTTPS and forwards to the host's loopback interface

## Set up

Run these commands from the project root:

```sh
cp deploy/prod/example.env deploy/prod/.env
openssl genpkey -algorithm RSA -out deploy/prod/private.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in deploy/prod/private.pem -out deploy/prod/public.pem
```

Edit `deploy/prod/.env` before starting:

- Set `BASE_URL` to the public HTTPS origin, for example `https://shop.example.com`. The application uses it for links; it is not Nginx configuration. Leave `VITE_API_URL=/api` for same-origin requests.
- Set a unique strong `POSTGRES_PASSWORD` and generate `HMAC_SECRET`, `IMGPROXY_KEY` and `IMGPROXY_SALT` using `openssl rand -hex 32` (run it separately for each value). Keep the RSA keys and these secrets safe and stable across restarts.
- Configure email and payment settings for your installation. For pay on delivery, set `PAYMENT_ON_DELIVERY_ENABLED=true` and `STRIPE_ENABLED=false`; for Stripe, set the server and frontend keys and configure a webhook at `https://YOUR-DOMAIN/api/payment/events`.
- Set `MARKETPLACE_PORT` if port 8080 is in use. It is published on `127.0.0.1` only. Set `MARKETPLACE_TRUSTED_PROXY` as described below.

Start the app:

```sh
docker compose -f deploy/prod/docker-compose.yaml up -d --build
```

The initial admin login is `admin@marketplace.com` / `admin`. Change its password immediately, before exposing the site. The database and images live in Docker volumes; back them up together with your configuration and keys. Do not remove the volumes when upgrading.

## Connect a reverse proxy

Point your HTTPS virtual host to `http://127.0.0.1:8080`, adjusting the port to match `MARKETPLACE_PORT`. For example, in an existing host Nginx HTTPS `server` block:

```nginx
client_max_body_size 30m;
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

The proxy must overwrite `X-Real-IP` with the actual client address and `X-Forwarded-Proto` with the original scheme; do not pass arbitrary incoming values. Configure `MARKETPLACE_TRUSTED_PROXY` to the **source IP or CIDR of that proxy as seen by the Marketplace Nginx container**, and trust only that proxy. For a host proxy connecting through Docker's published port this is often the Docker bridge gateway, not `127.0.0.1`. Inspect the gateway after starting with:

```sh
docker network inspect marketplace_edge --format '{{(index .IPAM.Config 0).Gateway}}'
```

Verify it is the address actually used by your proxy, put it in `.env`, and run `docker compose -f deploy/prod/docker-compose.yaml up -d` again. Until the trusted address is set, client IPs may all appear as the proxy, sharing rate limits. The default trusts only loopback. Do not trust a broad network containing untrusted clients. For a containerized reverse proxy, either reach the loopback port via the host or explicitly attach it and Marketplace Nginx to a private shared Docker network, use unique service aliases, and trust only the proxy's source address/network. A container's `127.0.0.1` refers to itself.

Use a separate checkout/configuration and `docker compose -p <unique-name> -f deploy/prod/docker-compose.yaml up -d --build` for each instance, with a distinct port, `APP_ID`, secrets and public `BASE_URL`. Compose project names isolate the networks and volumes. Check each instance's actual network name when inspecting its gateway.

## Existing installations

This replaces the old app-managed TLS endpoint. Keep your existing volumes, `.env`, keys and certificates; move certificate management and domain routing to the public proxy first. Set the new port and trusted proxy address, test via the private port, and then switch public traffic. The old SSL scripts are retained for existing installations but are not required for this setup. Test login, images, uploads, payment webhooks and client-IP rate limiting before retiring the old proxy.

## Resource tuning

Container limits are in [docker-compose.yaml](docker-compose.yaml). If you change PostgreSQL resources, review `shared_buffers`, `work_mem`, `maintenance_work_mem` and `max_connections` in [postgresql.conf](postgresql.conf).
