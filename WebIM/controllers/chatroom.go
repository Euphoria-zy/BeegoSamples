// Copyright 2013 Beego Samples authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package controllers

import (
	"container/list"
	"time"

	"github.com/astaxie/beego"
	"github.com/beego/samples/WebIM/models"
	"github.com/gorilla/websocket"
)

type Subscription struct {
	Archive []models.Event      // All the events from the archive.
	New     <-chan models.Event // New events coming in.
}

func newEvent(ep models.EventType, user, msg string) models.Event {
	return models.Event{ep, user, int(time.Now().Unix()), msg}
}

func Join(user string, ws *websocket.Conn) {
	subscribe <- Subscriber{Name: user, Conn: ws}
}

func Leave(user string) {
	unsubscribe <- user
}

type Subscriber struct {
	Name string
	Conn *websocket.Conn // Only for WebSocket users; otherwise nil.
}

var (
	// Channel for new join users.
	subscribe = make(chan Subscriber, 10) //订阅通道，加入聊天室的用户
	// Channel for exit users.
	unsubscribe = make(chan string, 10) //离开的用户：用户名
	// Send events here to publish them.
	publish = make(chan models.Event, 10) //发送消息的通道
	// Long polling waiting list.
	waitingList = list.New() //链表
	subscribers = list.New()
)

// This function handles all incoming chan messages.处理所有的消息
func chatroom() {
	for {
		select {
		case sub := <-subscribe:
			if !isUserExist(subscribers, sub.Name) {
				subscribers.PushBack(sub) // Add user to the end of list.订阅者队列
				// Publish a JOIN event.
				publish <- newEvent(models.EVENT_JOIN, sub.Name, "")              //新加入的用户，发送新用户加入的系统消息
				beego.Info("New user:", sub.Name, ";WebSocket:", sub.Conn != nil) //打印log
			} else {
				beego.Info("Old user:", sub.Name, ";WebSocket:", sub.Conn != nil)
			}
		case event := <-publish: //如果有需要发送的消息
			// Notify waiting list.清空等待队列
			for ch := waitingList.Back(); ch != nil; ch = ch.Prev() { //List.Back()：取出队尾节点;List.Prev()：取出上一个节点
				ch.Value.(chan bool) <- true //元素类型为bool 类型的chan
				waitingList.Remove(ch)
			}

			broadcastWebSocket(event) //websocket 发送消息
			models.NewArchive(event)  //longpooling发送消息

			if event.Type == models.EVENT_MESSAGE {
				beego.Info("Message from", event.User, ";Content:", event.Content)
			}
		case unsub := <-unsubscribe: //如果有用户退出
			for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
				if sub.Value.(Subscriber).Name == unsub { //list.element.value是接口类型interface{},通过Value.(类型)可转化为指定类型
					subscribers.Remove(sub)
					// Clone connection.
					ws := sub.Value.(Subscriber).Conn
					if ws != nil {
						ws.Close()
						beego.Error("WebSocket closed:", unsub)
					}
					publish <- newEvent(models.EVENT_LEAVE, unsub, "") // Publish a LEAVE event.
					break
				}
			}
		}
	}
}

// 在导入包的时候执行
func init() {
	go chatroom() //开启一个聊天室线程
}

// 判断是否为新加入聊天室的用户
func isUserExist(subscribers *list.List, user string) bool {
	for sub := subscribers.Front(); sub != nil; sub = sub.Next() {
		if sub.Value.(Subscriber).Name == user { //如果user和队首元素一致，则为新加入的用户
			return true
		}
	}
	return false
}
