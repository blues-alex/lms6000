package lms6000

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
	"uart"
)

const (
	// measure values offsets
	DATA_OFS     = 0x24
	TIMEDATE_OFS = 0xd4
	DATA_LENGTH  = 0x18
)

var (
	// Commands
	HALLO_MSG            = []byte{0x8c, 0x00}       //b'\x8c\x00'
	START_MEASURE        = []byte{0x8c, 0x0e, 0x01} // b'\x8c\x0e\x01'
	START_MULTI_MEASURES = []byte{0x8c, 0x0e}       // b'\x8c\x0e'
	GET_DATA             = []byte{0x8c, 0x13}       // b'\x8c\x0x13'
	SET_TIME_AVERAGE     = []byte{0x8c, 0x01}       // b'\x8c\x01'
	GET_TIME_AVERAGE     = []byte{0x8c, 0x05}       // b'\x8c\x05'

	// list of values
	MEASURE_KEYS = []string{
		"PAR", "PPFD", "YPFD", "Ep", "Eb", "Ey", "Er",
		"Erb Ratio", "E(lx)", "E(fc)", "CCT", "Duv",
		"x", "y", "u", "v", "u'", "v'", "SDCM", "Ra",
		"R1", "R2", "R3", "R4", "R5", "R6", "R7", "R8",
		"R9", "R10", "R11", "R12", "R13", "R14", "R15",
		"Ee", "S/P", "Dominant", "Purity (%)",
		"HalfWidth (nm)", "Peak", "Center", "Centroid",
	} // 43 values
)

type Measure struct {
	Values   map[string]float64 `json:"values"`
	Spectrum map[string]float64 `json:"spectrum"`
}

type LMS6000 struct {
	uart.SerialDevice
	TimeAverage time.Duration
}

func NewLMS6000(p string, b int) (*LMS6000, error) {
	port, err := uart.NewConnect(p, b)
	if err != nil {
		return nil, err
	}
	lms := &LMS6000{
		port,
		time.Second / 1000,
	}
	lms.TimeAverage, err = lms.GetTimeAverage()
	return lms, err
}

func (l *LMS6000) GetTimeAverage() (time.Duration, error) {
	l.ReadMessage()
	_, err := l.Write(GET_TIME_AVERAGE)
	if err != nil {
		return time.Second, err
	}

	ans, err := l.ReadMessage()
	if err != nil {
		return time.Second, err
	}

	if len(ans) < 6 {
		return time.Second, fmt.Errorf("response too short")
	}

	tm := binary.LittleEndian.Uint32(ans[2:6])
	timeAverage := time.Duration(int64(tm/1000) * int64(time.Millisecond))
	return timeAverage, nil
}

func (l *LMS6000) SetTimeAverage(t time.Duration) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(t.Microseconds()))
	res := append(SET_TIME_AVERAGE, buf...)

	_, err := l.Write(res)
	if err != nil {
		return err
	}
	l.TimeAverage = t
	_, err = l.ReadMessage()
	return err
}

func (l *LMS6000) GetMeasure() (*Measure, error) {
	err := l.StartMeasure()
	if err != nil {
		return nil, err
	}

	_, err = l.Write(GET_DATA)
	if err != nil {
		return nil, err
	}
	time.Sleep(l.TimeAverage)
	m, err := l.ReadMessage()
	if err != nil {
		return nil, err
	}

	if len(m) > DATA_OFS+TIMEDATE_OFS+DATA_LENGTH+(450*4) {

		return parseMeasure(m), nil
	}
	return nil, fmt.Errorf("ERROR: Failed to get measurement data.")
}

func (l *LMS6000) SwitchToMultiMeasureMode() error {
	_, err := l.Write(START_MULTI_MEASURES)
	if err != nil {
		return err
	}

	_, err = l.ReadMessage()
	return err
}

func (l *LMS6000) StartMeasure() error {
	_, err := l.Write(START_MEASURE)
	if err != nil {
		return err
	}

	_, err = l.ReadMessage()

	return err
}

func bytesToFloats(buf []byte) []float32 {
	if len(buf) == 0 {
		return nil
	}
	res := make([]float32, 0, len(buf)/4)

	for i := 4; i < len(buf)-4; i += 4 {
		bt := buf[i : i+4]
		bits := binary.LittleEndian.Uint32(bt)
		f := math.Float32frombits(bits)
		res = append(res, f)
	}
	return res
}

func newMeasure() *Measure {
	m := Measure{}
	m.Values = map[string]float64{}
	m.Spectrum = map[string]float64{}
	return &m
}

func parseMeasure(m []byte) *Measure {
	spectrumOfs := DATA_OFS + TIMEDATE_OFS + DATA_LENGTH
	minLen := spectrumOfs + 452*4
	if len(m) < minLen {
		return nil
	}
	data := bytesToFloats(m[DATA_OFS : TIMEDATE_OFS+4])
	spectrum := bytesToFloats(m[spectrumOfs : spectrumOfs+(452*4)])
	meas := newMeasure()
	for n, v := range data {
		if n >= len(MEASURE_KEYS) {
			break
		}
		meas.Values[MEASURE_KEYS[n]] = float64(v)
	}
	for n, v := range spectrum {
		meas.Spectrum[fmt.Sprintf("%d", n+350)] = float64(v)
	}
	return meas
}
