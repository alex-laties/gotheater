package main

import (
	"encoding/json"
	"sync"

	"github.com/alex-laties/gotheater/pkg/message"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"gopkg.in/olahol/melody.v1"
)

var sessions = make(map[string]*melody.Session)
var sessionsLock, currentRulerLock, currentMediaLock sync.Mutex
var currentRulerID string
var currentMediaURL string
var currentMediaPaused bool
var currentMediaTimestamp int

/**
A webserver that provides access to:
 - User management
 - Content management
 - Transcode management
*/
func main() {
	router := gin.Default()
	websocketRouter := melody.New()

	router.GET("/ws", func(c *gin.Context) {
		websocketRouter.HandleRequest(c.Writer, c.Request)
	})
	router.Static("/", "/var/lib/gotheater/frontend")

	websocketRouter.HandleConnect(func(s *melody.Session) {
		// assign a userID
		id := xid.New().String()
		s.Set("id", id)
		sessionsLock.Lock()
		sessions[id] = s
		sessionsLock.Unlock()

		// if no ruler, set current user as ruler
		currentRulerLock.Lock()
		if currentRulerID == "" {
			currentRulerID = id
		}
		currRuler := currentRulerID
		currentRulerLock.Unlock()

		msg := message.NewConnect(id, currRuler, currentMediaURL)
		msgBytes, _ := json.Marshal(msg)
		websocketRouter.Broadcast(msgBytes)

		// TODO send current members + media to client
	})

	websocketRouter.HandleMessage(func(s *melody.Session, b []byte) {
		idTemp, exists := s.Get("id")
		if !exists {
			return
		}
		id := idTemp.(string)

		var msg message.Base
		err := json.Unmarshal(b, &msg)
		if err != nil {
			return
		}

		switch msg.Type {
		case "setMedia":
			if id != currentRulerID {
				return
			}
			var media message.SetMedia
			err := json.Unmarshal(msg.Data, &media)
			if err != nil {
				return
			}
			currentMediaLock.Lock()
			currentMediaURL = media.URL
			currentMediaLock.Unlock()

			websocketRouter.BroadcastBinaryOthers(b, s)
		case "status":
			// we pass these on immediately
			websocketRouter.BroadcastOthers(b, s)
			if id == currentRulerID {
				// TODO parse timestamp and set to currentMediaTimestamp
			}
		case "ping":
			// TODO pong
		case "setName":
			// TODO parse name and set
		case "pause":
			currentMediaLock.Lock()
			currentMediaPaused = true
			currentMediaLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
		case "play":
			currentMediaLock.Lock()
			currentMediaPaused = false
			currentMediaLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
		case "seek":
			currentMediaLock.Lock()
			// TODO parse timestamp
			currentMediaLock.Unlock()
		default:
			return
		}
	})

	websocketRouter.HandleDisconnect(func(s *melody.Session) {
		idTemp, exists := s.Get("id")
		if !exists {
			return
		}
		id := idTemp.(string)

		sessionsLock.Lock()
		delete(sessions, id)
		sessionsLock.Unlock()

		currentRulerLock.Lock()
		if currentRulerID == id {
			currentRulerID = ""
		}
		currentRulerLock.Unlock()

		msg := message.NewDisconnect(id)
		msgBytes, _ := json.Marshal(msg)
		websocketRouter.BroadcastOthers(msgBytes, s)
	})
	router.Run(":8080")
}
