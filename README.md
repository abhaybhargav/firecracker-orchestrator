# Firecracker Orchestrator

A minimal orchestration system built in Go that deploys Docker containers inside Firecracker virtual machines on DigitalOcean.

## Features

- 🔥 **Firecracker Integration**: Lightweight virtual machines with fast boot times (~150ms)
- 🐳 **Container Support**: Run Docker containers inside VMs
- 🎯 **Simple Orchestration**: Basic scheduling and resource management
- 🌐 **Web UI**: Clean, responsive interface built with Tailwind CSS
- 📊 **Monitoring**: Real-time status and metrics
- 🔌 **REST API**: Full API access for automation

## Architecture

```
┌─────────────────────────────────────┐
│           Web UI / REST API         │
├─────────────────────────────────────┤
│        Orchestration Engine         │
├─────────────────────────────────────┤
│       Firecracker Manager           │
├─────────────────────────────────────┤
│    Firecracker VMs (with Docker)    │
├─────────────────────────────────────┤
│        Host Linux (DigitalOcean)     │
└─────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- Linux host (Ubuntu 22.04+ recommended)
- KVM support
- Root or sudo access for networking

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/abhaybhargav/firecracker-orchestrator
   cd firecracker-orchestrator
   ```

2. **Build the application**
   ```bash
   go mod download
   
   # For development (with CGO)
   make build
   
   # For deployment (pure Go, no CGO required)
   make build-linux-static
   ```

3. **Install Firecracker**
   ```bash
   # Download Firecracker binary
   curl -L https://github.com/firecracker-microvm/firecracker/releases/download/v1.4.1/firecracker-v1.4.1-x86_64.tgz -o firecracker.tgz
   tar -xzf firecracker.tgz
   sudo cp release-v1.4.1-x86_64/firecracker-v1.4.1-x86_64 /usr/bin/firecracker
   sudo chmod +x /usr/bin/firecracker
   ```

4. **Set up networking** (requires root)
   ```bash
   # Create bridge interface
   sudo ip link add name fc-br0 type bridge
   sudo ip addr add 192.168.100.1/24 dev fc-br0
   sudo ip link set fc-br0 up
   ```

5. **Create VM images**
   ```bash
   mkdir -p vm-images
   # You'll need to create or download kernel and rootfs images
   # See the VM Images section below
   ```

6. **Run the orchestrator**
   ```bash
   ./bin/orchestrator
   ```

7. **Access the web UI**
   Open http://localhost:8080 in your browser

## Configuration

The application can be configured via environment variables:

```bash
# Server configuration
HOST=0.0.0.0
PORT=8080

# Database
DATABASE_PATH=./orchestrator.db

# Firecracker
FIRECRACKER_BINARY=/usr/bin/firecracker
KERNEL_PATH=./vm-images/vmlinux.bin
ROOTFS_PATH=./vm-images/rootfs.ext4
SOCKET_DIR=/tmp/firecracker

# Networking
BRIDGE_NAME=fc-br0
TAP_DEVICE_BASE=fc-tap

# VM defaults
DEFAULT_MEMORY_MB=512
DEFAULT_CPUS=1
DEFAULT_DISK_GB=2

# Logging
LOG_LEVEL=info
```

## VM Images

You need Linux kernel and rootfs images to run Firecracker VMs. Here are two options:

### Option 1: Use Pre-built Images

Download pre-built images (when available):
```bash
# Download kernel
curl -L https://github.com/firecracker-microvm/firecracker/releases/download/v1.4.1/vmlinux.bin -o vm-images/vmlinux.bin

# Download rootfs (you'll need to find or build one with Docker pre-installed)
```

### Option 2: Build Your Own

Create a minimal Ubuntu rootfs with Docker:
```bash
# This is a simplified example - you'll need a proper build process
sudo debootstrap focal vm-images/rootfs
sudo chroot vm-images/rootfs /bin/bash
# Install Docker and other dependencies inside the chroot
# Create an ext4 filesystem from the rootfs directory
```

## API Reference

### Virtual Machines

- `GET /api/v1/vms` - List all VMs
- `POST /api/v1/vms` - Create a new VM
- `GET /api/v1/vms/{id}` - Get VM details
- `PUT /api/v1/vms/{id}` - Update VM
- `DELETE /api/v1/vms/{id}` - Delete VM
- `POST /api/v1/vms/{id}/start` - Start VM
- `POST /api/v1/vms/{id}/stop` - Stop VM

### Containers

- `GET /api/v1/containers` - List all containers
- `POST /api/v1/containers` - Deploy a new container
- `GET /api/v1/containers/{id}` - Get container details
- `DELETE /api/v1/containers/{id}` - Delete container

### System

- `GET /api/v1/status` - System status
- `GET /api/v1/health` - Health check
- `GET /api/v1/stats` - System statistics

## Example Usage

### Create a VM via API

```bash
curl -X POST http://localhost:8080/api/v1/vms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server",
    "memory": 1024,
    "cpus": 2,
    "disk_size": 5
  }'
```

### Deploy a Container

```bash
curl -X POST http://localhost:8080/api/v1/containers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx",
    "image": "nginx:latest",
    "vm_id": "vm-id-here",
    "ports": {"80": "8080"},
    "environment": {"ENV": "production"}
  }'
```

## Production Deployment

### DigitalOcean Setup

1. **Create a droplet**
   ```bash
   # Recommended: 8GB RAM, 4 CPUs, SSD storage
   # Ubuntu 22.04 LTS
   ```

2. **Configure the host**
   ```bash
   # Enable KVM
   sudo modprobe kvm_intel  # or kvm_amd
   sudo usermod -a -G kvm $USER
   
   # Set up networking
   sudo apt install bridge-utils
   # Configure bridge as shown above
   ```

3. **Deploy with systemd**
   ```bash
   sudo cp deployments/systemd/firecracker-orchestrator.service /etc/systemd/system/
   sudo systemctl enable firecracker-orchestrator
   sudo systemctl start firecracker-orchestrator
   ```

## Development

### Project Structure

```
firecracker-orchestrator/
├── cmd/orchestrator/          # Main application
├── pkg/
│   ├── api/                   # REST API handlers
│   ├── firecracker/           # VM management
│   ├── container/             # Container management
│   └── scheduler/             # Scheduling logic
├── internal/
│   ├── config/                # Configuration
│   └── database/              # Database models
├── web/
│   ├── templates/             # HTML templates
│   └── static/                # Static assets
└── deployments/               # Deployment configs
```

### Building

```bash
# Build for current platform
go build -o bin/orchestrator ./cmd/orchestrator

# Build for Linux (if developing on macOS)
GOOS=linux GOARCH=amd64 go build -o bin/orchestrator-linux ./cmd/orchestrator
```

### Testing

```bash
go test ./...
```

## Limitations & Future Improvements

Current limitations:
- Basic scheduling (round-robin)
- No container runtime integration yet (Docker API calls need implementation)
- Simple networking setup
- No persistent storage management
- Limited monitoring and logging

Planned improvements:
- Advanced scheduling algorithms
- Container runtime integration
- Service discovery
- Load balancing
- Persistent volumes
- Metrics and alerting
- Horizontal scaling
- Security enhancements

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- Open an issue on GitHub
- Check the documentation
- Review the API reference