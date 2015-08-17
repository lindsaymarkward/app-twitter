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

// TODO maybe - scrolling text for tweet name / messages?
// TODO - start timer on tap, stop it on double-tap, do tap action if it fires

var tapInterval = time.Millisecond * 450
var updateFrequency = time.Second * 2

// states
const (
	ErrorAccount = iota
	Choosing
	Tweeting
	TweetFailed
	TweetSucceeded
)

// load state images
var images map[string]util.Image

// init runs before anything else, and loads the images for the LED pane
func init() {
	images = make(map[string]util.Image)
	images["logo"] = util.LoadImage(util.ResolveImagePath("twitter-bird.png"))
	images["animated"] = util.LoadImage(util.ResolveImagePath("twitter-animated.gif"))
	images["error"] = util.LoadImage(util.ResolveImagePath("errorX.gif"))
	images["at"] = util.LoadImage(util.ResolveImagePath("at.gif"))
	images["tick"] = util.LoadImage(util.ResolveImagePath("tick.gif"))
}

// LEDPane stores the data we want to access
type LEDPane struct {
	lastTap              time.Time
	lastDoubleTap        time.Time
	lastTapLocation      gestic.Location
	changeTweetDirection int
	currentImage         util.Image
	app                  *TwitterApp
	state                int
	hasStoredTweets      bool
	numberOfTweets       int
	currentTweetNumber   int
	updateTimer          *time.Timer
	tapTimer             *time.Timer
}

// NewLEDPane creates an LEDPane with the data and timers initialised
// the app is passed in so that the pane can access the data and methods in it
func NewLEDPane(a *TwitterApp) *LEDPane {
	p := &LEDPane{
		lastTap:         time.Now(),
		lastDoubleTap:   time.Now(),
		app:             a,
		hasStoredTweets: false,
		numberOfTweets:  1, // to avoid divide by zero error the first time it's run
	}

	p.updateTimer = time.AfterFunc(0, p.UpdateStatus)
	p.tapTimer = time.AfterFunc(0, p.TapAction)
	return p
}

// Gesture is called by the system when the LED matrix receives any kind of gesture
func (p *LEDPane) Gesture(gesture *gestic.GestureMessage) {
	//	log.Infof("gesture received - %v, %v", gesture.Touch, gesture.Position)
	//	log.Infof("Touch %v, Tap %v, Since: %v, Double %v, Since %v", gesture.Touch.Active(), gesture.Tap.Active(), time.Since(p.lastTap), gesture.DoubleTap.Active(), time.Since(p.lastDoubleTap))

	// check the second last touch location because the most recent one before a tap is usually blank it seems
	lastLocation := p.lastTapLocation
	p.lastTapLocation = gesture.Touch

	if gesture.Tap.Active() && time.Since(p.lastTap) > tapInterval {
		p.lastTap = time.Now()
		log.Infof("Tap! %v", lastLocation)

		// do tap action only if we are in the right state
		if p.state == Choosing && p.hasStoredTweets {
			// start timer that will be stopped if double tap happens in time
			// this avoids the problem of the first tap of a double being actioned as a tap
			p.tapTimer.Reset(tapInterval)
			// change between images - right or left
			if lastLocation.West && !lastLocation.East {
				p.changeTweetDirection = -1
			} else {
				p.changeTweetDirection = 1
			}
		}
	}

	if gesture.DoubleTap.Active() && time.Since(p.lastDoubleTap) > tapInterval {
		p.lastDoubleTap = time.Now()
		log.Infof("Double Tap!")
		if p.state == Choosing {
			// don't do tap action since we're double tapping
			p.tapTimer.Stop()
			// TODO - learn why I need "go" here or the LED connection gets lost
			// ("WARNING matrix RemoteMatrix.go:70 Lost connection to led controller: EOF")
			//		go p.app.PostDirectMessage("Nice one? I hope so!", "@lindsaymarkward")

			go p.tweetIt()
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

	// create an empty 16*16 RGBA image for the Draw function to draw into (to be returned)
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	switch p.state {
	case Tweeting:
		draw.Draw(img, img.Bounds(), images["animated"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		O4b03b.Font.DrawString(img, 6, 3, fmt.Sprintf("%d", p.currentTweetNumber+1), color.RGBA{20, 154, 233, 255})
	case Choosing:
		// different tweet numbers and DM or TWT text
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
	case ErrorAccount:
		// @ with animated cross through it
		draw.Draw(img, img.Bounds(), images["at"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		draw.Draw(img, img.Bounds(), images["error"].GetNextFrame(), image.Point{0, 0}, draw.Over)
	case TweetSucceeded:
		// bird with animated tick and tweet number
		draw.Draw(img, img.Bounds(), images["logo"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		draw.Draw(img, img.Bounds(), images["tick"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		O4b03b.Font.DrawString(img, 6, 3, fmt.Sprintf("%d", p.currentTweetNumber+1), color.RGBA{255, 255, 255, 255})
	case TweetFailed:
		// bird with animated cross through it
		draw.Draw(img, img.Bounds(), images["logo"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		draw.Draw(img, img.Bounds(), images["error"].GetNextFrame(), image.Point{0, 0}, draw.Over)
		O4b03b.Font.DrawString(img, 6, 3, fmt.Sprintf("%d", p.currentTweetNumber+1), color.RGBA{255, 255, 255, 255})
	}
	// return the image we've created to be rendered to the matrix
	return img, nil
}

// drawText is a helper function to draw a string of text into an image
// this actually draws black to determine width, then aligns to the right
func drawText(text string, col color.RGBA, top int, img *image.RGBA) {
	width := O4b03b.Font.DrawString(img, 0, 8, text, color.Black)
	start := int(16 - width - 1)
	O4b03b.Font.DrawString(img, start, top, text, col)
}

// UpdateStatus (regularly) checks the account (API) initialisation status and number of tweets stored
// and sets the pane state accordingly.
// This gets updated regularly so you don't have to restart the app when you update the config
func (p *LEDPane) UpdateStatus() {
	if !p.app.Initialised {
		p.state = ErrorAccount
	} else {
		p.state = Choosing
		p.numberOfTweets = len(p.app.config.Tweets)
		if p.numberOfTweets == 0 {
			p.currentTweetNumber = -1
			p.hasStoredTweets = false
		} else if p.hasStoredTweets == false {
			// this is the first update where there are now stored tweets
			p.currentTweetNumber = 0
			p.hasStoredTweets = true
		}
	}
	//	log.Infof("update. State is %v", p.state)
	p.updateTimer.Reset(updateFrequency)
}

// TapAction changes to the next/previous stored tweet (run on a timer when tapped)
func (p *LEDPane) TapAction() {
	p.currentTweetNumber += p.changeTweetDirection
	p.currentTweetNumber %= p.numberOfTweets
	if p.currentTweetNumber < 0 {
		p.currentTweetNumber = p.numberOfTweets - 1
	}
}

// tweetIt calls app's appropriate function to post tweet or direct message
// should handle result and set state
func (p *LEDPane) tweetIt() {
	var err error
	// stop the regular status updating while we tweet and handle success/failure
	p.updateTimer.Stop()
	p.state = Tweeting
	// TODO - timer for tweeting that animates until result determined - tick or cross for success or fail

	tweet := p.app.config.Tweets[p.app.config.TweetNames[p.currentTweetNumber]]
	tweet.Number += 1
	// update config to update this number (to avoid Twitter rejecting duplicate tweets/messages)
	p.app.config.Tweets[p.app.config.TweetNames[p.currentTweetNumber]] = tweet
	p.app.SendEvent("config", p.app.config)

	log.Infof("Tweeting: %v to %v (%v)", tweet.Message, tweet.To, tweet.Number)

	message := fmt.Sprintf("%s %d", tweet.Message, tweet.Number)
	if tweet.To == "" {
		// post tweet
		err = p.app.PostTweet(message)
	} else {
		// send direct message
		err = p.app.PostDirectMessage(message, tweet.To)
	}
	// handle error - success/fail display
	if err != nil {
		//		log.Errorf(fmt.Sprintf("Tweetit error: %v", err))
		p.state = TweetFailed
	} else {
		p.state = TweetSucceeded
	}
	// reset usual timer which will set state (so it displays success/fail for 2 seconds)
	p.updateTimer.Reset(updateFrequency)
}
