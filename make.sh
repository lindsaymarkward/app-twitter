# simple make for new packaged app-twitter
GOOS=linux GOARCH=arm go build .
tar -czf app-twitter.tar.gz app-twitter package.json images/*
