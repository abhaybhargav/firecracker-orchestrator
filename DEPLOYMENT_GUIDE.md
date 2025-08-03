# Deployment Guide - Firecracker Orchestrator

## SQLite CGO Issue Resolution

If you encounter the error:
```
Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

We provide **two solutions**:

## Solution 1: Use Pure Go SQLite (Recommended for Deployment)

This is the **easiest solution** for deployment as it doesn't require CGO and creates fully static binaries.

### Build and Run
```bash
# Build static binary (no CGO required)
make build-linux-static

# Run with pure Go SQLite driver (default)
./bin/orchestrator-linux-static
```

The application will automatically use the pure Go SQLite driver which doesn't require CGO.

### Environment Variables
```bash
# Explicitly set pure Go driver (default)
export DATABASE_DRIVER=sqlite

# Or use CGO-based driver for better performance
export DATABASE_DRIVER=sqlite3
```

## Solution 2: Enable CGO for SQLite3

If you want better database performance, use the CGO-based SQLite driver:

### Build with CGO
```bash
# For local development
make build

# For Linux deployment with CGO
make build-linux

# Run with CGO-based driver
DATABASE_DRIVER=sqlite3 ./bin/orchestrator-linux
```

### Install Build Dependencies (Linux)
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install gcc libc6-dev

# CentOS/RHEL
sudo yum install gcc glibc-devel

# Alpine
sudo apk add gcc musl-dev
```

## Comparison

| Feature | Pure Go SQLite | CGO SQLite |
|---------|---------------|------------|
| **Deployment** | ✅ Single static binary | ❌ Requires C compiler |
| **Cross-compilation** | ✅ Easy | ❌ Complex |
| **Performance** | ⚠️ Slower | ✅ Faster |
| **Dependencies** | ✅ None | ❌ Requires gcc |
| **Binary size** | ⚠️ Larger (~25MB) | ✅ Smaller (~20MB) |

## Recommended Deployment Strategy

### For Production (DigitalOcean)
```bash
# Use pure Go for simplicity
make build-linux-static
scp bin/orchestrator-linux-static your-server:/opt/firecracker-orchestrator/
```

### For High Performance
```bash
# Build on the target server with CGO
git clone <repo>
make build
DATABASE_DRIVER=sqlite3 ./bin/orchestrator
```

## Complete Deployment Example

### 1. Build Static Binary
```bash
# On your development machine
make build-linux-static
```

### 2. Deploy to Server
```bash
# Copy to server
scp bin/orchestrator-linux-static user@your-server:/tmp/

# On the server
sudo mkdir -p /opt/firecracker-orchestrator
sudo mv /tmp/orchestrator-linux-static /opt/firecracker-orchestrator/orchestrator
sudo chmod +x /opt/firecracker-orchestrator/orchestrator
```

### 3. Create Service
```bash
# Copy service file
sudo cp deployments/systemd/firecracker-orchestrator.service /etc/systemd/system/

# Update the service file to use the static binary
sudo systemctl daemon-reload
sudo systemctl enable firecracker-orchestrator
sudo systemctl start firecracker-orchestrator
```

### 4. Verify
```bash
# Check status
sudo systemctl status firecracker-orchestrator

# View logs
sudo journalctl -u firecracker-orchestrator -f

# Test API
curl http://localhost:8080/api/v1/health
```

## Environment Variables

```bash
# Database configuration
DATABASE_DRIVER=sqlite        # Use pure Go SQLite (default)
DATABASE_PATH=/var/lib/firecracker-orchestrator/orchestrator.db

# Server configuration
HOST=0.0.0.0
PORT=8080

# Firecracker paths
FIRECRACKER_BINARY=/usr/bin/firecracker
KERNEL_PATH=/opt/firecracker-orchestrator/vm-images/vmlinux.bin
ROOTFS_PATH=/opt/firecracker-orchestrator/vm-images/rootfs.ext4

# Logging
LOG_LEVEL=info
```

## Troubleshooting

### Issue: CGO compilation errors
**Solution**: Use `make build-linux-static` for pure Go build

### Issue: "driver sqlite not found" 
**Solution**: Ensure you're using the static binary built with pure Go SQLite

### Issue: Database performance is slow
**Solution**: Switch to CGO version with `DATABASE_DRIVER=sqlite3`

### Issue: Binary won't run on Alpine Linux
**Solution**: Use the static build (`build-linux-static`) which is compatible with musl libc

## Performance Notes

- **Pure Go SQLite**: ~10-20% slower than CGO version
- **CGO SQLite**: Faster but requires compilation environment
- For most use cases, the pure Go version provides adequate performance
- Use CGO version only if you need maximum database performance

The pure Go version is recommended for most deployments due to its simplicity and portability.