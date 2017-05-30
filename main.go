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
	channel  = flag.String("channel", "", "Target to send notifications to, likely a channel.")
	nickname = flag.String("nickname", "go-prom-irc", "Nickname to assume once connected")
	gecos    = flag.String("gecos", "go-prom-irc", "Realname to assume once connected")
	cafile   = flag.String("cafile", "", "Path to the ca file that verifies the server certificate.")
)

func CreateFunctionNotifyFunction(bot *irc.Connection) http.HandlerFunc {

	const templateString = "{{ .Alert.Labels.instance }} {{ .ColorStart }}{{ .Alert.Labels.alertname}}{{ .ColorEnd }} - {{ .Alert.Annotations.description}}"

	notificationTemplate, err := template.New("notification").Parse(templateString)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	return func(wr http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		decoder := json.NewDecoder(req.Body)

		type Alert struct {
			Labels      map[string]interface{} `json:"labels"`
			Annotations map[string]interface{} `json:"annotations"`
			StartsAt    string                 `json:"startsAt"`
			EndsAt      string                 `json:"endsAt"`
		}

		type Notification struct {
			Version           string                 `json:"version"`
			GroupKey          uint64                 `json:"groupKey"`
			Status            string                 `json:"status"`
			Receiver          string                 `json:"receiver"`
			GroupLables       map[string]interface{} `json:"groupLabels"`
			CommonLabels      map[string]interface{} `json:"commonLabels"`
			CommonAnnotations map[string]interface{} `json:"commonAnnotations"`
			ExternalURL       string                 `json:"externalURL"`
			Alerts            []Alert                `json:"alerts"`
		}

		var notification Notification

		if err := decoder.Decode(&notification); err != nil {
			log.Println(err)
			return
		}

		body, err := json.Marshal(&notification)

		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("JSON: %v", string(body))
		for _, alert := range notification.Alerts {
			type NotificationContext struct {
				Alert        *Alert
				Notification *Notification
				ColorStart   string
				ColorEnd     string
			}
			context := NotificationContext{
				Alert:        &alert,
				Notification: &notification,
				ColorStart:   notification.Status,
				ColorEnd:     "\x03",
			}

			var buf bytes.Buffer
			err = notificationTemplate.Execute(&buf, &context)
			bot.Privmsg(*channel, buf.String())
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
		log.Printf("Joining %v", channel)
		irccon.Join(*channel)
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
}
