# OCI Always Free Deployment & Operations Guide

This guide details the end-to-end setup and production operations for the **NITC PYQ Archive** on an **Oracle Cloud Infrastructure (OCI) Always Free** Ampere A1 VM.

---

## 1. OCI Ampere A1 VM Provisioning

1. Log in to the OCI Console (ensure you are in your tenancy's **Home Region**).
2. Navigate to **Compute &rsaquo; Instances &rsaquo; Create Instance**.
3. Choose:
   - **Name**: `nitc-pyq-archive-vm`
   - **Image**: Ubuntu 24.04 LTS (or Ubuntu 22.04 LTS) **ARM64**
   - **Shape**: `VM.Standard.A1.Flex`
   - **OCPUs**: `2`
   - **Memory**: `12 GB`
   - **Boot Volume**: `50 GB`
4. Add your SSH Public Key (`~/.ssh/id_ed25519.pub`).
5. Click **Create** and note the Assigned Public IP.

---

## 2. VCN & Firewall Security Rules

1. In OCI Console, go to **Networking &rsaquo; Virtual Cloud Networks &rsaquo; [Your VCN] &rsaquo; Security Lists &rsaquo; Default Security List**.
2. Add the following **Ingress Rules** (CIDR `0.0.0.0/0`):
   - **Port 22** (TCP): SSH access (or restrict to your personal IP)
   - **Port 80** (TCP): HTTP (for Let's Encrypt ACME challenges)
   - **Port 443** (TCP): HTTPS (secure public traffic)
3. SSH into the VM:
   ```bash
   ssh -i ~/.ssh/id_ed25519 ubuntu@<PUBLIC_IP>
   ```
4. Open the host-level Ubuntu `iptables` / `ufw` firewall:
   ```bash
   sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80 -j ACCEPT
   sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
   sudo netfilter-persistent save
   ```

---

## 3. Install Docker & Docker Compose

Run on the VM:
```bash
sudo apt-get update && sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker ubuntu
```

---

## 4. DNS Configuration

Add an `A` record in your DNS provider pointing your domain or subdomain (e.g. `pyq.nitc.ac.in` or `archive.example.org`) to the OCI VM's Public IP.

---

## 5. Deployment Setup

1. Clone the repository into `/opt/archive`:
   ```bash
   sudo mkdir -p /opt/archive && sudo chown ubuntu:ubuntu /opt/archive
   git clone <REPO_URL> /opt/archive
   cd /opt/archive
   ```
2. Create production `.env`:
   ```bash
   cp .env.example .env
   ```
3. Edit `.env` with secure secrets:
   ```bash
   APP_ENV=production
   APP_ADDR=:8080
   APP_BASE_URL=https://your-domain.org
   DATABASE_URL=postgres://archive:STRONG_RANDOM_PASSWORD@postgres:5432/archive?sslmode=disable
   POSTGRES_USER=archive
   POSTGRES_PASSWORD=STRONG_RANDOM_PASSWORD
   POSTGRES_DB=archive
   SESSION_SECRET=$(openssl rand -hex 32)
   LOG_LEVEL=info
   ```

---

## 6. Production Caddy Reverse Proxy Configuration

Create `Caddyfile` in `/opt/archive/Caddyfile`:
```caddy
your-domain.org {
    reverse_proxy archive-web:8080
    encode zstd gzip

    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "SAMEORIGIN"
        Referrer-Policy "strict-origin-when-cross-origin"
    }
}
```

Add Caddy service to production compose:
```yaml
  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - archive-web
```

---

## 7. Starting the Stack & Migrations

```bash
docker compose up -d postgres
# Migrations and schema creation run automatically upon web app boot:
docker compose up -d --build archive-web
```

To seed initial branches, courses, and sample data:
```bash
docker compose run --rm archive-web /app/archive -seed
```

---

## 8. Backup & Restore Operations

### Daily PostgreSQL Backup
Automate via crontab:
```bash
crontab -e
```
Add:
```bash
0 2 * * * docker compose -f /opt/archive/compose.yaml exec -T postgres pg_dump -U archive archive | gzip > /opt/backups/pg_dump_$(date +\%F).sql.gz
```

### Restoring from Backup
```bash
gunzip < /opt/backups/pg_dump_YYYY-MM-DD.sql.gz | docker compose exec -T postgres psql -U archive -d archive
```

---

## 9. Inspecting Logs & Monitoring

- View web logs:
  ```bash
  docker compose logs -f archive-web
  ```
- View database logs:
  ```bash
  docker compose logs -f postgres
  ```
- Check application health:
  ```bash
  curl -i http://localhost:8080/healthz
  ```
