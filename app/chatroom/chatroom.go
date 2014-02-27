package chatroom

import (
	"container/list"
	"time"
)

type Event struct {
	Type      string
	User      string
	Timestamp int
	Text      string
}

type Subscription struct {
	Archive []Event
	New     <-chan Event
}

func (s Subscription) Cancel() {
	unsubscribe <- s.New
	drain(s.New)
}

func newEvent(typ, usr, msg String) Event {
	return Event{typ, usr, int(time.Now().Unix()), msg}
}

func Subscribe(room ChatRoom) Subscription {
	resp := make(chan Subscription)
	room.subscribe <- resp //this means push response onto subscribe
	return <-resp     //this means wait on response, then return whatever value comes out of resp
}

func JoinRoom(room ChatRoom, user string) {
	room.publish <- newEvent("join", user, "")
}

func SayRoom(room ChatRoom, user, message string) {
	room.publish <- newEvent("message", user, message)
}

func LeaveRoom(room ChatRoom, user string) {
	room.publish <- newEvent("leave", user, "")
}

func CommandRoom(room ChatRoom, user, command string) {
    room.publish <- newEvent("command", user, command)
}

type ChatRoom struct {
	subscribe   chan (chan<- Subscription)
	unsubscribe chan (<-chan Event)
	publish     chan Event
}

func NewChatRoom() ChatRoom {
	return ChatRoom{
		make(chan (chan<- Subscription), 10),
		make(chan (<-chan Event), 10),
		make(chan Event, 10),
	}
}

var Rooms map[string]ChatRoom

func GetOrCreateRoom(key string) ChatRoom {
    r = Rooms[key]
    if r.subscribe == nil && r.unsubscribe == nil && r.publish == nil {
        //Room does not exist and we need to initialize it
        Rooms[key] = NewChatRoom()
        go runChatRoom(Rooms[key])
    }
    return r[key]
}

const archiveSize = 10

func runChatRoom(r ChatRoom) {
    //Runs a chat room. 
	var r = GetOrCreateRoom(key)
	subscribe, publish, unsubscribe := r
	archive := list.New()
	subscribers := list.New()
	for {
		select {
		case ch := <-subscribe:
			var events []Event
			for e := archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription(events, subscriber)

		case event := <-publish:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				ch.Value.(chan Event) <- event
			}
			if archive.Len() >= archiveSize {
				archive.Remove(archive.Front())
			}
			archive.PushBack(event)

		case unsub := <-unsubscribe:
			for ch := subscribers.Front(); ch != nil; ch = ch.Next() {
				if ch.Value.(chan Event) == unsub {
					subscribers.Remove(ch)
					break
				}
			}
		}
	}
}

//empties out event channel
func drain(ch <-chan Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
