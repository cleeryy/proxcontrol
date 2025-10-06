# proxcontrol 🛡️

Discord bot for managing Proxmox Virtual Machines via slash commands.

[![Docker Build](https://img.shields.io/github/actions/workflow/status/cleeryy/proxcontrol/docker-publish.yml?branch=main)](https://github.com/cleeryy/proxcontrol/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/cleeryy/proxcontrol)](go.mod)

## Features

- ✅ Start/Stop Proxmox VMs from Discord
- 📊 Display VM status (CPU, RAM, uptime)
- 🔍 VM name autocomplete
- 🔒 VM whitelist for security
- 🐳 Docker containerized
- 🚀 CI/CD with GitHub Actions

## Prerequisites

- Proxmox VE 7.0+
- Proxmox API Token
- Discord Bot Token
- Docker & Docker Compose (for deployment)

## Quick Start

### 1. Proxmox Configuration

Create an API Token in Proxmox:

1. Navigate to Datacenter > Permissions > API Tokens
2. Create a token with `VM.PowerMgmt` permissions
3. Copy the Token ID and Secret

### 2. Discord Bot Setup

1. Create an application at https://discord.com/developers/applications
2. Create a bot and copy the token
3. Enable required intents (Message Content, Server Members)
4. Invite the bot with `bot` and `applications.commands` scopes

### 3. Deployment with Docker

```
# Clone the repository
git clone https://github.com/cleeryy/proxcontrol.git
cd proxcontrol

# Configure environment variables
cp .env.example .env
vim .env

# Start with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

### 4. Environment Variables

```
DISCORD_TOKEN=your_discord_token
DISCORD_GUILD_ID=your_server_id
PROXMOX_URL=https://proxmox.local:8006/api2/json
PROXMOX_TOKEN_ID=root@pam!discord-bot
PROXMOX_SECRET=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
PROXMOX_NODE=pve
ALLOWED_VMS=100,101,102
```

## Usage

Available commands:

- `/vm start <vm>` - Start a virtual machine
- `/vm stop <vm>` - Stop a virtual machine (graceful shutdown)
- `/vm status <vm>` - Display VM status and metrics
- `/vm list` - List all authorized VMs

The `<vm>` parameter supports both VM names and IDs with autocomplete suggestions.

## Development

```
# Install dependencies
go mod download

# Run in development mode
go run cmd/bot/main.go

# Build binary
go build -o proxcontrol ./cmd/bot

# Run tests
go test ./...

# Run linter
go vet ./...
```

## Project Structure

```
proxcontrol/
├── cmd/
│   └── bot/              # Application entry point
├── internal/
│   ├── proxmox/          # Proxmox API client
│   └── config/           # Configuration management
├── Dockerfile            # Multi-stage Docker build
├── docker-compose.yml    # Docker Compose configuration
└── .github/
    └── workflows/        # CI/CD pipelines
```

## Docker Build

```
# Build image
docker build -t proxcontrol:latest .

# Run container
docker run -d --name proxcontrol --env-file .env proxcontrol:latest

# Stop container
docker stop proxcontrol
```

## Security

- API tokens should never be committed to version control
- Use environment variables or secrets management
- Restrict VM whitelist to only necessary VMs
- Consider implementing role-based access control via Discord roles

## Troubleshooting

### Bot doesn't respond to commands

- Verify the bot has `applications.commands` scope
- Check if the bot has proper permissions in the Discord server
- Review bot logs: `docker-compose logs -f`

### Proxmox connection errors

- Ensure the Proxmox API is accessible from the bot's network
- Verify API token has correct permissions
- Check if VM IDs in `ALLOWED_VMS` exist

### Autocomplete not working

- Restart the bot after updating `ALLOWED_VMS`
- Verify the bot can connect to Proxmox API
- Check that VMs are in a valid state

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [DiscordGo](https://github.com/bwmarrin/discordgo)
- Proxmox VE API documentation
- Docker multi-stage builds for optimized images

## Roadmap

- [ ] Role-based permissions
- [ ] Activity logging and audit trail
- [ ] VM reboot command
- [ ] Multi-node Proxmox support
- [ ] Scheduled VM operations
- [ ] Resource usage alerts

## Support

For issues, questions, or suggestions, please open an issue on GitHub.
