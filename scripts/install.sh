#!/bin/bash

# Cytology Viewer Pro - Installation Script
# For Ubuntu 22.04 with CUDA support

set -e

echo "======================================"
echo "Cytology Viewer Pro - Installation"
echo "======================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root (use sudo)"
  exit 1
fi

# Check for NVIDIA GPU
if ! command -v nvidia-smi &> /dev/null; then
    echo "ERROR: nvidia-smi not found. Please install NVIDIA drivers first."
    exit 1
fi

echo "✓ NVIDIA GPU detected"
nvidia-smi --query-gpu=name --format=csv,noheader

# Install dependencies
echo ""
echo "Installing system dependencies..."
apt-get update
apt-get install -y \
    wget \
    git \
    build-essential \
    libwebp-dev \
    cuda-toolkit-12-3

# Install Go
if ! command -v go &> /dev/null; then
    echo ""
    echo "Installing Go 1.21..."
    wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    rm go1.21.6.linux-amd64.tar.gz
    
    # Add Go to PATH for current session
    export PATH=$PATH:/usr/local/go/bin
    
    # Add Go to PATH permanently
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    
    echo "✓ Go installed"
else
    echo "✓ Go already installed"
fi

# Create application user
if ! id "cytoviewer" &>/dev/null; then
    echo ""
    echo "Creating cytoviewer user..."
    useradd -r -s /bin/false cytoviewer
    echo "✓ User created"
fi

# Create directories
echo ""
echo "Creating directories..."
mkdir -p /opt/cyto-viewer
mkdir -p /data/slides
mkdir -p /data/temp
mkdir -p /etc/cyto-viewer
mkdir -p /var/log/cyto-viewer

chown -R cytoviewer:cytoviewer /opt/cyto-viewer
chown -R cytoviewer:cytoviewer /data/slides
chown -R cytoviewer:cytoviewer /data/temp
chown -R cytoviewer:cytoviewer /var/log/cyto-viewer

echo "✓ Directories created"

# Build application
echo ""
echo "Building application..."
cd "$(dirname "$0")/.."
make build

# Install binary
echo ""
echo "Installing binary..."
cp bin/cyto-viewer /usr/local/bin/
chmod +x /usr/local/bin/cyto-viewer
echo "✓ Binary installed"

# Install web files
echo ""
echo "Installing web files..."
cp -r web /opt/cyto-viewer/
chown -R cytoviewer:cytoviewer /opt/cyto-viewer/web
echo "✓ Web files installed"

# Install configuration
if [ ! -f /etc/cyto-viewer/config.env ]; then
    echo ""
    echo "Installing configuration..."
    cp config.env.example /etc/cyto-viewer/config.env
    
    # Generate JWT secret
    JWT_SECRET=$(openssl rand -base64 32)
    sed -i "s/CHANGE_THIS_TO_A_RANDOM_SECRET/$JWT_SECRET/" /etc/cyto-viewer/config.env
    
    chmod 600 /etc/cyto-viewer/config.env
    chown cytoviewer:cytoviewer /etc/cyto-viewer/config.env
    echo "✓ Configuration installed"
    echo ""
    echo "⚠️  IMPORTANT: Edit /etc/cyto-viewer/config.env to configure your system"
else
    echo "✓ Configuration already exists"
fi

# Install systemd service
echo ""
echo "Installing systemd service..."
cp scripts/cyto-viewer.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable cyto-viewer
echo "✓ Service installed"

# Test GPU
echo ""
echo "Testing CUDA installation..."
nvidia-smi
nvcc --version

echo ""
echo "======================================"
echo "Installation Complete!"
echo "======================================"
echo ""
echo "Next steps:"
echo "1. Edit configuration: nano /etc/cyto-viewer/config.env"
echo "2. Configure scanner address and credentials"
echo "3. Generate password hash: cd /opt/cyto-viewer && make gen-password"
echo "4. Start service: systemctl start cyto-viewer"
echo "5. Check status: systemctl status cyto-viewer"
echo "6. View logs: journalctl -u cyto-viewer -f"
echo "7. Access viewer: http://localhost:8080"
echo ""
echo "For help, see README.md"
echo ""
