package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/alex-laties/gotheater/pkg/message"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"gopkg.in/olahol/melody.v1"
)

var sessions = make(map[string]*melody.Session)
var sessionsLock, currentRulerLock, currentMediaLock sync.Mutex
var currentRulerID string
var currentMediaURL string
var currentMediaPaused bool = true
var currentMediaTimestamp int

/**
A webserver that provides access to:
 - User management
 - Content management
 - Transcode management
*/
func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugarLog := logger.Sugar()
	router := gin.Default()
	websocketRouter := melody.New()

	router.GET("/ws", func(c *gin.Context) {
		websocketRouter.HandleRequest(c.Writer, c.Request)
	})

	// we override the NoRoute handler to support serving react files from here
	// we can't use static since we already set "/ws"
	fs := gin.Dir("/var/lib/gotheater/frontend", false)
	fileServer := http.StripPrefix("/", http.FileServer(fs))
	router.NoRoute(func(c *gin.Context) {
		f, err := fs.Open(c.Request.URL.Path)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
			return
		}
		f.Close()

		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	websocketRouter.HandleConnect(func(s *melody.Session) {
		// assign a userID
		id := xid.New().String()
		s.Set("id", id)
		sessionsLock.Lock()
		sessions[id] = s
		sessionsLock.Unlock()
		sugarLog.Infow("newUser",
			"id", id)
		// if no ruler, set current user as ruler
		currentRulerLock.Lock()
		if currentRulerID == "" {
			currentRulerID = id
			sugarLog.Infow("newRuler", "id", id)
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
			sugarLog.Warnw("message without ID",
				"message", string(b),
			)
			return
		}
		id := idTemp.(string)

		var msg message.Base
		err := json.Unmarshal(b, &msg)
		if err != nil {
			sugarLog.Error(err)
			return
		}

		switch msg.Type {
		case "setLeader":
			if id != currentRulerID {
				sugarLog.Warn("attempt to change ruler from non-ruler")
				return
			}

			var newRuler message.SetRuler
			err := json.Unmarshal(msg.Data, &newRuler)
			if err != nil {
				sugarLog.Error(err)
				return
			}
			// verify the session exists
			sessionsLock.Lock()
			if _, exists := sessions[newRuler.NewRulerID]; !exists {
				sugarLog.Warn("attempt to change ruler to a non-existant user")
				sessionsLock.Unlock()
				return
			}
			sessionsLock.Unlock()

			currentRulerLock.Lock()
			currentRulerID = newRuler.NewRulerID
			currentRulerLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
		case "setMedia":
			// only the ruler can set media
			if id != currentRulerID {
				sugarLog.Warn("attempted setMedia from non-ruler")
				return
			}
			var media message.SetMedia
			err := json.Unmarshal(msg.Data, &media)
			if err != nil {
				sugarLog.Error(err)
				return
			}
			currentMediaLock.Lock()
			currentMediaURL = media.URL
			currentMediaLock.Unlock()

			sugarLog.Infow(string(b))
			websocketRouter.BroadcastBinaryOthers(b, s)
		case "status":
			var status message.Status
			err := json.Unmarshal(msg.Data, &status)
			if err != nil {
				sugarLog.Error(err)
				return
			}
			if status.Name != "" {
				s.Set("name", status.Name)
			}
			websocketRouter.BroadcastOthers(b, s)
		case "playbackStatus":
			if id != currentRulerID {
				sugarLog.Warn("attempt to send playbackStatus by non-ruler")
				return
			}
			// capture current playback timestamp
			var playbackStatus message.RulerPlaybackStatus
			err := json.Unmarshal(msg.Data, &playbackStatus)
			if err != nil {
				sugarLog.Error(err)
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
				sugarLog.Error(err)
				return
			}

			currTime := int64(time.Nanosecond) * time.Now().UnixNano() / int64(time.Millisecond)
			payload, err := json.Marshal(map[string]interface{}{
				"id":   "god",
				"type": "pong",
				"data": message.Pong{
					Ping:       ping,
					ReceivedAt: int(currTime),
				},
			})
			if err != nil {
				sugarLog.Error(err)
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
				sugarLog.Error(err)
				return
			}

			currentMediaLock.Lock()
			currentMediaPaused = true
			currentMediaTimestamp = seekTo.MediaTimestamp
			currentMediaLock.Unlock()
			websocketRouter.BroadcastOthers(b, s)
		default:
			sugarLog.Warnf("undefined message type",
				"message", string(b),
			)
			return
		}
	})

	websocketRouter.HandleDisconnect(func(s *melody.Session) {
		idTemp, exists := s.Get("id")
		if !exists {
			sugarLog.Warn("disconnect from session that no longer exists")
			return
		}
		id := idTemp.(string)

		sessionsLock.Lock()
		delete(sessions, id)
		sessionsLock.Unlock()

		currentRulerLock.Lock()
		if currentRulerID == id {
			// select a new ruler at random if possible
			currentRulerID = ""
			sessionsLock.Lock()
			if len(sessions) > 0 {
				for randID, _ := range sessions {
					// first is kind of random, right?
					currentRulerID = randID
					break
				}
			}
			msgBytes, _ := json.Marshal(message.NewRuler(currentRulerID))
			go websocketRouter.BroadcastOthers(msgBytes, s)
			sessionsLock.Unlock()
		}
		currentRulerLock.Unlock()

		msg := message.NewDisconnect(id)
		msgBytes, _ := json.Marshal(msg)
		websocketRouter.BroadcastOthers(msgBytes, s)
	})
	router.Run(":8080")
}
