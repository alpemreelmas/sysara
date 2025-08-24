#!/bin/bash

# Sysara System Management Platform Installer
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check root
if [[ $EUID -ne 0 ]]; then
    print_error "Run as root: sudo $0"
    exit 1
fi

print_status "Installing Sysara System Management Platform..."

# Install dependencies
print_status "Installing system dependencies..."
if command -v apt &> /dev/null; then
    apt update && apt install -y curl wget git build-essential sqlite3
elif command -v yum &> /dev/null; then
    yum install -y curl wget git gcc make sqlite
elif command -v dnf &> /dev/null; then
    dnf install -y curl wget git gcc make sqlite
fi

# Install Go if not present
if ! command -v go &> /dev/null; then
    print_status "Installing Go..."
    GO_VERSION="1.21.0"
    ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
    wget -q "https://golang.org/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin
    rm "go${GO_VERSION}.linux-${ARCH}.tar.gz"
fi

# Create sysara user
print_status "Creating sysara user..."
useradd -r -s /bin/false -d /opt/sysara sysara || true

# Create directories
print_status "Creating directories..."
mkdir -p /opt/sysara/{data,logs,static,templates}
mkdir -p /etc/sysara

# Build application (assuming source is in current directory)
if [[ -f "main.go" ]]; then
    print_status "Building Sysara..."
    /usr/local/go/bin/go mod tidy
    /usr/local/go/bin/go build -o /opt/sysara/sysara .
    
    # Copy static files and templates
    cp -r static/* /opt/sysara/static/ 2>/dev/null || true
    cp -r templates/* /opt/sysara/templates/ 2>/dev/null || true
fi

# Set permissions
chown -R sysara:sysara /opt/sysara
chmod +x /opt/sysara/sysara

# Create systemd service
print_status "Creating systemd service..."
cat > /etc/systemd/system/sysara.service << 'EOF'
[Unit]
Description=Sysara System Management Platform
After=network.target

[Service]
Type=simple
User=sysara
Group=sysara
WorkingDirectory=/opt/sysara
ExecStart=/opt/sysara/sysara
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
systemctl daemon-reload
systemctl enable sysara
systemctl start sysara

print_success "Sysara installation completed!"
print_status "Service status: $(systemctl is-active sysara)"
print_status "Access Sysara at: http://localhost:8080"
print_status "Default login will be created on first run"

# Show service logs
print_status "Recent logs:"
journalctl -u sysara --no-pager -n 5