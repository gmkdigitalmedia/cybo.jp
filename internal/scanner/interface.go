package scanner

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"cyto-viewer/internal/config"
)

type Interface struct {
	config    *config.ScannerConfig
	conn      net.Conn
	mu        sync.Mutex
	connected bool
	layerData map[int]*LayerInfo
}

type LayerInfo struct {
	LayerIndex int
	Width      int
	Height     int
	FocusDepth float64
	TileSize   int
	Format     string
}

type ScanRequest struct {
	StartX int
	StartY int
	Width  int
	Height int
	Layers []int // Which focus layers to capture
}

type ScanResult struct {
	SlideID   string
	Timestamp time.Time
	Layers    []*LayerData
	Metadata  map[string]interface{}
}

type LayerData struct {
	Layer      int
	RawData    []byte
	Width      int
	Height     int
	TilesX     int
	TilesY     int
	TileSize   int
	Compressed bool
}

const (
	// Scanner communication protocol commands
	CMD_CONNECT     = 0x01
	CMD_DISCONNECT  = 0x02
	CMD_SCAN        = 0x03
	CMD_GET_LAYERS  = 0x04
	CMD_CALIBRATE   = 0x05
	CMD_STATUS      = 0x06
	CMD_SET_FOCUS   = 0x07
	CMD_GET_IMAGE   = 0x08
)

func NewInterface(cfg *config.ScannerConfig) (*Interface, error) {
	iface := &Interface{
		config:    cfg,
		layerData: make(map[int]*LayerInfo),
	}

	if err := iface.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to scanner: %w", err)
	}

	// Get layer information from scanner
	if err := iface.queryLayerInfo(); err != nil {
		return nil, fmt.Errorf("failed to query layer info: %w", err)
	}

	return iface, nil
}

func (s *Interface) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Connect to scanner via TCP or serial
	var err error
	switch s.config.Protocol {
	case "tcp":
		s.conn, err = net.DialTimeout("tcp", s.config.Address, 10*time.Second)
	case "serial":
		// For serial connections, use a serial library
		return fmt.Errorf("serial protocol not yet implemented")
	default:
		return fmt.Errorf("unsupported protocol: %s", s.config.Protocol)
	}

	if err != nil {
		return err
	}

	// Send connection handshake
	if err := s.sendCommand(CMD_CONNECT, nil); err != nil {
		s.conn.Close()
		return err
	}

	s.connected = true
	return nil
}

func (s *Interface) queryLayerInfo() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return fmt.Errorf("not connected to scanner")
	}

	// Request layer information
	if err := s.sendCommand(CMD_GET_LAYERS, nil); err != nil {
		return err
	}

	// Read response
	response, err := s.readResponse()
	if err != nil {
		return err
	}

	// Parse layer information
	// Format: [num_layers][layer_1_info][layer_2_info]...
	numLayers := int(binary.BigEndian.Uint32(response[:4]))
	offset := 4

	for i := 0; i < numLayers; i++ {
		layerInfo := &LayerInfo{
			LayerIndex: int(binary.BigEndian.Uint32(response[offset:])),
			Width:      int(binary.BigEndian.Uint32(response[offset+4:])),
			Height:     int(binary.BigEndian.Uint32(response[offset+8:])),
			FocusDepth: float64(binary.BigEndian.Uint32(response[offset+12:])) / 1000.0,
			TileSize:   int(binary.BigEndian.Uint32(response[offset+16:])),
		}
		s.layerData[layerInfo.LayerIndex] = layerInfo
		offset += 20
	}

	return nil
}

func (s *Interface) StartScan(ctx context.Context, req *ScanRequest) (*ScanResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil, fmt.Errorf("not connected to scanner")
	}

	// Prepare scan command
	cmdData := make([]byte, 20)
	binary.BigEndian.PutUint32(cmdData[0:], uint32(req.StartX))
	binary.BigEndian.PutUint32(cmdData[4:], uint32(req.StartY))
	binary.BigEndian.PutUint32(cmdData[8:], uint32(req.Width))
	binary.BigEndian.PutUint32(cmdData[12:], uint32(req.Height))
	binary.BigEndian.PutUint32(cmdData[16:], uint32(len(req.Layers)))

	// Add layer indices
	for _, layer := range req.Layers {
		layerBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(layerBuf, uint32(layer))
		cmdData = append(cmdData, layerBuf...)
	}

	// Send scan command
	if err := s.sendCommand(CMD_SCAN, cmdData); err != nil {
		return nil, err
	}

	// Wait for scan to complete and receive data
	result := &ScanResult{
		SlideID:   fmt.Sprintf("slide_%d", time.Now().Unix()),
		Timestamp: time.Now(),
		Layers:    make([]*LayerData, 0, len(req.Layers)),
		Metadata:  make(map[string]interface{}),
	}

	// Receive layer data
	for range req.Layers {
		layerData, err := s.receiveLayerData(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to receive layer data: %w", err)
		}
		result.Layers = append(result.Layers, layerData)
	}

	return result, nil
}

func (s *Interface) receiveLayerData(ctx context.Context) (*LayerData, error) {
	// Read layer header
	header := make([]byte, 32)
	if _, err := s.conn.Read(header); err != nil {
		return nil, err
	}

	layerData := &LayerData{
		Layer:      int(binary.BigEndian.Uint32(header[0:])),
		Width:      int(binary.BigEndian.Uint32(header[4:])),
		Height:     int(binary.BigEndian.Uint32(header[8:])),
		TilesX:     int(binary.BigEndian.Uint32(header[12:])),
		TilesY:     int(binary.BigEndian.Uint32(header[16:])),
		TileSize:   int(binary.BigEndian.Uint32(header[20:])),
		Compressed: header[24] == 1,
	}

	dataSize := int(binary.BigEndian.Uint32(header[28:]))
	layerData.RawData = make([]byte, dataSize)

	// Read the actual image data
	totalRead := 0
	for totalRead < dataSize {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			n, err := s.conn.Read(layerData.RawData[totalRead:])
			if err != nil {
				return nil, err
			}
			totalRead += n
		}
	}

	return layerData, nil
}

func (s *Interface) GetLayerInfo() map[int]*LayerInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Return a copy
	result := make(map[int]*LayerInfo)
	for k, v := range s.layerData {
		result[k] = v
	}
	return result
}

func (s *Interface) SetFocusLayer(layer int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmdData := make([]byte, 4)
	binary.BigEndian.PutUint32(cmdData, uint32(layer))

	return s.sendCommand(CMD_SET_FOCUS, cmdData)
}

func (s *Interface) GetStatus() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sendCommand(CMD_STATUS, nil); err != nil {
		return nil, err
	}

	response, err := s.readResponse()
	if err != nil {
		return nil, err
	}

	// Parse status response
	status := map[string]interface{}{
		"connected":    s.connected,
		"temperature":  float64(binary.BigEndian.Uint32(response[0:])) / 100.0,
		"ready":        response[4] == 1,
		"error_code":   int(binary.BigEndian.Uint32(response[5:])),
		"current_layer": int(binary.BigEndian.Uint32(response[9:])),
	}

	return status, nil
}

func (s *Interface) sendCommand(cmd byte, data []byte) error {
	// Command format: [CMD][LENGTH][DATA]
	cmdPacket := []byte{cmd}
	
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))
	cmdPacket = append(cmdPacket, length...)
	
	if data != nil {
		cmdPacket = append(cmdPacket, data...)
	}

	_, err := s.conn.Write(cmdPacket)
	return err
}

func (s *Interface) readResponse() ([]byte, error) {
	// Read length first
	lengthBuf := make([]byte, 4)
	if _, err := s.conn.Read(lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	response := make([]byte, length)

	totalRead := 0
	for totalRead < int(length) {
		n, err := s.conn.Read(response[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}

	return response, nil
}

func (s *Interface) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.connected {
		return nil
	}

	s.sendCommand(CMD_DISCONNECT, nil)
	s.connected = false

	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}
