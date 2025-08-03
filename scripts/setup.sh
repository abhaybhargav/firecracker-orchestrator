#!/bin/bash

# Firecracker Orchestrator Setup Script
# This script sets up the environment for running Firecracker VMs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root directly."
        log_info "Some commands will use sudo when needed."
        exit 1
    fi
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."
    
    # Check if KVM is available
    if ! lsmod | grep -q kvm; then
        log_error "KVM is not available. Please ensure virtualization is enabled in BIOS."
        exit 1
    fi
    
    # Check if user is in kvm group
    if ! groups | grep -q kvm; then
        log_warn "User is not in kvm group. Adding user to kvm group..."
        sudo usermod -a -G kvm $USER
        log_warn "Please log out and log back in for group changes to take effect."
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    log_info "System requirements check passed."
}

# Install Firecracker
install_firecracker() {
    log_info "Installing Firecracker..."
    
    if command -v firecracker &> /dev/null; then
        log_info "Firecracker is already installed."
        return
    fi
    
    local version="1.4.1"
    local arch="x86_64"
    local url="https://github.com/firecracker-microvm/firecracker/releases/download/v${version}/firecracker-v${version}-${arch}.tgz"
    
    # Download and extract
    curl -L "$url" -o /tmp/firecracker.tgz
    tar -xzf /tmp/firecracker.tgz -C /tmp
    
    # Install binary
    sudo cp "/tmp/release-v${version}-${arch}/firecracker-v${version}-${arch}" /usr/bin/firecracker
    sudo chmod +x /usr/bin/firecracker
    
    # Cleanup
    rm -rf /tmp/firecracker.tgz "/tmp/release-v${version}-${arch}"
    
    log_info "Firecracker installed successfully."
}

# Setup networking
setup_networking() {
    log_info "Setting up networking..."
    
    # Check if bridge already exists
    if ip link show fc-br0 &> /dev/null; then
        log_info "Bridge fc-br0 already exists."
        return
    fi
    
    # Create bridge
    sudo ip link add name fc-br0 type bridge
    sudo ip addr add 192.168.100.1/24 dev fc-br0
    sudo ip link set fc-br0 up
    
    # Enable IP forwarding
    echo 'net.ipv4.ip_forward = 1' | sudo tee -a /etc/sysctl.conf
    sudo sysctl -p
    
    # Setup iptables for NAT (basic setup)
    sudo iptables -t nat -A POSTROUTING -s 192.168.100.0/24 ! -d 192.168.100.0/24 -j MASQUERADE
    
    log_info "Networking setup completed."
}

# Create directories
create_directories() {
    log_info "Creating necessary directories..."
    
    mkdir -p vm-images
    mkdir -p logs
    sudo mkdir -p /tmp/firecracker
    sudo chmod 755 /tmp/firecracker
    
    log_info "Directories created."
}

# Build the application
build_app() {
    log_info "Building the orchestrator..."
    
    go mod download
    go build -o bin/orchestrator ./cmd/orchestrator
    
    log_info "Build completed successfully."
}

# Download sample VM images (if available)
setup_vm_images() {
    log_info "Setting up VM images..."
    
    # Create placeholder files for now
    if [ ! -f vm-images/vmlinux.bin ]; then
        log_warn "Kernel image not found. You need to provide vm-images/vmlinux.bin"
        log_info "You can download it from: https://github.com/firecracker-microvm/firecracker/releases/"
    fi
    
    if [ ! -f vm-images/rootfs.ext4 ]; then
        log_warn "Root filesystem not found. You need to provide vm-images/rootfs.ext4"
        log_info "You need to create a rootfs with Docker pre-installed."
    fi
}

# Create systemd service file
create_service() {
    log_info "Creating systemd service file..."
    
    local service_content="[Unit]
Description=Firecracker Orchestrator
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$(pwd)
ExecStart=$(pwd)/bin/orchestrator
Restart=always
RestartSec=5
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=LOG_LEVEL=info

[Install]
WantedBy=multi-user.target"

    echo "$service_content" | sudo tee /etc/systemd/system/firecracker-orchestrator.service
    sudo systemctl daemon-reload
    
    log_info "Systemd service created. You can enable it with:"
    log_info "  sudo systemctl enable firecracker-orchestrator"
    log_info "  sudo systemctl start firecracker-orchestrator"
}

# Main setup function
main() {
    log_info "Starting Firecracker Orchestrator setup..."
    
    check_root
    check_requirements
    install_firecracker
    setup_networking
    create_directories
    build_app
    setup_vm_images
    create_service
    
    log_info "Setup completed successfully!"
    log_info ""
    log_info "Next steps:"
    log_info "1. Obtain VM images (kernel and rootfs) and place them in vm-images/"
    log_info "2. Run the orchestrator: ./bin/orchestrator"
    log_info "3. Open http://localhost:8080 in your browser"
    log_info ""
    log_warn "Note: You may need to log out and log back in for group changes to take effect."
}

# Run main function
main "$@"