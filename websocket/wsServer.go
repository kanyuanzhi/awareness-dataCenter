package websocket

import (
	"dataCenter/models"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const MAXMESSAGESIZE = 1024

var upgrader = websocket.Upgrader{
	ReadBufferSize:  MAXMESSAGESIZE,
	WriteBufferSize: MAXMESSAGESIZE,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket服务器
type WsServer struct {
	wsClients *WsClients
}

func NewWsServer(wsClients *WsClients) *WsServer {
	return &WsServer{
		wsClients: wsClients,
	}
}

func (ws *WsServer) Start() {
	addr := flag.String("addr", "localhost:9091", "http service address")
	http.HandleFunc("/ws", ws.serveWs)
	log.Printf("start WsServer on port 9091")
	http.ListenAndServe(*addr, nil)
}

func (ws *WsServer) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)

	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		conn.Close()
		return
	}
	ws.wsClients.Register <- conn // 在Clients中注册
	var m map[string]interface{}
	err = json.Unmarshal(message, &m)
	if err != nil {
		log.Println("json:", err)
		conn.Close()
		return
	}
	response, _ := json.Marshal(models.NewWsResponse(true))
	err = conn.WriteMessage(websocket.TextMessage, response)
	if err != nil {
		log.Printf("write errro: %s", err)
	}

	switch int(m["request_type"].(float64)) {
	case 10:
		var requestForPeople models.WsRequestForPeople
		json.Unmarshal(message, &requestForPeople)
		requestForPeople.Conn = conn
		ws.wsClients.requestForPeople <- &requestForPeople
	case 11:
		var requestForPerson models.WsRequestForPerson
		json.Unmarshal(message, &requestForPerson)
		requestForPerson.Conn = conn
		ws.wsClients.requestForPerson <- &requestForPerson
	default:
		break
	}
}