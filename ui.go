package main

import (
	"encoding/json"
	"fmt"

	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

// TODO: (if useful) make config handle multiple accounts

// TweetDetails stores the values for one tweet or direct message
// Number is the auto-incrementing value to add to tweets/messages so that Twitter won't reject as duplicates
type TweetDetails struct {
	Name 	string `json:name`
	Message string `json:message`
	To      string `json:to`
	Number  int    `json:number`
}

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
	case "list":
		return c.list()
	case "":
		// present the existing or new Twitter Account screen
		if c.app.config.Username != "" {
			return c.list()
		}
		fallthrough
	case "new":
		return c.edit(&TwitterAppModel{})

	case "edit":
		return c.edit(c.app.config)

	case "save":
		configData := &TwitterAppModel{}
		err := json.Unmarshal(request.Data, configData)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		err = c.app.SaveAccount(*configData)
		if err != nil {
			return c.error(fmt.Sprintf("Could not save Twitter Account: %s", err))
		}

		return c.list()

	case "confirmDelete":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		return c.confirmDelete(values["username"])

	case "delete":
		var values map[string]string
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		err = c.app.DeleteAccount(values["username"])
		if err != nil {
			return c.error(fmt.Sprintf("Failed to delete Twitter Account: %s", err))
		}

		return c.edit(&TwitterAppModel{})

	case "newTweet":
		return c.newTweet()

	case "saveTweet":
		var values TweetDetails
		//		var result
		err := json.Unmarshal(request.Data, &values)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		//		log.Infof("values: %v", values)

		if contains(values.Options, "direct") {
			result, err := c.app.twitterAPI.PostDMToScreenName(values.Message, values.To)
			if err != nil {
				log.Errorf("Error sending DM %v", err)
			}
			log.Infof("%v", result)
		} else {
			result, err := c.app.twitterAPI.PostTweet(values.Message, nil)
			if err != nil {
				log.Errorf("Error posting Tweet %v", err)
			}
			log.Infof("%v", result)
		}
		//		fmt.Printf("%#v\n\n%v\n", result, err)
		return c.newTweet()

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
				Name:  "list",
			},
		},
	}, nil
}

// list is a config screen for displaying accounts with options for editing, deleting and controlling
func (c *ConfigService) list() (*suit.ConfigurationScreen, error) {
	// TODO: currently this displays "undefined" when there's no account; need to check
	screen := suit.ConfigurationScreen{
		Title: "Twitter App Config",
		Sections: []suit.Section{
			suit.Section{
				Title: "Edit",
				Contents: []suit.Typed{
					suit.ActionList{
						Name: "account",
						Options: []suit.ActionListOption{
							suit.ActionListOption{
								Title: c.app.config.Username,
							},
						},
						PrimaryAction: &suit.ReplyAction{
							Name:        "edit",
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
				Label:       "Tweets",
				Name:        "tweets",
				DisplayIcon: "twitter",
			},
			suit.ReplyAction{
				Label:        "New Account",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func (c *ConfigService) newTweet() (*suit.ConfigurationScreen, error) {
	screen := suit.ConfigurationScreen{
		Title: "New Tweet/Message",
		Sections: []suit.Section{

			// tweet!
			suit.Section{
				Title: "Tweet!",
				Contents: []suit.Typed{
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "Give this tweet/message a name to identify it",
					},
					suit.InputText{
						Name:        "message",
						Before:      "Message",
						Placeholder: "Up to 140 characters",
					},
					suit.InputText{
						Name:        "to",
						Before:      "To",
						Placeholder: "Complete the \"To\" field to make it a direct message instead of a public tweet",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
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

// edit is a config screen for editing the config of a Twitter Account
func (c *ConfigService) edit(config *TwitterAppModel) (*suit.ConfigurationScreen, error) {

	var title string
	if config.Username != "" {
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
						Value:       config.Username,
					},
					suit.StaticText{
						Value: "See: https://dev.twitter.com/oauth/overview/application-owner-access-tokens",
					},
					suit.InputText{
						Name:   "consumerkey",
						Before: "Consumer Key",
						Value:  config.ConsumerKey,
					},
					suit.InputText{
						Name:   "consumersecret",
						Before: "Consumer Secret",
						Value:  config.ConsumerSecret,
					},
					suit.InputText{
						Name:   "accesstoken",
						Before: "Access Token",
						Value:  config.AccessToken,
					},
					suit.InputText{
						Name:   "accesstokensecret",
						Before: "Access Token Secret",
						Value:  config.AccessTokenSecret,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "list",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

// confirmDelete is a config screen for confirming/cancelling deleting of Twitter Account
func (c *ConfigService) confirmDelete(id string) (*suit.ConfigurationScreen, error) {
	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Title: "Confirm Deletion of " + c.app.config.Username,
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
				Name:        "list",
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
