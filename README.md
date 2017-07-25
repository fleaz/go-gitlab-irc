# Acknowledgment
Thanks to @andir for the original codebase: https://github.com/andir/go-prom-irc

# go-gitlab-irc

Implements GitLab API and pipe output into the configured IRC channels.

# Set-Up

  `go get github.com/f-breidenstein/go-gitlab-irc`
  
  `go-gitlab-irc -host irc.hackint.org -sslport 6697 -nickname my-go-gitlab-irc-bot -cafile hackint-rootca.crt -channelmapping channelmapping.yml`

# Configuration
  ```
 Usage of ./go-rom-irc:
  -cafile string
    	Path to the ca file that verifies the server certificate.
  -channelmapping string
    	Path to the channel mapping file that mapps repository names to irc channels.
  -gecos string
    	Realname to assume once connected (default "go-gitlab-irc")
  -host string
    	Hostname of the IRC server (default "irc.hackint.org")
  -nickname string
    	Nickname to assume once connected (default "go-gitlab-irc")
  -sslport int
    	SSL capable port of the IRC server (default 6697)
```

go-gitlab-irc only supports connecting to IRC via SSL so far. Make sure you provide the proper `-cafile` option for your network.

# Webhook Set-Up

By default, the bot will listen on localhost at port 8084. use the following URL
to add it to your webhooks: `http://127.0.0.1:8084/notify`
