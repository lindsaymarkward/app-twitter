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

var tapInterval = time.Millisecond * 500

// states
const (
	Intro = iota
	ErrorAccount
	ErrorTweet
	Tweeting
	Choosing
)

// load state images
var images map[string]util.Image
func init() {
	images = make(map[string]util.Image)
	images["logo"] = util.LoadImage(util.ResolveImagePath("twitter-bird.png"))
	images["animated"] = util.LoadImage(util.ResolveImagePath("twitter-animated.gif"))
	images["error"] = util.LoadImage(util.ResolveImagePath("canceling.gif")) // not...
	images["at"] = util.LoadImage(util.ResolveImagePath("at.png"))
}

// LEDPane stores the data we want to access
type LEDPane struct {
	lastTap         time.Time
	lastDoubleTap   time.Time
	lastTapLocation gestic.Location
	currentImage    util.Image
	app             *TwitterApp
	state           int
	updateTimer		*time.Timer
}

// NewLEDPane creates an LEDPane with the data and timers initialised
// the app is passed in so that the pane can access the data and methods in it
func NewLEDPane(a *TwitterApp) *LEDPane {

	pane := &LEDPane{
		lastTap: time.Now(),
		app:     a,
		state:   Intro,
	}

	if !a.Initialised {
		pane.state = ErrorAccount
	}

	pane.updateTimer = time.AfterFunc(0, func() {
		if !a.Initialised {
			pane.state = ErrorAccount
		} else {
			pane.state = Choosing
		}
		log.Infof("update timer running... state is %v", pane.state)
	})

//	pane.introTimeout = time.AfterFunc(0, func() {
//		pane.state = Choosing
//	})

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
			p.currentImage = images["logo"]
		} else {
			p.currentImage = images["animated"]
		}
	}

	if gesture.DoubleTap.Active() && time.Since(p.lastDoubleTap) > tapInterval {
		p.lastDoubleTap = time.Now()
		log.Infof("Double Tap!")

		// TODO - state with timer

		p.currentImage = images["animated"]
		// TODO - learn why I need "go" here or the LED connection gets lost
		// ("WARNING matrix RemoteMatrix.go:70 Lost connection to led controller: EOF")
		//		go p.app.PostDirectMessage("Nice one? Well, I hope so!", "@lindsaymarkward")
	}
}

// KeepAwake sets whether the display fades after 30 seconds (false) or stays on (true)
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

	log.Infof("State: %v", p.state)

//	p.UpdateStatus() // ??

	// create an empty 16*16 RGBA image for the Draw function to draw into (to be returned)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	switch p.state {
	case Intro:
		return images["logo"].GetNextFrame(), nil
	case Tweeting:
	// animated
	case Choosing:
		// different tweet images/text
		if len(p.app.config.Tweets) == 0 {
			// no tweets are saved

		}
		p.currentImage = images["logo"]
	case ErrorAccount:
		// @ with cross through it
		draw.Draw(img, img.Bounds(), images["at"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		draw.Draw(img, img.Bounds(), images["error"].GetNextFrame(), image.Point{0, 0}, draw.Over)
	case ErrorTweet:
		// (animated?) bird with cross
	}

	// Draw (built-in Go function) draws the frame from stateImg into the img 'image' starting at 4th parameter, "Over" the top
//	draw.Draw(img, img.Bounds(), p.currentImage.GetNextFrame(), image.Point{0, 0}, draw.Over)

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

func (p *LEDPane) UpdateStatus() {
	// TODO - where's the best place for this? Update function called less frequently?
	if !p.app.Initialised {
		p.state = ErrorAccount
	} else {
		p.state = Choosing
	}
}
