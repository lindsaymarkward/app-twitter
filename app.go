package main

// TODO - Ninja channels stuff so this can be used by other apps/drivers

import (
	"fmt"

	"github.com/ChimeraCoder/anaconda"
	"github.com/lindsaymarkward/go-ninja/config"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
	"github.com/ninjasphere/sphere-go-led-controller/remote"
)

var info = ninja.LoadModuleInfo("./package.json")
var host = config.String("localhost", "led.host")
var port = config.Int(3115, "led.remote.port")

// TwitterApp stores the app's core details including the Initialised boolean for whether authentication (API) worked
type TwitterApp struct {
	support.AppSupport
	led            *remote.Matrix
	config         *TwitterAppModel
	twitterAPI     *anaconda.TwitterApi
	Initialised    bool
}

// Start the app, set up Twitter API, create LED pane
func (a *TwitterApp) Start(m *TwitterAppModel) error {
	log.Infof("Starting Twitter app with config: %v", m)
	if m != nil {
		a.config = m
	} else {
		a.config =&TwitterAppModel{ }
	}
	// TODO - (temporary solution) - configure update tweets frequency
	a.config.CheckTweetsFrequency = 8

	// for clearing tweets (testing)
//		a.config.TweetNames = nil
//		a.config.Tweets = nil

	// initialise Twitter API and set Initialised state (don't try if account isn't set)
	if a.config.Account.Username != "" {
		// check
		go a.InitTwitterAPI(a.config.Account)
	} else {
		a.Initialised = false
	}

	a.Conn.MustExportService(&ConfigService{a}, "$app/"+a.Info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	log.Infof("Making new pane for Twitter...")
	pane := NewLEDPane(a)

	// Export our newly made pane
	a.led = remote.NewTCPMatrix(pane, fmt.Sprintf("%s:%d", host, port))

	return nil
}

// Stop the app - sort of, not really
func (a *TwitterApp) Stop() error {
	return nil
}

// SaveAccount saves the account to the config and initialises the Twitter API
func (a *TwitterApp) SaveAccount(account AccountDetails) error {
	log.Infof("Saving account with username %v\n", account.Username)

	// for multiple accounts...
	//	if a.config.Accounts == nil {
	//		a.config.Accounts = make(map[string]AccountDetails)
	//	}
	//	a.config.Accounts[account.Username] = account

	a.config.Account = account
	// create Twitter API (anaconda) object
	a.InitTwitterAPI(account)
	return a.SendEvent("config", a.config)
}

// InitTwitterAPI creates a new Twitter API object using the account details
func (a *TwitterApp) InitTwitterAPI(account AccountDetails) error {
	anaconda.SetConsumerKey(account.ConsumerKey)
	anaconda.SetConsumerSecret(account.ConsumerSecret)
	a.twitterAPI = anaconda.NewTwitterApi(account.AccessToken, account.AccessTokenSecret)
	user, err := a.twitterAPI.GetSelf(nil)
	if err != nil {
		log.Infof("Error initialising Twitter API: %v", err)
		a.Initialised = false
		return err
	}
	log.Infof("Initialised Twitter API with username: %v", user.ScreenName)
	a.Initialised = true
	return nil
}

// PostTweet sends message as a regular public tweet
func (a *TwitterApp) PostTweet(message string) error {
	_, err := a.twitterAPI.PostTweet(message, nil)
	if err != nil {
		log.Errorf("Error posting Tweet: %v", err)
		//		log.Infof("Twitter API result: %#v", result)
	}
	return err
}

// PostDirectMessage sends message to user as a direct message
func (a *TwitterApp) PostDirectMessage(message, user string) error {
	_, err := a.twitterAPI.PostDMToScreenName(message, user)
	if err != nil {
		log.Errorf("Error sending direct message: %v", err)
		//		log.Infof("Twitter API result: %#v", result)
	}
	return err
}
