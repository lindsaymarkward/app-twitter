package main

type TwitterAppModel struct {
	Username          string `json:"username"`
	ConsumerKey       string `"json:consumerkey"`
	ConsumerSecret    string `"json:consumersecret"`
	AccessToken       string `json:"accesstoken"`
	AccessTokenSecret string `json:"accesstokensecret"`
}
