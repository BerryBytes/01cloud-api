package websocket

import (
	"01cloud-api/api/models"
	"encoding/json"

	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

func WebsocketConn(room string) (*websocket.Conn, *sync.Mutex) {
	url := os.Getenv("API_SERVER") + "/ws-server?room=" + room + "&token=" + os.Getenv("API_SECRET")
	log.Printf("connecting to %s", url)

	var c *websocket.Conn
	var err error
	// Output: Retry for 3 times if failed
	for i := 1; i <= 3; i++ {
		c, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			if i == 3 {
				log.Warnf("Error in Websocket connection \n err ====> %v", err)
			}
			log.Printf("Retrying Websocket connection = %v", i)
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	mutex := &sync.Mutex{}
	return c, mutex
}

func EmitMessage(event string, namespace string, name string, wsConn *websocket.Conn, mutex *sync.Mutex) {
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	mp := &map[string]interface{}{
		"Type":      "domain-status",
		"name":      name,
		"Data":      event,
		"Namespace": namespace,
	}
	data, err := json.Marshal(mp)
	if err != nil {
		log.Printf("error %v", err)
		return
	}
	err = wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Printf("error %v\n", err)
		return
	}
}

func EmitErrorMessage(wsConn *websocket.Conn, mutex *sync.Mutex, errorData *models.ErrorMessage) {
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	mp := &map[string]interface{}{
		"type":           "error",
		"cluster_id":     errorData.ClusterId,
		"environment_id": errorData.EnvironmentId,
		"code":           errorData.Code,
		"message":        errorData.Message,
		"source":         errorData.Source,
		"time":           errorData.Time,
	}
	data, err := json.Marshal(mp)
	if err != nil {
		log.Printf("error %v", err)
		return
	}
	err = wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Printf("error %v\n", err)
		return
	}
}

func EmitStatusMessage(status string, namespace string, name string, wsConn *websocket.Conn, mutex *sync.Mutex) {
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	mp := &map[string]interface{}{
		"type":      "fetch-status",
		"name":      name,
		"data":      status,
		"namespace": namespace,
	}
	data, err := json.Marshal(mp)
	if err != nil {
		return
	}
	err = wsConn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return
	}
}
