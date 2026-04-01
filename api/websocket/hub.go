package websocket

type Message struct {
	data []byte
	room string
}

type Subscription struct {
	Conn *Connection
	Room string
}

type Hub struct {
	Rooms      map[string]map[*Connection]bool
	Broadcast  chan Message
	Register   chan Subscription
	Unregister chan Subscription
}

var H = Hub{
	Broadcast:  make(chan Message),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
	Rooms:      make(map[string]map[*Connection]bool),
}

func (h *Hub) Run() {
	for {
		select {
		case s := <-h.Register:
			Connections := h.Rooms[s.Room]
			if Connections == nil {
				Connections = make(map[*Connection]bool)
				h.Rooms[s.Room] = Connections
			}
			h.Rooms[s.Room][s.Conn] = true
		case s := <-h.Unregister:
			Connections := h.Rooms[s.Room]
			if Connections != nil {
				if _, ok := Connections[s.Conn]; ok {
					delete(Connections, s.Conn)
					close(s.Conn.send)
					if len(Connections) == 0 {
						delete(h.Rooms, s.Room)
					}
				}
			}
		case m := <-h.Broadcast:
			Connections := h.Rooms[m.room]
			for c := range Connections {
				select {
				case c.send <- m.data:
				default:
					close(c.send)
					delete(Connections, c)
					if len(Connections) == 0 {
						delete(h.Rooms, m.room)
					}
				}
			}
		}
	}
}
