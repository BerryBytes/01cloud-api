package websocket

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

const (
	// writeWait      = 60 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Connection struct {
	ws   *websocket.Conn
	send chan []byte
}

func (s Subscription) readPump() {
	c := s.Conn
	defer func() {
		H.Unregister <- s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	//c.ws.SetReadDeadline(time.Now().Add(pongWait))
	//c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway)
			break
		}

		m := Message{msg, s.Room}
		H.Broadcast <- m
	}
}

func (c *Connection) write(mt int, payload []byte) error {
	//c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (s *Subscription) writePump() {
	c := s.Conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case Message, ok := <-c.send:
			if !ok {
				err := c.write(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Error("websocket write error :: ", err)
				}
				return
			}
			if err := c.write(websocket.TextMessage, Message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	//vars := mux.Vars(r)
	//log.Println(r.FormValue("room"))
	if err != nil {
		log.Error(err)
		return
	}
	c := &Connection{send: make(chan []byte, 256), ws: ws}
	s := Subscription{c, r.FormValue("room")}

	H.Register <- s
	go s.writePump()
	s.readPump()
}
