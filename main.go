//go:build example1
// +build example1

package main

import (
	"fmt"
	"log"
	"net/http"

	// "strconv"
	"time"

	net "github.com/subchord/go-sse"
)

type API struct {
	broker *net.Broker
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func main() {
	sseClientBroker := net.NewBroker(map[string]string{
		"Access-Control-Allow-Origin": "*",
	})

	api := &API{broker: sseClientBroker}

	http.HandleFunc("/sse", api.sseHandler)

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		log.Println("In de /start handlerfunc")
		http.ServeFile(w, r, "web_sse_example.html")
	})

	count := 0

	http.HandleFunc("/msg", func(w http.ResponseWriter, r *http.Request) {
		log.Println("In de /msg handlerfunc")
		enableCors(&w)

		switch r.Method {
		case "POST":
			// err := r.ParseForm()
			// if err != nil{
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }
			msg := r.PostFormValue("msg")
			count++
			api.broker.Broadcast(net.StringEvent{
				Id:    fmt.Sprintf("event-id-%v", count),
				Event: "message",
				Data:  msg,
			})

		default:
			fmt.Fprintf(w, "Sorry, only POST method supported.")
		}
	})

	// Broadcast message to all clients every 5 seconds
	go func() {
		tick := time.Tick(60 * time.Second)
		for {
			select {
			case <-tick:
				count++
				api.broker.Broadcast(net.StringEvent{
					Id:    fmt.Sprintf("event-id-%v", count),
					Event: "ping",
					// Data:  strconv.Itoa(count),
				})
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}

func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	client, err := api.broker.Connect(fmt.Sprintf("%v", time.Now().Unix()), writer, request)
	if err != nil {
		log.Println(err)
		return
	}
	<-client.Done()
	log.Printf("connection with client %v closed", client.Id())
}
