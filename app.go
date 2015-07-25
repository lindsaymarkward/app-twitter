package main

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")

type TwitterApp struct {
	support.AppSupport
	config     *TwitterAppModel
	twitterAPI *anaconda.TwitterApi
}

// Start is called after the ExportApp call is complete.
func (a *TwitterApp) Start(m *TwitterAppModel) error {
	log.Infof("Starting Twitter app with config: %v", m)
	a.config = m

	a.SendEvent("config", m)
	a.InitTwitterAPI()

	a.Conn.MustExportService(&ConfigService{a}, "$app/"+a.Info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	return a.SendEvent("config", a.config)
}

// Stop
func (a *TwitterApp) Stop() error {
	return nil
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
	return nil
}
