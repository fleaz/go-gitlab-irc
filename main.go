package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"html/template"
	"io/ioutil"
	"log"

	"github.com/thoj/go-ircevent"

	"net/http"
	"strconv"
)

var (
	host     = flag.String("host", "irc.hackint.org", "Hostname of the IRC server")
	port     = flag.Int("sslport", 6697, "SSL capable port of the IRC server")
	nickname = flag.String("nickname", "go-prom-irc", "Nickname to assume once connected")
	gecos    = flag.String("gecos", "go-prom-irc", "Realname to assume once connected")
	cafile   = flag.String("cafile", "", "Path to the ca file that verifies the server certificate.")
)

func CreateFunctionNotifyFunction(bot *irc.Connection) http.HandlerFunc {

	const pushString = "[\x0311 {{ .Project.Name }} \x03] {{ .UserName }} pushed {{ .TotalCommits }} new commits to \x0305{{ .Project.Branch }}\x03"
	const commitString = "\x0315{{ .ShortID }}\x03 (\x0303{{ .AddedFiles }}\x03,\x0308{{ .ModifiedFiles }}\x03,\x0304{{ .RemovedFiles }}\x03) - {{ .Message }}"

	pushTemplate, err := template.New("push notification").Parse(pushString)
	if err != nil {
		log.Fatalf("Failed to parse pushEvent template: %v", err)
	}

	commitTemplate, err := template.New("commit notification").Parse(commitString)
	if err != nil {
		log.Fatalf("Failed to parse commitString template: %v", err)
	}

	return func(wr http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		decoder := json.NewDecoder(req.Body)

		var eventType = req.Header["X-Gitlab-Event"][0]

		type Project struct {
			Name   string `json:"name"`
			Branch string `json:"default_branch"`
		}

		type Commit struct {
			Id       string   `json:"id"`
			Message  string   `json:"message"`
			Added    []string `"json:"added"`
			Modified []string `"json:"modified"`
			Removed  []string `"json:"removed"`
		}

		type PushEvent struct {
			UserName     string   `json:"user_name"`
			Project      Project  `json:"project"`
			Commits      []Commit `json:"commits"`
			TotalCommits int      `json:"total_commits_count"`
		}

		switch eventType {

		case "Push Hook":
			var pushEvent PushEvent
			if err := decoder.Decode(&pushEvent); err != nil {
				log.Println(err)
				return
			}
			var buf bytes.Buffer
			err = pushTemplate.Execute(&buf, &pushEvent)
			bot.Privmsg("#fleaz", buf.String())

			for _, commit := range pushEvent.Commits {
				type CommitContext struct {
					ShortID       string
					Message       string
					AddedFiles    int
					ModifiedFiles int
					RemovedFiles  int
				}

				context := CommitContext{
					ShortID:       commit.Id[len(commit.Id)-8:],
					Message:       commit.Message,
					AddedFiles:    len(commit.Added),
					ModifiedFiles: len(commit.Modified),
					RemovedFiles:  len(commit.Removed),
				}

				var buf bytes.Buffer
				err = commitTemplate.Execute(&buf, &context)
				if err != nil {
					log.Printf("ERROR: %v", err)
				}
				bot.Privmsg("#fleaz", buf.String())

			}

		default:
			log.Printf("Unknown event: %s", eventType)
		}

	}

}

func main() {
	flag.Parse()

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(*cafile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	irccon := irc.IRC(*nickname, *gecos)

	irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = tlsConfig

	RegisterHandlers(irccon)

	var server bytes.Buffer
	server.WriteString(*host)
	server.WriteString(":")
	server.WriteString(strconv.Itoa(*port))

	err = irccon.Connect(server.String())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		http.HandleFunc("/notify", CreateFunctionNotifyFunction(irccon))
		http.ListenAndServe("127.0.0.1:8084", nil)
	}()

	irccon.Loop()
}

func RegisterHandlers(irccon *irc.Connection) {
	irccon.AddCallback("001", func(e *irc.Event) {
		log.Printf("Joining %v", *channel)
		irccon.Join(*channel)
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
}
