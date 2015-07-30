package main

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/ninjasphere/gestic-tools/go-gestic-sdk"
	"github.com/ninjasphere/sphere-go-led-controller/fonts/O4b03b"
	"github.com/ninjasphere/sphere-go-led-controller/util"
)

// TODO - states: error, each tweet, sending (with timer)
// error: check API - can I draw animated red X over top of Twitter pic?

var tapInterval = time.Millisecond * 500
var introDuration = time.Millisecond * 1500

// load state images
var imageLogo = util.LoadImage(util.ResolveImagePath("twitter-bird.png"))
var imageAnimated = util.LoadImage(util.ResolveImagePath("twitter-animated.gif"))
var imageError = util.LoadImage(util.ResolveImagePath("error.gif"))

// states
const (
	Intro = iota
	ErrorAccount
	ErrorTweet
	Tweeting
	Choosing
)

// LEDPane stores the data we want to access
type LEDPane struct {
	lastTap         time.Time
	lastDoubleTap   time.Time
	lastTapLocation gestic.Location
	introTimeout    *time.Timer
	visible         bool
	currentImage    util.Image
	app             *TwitterApp
	state           int
}

// NewLEDPane creates an LEDPane with the data and timers initialised
// the app is passed in so that the pane can access the data and methods in it
func NewLEDPane(a *TwitterApp) *LEDPane {

	pane := &LEDPane{
		lastTap: time.Now(),
		app:     a,
		state:   Intro,
	}

	pane.introTimeout = time.AfterFunc(0, func() {
		pane.state = Choosing
	})

	return pane
}

// Gesture is called by the system when the LED matrix receives any kind of gesture
func (p *LEDPane) Gesture(gesture *gestic.GestureMessage) {
	//	log.Infof("gesture received - %v, %v", gesture.Touch, gesture.Position)

	// check the second last touch location since the most recent one before a tap is usually blank it seems
	lastLocation := p.lastTapLocation
	p.lastTapLocation = gesture.Touch

	if gesture.Tap.Active() && time.Since(p.lastTap) > tapInterval {
		p.lastTap = time.Now()

		log.Infof("Tap! %v", lastLocation)

		// check state of display to know what action to do
		// TODO - states for actions (configured tweets)

		// change between images - right or left
		if lastLocation.East && !lastLocation.West {
			p.currentImage = imageLogo
		} else {
			p.currentImage = imageAnimated
		}
	}

	if gesture.DoubleTap.Active() && time.Since(p.lastDoubleTap) > tapInterval {
		p.lastDoubleTap = time.Now()
		log.Infof("Double Tap!")

		// TODO - state with timer

		p.currentImage = imageAnimated
		// TODO - learn why I need "go" here or the LED connection gets lost
		// ("WARNING matrix RemoteMatrix.go:70 Lost connection to led controller: EOF")
		//		go p.app.PostDirectMessage("Nice one? Well, I hope so!", "@lindsaymarkward")
	}
}

// KeepAwake sets whether the display fades (false) or not (true)
func (p *LEDPane) KeepAwake() bool {
	return false
}

// IsEnabled is needed as it's part of the remote.pane interface
func (p *LEDPane) IsEnabled() bool {
	return true
}

// Render is called by the system repeatedly when the pane is visible
// It should return the RGBA image to be rendered on the LED matrix
func (p *LEDPane) Render() (*image.RGBA, error) {

	if !p.visible {
		p.visible = true
		p.state = Intro
		p.introTimeout.Reset(introDuration)
	}

	if p.state == Intro {
		p.currentImage = imageLogo
		// simply return the logo image
		return imageLogo.GetNextFrame(), nil
	}
	// create an empty 16*16 RGBA image for the Draw function to draw into (to be returned)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	// TODO - where's the best place for this? Update function called less frequently?
	if !p.app.Initialised {
		p.state = ErrorAccount
	} else {
		p.state = Choosing
	}

	switch p.state {

	case Tweeting:
	// animated
	case Choosing:
		// different tweet images/text
		p.currentImage = imageLogo
	case ErrorAccount:
		// @ with cross through it
		p.currentImage = imageError
	case ErrorTweet:
		// (animated?) bird with cross
	}

	// Draw (built-in Go function) draws the frame from stateImg into the img 'image' starting at 4th parameter, "Over" the top
	draw.Draw(img, img.Bounds(), p.currentImage.GetNextFrame(), image.Point{0, 0}, draw.Over)

	//	// draw the index up the top
	//	drawText(fmt.Sprintf("%2d", p.imageIndex), color.RGBA{10, 250, 250, 255}, 2, img)
	//	// draw the text from app down the bottom
	//	drawText(p.app.config.Account.Username, color.RGBA{253, 151, 32, 255}, 9, img)

	// return the image we've created to be rendered to the matrix
	return img, nil
}

// drawText is a helper function to draw a string of text into an image
func drawText(text string, col color.RGBA, top int, img *image.RGBA) {
	width := O4b03b.Font.DrawString(img, 0, 8, text, color.Black)
	start := int(16 - width - 1)

	O4b03b.Font.DrawString(img, start, top, text, col)
}
