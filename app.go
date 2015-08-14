package main

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

type TwitterApp struct {
	support.AppSupport
	led         *remote.Matrix
	config      *TwitterAppModel
	twitterAPI  *anaconda.TwitterApi
	Initialised bool
}

// Start is called after the ExportApp call is complete.
func (a *TwitterApp) Start(m *TwitterAppModel) error {
	log.Infof("Starting Twitter app with config: %v", m)
	a.config = m

	// for clearing tweets (testing)
//	a.config.TweetNames = nil
//	a.config.Tweets = nil

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

	return a.SetupPane()
}

// Stop
func (a *TwitterApp) Stop() error {
	return nil
}

func (a *TwitterApp) SetupPane() error {
	log.Infof("Making new pane...")
	// The pane must implement the remote.pane interface
	pane := NewLEDPane(a)

	// Export our pane over the TCP connection we just made
	a.led = remote.NewTCPMatrix(pane, fmt.Sprintf("%s:%d", host, port))
	return nil
}

func (a *TwitterApp) SaveAccount(account AccountDetails) error {
	log.Infof("Saving username %v\n", account.Username)

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

// TODO maybe - multiple Twitter accounts would require multiple APIs (or switching between)... (just using one for now)

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
	log.Infof("Initialised Twitter API with self: %v", user.ScreenName)
	a.Initialised = true
	return nil
}

func (a *TwitterApp) PostTweet(tweet string) error {
	result, err := a.twitterAPI.PostTweet(tweet, nil)
	if err != nil {
		log.Errorf("Error posting Tweet %v", err)
	}
	// TODO - probably remove this result logging
	log.Infof("%v", result)
	return err
}

func (a *TwitterApp) PostDirectMessage(message, user string) error {
	result, err := a.twitterAPI.PostDMToScreenName(message, user)
	if err != nil {
		log.Errorf("Error sending direct message %v", err)
	}
	log.Infof("%v", result)
	return err
}
