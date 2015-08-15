package main

import (
	"encoding/json"
	"fmt"

	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

// TODO: (if useful) make config handle multiple accounts

type ConfigService struct {
	app *TwitterApp
}

// GetActions is called by the Ninja Sphere system and returns the actions that this driver performs
func (c *ConfigService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Label:       "Twitter",
			DisplayIcon: "twitter",
		},
	}, nil
}

// Configure is the handler for all configuration screen requests
func (c *ConfigService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	log.Infof("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "":
		if c.app.config.Account.Username != "" {
			return c.listTweets()
		}
		fallthrough
	case "listAccounts":
		// present the existing or new Twitter Account screen
		if c.app.config.Account.Username != "" {
			return c.listAccounts()
		}
		fallthrough
	case "newAccount":
		return c.editAccount(&TwitterAppModel{})

	case "editAccount":
		return c.editAccount(c.app.config)

	case "saveAccount":
		configData := &AccountDetails{}
		err := json.Unmarshal(request.Data, configData)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// check and add @ if needed
		if len(configData.Username) > 0 && configData.Username[0] != '@' {
			configData.Username = "@" + configData.Username
		}
		err = c.app.SaveAccount(*configData)
		if err != nil {
			return c.error(fmt.Sprintf("Could not save Twitter Account: %s", err))
		}

		return c.listAccounts()

	case "confirmDelete":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal confirm delete config request %s: %s", request.Data, err))
		}
		return c.confirmDeleteAccount(values["username"])

	case "delete":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal delete config request %s: %s", request.Data, err))
		}
		// set username to blank, save config, load new account screen
		c.app.config.Account.Username = ""
		c.app.SendEvent("config", c.app.config)
		return c.editAccount(&TwitterAppModel{})

	case "confirmDeleteTweet":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal confirm delete tweet config request %s: %s", request.Data, err))
		}
		return c.confirmDeleteTweet(values["tweetName"])

	case "deleteTweet":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal delete tweet config request %s: %s", request.Data, err))
		}
		// remove tweet from map and slice, save config
		delete(c.app.config.Tweets, values["tweetName"])

		i := indexOf(c.app.config.TweetNames, values["tweetName"])
		c.app.config.TweetNames = append(c.app.config.TweetNames[:i], c.app.config.TweetNames[i+1:]...)

		c.app.SendEvent("config", c.app.config)
		return c.listTweets()

	case "listTweets":
		return c.listTweets()

	case "newTweet":
		return c.editTweet("")

	case "editTweet":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal editTweet config request %s: %s", request.Data, err))
		}
		return c.editTweet(values["tweetName"])

	case "saveTweet":
		var values TweetDetails
		//		var result
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// We could check the length of the message here but giving an error would mean user had to start again
		// So we just label it when displaying it

		// check and add @ to To field if needed
		if len(values.To) > 0 && values.To[0] != '@' {
			values.To = "@" + values.To
		}

		// add tweet (map and slice) and save config (make new map &slice if no tweets exist yet)
		if c.app.config.Tweets == nil {
			c.app.config.Tweets = make(map[string]TweetDetails)
			c.app.config.TweetNames = make([]string, 0)
		}
		c.app.config.Tweets[values.Name] = values
		c.app.config.TweetNames = append(c.app.config.TweetNames, values.Name)
		c.app.SendEvent("config", c.app.config)
		return c.listTweets()

	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

// error is a generic config screen for displaying error messages
func (c *ConfigService) error(message string) (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Error",
						Subtitle:     message,
						DisplayClass: "danger",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Back",
				Name:  "listAccounts",
			},
		},
	}, nil
}

// list is a config screen for displaying accounts with options for editing, deleting and controlling
func (c *ConfigService) listAccounts() (*suit.ConfigurationScreen, error) {
	subtitle := ""
	if !c.app.Initialised {
		subtitle = "INVALID ACCOUNT!"
	}
	screen := suit.ConfigurationScreen{
		Title: "Twitter App Config",
		Sections: []suit.Section{
			suit.Section{
				Title: "Edit Account",
				Contents: []suit.Typed{
					suit.ActionList{
						Name: "account",
						Options: []suit.ActionListOption{
							suit.ActionListOption{
								Title:    c.app.config.Account.Username,
								Subtitle: subtitle,
							},
						},
						PrimaryAction: &suit.ReplyAction{
							Name:        "editAccount",
							DisplayIcon: "pencil",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "confirmDelete",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label:        "Tweets",
				Name:         "listTweets",
				DisplayIcon:  "twitter",
				DisplayClass: "info",
			},
			suit.ReplyAction{
				Label:        "New Account",
				Name:         "newAccount",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

// listTweets is a config screen for displaying tweets with options for editing, deleting and creating new ones
func (c *ConfigService) listTweets() (*suit.ConfigurationScreen, error) {
	var tweetOptions []suit.ActionListOption
	for i, tweetName := range c.app.config.TweetNames {
		subtitle := ""
		tweet := c.app.config.Tweets[c.app.config.TweetNames[i]]
		// create edit actions
		if len(tweet.Message) > 137 {
			subtitle = "TOO LONG!"
		} else if tweet.To != "" {
			subtitle = "DM"
		}
		tweetOptions = append(tweetOptions, suit.ActionListOption{
			Title:    fmt.Sprintf("%d-%s", i+1, tweetName),
			Subtitle: subtitle,
			Value:    tweetName,
		})
	}
	screen := suit.ConfigurationScreen{
		Title: "Tweets",
		Sections: []suit.Section{
			suit.Section{
				Title: "Create or Edit Tweets",
				Contents: []suit.Typed{
					suit.StaticText{
						// TODO - could improve this process if needed
						Value: "To rename a tweet, edit it, save with a different name, then delete the one with the old name",
					},
					suit.ActionList{
						Name:    "tweetName",
						Options: tweetOptions,
						PrimaryAction: &suit.ReplyAction{
							Name:        "editTweet",
							DisplayIcon: "pencil",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "confirmDeleteTweet",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label:        "Accounts",
				Name:         "listAccounts",
				DisplayClass: "info",
				DisplayIcon:  "at",
			},
			suit.ReplyAction{
				Label:        "New Tweet",
				Name:         "newTweet",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}
	return &screen, nil
}

func (c *ConfigService) editTweet(tweetName string) (*suit.ConfigurationScreen, error) {
	tweet := TweetDetails{}
	title := "New Tweet/Message"
	if tweetName != "" {
		title = "Edit Tweet/Message"
		tweet = c.app.config.Tweets[tweetName]
	}
	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				//				Title: "Tweet",
				Contents: []suit.Typed{
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "Give this tweet/message a name to identify it",
						Value:       tweet.Name,
					},
					suit.InputText{
						Name:        "message",
						Before:      "Message",
						Placeholder: "Up to 140 characters",
						Value:       tweet.Message,
					},
					suit.InputText{
						Name:        "to",
						Before:      "To",
						Placeholder: "Complete this field to make it a direct message instead of a public tweet",
						Value:       tweet.To,
					},
					suit.InputHidden{
						Name:  "number",
						Value: fmt.Sprintf("%d", tweet.Number),
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "listTweets",
			},
			suit.ReplyAction{
				Label:        "Save Tweet",
				Name:         "saveTweet",
				DisplayIcon:  "save",
				DisplayClass: "success",
			},
		},
	}
	return &screen, nil
}

// editAccount is a config screen for editing the config of a Twitter Account
func (c *ConfigService) editAccount(config *TwitterAppModel) (*suit.ConfigurationScreen, error) {
	//	var cancelClose := &suit.Typed{
	//		suit.ReplyAction{
	//			Label: "Cancel",
	//			Name:  "listAccounts",
	//		},
	//	}
	var title string
	if config.Account.Username != "" {
		title = "Editing Twitter Account"
	} else {
		title = "New Twitter Account"
	}

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputText{
						Name:        "username",
						Before:      "Username",
						Placeholder: "@...",
						Value:       config.Account.Username,
					},
					suit.StaticText{
						Value: "See: https://dev.twitter.com/oauth/overview/application-owner-access-tokens",
					},
					suit.InputText{
						Name:   "consumerkey",
						Before: "Consumer Key",
						Value:  config.Account.ConsumerKey,
					},
					suit.InputText{
						Name:   "consumersecret",
						Before: "Consumer Secret",
						Value:  config.Account.ConsumerSecret,
					},
					suit.InputText{
						Name:   "accesstoken",
						Before: "Access Token",
						Value:  config.Account.AccessToken,
					},
					suit.InputText{
						Name:   "accesstokensecret",
						Before: "Access Token Secret",
						Value:  config.Account.AccessTokenSecret,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
				//				Name:  "listAccounts",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "saveAccount",
				DisplayClass: "success",
				DisplayIcon:  "save",
			},
		},
	}
	return &screen, nil
}

// confirmDeleteAccount is a config screen for confirming/cancelling deleting of Twitter Account
func (c *ConfigService) confirmDeleteAccount(id string) (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Title: "Confirm Deletion of " + c.app.config.Account.Username,
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Do you really want to delete this Twitter Account?",
						DisplayClass: "danger",
						DisplayIcon:  "warning",
					},
					suit.InputHidden{
						Name:  "username",
						Value: id,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:       "Cancel",
				Name:        "listAccounts",
				DisplayIcon: "close",
			},
			suit.ReplyAction{
				Label:        "Confirm - Delete",
				Name:         "delete",
				DisplayClass: "warning",
				DisplayIcon:  "check",
			},
		},
	}, nil
}

// confirmDeleteAccount is a config screen for confirming/cancelling deleting of a stored tweet/message
func (c *ConfigService) confirmDeleteTweet(name string) (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Title: "Confirm Deletion of tweet: " + name,
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Do you really want to delete this tweet?",
						DisplayClass: "danger",
						DisplayIcon:  "warning",
					},
					suit.InputHidden{
						Name:  "tweetName",
						Value: name,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:       "Cancel",
				Name:        "listTweets",
				DisplayIcon: "close",
			},
			suit.ReplyAction{
				Label:        "Confirm - Delete",
				Name:         "deleteTweet",
				DisplayClass: "warning",
				DisplayIcon:  "check",
			},
		},
	}, nil
}

//func contains(s []string, e string) bool {
//	for _, a := range s {
//		if a == e {
//			return true
//		}
//	}
//	return false
//}

// pos finds the position of a value in a slice, returns -1 if not found
func indexOf(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}
