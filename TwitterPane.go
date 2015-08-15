package main

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"fmt"

	"github.com/ninjasphere/gestic-tools/go-gestic-sdk"
	"github.com/ninjasphere/sphere-go-led-controller/fonts/O4b03b"
	"github.com/ninjasphere/sphere-go-led-controller/util"
)

// TODO - states: error, each tweet, sending (with timer)
// TODO maybe - scrolling text for tweet?

var tapInterval = time.Millisecond * 900
var updateFrequency = time.Second * 2

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
	images["error"] = util.LoadImage(util.ResolveImagePath("errorX.gif")) // not...
	images["at"] = util.LoadImage(util.ResolveImagePath("at.gif"))
}

// LEDPane stores the data we want to access
type LEDPane struct {
	lastTap            time.Time
	lastDoubleTap      time.Time
	lastTapLocation    gestic.Location
	currentImage       util.Image
	app                *TwitterApp
	state              int
	hasStoredTweets    bool
	numberOfTweets     int
	currentTweetNumber int
	updateTimer        *time.Timer
}

// NewLEDPane creates an LEDPane with the data and timers initialised
// the app is passed in so that the pane can access the data and methods in it
func NewLEDPane(a *TwitterApp) *LEDPane {

	pane := &LEDPane{
		state:           Intro,
		lastTap:         time.Now(),
		app:             a,
		hasStoredTweets: false,
	}

	pane.updateTimer = time.AfterFunc(0, pane.UpdateStatus)
	//	pane.updateTimer = time.AfterFunc(0, func() {
	//		if !a.Initialised {
	//			pane.state = ErrorAccount
	//		} else {
	//			pane.state = Choosing
	//		}
	//		log.Infof("update timer running... state is %v", pane.state)
	//	})

	//	pane.introTimeout = time.AfterFunc(0, func() {
	//		pane.state = Choosing
	//	})

	return pane
}

// Gesture is called by the system when the LED matrix receives any kind of gesture
func (p *LEDPane) Gesture(gesture *gestic.GestureMessage) {
	//	log.Infof("gesture received - %v, %v", gesture.Touch, gesture.Position)

	// check the second last touch location because the most recent one before a tap is usually blank it seems
	lastLocation := p.lastTapLocation
	p.lastTapLocation = gesture.Touch

	log.Infof("Tap %v, Since: %v, Double %v, Since %v", gesture.Tap.Active(), time.Since(p.lastTap), gesture.DoubleTap.Active(), time.Since(p.lastDoubleTap))

	if gesture.Tap.Active() && time.Since(p.lastTap) > tapInterval {
		p.lastTap = time.Now()

		log.Infof("Tap! %v", lastLocation)

		// check state of display to know what action to do
		// TODO - states for actions (configured tweets)

		if p.state == Choosing && p.hasStoredTweets {
			// change between images - right or left
			if lastLocation.West && !lastLocation.East {
				p.currentTweetNumber -= 1
				if p.currentTweetNumber < 0 {
					p.currentTweetNumber = p.numberOfTweets - 1
				}
			} else {
				p.currentTweetNumber += 1
				p.currentTweetNumber %= p.numberOfTweets
			}
		}
	}

	if gesture.DoubleTap.Active() && time.Since(p.lastDoubleTap) > tapInterval {
		p.lastDoubleTap = time.Now()
		log.Infof("Double Tap!")
		if p.state == Choosing {
			p.state = Tweeting
			// TODO - timer for tweeting, return to Choosing - DONE?
			tweet := p.app.config.Tweets[p.app.config.TweetNames[p.currentTweetNumber]]
			tweet.Number += 1
			// update config to update this number
			p.app.config.Tweets[p.app.config.TweetNames[p.currentTweetNumber]] = tweet
			p.app.SendEvent("config", p.app.config)

			log.Infof("Tweeting: %v to %v (%v)", tweet.Message, tweet.To, tweet.Number)

			// TODO - learn why I need "go" here or the LED connection gets lost
			// ("WARNING matrix RemoteMatrix.go:70 Lost connection to led controller: EOF")
			//		go p.app.PostDirectMessage("Nice one? Well, I hope so!", "@lindsaymarkward")
			go p.tweetIt(tweet)
		}
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

	//	log.Infof("State: %v", p.state)

	//	p.UpdateStatus() // ??

	// create an empty 16*16 RGBA image for the Draw function to draw into (to be returned)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	switch p.state {
	case Intro:
		return images["logo"].GetNextFrame(), nil
	case Tweeting:
		draw.Draw(img, img.Bounds(), images["animated"].GetNextFrame(), image.Point{0, 0}, draw.Over)
	// ?? TODO ??
	// restart timer for update so animation shows for full time
	//		p.updateTimer.Reset(updateFrequency)

	case Choosing:
		// different tweet images/text
		draw.Draw(img, img.Bounds(), images["logo"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		if !p.hasStoredTweets {
			O4b03b.Font.DrawString(img, 4, 5, "NO", color.RGBA{255, 0, 0, 255})
//			drawText("NO", color.RGBA{255, 250, 0, 255}, 2, img)
		} else {
			// display tweet number and type on Spheramid
//			drawText(fmt.Sprintf("%d", p.currentTweetNumber+1), color.RGBA{255, 250, 0, 255}, 2, img)
			O4b03b.Font.DrawString(img, 6, 3, fmt.Sprintf("%d", p.currentTweetNumber+1), color.RGBA{255, 250, 0, 255})
			if p.app.config.Tweets[p.app.config.TweetNames[p.currentTweetNumber]].To == "" {
				O4b03b.Font.DrawString(img, 2, 10, "TWT", color.RGBA{20, 255, 20, 255})
			} else {
				O4b03b.Font.DrawString(img, 3, 10, "DM", color.RGBA{20, 255, 250, 255})
			}
		}
		//	p.app.config.Tweets
	case ErrorAccount:
		// @ with cross through it
		draw.Draw(img, img.Bounds(), images["at"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		draw.Draw(img, img.Bounds(), images["error"].GetNextFrame(), image.Point{0, 0}, draw.Over)
	case ErrorTweet:
		// (animated?) bird with cross
	}
	// return the image we've created to be rendered to the matrix
	return img, nil
}

// drawText is a helper function to draw a string of text into an image
func drawText(text string, col color.RGBA, top int, img *image.RGBA) {
	// this actually draws black to determine width, then aligns to the right
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
		p.numberOfTweets = len(p.app.config.Tweets)
		if p.numberOfTweets == 0 {
			p.currentTweetNumber = -1
			p.hasStoredTweets = false
		} else {
			if p.hasStoredTweets == false {
				// this is the first update where there are now stored tweets
				p.currentTweetNumber = 0
				p.hasStoredTweets = true
			}
		}
	}
	//	log.Infof("update. State is %v", p.state)
	p.updateTimer.Reset(updateFrequency)
}

func (p *LEDPane) tweetIt(tweet TweetDetails) {
	var err error
	message := fmt.Sprintf("%s %d", tweet.Message, tweet.Number)
	if tweet.To == "" {
		// post tweet
		err = p.app.PostTweet(message)
		//		log.Infof(fmt.Sprintf("%s %d", tweet.Message, tweet.Number))
	} else {
		// send direct message
		err = p.app.PostDirectMessage(message, tweet.To)
	}
	// handle error
	log.Errorf(fmt.Sprintf("%v", err))
	// TODO - this isn't the right error... needs to look at result returned from Twitter (e.g. "you already said that")
}
