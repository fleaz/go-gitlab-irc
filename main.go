package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"

	"github.com/thoj/go-ircevent"

	"net/http"
)

//func RegisterConnect(s ircx.Sender, m *irc.Message) {
//	log.Println("Connected...")
//	s.Send(&irc.Message{
//		Command:       irc.JOIN,
//		Params:        []string{"#ffda-noc"},
//		EmptyTrailing: true,
//	})
//}
//
//func PingHandler(s ircx.Sender, m *irc.Message) {
//	log.Println("P[IO]NG")
//	s.Send(&irc.Message{
//		Command:  irc.PONG,
//		Params:   m.Params,
//		Trailing: m.Trailing,
//	})
//}
//
//func RegisterHandlers(bot *ircx.Bot) {
//	bot.HandleFunc(irc.RPL_WELCOME, RegisterConnect)
//	bot.HandleFunc(irc.PING, PingHandler)
//}

func CreateFunctionNotifyFunction(bot *irc.Connection) http.HandlerFunc {

	const templateString = "{{ .Alert.Labels.instance }} {{ .ColorStart }}{{ .Alert.Labels.alertname}}{{ .ColorEnd }} - {{ .Alert.Annotations.description}}"

	notificationTemplate, err := template.New("notification").Parse(templateString)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	return func(wr http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		decoder := json.NewDecoder(req.Body)
		/*

			   {
			     "version": "3",
			     "groupKey": <number>     // key identifying the group of alerts (e.g. to deduplicate)
			     "status": "<resolved|firing>",
			     "receiver": <string>,
			     "groupLabels": <object>,
			     "commonLabels": <object>,
			     "commonAnnotations": <object>,
			     "externalURL": <string>,  // backling to the Alertmanager.
			     "alerts": [
			       {
			         "labels": <object>,
			         "annotations": <object>,
			         "startsAt": "<rfc3339>",
			         "endsAt": "<rfc3339>"
			       },
			       ...
			     ]
			   }

			---


			{"version":"3",
			 "groupKey":16716837308297233527,
			  "status":"firing",
			  "receiver":"irc",
			  "groupLabels":{
				  "alertname":"InstanceHighCpu"
			  },
			  "commonLabels":{
				  "alertname":"InstanceHighCpu",
				  "severity":"page"
			  },
			  "commonAnnotations":{
				  "description":" has high cpu activity",
				  "summary":"Instance : cpu high"
			  },
			  "externalURL":"http://elsa.darmstadt.freifunk.net:9093",
			  "alerts":[
			  	{"labels":{
					"alertname":"InstanceHighCpu",
					"severity":"page"
				},
				"annotations":{
					"description":
					" has high cpu activity",
					"summary":"Instance : cpu high"
				},
				"startsAt":"2017-02-28T03:00:22.803+01:00",
				"endsAt":"0001-01-01T00:00:00Z"
				}
			  ]
			}


		*/

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
				ColorStart:   getColorcode(notification.Status),
				ColorEnd:     "\x03",
			}

			var buf bytes.Buffer
			err = notificationTemplate.Execute(&buf, &context)
			bot.Privmsg("#ffda-mon", buf.String())
		}

	}

}

func getColorcode(status string) string {
	switch status {
	case "firing":
		return "\x0305"
	case "resolved":
		return "\x0303"
	default:
		return "\x0300"
	}
}

func main() {

	caCertPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile("./hackint-rootca.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	irccon := irc.IRC("ffda-prom-irc", "ffda-prometheus-notifier")

	irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = tlsConfig

	RegisterHandlers(irccon)

	err = irccon.Connect("irc.hackint.org:6697")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		http.HandleFunc("/notify", CreateFunctionNotifyFunction(irccon))
		http.ListenAndServe("127.0.0.1:8083", nil)
	}()

	irccon.Loop()
}

func RegisterHandlers(irccon *irc.Connection) {
	irccon.AddCallback("001", func(e *irc.Event) {
		const channel = "#ffda-mon"
		log.Printf("Joining %v", channel)
		irccon.Join(channel)
	})
	irccon.AddCallback("366", func(e *irc.Event) {})
}
