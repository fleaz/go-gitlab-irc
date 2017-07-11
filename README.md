# go-gitlab-irc

Implements GitLab API and pipe output into the configured IRC channels.

# Set-Up

  `go get github.com/f-breidenstein/go-gitlab-irc`
  
  `go-gitlab-irc -host irc.hackint.org -sslport 6697 -nickname my-go-gitlab-irc-bot -cafile hackint-rootca.crt`
  
# Configuration
  ```
 Usage of ./go-prom-irc:
  -cafile string
    	Path to the ca file that verifies the server certificate.
  -gecos string
    	Realname to assume once connected (default "go-prom-irc")
  -host string
    	Hostname of the IRC server (default "irc.hackint.org")
  -nickname string
    	Nickname to assume once connected (default "go-prom-irc")
  -sslport int
    	SSL capable port of the IRC server (default 6697)
```

go-gitlab-irc only supports connecting to IRC via SSL so far. Make sure you provide the proper `-cafile` option for your network.
