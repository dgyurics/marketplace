# Production Deployment

> This document will contain the full production deployment instructions.
> The deployment runbook is not finalized yet.

## Host Requirements

This production setup targets a **Linux host with cgroups enabled**.

- Minimum: **2 vCPU / 6 GB RAM**
- Recommended: **4 vCPU / 8 GB RAM**

## Resource Tuning Notes

Container resource settings are defined in [docker-compose.yaml](docker-compose.yaml).

When you change resource settings (especially memory/CPU for `postgres`), update PostgreSQL tuning in [postgresql.conf](postgresql.conf) as well.

At minimum, review and retune:

- `shared_buffers`
- `work_mem`
- `maintenance_work_mem`
- `max_connections`