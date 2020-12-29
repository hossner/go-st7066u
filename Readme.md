## About The Project

go-st7066u is a minimalistic Go driver for 1 or 2 rows, 32 or 40 columns, 5x8 or 5x11 dots LCD displays using the [ST7066](https://www.newhavendisplay.com/app_notes/ST7066U.pdf) chip.

I wrote this because
* I wanted a driver for Go
* I noticed small differences (mostly regarding timing) in the available drivers for the HD4470 chip

So I reconned I'd write my own :)

Note! I have not been able to test this on any other display than the one I'm using, and in any other configuration than the one I've used to hook it up to my Raspberry Pi, so I make no claim that it will work in your case. I have my [ST7066](https://www.newhavendisplay.com/app_notes/ST7066U.pdf) chip based 1602 LCD display hooked up to the Pi in the 4-bit configuration (using 4 only data wires) and not using the R/W (i.e. keeping it low/to gnd).

...OK, to be honest; I haven't tested it much at all...

A big shout out and thanks to the guys and gals behind:
* [go-rpio](https://github.com/stianeikeland/go-rpio)

on which this is built.

### Installation

```shell
go get github.com/hossner/go-st7066u
```

## Usage

```shell
package main

import (
	"time"
    "log"

	"github.com/hossner/go-st7066u"
)

func main() {
    lcd, err := st7066u.New(
        2,  // nr of rows
        16, // nr of columns/characters
        st7066u.DOTS5x8,    // font is 5x8 dots
        st7066u.BITMODE4,   // using 4-bit mode
        7,  // RS pin
        8,  // E pin
        15, // L (or LED) pin
        18, // D4 pin
        23, // D5 pin
        24, // D6 pin
        25  // D7 pin
    )
    if err != nil {
        log.Fatalln(err)
    }
    // Turn the backlight on
    lcd.LedOn(true)
    // Cursor starts by default positioned at row 0, col 0
    // Print something..
    lcd.Print("Hello")
    // Move cursor to second row, at first column/character position
    lcd.SetCursor(1, 0)
    // Print...
    lcd.Print("World")
    time.Sleep(time.Second * 5)
    lcd.LedOn(false)
    lcd.Close()
}
```

## API
Struct ```Device``` is the basic representation of the LCD display. Retrieve a new instance using the ```New``` function. Then the following functions can be used:

```Clear()```

Clears the display and positions the cursor at row 0, column 0.

```Close()```

Closes the gpio. Call this last.

```CursorBlink(on bool)```

Makes the cursor blink.

```CursorOn(on bool)```

Shows/hides the cursor. Default is hidden.

```Home()```

Returns the cursor at row 0, column 0.

```LedON(on bool)```

Turns on/off the backlight.

```MoveLeft(steps uint8)```

Moves the position of the cursor the provided nr of steps to the left.

```New(nrOfRows, nrOfCols uint8, charSym uint8, mode uint8, pinRS, pinE, pinL rpio.Pin, pins ...rpio.Pin) (*Device, error)```

New returns a pointer to a new device struct. The parameters are:
- Number of rows and columns on the display. 1 and 2 rows, and up to 40 columns are supported.
- The symmetry of the characters on the display. 5 x 8 (DOTS5x8) and 5 x 11 (DOTS5x11) are supported. Note that only 5 x 8 dot characters are supported on displays with two rows.
- If the display is connected using 4 (BITMODE4) or 8 (BITMODE8) data wires.
- The pins for the RS, E and L (or A/anode) wires. Note that this driver does not support usage of the R/W pin, this needs to be held low (to ground).
- The 4 or 8 pins for datatransfer. Start with the lowest D-pin (on the display), i.e. D0 (in BITMODE8) or D4 (in BITMODE4).

The function returns nil if an error is returned.

```Print(text string)```

Prints the provided string. Note the character set supported by the display (see [datasheet](https://www.newhavendisplay.com/app_notes/ST7066U.pdf) at page 14).

```PrintAt(row, col uint8, text string)```

Prints the provided string at the provided postion.

```PrintByte(byte)```

Prints the provided byte.

```PrintRune(rune)```

Prints the provided rune.

```SetCursor(row, col uint8)```

Moves the cursor to the provied location.

```TurnOn(on bool)```

Turns the display on/off. Note that this doesn't close the gpio's - use ```Close()``` for that.

## Issues / TBA
- Use of the R/W pin is not implemented (which would probably make it both faster and more stable), so the R/W pin must be held low (to gnd)
- The ST7066 chip has support for user-provided characters, but that functionality is not implemented
- More tests and testing is needed!

## License

Distributed under the MIT License. See `LICENSE` for more information.

