package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
  "os"
  "bytes"
  "time"

  "github.com/prometheus/alertmanager/template"
)

type Attachment struct {
  Title     string `json:"title"`
  TitleLink string `json:"titleLink"`
  Text      string `json:"text"`
  Color     string `json:"color"`
}

type responseJSON struct {
  Status  int
  Message string
}

func asJson(w http.ResponseWriter, status int, message string) {
  data := responseJSON{
    Status:  status,
    Message: message,
  }
  bytes, _ := json.Marshal(data)
  json := string(bytes[:])

  w.WriteHeader(status)
  fmt.Fprint(w, json)
}

func webhook(w http.ResponseWriter, r *http.Request) {
  defer r.Body.Close()

  data := template.Data{}
  if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
    asJson(w, http.StatusBadRequest, err.Error())
    return
  }

  // status condition
  var title string = "title"
  var titlelink string = ""
  var color = "blue"
  var text string

  if data.Status == "resolved" {
     color = "green"
     title = "Alert resolved!"
  } else if data.Status == "firing" {
     color = "red"
     title = "(Click me!) Please check ALERTMANAGER and silence it if you need."
     titlelink = "http://ALERT_MANAGER:PORT"
  }

  var attachments []Attachment
  for _, alert := range data.Alerts {
    text = ""
    for label := range alert.Labels {
      text = text + label + ": " + alert.Labels[label]  + "\n"
    }
    text = text + "StartsAt: " + alert.StartsAt.Format(time.RFC3339) + "\n" + "EndsAt: " + alert.EndsAt.Format(time.RFC3339) + "\n"
    attachments = append(attachments, Attachment{title, titlelink, text, color})
  }

  //enc := json.NewEncoder(os.Stdout)
  //enc.Encode(attachments)
  sendToDooray(data.Status, attachments)
}

func main() {
  http.HandleFunc("/webhook", webhook)
  listenAddress := ":8080"
  if os.Getenv("PORT") != "" {
    listenAddress = ":" + os.Getenv("PORT")
  }
  fmt.Printf("listening on: %v", listenAddress)
  log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func sendToDooray(status string, atts []Attachment) {
  var botName string = `"Prometheus AlertManager"`
  var botIconImage string = `"https://static.dooray.com/static_images/dooray-bot.png"`
  var text string = "\"[Alerting] " + status + "\""
  attsByt,  err := json.Marshal(atts)

  var myjson string = "{\"botName\":" + botName + ", \"botIconImage\":" + botIconImage + ", \"text\":" + text + ", \"attachments\": " + string(attsByt) + "}"

  reqBody := bytes.NewBufferString(myjson)
  fmt.Println(reqBody)
  resp, err := http.Post("DOORAY_LINK", "application/json", reqBody)

  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
}
