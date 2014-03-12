package controllers

import (
    "gotheater/app/chatroom"
    "code.google.com/p/go.net/websocket"
    "github.com/robfig/revel"
    "log"
    "net/http"
)

type Room struct {
    *revel.Controller
}

func (c Room) Get(room string) revel.Result {
    r, err := chatroom.GetChatRoom(room)
    if err != nil {
        c.Response.Status = http.StatusNotFound
        c.Response.ContentType = "application/text"
        log.Println("404")
        return c.Render()
    }

    log.Println("200")
    return c.Render(r)
}

func (c Room) Create(room string) revel.Result {
    r, err := chatroom.CreateChatRoom(room)
    if err != nil {
        c.Response.Status = http.StatusBadRequest
        c.Response.ContentType = "application/text"
        log.Println("400")
        return c.Render()
    }

    log.Println("200")
    return c.Render(r)
}

func (c Room) Delete(room string) revel.Result {
    r, err := chatroom.DeleteChatRoom(room)
    if err != nil {
        c.Response.Status = http.StatusBadRequest
        c.Response.ContentType = "application/text"
        log.Println("400")
        return c.Render()
    }

    log.Println("200")
    return c.Render(r)
}

func (c Room) Socket(room string, ws *websocket.Conn) revel.Result {
    // Join the room if it exists
    r, err := chatroom.GetChatRoom(room)
    if err != nil {
        c.Response.Status = http.StatusBadRequest
        return c.Render()
    }

    user := c.Params.Values.Get("user")

    subscription := chatroom.Subscribe(r)
    defer subscription.Cancel(r)

    chatroom.JoinRoom(r, user)
    defer chatroom.LeaveRoom(r, user)

    //send down archive
    for _, event := range subscription.Archive {
        if websocket.JSON.Send(ws, &event) != nil {
            //They dc'd
            return nil
        }
    }

    //need to multiplex between websocket messages and subscription events
    newMessages := make(chan string)
    go func() {
        var msg string
        for {
            err := websocket.Message.Receive(ws, &msg)
            if err != nil {
                close(newMessages)
                return
            }
            newMessages <- msg
        }
    }()

    //select between subscription and websocket
    for {
        select {
        case event := <-subscription.New:
            if websocket.JSON.Send(ws, &event) != nil {
                //they dc'd
                return nil
            }
        case msg, ok := <-newMessages:
            if !ok {
                //channel's not open, so user must have left
                return nil
            }

            //otherwise, say
            chatroom.SayRoom(r, user, msg)
        }
    }

    return nil
}
