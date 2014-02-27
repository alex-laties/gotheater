package chatroom

import (
    "errors"
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

func (s Subscription) Cancel(room ChatRoom) {
	room.unsubscribe <- s.New
	drain(s.New)
}

func newEvent(typ, usr, msg string) Event {
	return Event{typ, usr, int(time.Now().Unix()), msg}
}

func Subscribe(room ChatRoom) Subscription {
	resp := make(chan Subscription)
	room.subscribe <- resp
	return <-resp
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

func CreateChatRoom(key string) (ChatRoom, error) {
    r := Rooms[key]
    i := ChatRoom{}
    if r == i {
        //Room does not exist and we need to initialize it
        Rooms[key] = NewChatRoom()
        go runChatRoom(key)
        return Rooms[key], nil
    } else {
        return ChatRoom{}, errors.New("Room already exists")
    }
}

func GetChatRoom(key string) (ChatRoom, error) {
    r := Rooms[key]
    i := ChatRoom{}
    if r == i {
        return ChatRoom{}, errors.New("Room does not exist")
    }
    return r, nil
}

func DeleteChatRoom(key string) (ChatRoom, error) {
    r := Rooms[key]
    i := ChatRoom{}
    if r == i {
        return i, errors.New("Room does not exist")
    }

    delete(Rooms, key)
    return r, nil
}

const archiveSize = 10

func runChatRoom(key string) {
    //Runs a chat room. 
    r, err := GetChatRoom(key)
    if err != nil {
        return
    }
	subscribe, publish, unsubscribe := r.subscribe, r.publish, r.unsubscribe
	archive := list.New()
	subscribers := list.New()
	for {
        //Check to see if room is still active
        if _, err := GetChatRoom(key); err != nil {
            return
        }
		select {
		case ch := <-subscribe:
			var events []Event
			for e := archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			subscriber := make(chan Event, 10)
			subscribers.PushBack(subscriber)
			ch <- Subscription{events, subscriber}

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
