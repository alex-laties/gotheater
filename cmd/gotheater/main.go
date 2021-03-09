package main

import (
	"encoding/json"
	"sync"
	"time"

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

		var currSessions []map[string]string
		sessionsLock.Lock()
		for id, sess := range sessions {
			var name string
			if nameTemp, exists := sess.Get("name"); exists {
				name = nameTemp.(string)
			}
			currSessions = append(currSessions, map[string]string{
				"id":   id,
				"name": name,
			})
		}
		sessionsLock.Unlock()

		msg := message.NewConnect(id, map[string]interface{}{
			"id":                    id,
			"currentRulerID":        currRuler,
			"currentMediaURL":       currentMediaURL,
			"currentMediaTimestamp": currentMediaTimestamp,
			"currentMediaPaused":    currentMediaPaused,
			"currentSessions":       currSessions,
		})
		msgBytes, _ := json.Marshal(msg)
		websocketRouter.Broadcast(msgBytes)
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
			// only the ruler can set media
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
			var status message.Status
			err := json.Unmarshal(msg.Data, &status)
			if err != nil {
				return
			}
			if status.Name != "" {
				s.Set("name", status.Name)
			}
			websocketRouter.BroadcastOthers(b, s)
		case "playbackStatus":
			if id != currentRulerID {
				return
			}
			// capture current playback timestamp
			var playbackStatus message.RulerPlaybackStatus
			err := json.Unmarshal(msg.Data, &playbackStatus)
			if err != nil {
				return
			}
			currentMediaLock.Lock()
			currentMediaPaused = !playbackStatus.Playing
			currentMediaTimestamp = playbackStatus.CurrentMediaTimestamp
			currentMediaLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
		case "ping":
			var ping message.Ping
			if err := json.Unmarshal(msg.Data, &ping); err != nil {
				return
			}

			currTime := int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)
			payload, err := json.Marshal(map[string]interface{}{
				"id": "god",
				"data": message.Pong{
					Ping:       ping,
					ReceivedAt: int(currTime),
				},
			})
			if err != nil {
				return
			}
			s.Write(payload)
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
			var seekTo message.Seek
			err := json.Unmarshal(msg.Data, &seekTo)
			if err != nil {
				return
			}

			currentMediaLock.Lock()
			currentMediaPaused = true
			currentMediaTimestamp = seekTo.MediaTimestamp
			currentMediaLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
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
