package main

type TwitterAppModel struct {
	Account AccountDetails          `json:"account"`
	Tweets  map[string]TweetDetails `json:"tweets"`
}

// TweetDetails stores the values for one tweet or direct message
// Number is the auto-incrementing value to add to tweets/messages so that Twitter won't reject as duplicates
type TweetDetails struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	To      string `json:"to"`
	Number  int    `json:"number,string"`
}

type AccountDetails struct {
	Username          string `json:"username"`
	ConsumerKey       string `json:"consumerkey"`
	ConsumerSecret    string `json:"consumersecret"`
	AccessToken       string `json:"accesstoken"`
	AccessTokenSecret string `json:"accesstokensecret"`
}
