# app-twitter
Ninja Sphere app (Go) for Twitter, by Lindsay Ward

Setup
-----

 - Use the config in Labs (ninjasphere.local) to set your username (screen name) + authentication details, which you can generate via Twitter - see: [Twitter auth tokens help](https://dev.twitter.com/oauth/overview/application-owner-access-tokens)
 - Then create and save tweets or direct messages, which will be given numbers (1, 2...). 
 - To make a direct message, enter the recipient's Twitter handle in the "To" field.
 - To make a public tweet, leave the "To" field blank.

Usage
-----
When the app is running, the spheramid shows either:
 
  - a red X over an @ symbol means that the authentication details are invalid and the API can't be setup properly - fix this in Labs (you don't need to restart the app)
  - a red "NO" over the Twitter bird means no tweets have been stored - create some in Labs
  - a yellow number over the bird shows the current tweet (with "TWT") or direct message (with "DM"). 
  
When the spheramid shows a numbered tweet:

 - tap the right or left side to select the next/previous tweet
 - double tap to send that tweet

When you send a tweet you will see either a green tick for success or a red X for failure.    
The tweets/messages have a number appended that increases with each use so that Twitter doesn't reject them as duplicates.

Running
-------

To run/test from a Mac, build with `go build .` then use the command (replace XXX with your Sphere serial number):

`DEBUG=* ./app-twitter --mqtt.host=ninjasphere.local --mqtt.port=1883 --serial=XXX --led.host=ninjasphere.local`

TODO
----

  - need to implement some channels so it can be used by Ninja for notifications... and stuff
  - maybe make scrolling text to show the stored tweet names instead of just number
  - maybe support multiple Twitter accounts (let me know if this would be useful for anyone)
  - maybe implement Sign in with Twitter (Web) instead of copying keys/tokens - see: https://dev.twitter.com/web/sign-in/implementing
