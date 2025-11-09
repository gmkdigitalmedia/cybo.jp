#!/usr/bin/env python3
"""
Development server for Cytology Viewer Pro
Serves the frontend and provides mock API responses for testing
"""

from http.server import HTTPServer, SimpleHTTPRequestHandler
from urllib.parse import urlparse, parse_qs
import json
import io
from PIL import Image, ImageDraw, ImageFont
import os

class CytologyDevServer(SimpleHTTPRequestHandler):
    def do_GET(self):
        parsed_path = urlparse(self.path)

        # Serve the login page at root
        if parsed_path.path == '/':
            self.serve_page('web/index.html')
            return

        # Serve the slides index page
        if parsed_path.path.startswith('/slides.html'):
            self.serve_page('web/slides.html')
            return

        # Serve the viewer page
        if parsed_path.path.startswith('/viewer.html'):
            self.serve_page('web/viewer.html')
            return

        # Mock tile API
        if parsed_path.path.startswith('/api/tiles/'):
            self.serve_mock_tile(parsed_path)
            return

        # Mock slide list API
        if parsed_path.path == '/api/slides':
            self.serve_slide_list()
            return

        # Mock system stats API
        if parsed_path.path == '/api/system/stats':
            self.serve_system_stats()
            return

        # Serve static files normally
        super().do_GET()

    def serve_page(self, file_path):
        """Serve an HTML page"""
        try:
            with open(file_path, 'rb') as f:
                content = f.read()

            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.send_header('Content-Length', len(content))
            self.end_headers()
            self.wfile.write(content)
        except Exception as e:
            self.send_error(500, f'Error serving page: {str(e)}')

    def serve_mock_tile(self, parsed_path):
        """Generate a mock tile image with parameters displayed"""
        try:
            # Parse query parameters
            query = parse_qs(parsed_path.query)
            layer = query.get('layer', ['0'])[0]
            x = query.get('x', ['0'])[0]
            y = query.get('y', ['0'])[0]
            z = query.get('z', ['1'])[0]

            # Create a mock tile image
            size = 512
            img = Image.new('RGB', (size, size), color=(30, 35, 50))
            draw = ImageDraw.Draw(img)

            # Draw grid pattern
            grid_size = 64
            for i in range(0, size, grid_size):
                draw.line([(i, 0), (i, size)], fill=(50, 60, 80), width=1)
                draw.line([(0, i), (size, i)], fill=(50, 60, 80), width=1)

            # Draw some random "cells" to simulate cytology image
            import random
            random.seed(int(layer) * 1000 + int(x) * 100 + int(y))

            for _ in range(20):
                cx = random.randint(50, size - 50)
                cy = random.randint(50, size - 50)
                radius = random.randint(10, 30)
                color = (
                    random.randint(100, 200),
                    random.randint(80, 160),
                    random.randint(120, 200)
                )
                draw.ellipse([cx-radius, cy-radius, cx+radius, cy+radius],
                           fill=color, outline=(255, 255, 255))

            # Draw tile info
            info_text = f"Layer:{layer} X:{x} Y:{y} Z:{z}"
            try:
                font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 16)
            except:
                font = ImageFont.load_default()

            # Draw text background
            bbox = draw.textbbox((0, 0), info_text, font=font)
            text_width = bbox[2] - bbox[0]
            text_height = bbox[3] - bbox[1]
            draw.rectangle([10, 10, 20 + text_width, 20 + text_height],
                         fill=(0, 0, 0, 128))
            draw.text((15, 15), info_text, fill=(0, 200, 100), font=font)

            # Convert to JPEG
            buffer = io.BytesIO()
            img.save(buffer, format='JPEG', quality=85)
            content = buffer.getvalue()

            self.send_response(200)
            self.send_header('Content-type', 'image/jpeg')
            self.send_header('Content-Length', len(content))
            self.send_header('Cache-Control', 'public, max-age=31536000')
            self.end_headers()
            self.wfile.write(content)

        except Exception as e:
            print(f"Error generating tile: {e}")
            self.send_error(500, f'Error generating tile: {str(e)}')

    def serve_slide_list(self):
        """Serve mock slide list"""
        slides = [
            {
                "id": "sample-001",
                "name": "Sample-001",
                "created": "2025-11-08T10:30:00Z",
                "layers": 40,
                "width": 10000,
                "height": 10000,
                "status": "ready"
            },
            {
                "id": "sample-002",
                "name": "Sample-002",
                "created": "2025-11-07T14:20:00Z",
                "layers": 40,
                "width": 12000,
                "height": 12000,
                "status": "ready"
            }
        ]

        content = json.dumps(slides, indent=2).encode()

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.send_header('Content-Length', len(content))
        self.end_headers()
        self.wfile.write(content)

    def serve_system_stats(self):
        """Serve mock system stats"""
        stats = {
            "uptime": 12345,
            "version": "1.0.0-dev",
            "gpu": {
                "available": True,
                "device": "Mock GPU",
                "memory_used": 2048,
                "memory_total": 16384,
                "utilization": 45
            },
            "cache": {
                "size": 150,
                "hit_rate": 0.95,
                "evictions": 42
            },
            "tiles": {
                "served": 15234,
                "errors": 12,
                "avg_time_ms": 8
            }
        }

        content = json.dumps(stats, indent=2).encode()

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.send_header('Content-Length', len(content))
        self.end_headers()
        self.wfile.write(content)

    def log_message(self, format, *args):
        """Custom log format"""
        print(f"[DEV SERVER] {self.address_string()} - {format % args}")


def run_server(port=8080):
    """Run the development server"""
    server_address = ('', port)
    httpd = HTTPServer(server_address, CytologyDevServer)

    print("=" * 70)
    print("Cytology Viewer Pro - Development Server")
    print("=" * 70)
    print(f"\nServer running at: http://localhost:{port}")
    print(f"\nOpen your browser to: http://localhost:{port}")
    print("\nFeatures:")
    print("   - Professional enterprise-grade UI")
    print("   - GPU-accelerated WebGL renderer")
    print("   - Mock tile generation for testing")
    print("   - Full viewer functionality")
    print("\nPress Ctrl+C to stop the server\n")
    print("=" * 70)
    print()

    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\n\nShutting down server...")
        httpd.shutdown()
        print("Server stopped")


if __name__ == '__main__':
    import sys
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8080
    run_server(port)
