# ProxGuard 🛡️

Bot Discord pour gérer les machines virtuelles Proxmox via des commandes slash.

## Fonctionnalités

- ✅ Démarrer/Arrêter des VMs Proxmox depuis Discord
- 📊 Afficher le statut des VMs (CPU, RAM, uptime)
- 🔍 Autocomplete avec noms de VMs
- 🔒 Whitelist de VMs autorisées
- 🐳 Conteneurisé avec Docker
- 🚀 CI/CD avec GitHub Actions

## Prérequis

- Proxmox VE 7.0+
- API Token Proxmox
- Discord Bot Token
- Docker & Docker Compose (pour le déploiement)

## Installation

### 1. Configuration Proxmox

Créer un API Token dans Proxmox :

1. Datacenter > Permissions > API Tokens
2. Créer un token avec les permissions `VM.PowerMgmt`
3. Copier le Token ID et le Secret

### 2. Configuration Discord

1. Créer une application sur https://discord.com/developers/applications
2. Créer un bot et copier le token
3. Activer les intents nécessaires
4. Inviter le bot avec les scopes `bot` et `applications.commands`

### 3. Déploiement avec Docker

```
# Cloner le repo
git clone https://github.com/ton-username/proxguard.git
cd proxguard

# Copier et configurer les variables d'environnement
cp .env.example .env
nano .env

# Lancer avec Docker Compose
docker-compose up -d
```

### 4. Variables d'environnement

```
DISCORD_TOKEN=votre_token_discord
DISCORD_GUILD_ID=votre_server_id
PROXMOX_URL=https://proxmox.local:8006/api2/json
PROXMOX_TOKEN_ID=root@pam!discord-bot
PROXMOX_SECRET=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
PROXMOX_NODE=pve
ALLOWED_VMS=100,101,102
```

## Utilisation

Commandes disponibles :

- `/vm start <vm>` - Démarrer une VM
- `/vm stop <vm>` - Arrêter une VM (graceful)
- `/vm status <vm>` - Afficher le statut
- `/vm list` - Lister les VMs autorisées

## Développement

```
# Installer les dépendances
go mod download

# Lancer en mode dev
go run cmd/bot/main.go

# Build
go build -o proxguard ./cmd/bot

# Tests
go test ./...
```

## Structure du projet

```
proxguard/
├── cmd/bot/          # Point d'entrée
├── internal/
│   ├── bot/          # Logique Discord
│   ├── proxmox/      # Client Proxmox
│   └── config/       # Configuration
├── Dockerfile
├── docker-compose.yml
└── .github/workflows/
```

## Licence

MIT
