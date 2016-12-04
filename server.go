package main

import (
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gopkg.in/redis.v3"
)

type Request struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Client struct {
	Id        int
	websocket *websocket.Conn
}

var Clients = make(map[int]Client)

func ConnectNewClient(request_channel chan Request) {

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pubsub, err := client.Subscribe("test1")
	if err != nil {
		fmt.Println("No es posible suscribirse al canal")
	}

	for {
		message, err := pubsub.ReceiveMessage()
		if err != nil {
			fmt.Println("No es posible leer el mensaje")
		}

		request := Request{}
		if err := json.Unmarshal([]byte(message.Payload), &request); err != nil {
			fmt.Println("No es posible leer el mensaje")
		}
		fmt.Println(request.Name)
		fmt.Println(request.Id)

		request_channel <- request

		//fmt.Println(message.Channel)
		//fmt.Println(message.Payload)
	}
}

func main() {
	channel_request := make(chan Request)
	go ConnectNewClient(channel_request)
	go ValidateChannel(channel_request)

	mux := mux.NewRouter()
	mux.HandleFunc("/subscribe/", Subscribe).Methods("Get")
	http.Handle("/", mux)
	fmt.Println("El servidor se encuentra en el puerto 8000")
	http.ListenAndServe(":8000", nil)
}

func Subscribe(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		return
	}
	fmt.Println("Nuevo web socket")
	count := len(Clients)
	new_client := Client{0, ws}
	Clients[count] = new_client
	fmt.Println("Nuevo cliente")
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			delete(Clients, new_client.Id)
			fmt.Println("Se fue el cliente")
			return
		}
	}
}

func ValidateChannel(request chan Request) {
	for {
		select {
		case r := <-request:
			SendMessage(r)
		}
	}
}

func SendMessage(request Request) {
	for _, client := range Clients {
		if err := client.websocket.WriteJSON(request); err != nil {
			return
		}
	}
}
