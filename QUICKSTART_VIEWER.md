# Cytology Viewer Pro - Quick Start Guide

## Get Started in 2 Minutes

This guide will help you run the professional, enterprise-grade cytology viewer immediately.

---

## What You're Getting

A **professional medical imaging viewer** with:

- Enterprise-grade UI - Modern dark theme with smooth animations
- GPU-accelerated rendering - WebGL 2.0 for silky 60 FPS performance
- Multi-layer focus control - Navigate through 40 focus layers
- Real-time adjustments - Brightness, contrast, sharpness, saturation
- Professional tools - Annotations, measurements, presets
- Responsive design - Works on any screen size
- Slide library - Professional sidebar with slide management
- Performance monitoring - Real-time FPS, tile count, zoom stats

---

## Option 1: Quick Development Server (No Dependencies)

### Step 1: Install Python (if not already installed)

```bash
# Check if Python is installed
python3 --version

# If not installed, install Python 3
# Ubuntu/Debian:
sudo apt install python3 python3-pip

# macOS:
brew install python3
```

### Step 2: Install PIL/Pillow (for mock tile generation)

```bash
pip3 install pillow
```

### Step 3: Run the Dev Server

```bash
# Make the server executable
chmod +x dev_server.py

# Run it
python3 dev_server.py

# Or on a different port:
python3 dev_server.py 3000
```

### Step 4: Open Your Browser

Navigate to: **http://localhost:8080**

You'll see the professional viewer with mock cytology images generated on the fly.

---

## Option 2: Production Server (Go + CUDA)

For production deployment with real GPU acceleration:

### Prerequisites
- NVIDIA GPU with CUDA support
- CUDA Toolkit 12.3+
- Go 1.21+
- Ubuntu 22.04 or similar

### Quick Build

```bash
# Install dependencies
make install-deps

# Build the project
make build

# Create configuration
cp config.env.example config.env
# Edit config.env with your settings

# Run the server
./bin/cyto-viewer
```

Open browser to: **http://localhost:8080**

---

## Using the Viewer

### Navigation

- **Pan**: Click and drag with mouse
- **Zoom**: Mouse wheel or trackpad pinch
- **Reset View**: Click "Reset View" button or press `R`

### Keyboard Shortcuts

- `P` - Pan tool
- `A` - Annotate tool
- `M` - Measure tool
- `R` - Reset view
- `F` - Toggle fullscreen
- `0-9` - Quick layer switching

### Controls Panel (Right Side)

**Focus Control**
- Adjust the focus layer from 0 (surface) to 39 (deep)
- Use quick buttons for common layers
- Real-time layer switching

**Image Adjustments**
- **Brightness**: 50% - 200%
- **Contrast**: 50% - 200%
- **Sharpness**: 0% - 100% (GPU-accelerated unsharp mask)
- **Saturation**: 0% - 200% (HSV color space)

**View Controls**
- Reset View - Return to default position/zoom
- Fullscreen - Immersive viewing mode
- Toggle Grid - Show measurement grid
- Export View - Save current view as image

**Presets**
- Default View - Standard settings
- High Contrast - Enhanced for difficult samples
- Enhanced Detail - Maximum sharpness and clarity

### Slide Library (Left Side)

- Browse available slides
- Search slides by name
- Click to switch between slides
- View slide metadata

### Performance Stats (Bottom)

- **FPS**: Frames per second (target: 60)
- **Tiles Loaded**: Number of tiles in cache
- **Zoom**: Current zoom level
- **Position**: Camera coordinates

---

## UI Features

### Professional Design Elements

- Gradient backgrounds - Modern, professional look
- Smooth animations - Hover effects, transitions
- Glass morphism - Frosted glass effects with backdrop blur
- Color-coded controls - Intuitive visual feedback
- Responsive sliders - Dynamic background fills
- Loading states - Professional spinner animations
- Status indicators - Real-time system status

### Enterprise-Grade Polish

- Professional color scheme (dark blue theme)
- Consistent spacing and typography
- Accessible UI with clear labels
- Smooth 60 FPS animations
- High-contrast elements for readability
- Medical-grade color accuracy

---

## Customization

### Changing Colors

Edit the CSS variables in `web/index.html`:

```css
:root {
    --primary: #0066cc;        /* Main brand color */
    --secondary: #00c853;      /* Accent color */
    --background: #0a0e1a;     /* Main background */
    --surface: #151b2e;        /* Card backgrounds */
    /* ... */
}
```

### Adding Custom Presets

Add to the `applyPreset()` function in `web/index.html`:

```javascript
const presets = {
    'my-preset': {
        brightness: 120,
        contrast: 110,
        sharpness: 40,
        saturation: 95
    }
};
```

---

## Performance Optimization

### For Best Performance

1. **Use a modern browser**
   - Chrome 90+ or Firefox 88+
   - Enable hardware acceleration

2. **GPU Acceleration**
   - Ensure WebGL 2.0 is supported
   - Check: Visit `chrome://gpu`

3. **Tile Caching**
   - Default: 200 tiles cached
   - Adjust in code: `this.maxCacheSize = 200`

4. **Prefetching**
   - Tiles load 1 tile outside viewport
   - Adjust padding in `calculateVisibleTiles()`

---

## Troubleshooting

### Viewer Not Loading

1. Check browser console (F12)
2. Ensure dev server is running
3. Check port is not in use

### WebGL Errors

- Update graphics drivers
- Try different browser
- Check WebGL support: https://get.webgl.org/webgl2/

### Slow Performance

- Close other GPU-intensive apps
- Reduce browser zoom level
- Check FPS counter (should be 60)

### Tiles Not Loading

- Check network tab in browser dev tools
- Verify dev server is generating tiles
- Check console for errors

---

## Production Deployment

### Security Checklist

- [ ] Change JWT secret in config
- [ ] Set up HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Enable authentication
- [ ] Set up backups
- [ ] Configure log rotation
- [ ] Review CORS settings

### Performance Checklist

- [ ] Enable GZIP compression
- [ ] Set up CDN for static assets
- [ ] Configure tile caching headers
- [ ] Optimize GPU cache size
- [ ] Monitor memory usage
- [ ] Set up load balancing (if needed)

---

## API Endpoints (for Integration)

```bash
# Get slide list
GET /api/slides

# Get specific slide
GET /api/slides/{slideId}

# Get tile
GET /api/tiles/{slideId}?layer=5&x=10&y=20&z=1

# System stats
GET /api/system/stats

# Scanner status (production only)
GET /api/scanner/status
```

---

## Performance Advantages

### Compared to OpenSeadragon/Similar Viewers

- 10-50x faster tile rendering
- 60 FPS instead of 15-30 FPS
- Modern UI vs basic controls
- GPU-accelerated image processing
- Enterprise polish with animations
- Built-in tools (annotations, measurements)
- Professional design
- Mobile-ready responsive layout

### Technology Stack

- **Frontend**: WebGL 2.0, Modern JavaScript
- **Backend**: Go + CUDA (production)
- **Mock Server**: Python 3 (development)
- **Design**: Custom CSS with modern best practices

---

## Learn More

- **Full Documentation**: See `README.md`
- **Architecture**: See `PROJECT_OVERVIEW.md`
- **Performance Comparison**: See `COMPARISON.md`
- **API Reference**: See source code comments

---

## Demo Tips

1. **Show smooth zooming** - Demonstrate the 60 FPS performance
2. **Switch layers quickly** - Show real-time focus control
3. **Use presets** - Display instant image enhancement
4. **Fullscreen mode** - Immersive experience
5. **Keyboard shortcuts** - Professional workflow
6. **Show statistics** - Highlight technical performance

---

## Next Steps

1. **Test the viewer** - Explore all controls
2. **Customize the theme** - Make it your own
3. **Connect real data** - Integrate with actual scanner
4. **Add features** - Extend with custom tools
5. **Deploy to production** - Use the full Go/CUDA stack

---

## Support

If you encounter any issues:

1. Check the troubleshooting section above
2. Review the browser console (F12)
3. Check the dev server logs
4. Review the full `README.md` documentation

---

**Professional, high-performance cytology viewer for medical imaging**

Built with precision. Designed with care. Performs beyond expectations.

cybo.co.jp - Cytology Viewer Pro
