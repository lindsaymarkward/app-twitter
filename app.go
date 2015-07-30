package main

import (
	"fmt"
	"net"
	"os"

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
	led        *remote.Matrix
	config     *TwitterAppModel
	twitterAPI *anaconda.TwitterApi
}

// Start is called after the ExportApp call is complete.
func (a *TwitterApp) Start(m *TwitterAppModel) error {
	log.Infof("Starting Twitter app with config: %v", m)
	a.config = m

	// TODO - get key of first one
	a.InitTwitterAPI(a.config.Account)

	a.Conn.MustExportService(&ConfigService{a}, "$app/"+a.Info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	a.SetupPane()

	// TODO - do I need to do this again (nothing's changed... persists?)
	return a.SendEvent("config", a.config)
}

// Stop
func (a *TwitterApp) Stop() error {
	return nil
}

func (a *TwitterApp) SetupPane() {
	log.Infof("Making new pane...")
	// The pane must implement the remote.pane interface
	pane := NewLEDPane(a)

	// Connect to the LED controller remote pane interface via TCP
	log.Infof("Connecting to LED controller...")
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	// This creates a TCP connection, conn
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("DialTCP failed:", err.Error())
		os.Exit(1)
	}

	log.Infof("Connected. Now making new matrix...")

	// Export our pane over the TCP connection we just made
	a.led = remote.NewMatrix(pane, conn)
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

// TODO - multiple Twitter accounts would require multiple APIs... (just using one for now)

func (a *TwitterApp) InitTwitterAPI(account AccountDetails) error {
	anaconda.SetConsumerKey(account.ConsumerKey)
	anaconda.SetConsumerSecret(account.ConsumerSecret)
	a.twitterAPI = anaconda.NewTwitterApi(account.AccessToken, account.AccessTokenSecret)
	user, err := a.twitterAPI.GetSelf(nil)
	if err != nil {
		log.Infof("Error initialising Twitter API: %v", err)
		return err
	}
	log.Infof("Initialised Twitter API with self: %v", user.ScreenName)
	return nil
}

func (a *TwitterApp) PostTweet(tweet string) {
	result, err := a.twitterAPI.PostTweet(tweet, nil)
	if err != nil {
		log.Errorf("Error posting Tweet %v", err)
	}
	log.Infof("%v", result)
}

func (a *TwitterApp) PostDirectMessage(message, user string) {
	result, err := a.twitterAPI.PostDMToScreenName(message, user)
	if err != nil {
		log.Errorf("Error sending direct message %v", err)
	}
	log.Infof("%v", result)
}
