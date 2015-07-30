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

	a.InitTwitterAPI()

	a.Conn.MustExportService(&ConfigService{a}, "$app/"+a.Info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	a.SetupPane()

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

func (a *TwitterApp) SaveAccount(data TwitterAppModel) error {
	log.Infof("Saving username %v\n", data.Username)
	a.config = &data
	// create Twitter API (anaconda) object
	a.InitTwitterAPI()
	return a.SendEvent("config", a.config)
}

func (a *TwitterApp) DeleteAccount(username string) error {
	a.config.Username = ""
	// ?? probably:
	//	a.config = &TwitterAppModel{}
	return a.SendEvent("config", a.config)
}

func (a *TwitterApp) InitTwitterAPI() error {
	anaconda.SetConsumerKey(a.config.ConsumerKey)
	anaconda.SetConsumerSecret(a.config.ConsumerSecret)
	a.twitterAPI = anaconda.NewTwitterApi(a.config.AccessToken, a.config.AccessTokenSecret)
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

func (a *TwitterApp) DoThing(value string) {
	log.Infof("Just doin' nothin' much, %v, %v", a.config.Username, value)
}