package lms6000

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
)

type mockSerialDevice struct {
	readData  []byte
	readDelay time.Duration
}

func (m *mockSerialDevice) Write(b []byte) (int, error) {
	return len(b), nil
}

func (m *mockSerialDevice) ReadMessage() ([]byte, error) {
	if m.readDelay > 0 {
		time.Sleep(m.readDelay)
	}
	return m.readData, nil
}

func createMeasureResponse() []byte {
	resp := make([]byte, 0, 2048)
	resp = append(resp, make([]byte, DATA_OFS)...)
	data := []float32{
		100.5, 200.0, 300.0, 400.0, 500.0, 600.0, 700.0,
		0.8, 1000.0, 100.0, 5000.0, 0.001, 0.3, 0.4,
		0.25, 0.3, 0.25, 0.3, 5.0, 90.0,
		85.0, 86.0, 87.0, 88.0, 89.0, 90.0, 91.0, 92.0,
		93.0, 94.0, 95.0, 96.0, 97.0, 98.0, 99.0,
		50.0, 1.5, 550.0, 50.0,
		20.0, 550.0, 555.0, 560.0,
	}
	for _, v := range data {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
		resp = append(resp, buf...)
	}
	resp = append(resp, make([]byte, TIMEDATE_OFS+DATA_LENGTH-4*len(data))...)
	spectrum := make([]float32, 452)
	for i := range spectrum {
		spectrum[i] = float32(i) * 0.1
	}
	for _, v := range spectrum {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
		resp = append(resp, buf...)
	}
	return resp
}

func TestBytesToFloats(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []float32
	}{
		{
			name:     "empty",
			input:    []byte{},
			expected: nil,
		},
		{
			name:     "valid floats",
			input:    []byte{0, 0, 0, 0, 0x00, 0x00, 0x80, 0x3f, 0x00, 0x00, 0x00, 0x40, 0, 0, 0, 0},
			expected: []float32{1.0, 2.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToFloats(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expected[%d] = %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestNewMeasure(t *testing.T) {
	m := newMeasure()

	if m == nil {
		t.Fatal("newMeasure returned nil")
	}

	if m.Values == nil {
		t.Error("Values map is nil")
	}

	if m.Spectrum == nil {
		t.Error("Spectrum map is nil")
	}
}

func TestParseMeasure(t *testing.T) {
	resp := createMeasureResponse()

	m := parseMeasure(resp)

	if m == nil {
		t.Fatal("parseMeasure returned nil")
	}

	if len(m.Values) == 0 {
		t.Error("Values map is empty")
	}

	if len(m.Spectrum) == 0 {
		t.Error("Spectrum map is empty")
	}

	expectedKeys := []string{"PAR", "CCT", "Ra"}
	for _, key := range expectedKeys {
		if _, ok := m.Values[key]; !ok {
			t.Errorf("expected key %s not found in Values", key)
		}
	}
}

func TestParseMeasureTooShort(t *testing.T) {
	shortData := make([]byte, 10)

	m := parseMeasure(shortData)

	if m != nil {
		t.Error("parseMeasure should return nil for too short data")
	}
}

func TestMeasureJSON(t *testing.T) {
	m := &Measure{
		Values:   map[string]float64{"CCT": 5000, "Ra": 90},
		Spectrum: map[string]float64{"550": 0.5},
	}

	if m.Values["CCT"] != 5000 {
		t.Errorf("expected CCT=5000, got %v", m.Values["CCT"])
	}

	if m.Spectrum["550"] != 0.5 {
		t.Errorf("expected spectrum[550]=0.5, got %v", m.Spectrum["550"])
	}
}
