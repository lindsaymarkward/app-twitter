package main

// TwitterAppModel stores the details for an account and the stored tweets
type TwitterAppModel struct {
	Account    AccountDetails          `json:"account"`
	Tweets     map[string]TweetDetails `json:"tweets"`
	TweetNames []string                `json:"tweetnames"`
}

// TweetDetails stores the values for one tweet or direct message
// Number is the auto-incrementing value to add to tweets/messages so that Twitter won't reject as duplicates
type TweetDetails struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	To      string `json:"to"`
	Number  int    `json:"number,string"`
}

// AccountDetails stores the authentication details for one user
// (get these from Twitter website, see README)
type AccountDetails struct {
	Username          string `json:"username"`
	ConsumerKey       string `json:"consumerkey"`
	ConsumerSecret    string `json:"consumersecret"`
	AccessToken       string `json:"accesstoken"`
	AccessTokenSecret string `json:"accesstokensecret"`
}
