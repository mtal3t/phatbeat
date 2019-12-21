package phatbeat

import (
	"github.com/warthog618/gpio"
	"log"
	"math"
	"time"
)

// Data
const (
	Dat           = 23
	Clk           = 24
	Pixels        = 16
	ChannelPixels = 8
	Brightness    = 7

	FastForward = 5
	Rewind      = 13
	PlayPause   = 6
	VolUp       = 16
	VolDown     = 26
	OnOff       = 12
)

// Pixel .
type Pixel struct {
	r          uint8
	g          uint8
	b          uint8
	brightness uint8
}

var pixels [16]Pixel

var buttonHandlers = map[int]func(int){}
var buttonRepeat = map[int]bool{}
var buttonHoldTime = map[int]int{}
var buttonHoldRepeat = map[int]bool{}
var buttonHoldHandler = map[int]func(int){}

var clearOnExit bool = true

var buttons = []int{
	FastForward,
	Rewind,
	PlayPause,
	VolUp,
	VolDown,
	OnOff,
}

func init() {
	err := gpio.Open()

	if err != nil {
		log.Fatal(err)
	}

	// Set default brightness value
	for i := range pixels {
		pixels[i].brightness = 7
	}

	// Setup once at intialization.
	log.Println("initializing gpio Pins")
	gpio.NewPin(Dat).Output()
	gpio.NewPin(Clk).Output()

	log.Println("setting up handlers")
	for _, v := range buttons {
		pin := gpio.NewPin(v)
		pin.Input()
		pin.PullUp()
		pin.Watch(gpio.EdgeFalling, handleButton)
	}
}

// Clean .
func Clean() {
	if clearOnExit {
		clear(nil)
		Show()
	}

	gpio.Close()
}

func handleHold(pin int) bool {

	handler, ok := buttonHoldHandler[pin]

	if ok {
		t, ok := buttonHoldTime[pin]
		if ok {
			end := int64(t*1e9) + time.Now().UnixNano()
			p := gpio.NewPin(pin)

			for time.Now().UnixNano() < end {
				if p.Read() {
					return false
				}

				time.Sleep(0.001 * 1e9)
			}

			handler(pin)

			return true
		}
	}

	return false
}

func handleButton(pin *gpio.Pin) {

	log.Println("handle button", pin.Pin())
	p := pin.Pin()
	go func(p int) {

		if handleHold(p) {
			return
		}

		handler, ok := buttonHandlers[p]

		if ok {
			handler(p)

			r, ok := buttonRepeat[p]
			if !ok || !r {
				return
			}

			// let's use a nanosecond precision
			delay := int64(0.25 * 1e9)
			rampRate := int64(0.9 * 1e9)
			tDelay := int64(0.5 * 1e9)
			tEnd := time.Now().UnixNano() + tDelay

			for time.Now().UnixNano() < tEnd {
				if gpio.NewPin(p).Read() {
					return
				}

				time.Sleep(0.001 * 1e9)
			}

			for !gpio.NewPin(p).Read() {
				buttonHandlers[p](p)
				time.Sleep(time.Duration(delay))
				delay *= rampRate
				delay = int64(math.Max(0.01*1e9, float64(delay)))
			}

		}

	}(p)

}

// Hold .
func Hold(b int, repeat bool, holdTime int, handler func(int)) {

	/*
			Register a button hold handler.

		    :param button: Individual pin or to watch
		    :param handler: handler function
		    :param repeat: Whether the handler should be repeated if the button is held
			:param hold_time: How long (in seconds) the button should be held before triggering
	*/

	log.Println("handling hold")

	buttonHoldRepeat[b] = repeat
	buttonHoldTime[b] = holdTime

	if handler != nil {
		buttonHoldHandler[b] = handler
	}
}

// On .
func On(button int, repeat bool, handle func(int)) {

	/*
			Attach a handler function a button.

		    :param button: Individual button pin to watch
		    :param handler: Optional handler function
		    :param repeat: Whether the handler should be repeated if the button is held
	*/

	log.Println("handling On")
	buttonRepeat[button] = repeat

	if handle != nil {
		buttonHandlers[button] = handle
	}

}

// SetBrightness .
func SetBrightness(brightness float32, channel *uint8) {

	/*
		Set the brightness of all pixels.

		:param brightness: Brightness: 0.0 to 1.0
	*/

	if brightness < 0 || brightness > 1 {
		log.Println("Brightness should be between 0.0 and 1.0")
		return
	}

	if channel == nil || *channel == 0 {
		for i := 0; i < ChannelPixels; i++ {
			pixels[i].brightness = uint8(31.0*brightness) & 0b11111
		}
	}

	if channel == nil || *channel == 1 {
		for i := 0; i < ChannelPixels; i++ {
			pixels[i+ChannelPixels].brightness = uint8(31.0*brightness) & 0b11111
		}
	}
}

func clear(channel *uint8) {

	if channel == nil || *channel == 0 {
		for i := 0; i < ChannelPixels; i++ {
			pixels[i].r = 0
			pixels[i].g = 0
			pixels[i].b = 0
		}
	}

	if channel == nil || *channel == 1 {
		for i := 0; i < ChannelPixels; i++ {
			pixels[i+ChannelPixels].r = 0
			pixels[i+ChannelPixels].g = 0
			pixels[i+ChannelPixels].b = 0
		}
	}
}

func writeByte(b byte) {
	// Fixme
	dat := gpio.NewPin(Dat)
	clk := gpio.NewPin(Clk)
	for i := 0; i < 8; i++ {
		dat.Write(b&0b10000000 > 1)
		clk.Write(gpio.High)
		time.Sleep(0.0000005 * 1e9)
		clk.Write(gpio.Low)
		time.Sleep(0.0000005 * 1e9)
	}

}

func eof() {

	dat := gpio.NewPin(Dat)
	clc := gpio.NewPin(Clk)
	dat.Write(gpio.High)
	for i := 0; i < 32; i++ {
		clc.Write(gpio.High)
		time.Sleep(0.0000005 * 1e9)
		clc.Write(gpio.Low)
		time.Sleep(0.0000005 * 1e9)
	}
}

func sof() {

	dat := gpio.NewPin(Dat)
	clc := gpio.NewPin(Clk)
	dat.Write(gpio.Low)
	for i := 0; i < 32; i++ {
		clc.Write(gpio.High)
		time.Sleep(0.0000005 * 1e9)
		clc.Write(gpio.Low)
		time.Sleep(0.0000005 * 1e9)
	}
}

// Show .
func Show() {

	sof()

	for _, p := range pixels {

		writeByte(0b11100000 | p.brightness)
		writeByte(p.b)
		writeByte(p.g)
		writeByte(p.r)
	}
}

// SetAll .
func SetAll(r uint8, g uint8, b uint8, brightness *float32, channel *int) {

	/*
			Set the RGB value and optionally brightness of all pixels
			If you don't supply a brightness value, the last value set for each pixel be kept.

		    :param r: Amount of red: 0 to 255
		    :param g: Amount of green: 0 to 255
		    :param b: Amount of blue: 0 to 255
		    :param brightness: Brightness: 0.0 to 1.0 (default around 0.2)
		    :param channel: Optionally specify which bar to set: 0 or 1
	*/

	if channel == nil || *channel == 0 {
		for i := 0; i < ChannelPixels; i++ {
			SetPixel(uint8(i), r, g, b, brightness, nil)
		}

	}

	if channel == nil || *channel == 1 {
		for i := uint8(0); i < ChannelPixels; i++ {
			SetPixel(i+ChannelPixels, r, g, b, brightness, nil)
		}
	}
}

// SetPixel .
func SetPixel(x uint8, r uint8, g uint8, b uint8, brightness *float32, channel *int) {

	/*
			Set the RGB value, and optionally brightness, of a single pixel
			If you don't supply a brightness value, the last value will be kept.

		    :param x: The horizontal position of the pixel: 0 to 7
		    :param r: Amount of red: 0 to 255
		    :param g: Amount of green: 0 to 255
		    :param b: Amount of blue: 0 to 255
		    :param brightness: Brightness: 0.0 to 1.0 (default around 0.2)
		    :param channel: Optionally specify which bar to set: 0 or 1
	*/

	var br uint8
	if brightness == nil {
		br = pixels[x].brightness
	} else {
		br = uint8(31.0*(*brightness)) & 0b11111
	}

	if *channel == 1 || *channel == 0 {
		if x > ChannelPixels {
			return
		}
		x += uint8(*channel * (ChannelPixels))
	}

	if x >= uint8(ChannelPixels) {
		x = Pixels - 1 - (x - ChannelPixels)
	}

	pixels[x].r = r & 0xff
	pixels[x].g = g & 0xff
	pixels[x].b = b & 0xff
	pixels[x].brightness = br
}
