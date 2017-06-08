# go-prom-irc

Implements Prometheus' Alertmanager API and pipes alerts into the configured IRC channel.

# Set-Up

  `go get github.com/andir/go-prom-irc`
  
  `go-prom-irc -host irc.hackint.org -sslport 6697 -channel #go-prom-irc -nickname my-go-prom-irc-bot -cafile hackint-rootca.crt`
  
# Configuration
  ```
 Usage of ./go-prom-irc:
  -cafile string
    	Path to the ca file that verifies the server certificate.
  -channel string
    	Target to send notifications to, likely a channel.
  -gecos string
    	Realname to assume once connected (default "go-prom-irc")
  -host string
    	Hostname of the IRC server (default "irc.hackint.org")
  -nickname string
    	Nickname to assume once connected (default "go-prom-irc")
  -sslport int
    	SSL capable port of the IRC server (default 6697)
```

go-prom-irc only supports connecting to IRC via SSL so far. Make sure you provide the proper `-cafile` option for your network.
