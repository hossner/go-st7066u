package st7066u

import (
	"errors"
	"time"

	"github.com/stianeikeland/go-rpio"
)

// BITMODE4 and BITMODE8; used to denote if the LCD display is used in 4 or 8 bit mode respectively
const (
	BITMODE4 uint8 = iota
	BITMODE8
)

// DOTS5x8 and DOTS5x11; used to specifify the dot matrix used by the LCD display
const (
	DOTS5x8 uint8 = iota
	DOTS5x11
)

const (
	cmdInstruction uint8 = iota
	cmdData
)

const (
	pinEDelay = time.Microsecond * 1
	pinEWait  = time.Microsecond * 70
	row1Addr  = 0x80
	row2Addr  = 0xC0
)

// Device is the basic struct representing the LCD display. Use func New to get a new struct
type Device struct {
	rows              uint8
	cols              uint8
	pinRS, pinE, pinL rpio.Pin
	pinDs             []rpio.Pin
	mode              uint8
	sym               uint8
	ledOn             bool
	masks             map[string]uint8
}

// New returns a Device struct used as a handler for the LCD display. Arguments are
//	nrOfRows:	(uint8) 1 or 2 rows LCD displayes are supported
//	nrOfCols:	(uint8) Nr of columns in the display. 16 and 20 are common values
//	charSym:	Symmetry of the characters on the LCD display. DOTS5x8 or DOTS5x11 are supported
//	mode:		In which "mode" the display is connected (w/ 4 or 8 data wires). BITMODE4 and BITMODE8 are supported
//	pinRS:		GPIO pin used for the RS (reset) pin on the LCD display
//	pinE:		GPIO pin used for the E (enable) pin on the LCD display
//	pinL:		GPIO pin used for the L (LED) pin on the LCD display
//	pins:		GPIO pins used for data, can be either 4 or 8 pins. Start with the lowest numbered pin on the LCD display (D0 or D4, depending on "mode" used)
func New(nrOfRows, nrOfCols uint8, charSym uint8, mode uint8, pinRS, pinE, pinL rpio.Pin, pins ...rpio.Pin) (*Device, error) {
	if err := validateSymm(nrOfRows, nrOfCols, charSym); err != nil {
		return nil, err
	}
	if err := rpio.Open(); err != nil {
		return nil, err
	}
	g := &Device{
		rows:  nrOfRows,
		cols:  nrOfCols,
		pinRS: pinRS,
		pinE:  pinE,
		pinL:  pinL,
		pinDs: pins,
		mode:  BITMODE8,
		sym:   charSym,
	}
	if err := validatePinMode(mode, len(pins)); err != nil {
		rpio.Close()
		return nil, err
	}
	if len(pins) == 4 {
		g.mode = BITMODE4
	}
	g.setDefaultMasks()
	g.init()
	g.Clear()
	return g, nil
}

// Clear clears the LCD
func (l *Device) Clear() {
	l.write(1<<0, cmdInstruction)
	time.Sleep(pinEWait * 100)
}

// Close closes the LCD display
func (l *Device) Close() {
	l.Clear()
	l.TurnOn(false)
	l.LedOn(false)

	for _, p := range append(l.pinDs, l.pinRS, l.pinE) {
		p.Low()
	}
	rpio.Close()
}

// CursorBlink sets the cursor to blink/not blink
func (l *Device) CursorBlink(on bool) {
	var mask uint8 = 1 << 0
	if !on {
		mask = ^mask
		l.masks["display"] &= mask
	} else {
		l.masks["display"] |= mask
	}
	l.write(l.masks["display"], cmdInstruction)
}

// CursorOn shows/hides the cursor
func (l *Device) CursorOn(on bool) {
	var mask uint8 = 1 << 1
	if !on {
		mask = ^mask
		l.masks["display"] &= mask
	} else {
		l.masks["display"] |= mask
	}
	l.write(l.masks["display"], cmdInstruction)
}

// Home moves the cursor to the home position, i.e. row 0, col 0
func (l *Device) Home() {
	l.write(1<<1, cmdInstruction)
}

// LedOn turns LCD LED on or off
func (l *Device) LedOn(on bool) {
	if on {
		l.pinL.High()
	} else {
		l.pinL.Low()
	}
}

// MoveLeft moves the caret 'steps' steps to the left
func (l *Device) MoveLeft(steps uint8) {
	var mask uint8 = 0b10000
	var a uint8
	for a = 0; a < steps; a++ {
		l.write(mask, cmdInstruction)
	}
}

// Print prints the provided text on the LCD display at the current position of the caret
func (l *Device) Print(text string) {
	txt := strToSt70660b(text)
	for _, c := range txt {
		l.write(c, cmdData)
	}
}

// PrintAt prints the provided text at the specified cursor position
func (l *Device) PrintAt(row, col uint8, text string) {
	l.SetCursor(row, col)
	l.Print(text)
}

// PrintByte prints just one byte character to the LCD display
func (l *Device) PrintByte(ch byte) {
	l.write(runeToSt70660b(rune(ch)), cmdData)
}

// PrintRune prints just one rune character to the LCD display
func (l *Device) PrintRune(ch rune) {
	l.write(runeToSt70660b(ch), cmdData)
}

// SetCursor moves the cursor to the provided row and col
func (l *Device) SetCursor(row, col uint8) {
	if row > l.rows-1 || col > l.cols-1 {
		return
	}
	offset := 0x40*row + col
	l.write(0x80|offset, cmdInstruction)
}

// TurnOn is used to turn whole LCD display on or off
func (l *Device) TurnOn(on bool) {
	var mask uint8 = 1 << 2
	if on {
		l.masks["display"] |= mask
	} else {
		mask := ^mask
		l.masks["display"] &= mask
	}
	l.write(l.masks["display"], cmdInstruction)
}

// enableWrite is the toggle sequence on pinE used to shift in the command
// to the LCD display
func (l *Device) enableWrite() {
	time.Sleep(pinEDelay)
	l.pinE.High()
	time.Sleep(pinEDelay)
	l.pinE.Low()
	time.Sleep(pinEWait)
}

// init initializes the LCD display with the default values
func (l *Device) init() {
	for _, p := range append(l.pinDs, l.pinRS, l.pinE) {
		rpio.PinMode(p, rpio.Output)
	}
	l.write(l.masks["functionSet"], cmdInstruction)
	time.Sleep(pinEWait)
	l.write(l.masks["display"], cmdInstruction)
	time.Sleep(pinEWait)
}

// setDefaultMasks sets the default values of the different instructions to be used at initialization
// of the display
func (l *Device) setDefaultMasks() {
	l.masks = make(map[string]uint8)
	l.masks["entryMode"] = 0b100
	l.masks["display"] = 0b1100
	l.masks["displayShift"] = 0b10100
	l.masks["functionSet"] = 0b100000
	if l.mode == BITMODE8 {
		l.masks["functionSet"] |= (1 << 4)
	}
	if l.rows == 2 {
		l.masks["functionSet"] |= (1 << 3)
	}
	if l.sym == DOTS5x11 {
		l.masks["functionSet"] |= (1 << 2)
	}
}

// validatePinMode validates input of nr of pins and mode requested
func validatePinMode(mode uint8, nrs int) error {
	if nrs != 4 && nrs != 8 {
		return errors.New("Number of pins must be either 4 or 8")
	}
	if mode != BITMODE4 && mode != BITMODE8 {
		return errors.New("Only BITMODE4 and BITMODE8 are supported")
	}
	if (mode == BITMODE4 && nrs != 4) || (mode == BITMODE8 && nrs != 8) {
		return errors.New("Missmatch between mode and number of pins used")
	}
	return nil
}

// validateSymm validates input of required nr of rows, cols and font symmetry
func validateSymm(rows, cols, font uint8) error {
	if rows > 2 || rows < 1 || cols < 1 || cols > 40 {
		return errors.New("Number of rows must be either 1 or 2, and number of columns must be greater than 0 and less than 40")
	}
	if font > DOTS5x11 {
		return errors.New("Only 5x8 and 5x11 dot characters are supported")
	}
	if font == DOTS5x11 && rows > 1 {
		return errors.New("5x11 dot characters not supported in multiline LCD displays")
	}
	return nil
}

// write writes data to the LCD display, either to be shown or as a command
func (l *Device) write(data uint8, cmd uint8) {
	l.pinRS.Write(rpio.State(cmd))
	if l.mode == BITMODE8 {
		for i := 0; i < 8; i++ {
			if data&(1<<i) == 1<<i {
				l.pinDs[i].High()
			} else {
				l.pinDs[i].Low()
			}
		}
		l.enableWrite()
		return
	}
	for nibble := 4; nibble >= 0; nibble -= 4 {
		for i := 0; i < 4; i++ {
			if data&(1<<(i+nibble)) == 1<<(i+nibble) {
				l.pinDs[i].High()
			} else {
				l.pinDs[i].Low()
			}
		}
		l.enableWrite()
	}
}
