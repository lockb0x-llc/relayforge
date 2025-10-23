# RelayForge

RelayForge is a decentralized infrastructure orchestration platform with a cloud control plane, federated runners, CLI, UI, and workflow templates.

## 🧭 Strategic Use Cases (Executive Summary)

### 1. **Cross-Platform Automation**
RelayForge lets your teams automate complex tasks across different systems—cloud, on-premise, or hybrid—without rewriting code or managing fragile integrations. It’s a universal conductor for infrastructure operations.

> Reduces manual overhead, accelerates deployment timelines, and ensures consistency across environments.

---

### 2. **Secure, Scalable Workflow Execution**
Instead of relying on centralized servers, RelayForge uses distributed “runners” that execute tasks wherever they’re needed. These runners are authenticated and monitored in real time.

>  Enables secure execution at scale, even across fragmented infrastructure or partner networks.

---

### 3. **Multi-Cloud Deployment Control**
RelayForge supports AWS, GCP, Azure, and more—allowing your teams to deploy applications, virtual machines, or containers across providers from a single interface.

>  Avoids cloud lock-in, supports cost optimization, and simplifies multi-cloud strategy.

---

### 4. **Real-Time Visibility and Auditability**
Every workflow run is tracked with live logs and historical records. You can see what happened, when, and where—whether triggered by a developer, a system, or an external API.

>  Improves compliance, supports forensic analysis, and strengthens operational transparency.

---

### 5. **Rapid Prototyping and Internal Platform Engineering**
RelayForge can be used to build internal tools, automate provisioning, or create modular workflows for new initiatives—without waiting on external vendors or long dev cycles.

>  Empowers teams to move fast, test ideas, and build operational leverage in-house.

---

### 6. **Integration with Existing Systems**
It plugs into GitHub, Docker, PostgreSQL, and other common tools—so your teams don’t need to abandon existing workflows. RelayForge becomes the orchestration layer on top.

>  Enhances existing investments, avoids disruption, and supports gradual adoption.


## Features

- 🚀 **Go API Server** - RESTful API with PostgreSQL backend
- 🏃 **Federated Runners** - Distributed workflow execution
- 🖥️ **Next.js Web UI** - Modern React-based interface
- 🛠️ **CLI Tool** - Command-line interface for workflow management
- 🔐 **GitHub OAuth** - Secure authentication
- 📝 **YAML Workflows** - Infrastructure as Code
- 🐚 **Shell Execution** - Run shell commands and scripts
- 📊 **Real-time Logs** - Live workflow execution monitoring
- 🐳 **Docker Support** - Containerized deployment
- ☁️ **Multi-cloud** - Support for AWS, GCP, Azure, and more

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### 1. Clone and Setup

```bash
git clone https://github.com/lockb0x-llc/relayforge.git
cd relayforge

# Copy environment variables
cp .env.example .env
# Edit .env with your GitHub OAuth credentials
```

### 2. GitHub OAuth Setup

1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Create a new OAuth App with:
   - Homepage URL: `http://localhost:3000`
   - Authorization callback URL: `http://localhost:8080/api/auth/callback`
3. Copy Client ID and Client Secret to `.env`

### 3. Start with Docker Compose (Recommended)

```bash
# Start all services
make docker-up

# Or manually
docker-compose up -d
```

### 4. Start Development Environment

```bash
# Install dependencies
make install

# Start development servers
make dev
```

### 5. Access the Application

- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Database**: localhost:5432

## Usage

### Web Interface

1. Open http://localhost:3000
2. Click "Login with GitHub"
3. Create your first workflow
4. Run and monitor workflows

### CLI Usage

```bash
# Build CLI
make build-cli

# Login
./bin/relayforge auth login

# Set token (after GitHub login)
./bin/relayforge auth set-token <your-token>

# List workflows
./bin/relayforge workflow list

# Create workflow
./bin/relayforge workflow create "My Workflow" examples/hello-world.yml

# Start workflow run
./bin/relayforge run start <workflow-id>

# List runs
./bin/relayforge run list <workflow-id>
```

### API Endpoints

#### Authentication
- `GET /api/auth/github` - GitHub OAuth login
- `GET /api/auth/callback` - OAuth callback
- `GET /api/auth/user` - Get current user

#### Workflows
- `GET /api/workflows` - List workflows
- `POST /api/workflows` - Create workflow
- `GET /api/workflows/:id` - Get workflow
- `PUT /api/workflows/:id` - Update workflow
- `DELETE /api/workflows/:id` - Delete workflow

#### Runs
- `GET /api/workflows/:id/runs` - List workflow runs
- `POST /api/workflows/:id/runs` - Start new run
- `GET /api/runs/:id` - Get run details
- `POST /api/runs/:id/cancel` - Cancel run

#### Runners
- `GET /api/runners` - List runners
- `POST /api/runners/register` - Register runner

#### WebSockets
- `WS /ws/logs/:runId` - Real-time log streaming

## Workflow YAML Format

```yaml
name: My Workflow
description: Description of what this workflow does

jobs:
  job1:
    runs-on: any  # or specific runner tags
    steps:
      - name: Step name
        run: |
          echo "Hello World"
          # Multi-line shell commands

  job2:
    runs-on: docker
    needs: [job1]  # Run after job1 completes
    steps:
      - name: Docker build
        run: docker build -t my-app .
      
      - name: Deploy
        run: docker run -d my-app
```

## Example Workflows

### Hello World
```yaml
name: Hello World
jobs:
  hello:
    runs-on: any
    steps:
      - name: Greet
        run: echo "Hello from RelayForge!"
```

### Multi-Cloud VM Deployment
See `examples/multi-cloud-vm.yml` for a comprehensive example of deploying VMs across AWS and GCP.

### Docker Application
See `examples/docker-deploy.yml` for containerized application deployment.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web UI        │    │   API Server    │    │   Database      │
│   (Next.js)     │◄──►│   (Go/Gin)      │◄──►│   (PostgreSQL)  │
│   Port 3000     │    │   Port 8080     │    │   Port 5432     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                │ HTTP/WebSocket
                                ▼
                       ┌─────────────────┐
                       │   Runners       │
                       │   (Federated)   │
                       │   Any Platform  │
                       └─────────────────┘
```

### Components

1. **API Server** (`cmd/api/`) - Central control plane
2. **Runner** (`cmd/runner/`) - Federated execution engine
3. **CLI** (`cmd/cli/`) - Command-line interface
4. **Web UI** (`web/`) - React/Next.js frontend
5. **Database** - PostgreSQL with GORM

## Development

### Project Structure

```
relayforge/
├── cmd/                    # Main applications
│   ├── api/               # API server
│   ├── runner/            # Workflow runner
│   └── cli/               # CLI tool
├── internal/              # Private Go packages
│   ├── api/              # API server logic
│   ├── auth/             # Authentication
│   ├── models/           # Data models
│   └── workflow/         # Workflow engine
├── pkg/                   # Public Go packages
│   └── types/            # Shared types
├── web/                   # Next.js frontend
├── migrations/            # Database migrations
├── examples/              # Sample workflows
├── docker-compose.yml     # Development setup
└── Makefile              # Build automation
```

### Make Commands

```bash
make help           # Show all available commands
make install        # Install dependencies
make dev            # Start development environment
make build          # Build all binaries
make test           # Run tests
make docker-up      # Start with Docker Compose
make docker-down    # Stop Docker services
make clean          # Clean build artifacts
```

### Database Schema

- **users** - GitHub OAuth user accounts
- **workflows** - YAML workflow definitions
- **runs** - Workflow executions
- **jobs** - Individual jobs within runs
- **steps** - Steps within jobs
- **logs** - Execution logs
- **runners** - Registered runner instances

## Deployment

### Production Docker Compose

```bash
# Set production environment variables
export GITHUB_CLIENT_ID=your_client_id
export GITHUB_CLIENT_SECRET=your_client_secret
export JWT_SECRET=your_secure_jwt_secret

# Start production stack
docker-compose -f docker-compose.yml up -d
```

### Kubernetes

Kubernetes manifests can be generated from the Docker Compose file or created manually for production deployments.

### Runner Deployment

Deploy runners on any infrastructure:

```bash
# Build runner
go build -o runner cmd/runner/main.go

# Configure environment
export API_URL=https://your-relayforge-api.com
export RUNNER_NAME=prod-runner-1
export RUNNER_TAGS=linux,aws,production

# Start runner
./runner
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | `localhost` |
| `DB_USER` | Database user | `relayforge` |
| `DB_PASSWORD` | Database password | `password` |
| `DB_NAME` | Database name | `relayforge` |
| `DB_PORT` | Database port | `5432` |
| `PORT` | API server port | `8080` |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | Required |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | Required |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `RUNNER_NAME` | Runner instance name | `relayforge-runner` |
| `RUNNER_TAGS` | Runner capability tags | `linux,shell` |
| `API_URL` | API server URL for runners | `http://localhost:8080` |

## Security

- GitHub OAuth for authentication
- JWT tokens for API access
- CORS protection
- SQL injection prevention with GORM
- Input validation and sanitization

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make test` and `make lint`
6. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- 📖 Documentation: [GitHub Wiki](https://github.com/lockb0x-llc/relayforge/wiki)
- 🐛 Issues: [GitHub Issues](https://github.com/lockb0x-llc/relayforge/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/lockb0x-llc/relayforge/discussions)
