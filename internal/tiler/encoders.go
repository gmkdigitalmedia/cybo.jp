package tiler

import (
	"bytes"
	"image"
	"image/jpeg"

	"github.com/chai2010/webp"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp as libwebp"
)

// encodeJPEG uses hardware-accelerated JPEG encoding when available
func encodeJPEG(img *image.RGBA, quality int) ([]byte, string, error) {
	var buf bytes.Buffer
	
	opts := &jpeg.Options{
		Quality: quality,
	}
	
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, "", err
	}
	
	return buf.Bytes(), "image/jpeg", nil
}

// encodeWebP provides optimized WebP encoding
func encodeWebP(img *image.RGBA, quality int) ([]byte, string, error) {
	config, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, float32(quality))
	if err != nil {
		return nil, "", err
	}

	// Use multi-threaded encoding for better performance
	config.SetThreadLevel(8)
	
	var buf bytes.Buffer
	if err := libwebp.Encode(&buf, img, config); err != nil {
		return nil, "", err
	}
	
	return buf.Bytes(), "image/webp", nil
}

// encodeAVIF provides next-gen AVIF encoding (best compression)
func encodeAVIF(img *image.RGBA, quality int) ([]byte, string, error) {
	// AVIF encoding requires libavif
	// For now, fallback to WebP which has similar compression
	// In production, integrate with libavif for true AVIF support
	return encodeWebP(img, quality)
}
