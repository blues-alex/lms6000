package lms6000

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/blues-alex/clog"
)

const (
	// meassure values offsets
	DATA_OFS     = 0x24
	TIMEDATE_OFS = 0xd4
	DATA_LENGTH  = 0x18
)

var (
	// Commands
	HALLO_MESS            = []byte{0x8c, 0x00}       //b'\x8c\x00'
	START_MEASSURE        = []byte{0x8c, 0x0e, 0x01} // b'\x8c\x0e\x01'
	START_MULTI_MEASSURES = []byte{0x8c, 0x0e}       // b'\x8c\x0e'
	GET_DATA              = []byte{0x8c, 0x13}       // b'\x8c\x13'
	SET_TIME_AVERAVE      = []byte{0x8c, 0x01}       // b'\x8c\x01'
	GET_TIME_AVERAVE      = []byte{0x8c, 0x05}       // b'\x8c\x05'

	// list of values
	MEASSURE_KEYS = []string{
		"PAR", "PPFD", "YPFD", "Ep", "Eb", "Ey", "Er",
		"Erb Ratio", "E(lx)", "E(fc)", "CCT", "Duv",
		"x", "y", "u", "v", "u'", "v'", "SDCM", "Ra",
		"R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8",
		"R9", "R10", "R11", "R12", "R13", "R14", "R15",
		"Ee", "S/P", "Dominant", "Purity (%)",
		"HalfWidth (nm)", "Peak", "Center", "Centroid",
	} // 43 values
)

type Meassure struct {
	Values   map[string]float64 `json:"values"`
	Spectrum map[string]float64 `json:"spectrum"`
}

type LMS6000 struct {
	device      SerialDevice
	TimeAverage time.Duration
}

type SerialDevice interface {
	Write([]byte) (int, error)
	Read([]byte) (int, error)
	ReadMessage() ([]byte, error)
	Close() error
}

func NewLMS6000(device SerialDevice) (*LMS6000, error) {
	lms := &LMS6000{
		device:      device,
		TimeAverage: time.Second / 1000,
	}

	timeAverage, err := lms.GetTimeAverage()
	if err != nil {
		clog.Error("Failed to get time average:", err)
		return nil, err
	}
	lms.TimeAverage = timeAverage

	return lms, nil
}

func (l *LMS6000) GetTimeAverage() (time.Duration, error) {
	_, err := l.device.Write(GET_TIME_AVERAVE)
	if err != nil {
		return time.Second, fmt.Errorf("failed to write GET_TIME_AVERAVE command: %w", err)
	}

	ans, err := l.device.ReadMessage()
	if err != nil {
		return time.Second, fmt.Errorf("failed to read response for GET_TIME_AVERAVE: %w", err)
	}

	tm := binary.LittleEndian.Uint32(ans[2:6])
	timeAverage := time.Duration(int64(tm/1000) * int64(time.Millisecond))
	return timeAverage, nil
}

func (l *LMS6000) SetTimeAverage(t time.Duration) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(t.Microseconds()))
	res := append(SET_TIME_AVERAVE, buf...)

	_, err := l.device.Write(res)
	if err != nil {
		return fmt.Errorf("failed to write SET_TIME_AVERAVE command: %w", err)
	}
	l.TimeAverage = t
	_, err = l.device.ReadMessage()
	return err
}

func (l *LMS6000) GetMeassure() (*Meassure, error) {
	err := l.StartMeassure()
	if err != nil {
		return nil, fmt.Errorf("failed to start measurement: %w", err)
	}

	time.Sleep((l.TimeAverage * 4) * time.Nanosecond)

	_, err = l.device.Write(GET_DATA)
	if err != nil {
		return nil, fmt.Errorf("failed to write GET_DATA command: %w", err)
	}

	m, err := l.device.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read measurement data: %w", err)
	}

	if len(m) > DATA_OFS+TIMEDATE_OFS+DATA_LENGTH+(450*4) {
		return parseMeassure(m), nil
	}
	return nil, fmt.Errorf("failed to get measurement data: insufficient data length")
}

func (l *LMS6000) SwithToMultiMeassureMod() error {
	_, err := l.device.Write(START_MULTI_MEASSURES)
	if err != nil {
		return fmt.Errorf("failed to write START_MULTI_MEASSURES command: %w", err)
	}

	_, err = l.device.ReadMessage()
	return err
}

func (l *LMS6000) StartMeassure() error {
	_, err := l.device.Write(START_MEASSURE)
	if err != nil {
		return fmt.Errorf("failed to write START_MEASSURE command: %w", err)
	}

	_, err = l.device.ReadMessage()
	return err
}

func bytesToFloats(buf []byte) []float32 {
	if len(buf) == 0 {
		return nil
	}
	res := []float32{}

	for i := 4; i < len(buf)-4; i += 4 {
		bt := buf[i : i+4]
		bits := binary.LittleEndian.Uint32(bt)
		float := math.Float32frombits(bits)
		res = append(res, float)
	}
	return res
}

func newMeassure() *Meassure {
	m := Meassure{}
	m.Values = map[string]float64{}
	m.Spectrum = map[string]float64{}
	return &m
}

func parseMeassure(m []byte) *Meassure {
	spectrumOfs := DATA_OFS + TIMEDATE_OFS + DATA_LENGTH
	data := bytesToFloats(m[DATA_OFS : TIMEDATE_OFS+4])
	spectrum := bytesToFloats(m[spectrumOfs : spectrumOfs+(452*4)])
	meass := newMeassure()
	for n, v := range data {
		meass.Values[MEASSURE_KEYS[n]] = float64(v)
	}
	for n, v := range spectrum {
		meass.Spectrum[fmt.Sprintf("%d", n+350)] = float64(v)
	}
	return meass
}
