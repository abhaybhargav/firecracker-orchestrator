# Firecracker Orchestrator - Development Guide

## Current Implementation Status

### ✅ Completed Components

1. **Project Structure**
   - Go module with proper package organization
   - Clean separation of concerns (API, Firecracker, Database, Config)
   - Web templates and static assets structure

2. **Database Layer**
   - SQLite integration with proper models
   - VM and Container management tables
   - CRUD operations for all entities

3. **Configuration Management**
   - Environment variable based configuration
   - Sensible defaults for all settings
   - Easy customization via env vars

4. **REST API Server**
   - Gin-based HTTP server
   - Complete API endpoints for VMs and containers
   - JSON responses with proper error handling

5. **Web UI**
   - Responsive design with Tailwind CSS
   - Alpine.js for interactivity
   - Dashboard with real-time stats
   - VM management interface

6. **Firecracker Integration**
   - VM lifecycle management (create, start, stop, delete)
   - Proper networking setup with TAP devices
   - JSON-based VM configuration
   - Resource allocation and tracking

### 🚧 In Progress / Planned

1. **Container Management**
   - Docker API integration within VMs
   - Container deployment and lifecycle
   - Port mapping and networking

2. **Scheduling Logic**
   - Resource-based placement
   - Load balancing across VMs
   - Health checking

3. **Production Features**
   - Persistent storage management
   - Service discovery
   - Monitoring and metrics
   - Security enhancements

## Architecture

The system follows a layered architecture:

```
┌─────────────────────────────────────┐
│ Web UI (Tailwind + Alpine.js)      │
├─────────────────────────────────────┤
│ REST API (Gin Router)               │
├─────────────────────────────────────┤
│ Business Logic                      │
│ ├─ VM Manager                       │
│ ├─ Container Manager                │
│ └─ Scheduler                        │
├─────────────────────────────────────┤
│ Data Layer (SQLite)                 │
├─────────────────────────────────────┤
│ Firecracker VMs                     │
└─────────────────────────────────────┘
```

## Key Components

### 1. Database Models (`internal/database/`)
- **VM**: Represents Firecracker virtual machines
- **Container**: Represents Docker containers running in VMs
- Full CRUD operations with proper relationships

### 2. Firecracker Manager (`pkg/firecracker/`)
- VM lifecycle management
- Networking configuration (TAP devices, bridge)
- Resource allocation and monitoring
- Firecracker API integration

### 3. API Server (`pkg/api/`)
- RESTful endpoints for all operations
- Web UI route handling
- JSON API responses
- Error handling and logging

### 4. Configuration (`internal/config/`)
- Environment-based configuration
- Sensible defaults
- Easy customization

## API Endpoints

### Virtual Machines
- `GET /api/v1/vms` - List all VMs
- `POST /api/v1/vms` - Create new VM
- `GET /api/v1/vms/{id}` - Get VM details
- `PUT /api/v1/vms/{id}` - Update VM
- `DELETE /api/v1/vms/{id}` - Delete VM
- `POST /api/v1/vms/{id}/start` - Start VM
- `POST /api/v1/vms/{id}/stop` - Stop VM

### Containers
- `GET /api/v1/containers` - List all containers
- `POST /api/v1/containers` - Deploy container
- `GET /api/v1/containers/{id}` - Get container details
- `DELETE /api/v1/containers/{id}` - Delete container

### System
- `GET /api/v1/status` - System status
- `GET /api/v1/health` - Health check
- `GET /api/v1/stats` - System statistics

## Development Workflow

### Building
```bash
# Build for current platform
make build

# Build for Linux (useful for deployment)
make build-linux

# Development build with debug logging
make dev
```

### Testing
```bash
# Run tests
make test

# Format code
make fmt
```

### Running
```bash
# Run with default settings
make run

# Run in development mode
LOG_LEVEL=debug ./bin/orchestrator
```

## Deployment

### Local Development
1. Build the application: `make build`
2. Ensure Firecracker is installed
3. Set up networking (requires sudo)
4. Run: `./bin/orchestrator`

### Production (DigitalOcean)
1. Use the setup script: `./scripts/setup.sh`
2. Deploy with systemd service
3. Configure firewall and security

## Configuration Options

Key environment variables:

```bash
# Server
HOST=0.0.0.0
PORT=8080

# Database
DATABASE_PATH=./orchestrator.db

# Firecracker
FIRECRACKER_BINARY=/usr/bin/firecracker
KERNEL_PATH=./vm-images/vmlinux.bin
ROOTFS_PATH=./vm-images/rootfs.ext4

# Networking
BRIDGE_NAME=fc-br0
TAP_DEVICE_BASE=fc-tap

# Logging
LOG_LEVEL=info
```

## Current Limitations

1. **Container Integration**: Docker API calls within VMs not yet implemented
2. **Persistent Storage**: No volume management yet
3. **Service Discovery**: No internal service discovery
4. **Monitoring**: Basic metrics only
5. **Security**: Minimal security features

## Next Steps

1. **Complete Container Management**
   - Implement Docker API integration
   - Add container health checking
   - Implement port mapping

2. **Enhanced Scheduling**
   - Resource-aware placement
   - Anti-affinity rules
   - Auto-scaling

3. **Production Features**
   - Persistent volumes
   - Service mesh
   - Monitoring dashboard
   - Security hardening

## VM Images

You need to provide:
- `vm-images/vmlinux.bin` - Linux kernel for Firecracker
- `vm-images/rootfs.ext4` - Root filesystem with Docker pre-installed

Example rootfs should include:
- Minimal Linux distribution (Alpine/Ubuntu)
- Docker daemon
- Basic networking tools
- SSH access (optional)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Update documentation
5. Submit pull request

## Troubleshooting

### Common Issues

1. **Permission Denied (Firecracker)**
   - Ensure user is in `kvm` group
   - Check file permissions on Firecracker binary

2. **Network Issues**
   - Verify bridge interface exists
   - Check iptables rules
   - Ensure IP forwarding is enabled

3. **VM Boot Failures**
   - Verify kernel and rootfs paths
   - Check file formats and permissions
   - Review Firecracker logs

### Debugging

Enable debug logging:
```bash
LOG_LEVEL=debug ./bin/orchestrator
```

Check system status:
```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/stats
```