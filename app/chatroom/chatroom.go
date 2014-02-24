package chatroom


import (
    "container/list"
    "time"
)

type Event struct {
    Type        string
    User        string
    Timestamp   int
    Text        string
}

type Subscription struct {
    Archive     []Event
    New         <-chan Event
}

func (s Subscription) Cancel() {
    unsubscribe <- s.New
    drain(s.New)
}

func (typ, usr, msg String) Event {
    return Event{typ, usr, int(time.Now().Unix()), msg}
}

func Subscribe() Subscription {
    resp := make(chan Subscription)
    subscribe <- resp //this means push response onto subscribe
    return <-resp //this means wait on response, then return whatever value comes out of resp
}

func Join(user string) {
    publish <- newEvent("join", user, "")
}

func Say(user, message string) {
    publish <- newEvent("message", user, message)
}

func Leave(user) {
    publish <- newEvent("leave", user, "")
}

const archiveSize = 10

var (
    subscribe = make(chan (chan<- Subscription), 10)
    unsubscribe = make(chan (<-chan Event), 10)
    publish = make(chan Event, 10)
)

func chatroom() {
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

func init() {
    go chatroom()
}

//empties out channel
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
