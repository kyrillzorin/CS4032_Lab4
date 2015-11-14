package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var ClientID = 0
var RoomID = 0
var Clients map[int]*ChatClient = make(map[int]*ChatClient)
var Rooms map[int]*ChatRoom = make(map[int]*ChatRoom)

type ChatClient struct {
	Name string
	Conn net.Conn
}

type ChatRoom struct {
	Name    string
	Clients map[int]struct{}
}

func NewClient(name string, conn net.Conn) *ChatClient {
	c := new(ChatClient)
	c.Name = name
	c.Conn = conn
	return c
}

func NewRoom(name string) *ChatRoom {
	r := new(ChatRoom)
	r.Name = name
	r.Clients = make(map[int]struct{})
	return r
}

func getClientID(clientName string, conn net.Conn) int {
	clientID := -1
	for key, value := range Clients {
		if value.Name == clientName {
			clientID = key
		}
	}
	if clientID == -1 {
		client := NewClient(clientName, conn)
		ClientID++
		clientID = ClientID
		Clients[clientID] = client
	}
	return clientID
}

func addClientToRoom(chatRoom string, clientID int) int {
	roomKey := -1
	for key, value := range Rooms {
		if value.Name == chatRoom {
			roomKey = key
		}
	}
	if roomKey == -1 {
		room := NewRoom(chatRoom)
		RoomID++
		roomKey = RoomID
		Rooms[roomKey] = room
	}
	Rooms[roomKey].Clients[clientID] = struct{}{}
	return roomKey
}

func removeClientFromRoom(clientID int, roomID int) {
	delete(Rooms[roomID].Clients, clientID)
}

func getClientChatrooms(clientID int) []int {
	chatRooms := []int{}
	for key, value := range Rooms {
		_, ok := value.Clients[clientID]
		if ok {
			chatRooms = append(chatRooms, key)
		}
	}
	return chatRooms
}

func sendMessageToChatroom(roomID int, clientName string, message string) {
	_, ok := Rooms[roomID]
	if ok {
		messageToSend:= "CHAT:" + strconv.Itoa(roomID) + "CLIENT_NAME:" + clientName + "MESSAGE:" + message
		for key, _ := range Rooms[roomID].Clients {
			fmt.Fprintf(Clients[key].Conn, messageToSend)
		}
	} else {
		fmt.Println("ERROR: chat room " + strconv.Itoa(roomID) + "not found in database")
	}
}

func handleClient(message string, conn net.Conn) {
	if strings.HasPrefix(message, "JOIN_CHATROOM:") {
		handleJoinRoom(message, conn)
	} else if strings.HasPrefix(message, "LEAVE_CHATROOM:") {
		handleLeaveRoom(message, conn)
	} else if strings.HasPrefix(message, "CHAT:") {
		handleChat(message, conn)
	} else if strings.HasPrefix(message, "DISCONNECT:") {
		handleDisconnect(message, conn)
	} else {
		handleDefault()
	}
}

func handleJoinRoom(message string, conn net.Conn) bool {
	status := true
	chatRoom := strings.TrimPrefix(message, "JOIN_CHATROOM:")
	chatRoom = strings.TrimSpace(chatRoom)
	connReader := bufio.NewReader(conn)
	message, _ = connReader.ReadString('\n')
	clientIP := ""
	if strings.HasPrefix(message, "CLIENT_IP:") && status {
		clientIP = strings.TrimPrefix(message, "CLIENT_IP:")
		clientIP = strings.TrimSpace(clientIP)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	clientPort := ""
	if strings.HasPrefix(message, "PORT:") && status {
		clientPort = strings.TrimPrefix(message, "PORT:")
		clientPort = strings.TrimSpace(clientPort)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	clientName := ""
	if strings.HasPrefix(message, "CLIENT_NAME:") && status {
		clientName = strings.TrimPrefix(message, "CLIENT_NAME:")
		clientName = strings.TrimSpace(clientName)
	} else {
		status = false
	}
	if !status {
		return status
	}
	clientID := getClientID(clientName, conn)
	roomID := addClientToRoom(chatRoom, clientID)
	messageToSend := "JOINED_CHATROOM:" + chatRoom + "\n" +
		"SERVER_IP:" + EXT_IP + "\n" +
		"PORT:" + PORT + "\n" +
		"ROOM_REF:" + strconv.Itoa(roomID) + "\n" +
		"JOIN_ID:" + strconv.Itoa(clientID) + "\n"
	fmt.Fprintf(conn, messageToSend)
	messageToSend = clientName + " has joined this chatroom."
	sendMessageToChatroom(roomID, clientName, messageToSend)
	return status
}

func handleLeaveRoom(message string, conn net.Conn) bool {
	status := true
	roomID := strings.TrimPrefix(message, "LEAVE_CHATROOM:")
	roomID = strings.TrimSpace(roomID)
	connReader := bufio.NewReader(conn)
	message, _ = connReader.ReadString('\n')
	clientID := ""
	if strings.HasPrefix(message, "JOIN_ID:") && status {
		clientID = strings.TrimPrefix(message, "JOIN_ID:")
		clientID = strings.TrimSpace(clientID)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	clientName := ""
	if strings.HasPrefix(message, "CLIENT_NAME:") && status {
		clientName = strings.TrimPrefix(message, "CLIENT_NAME:")
		clientName = strings.TrimSpace(clientName)
	} else {
		status = false
	}
	if !status {
		return status
	}
	messageToSend := "LEFT_CHATROOM:" + roomID + "\n" +
		"JOIN_ID:" + clientID + "\n"
	fmt.Fprintf(conn, messageToSend)
	messageToSend = clientName + " has left this chatroom."
	roomKey, _ := strconv.Atoi(roomID)
	clientKey, _ := strconv.Atoi(clientID)
	sendMessageToChatroom(roomKey, clientName, messageToSend)
	removeClientFromRoom(clientKey, roomKey)
	return status
}

func handleChat(message string, conn net.Conn) bool {
	status := true
	roomID := strings.TrimPrefix(message, "CHAT:")
	roomID = strings.TrimSpace(roomID)
	connReader := bufio.NewReader(conn)
	message, _ = connReader.ReadString('\n')
	clientID := ""
	if strings.HasPrefix(message, "JOIN_ID:") && status {
		clientID = strings.TrimPrefix(message, "JOIN_ID:")
		clientID = strings.TrimSpace(clientID)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	clientName := ""
	if strings.HasPrefix(message, "CLIENT_NAME:") && status {
		clientName = strings.TrimPrefix(message, "CLIENT_NAME:")
		clientName = strings.TrimSpace(clientName)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	text := ""
	if strings.HasPrefix(message, "MESSAGE:") && status {
		text = strings.TrimPrefix(message, "MESSAGE:")
		message, _ = connReader.ReadString('\n')
		message = strings.TrimSpace(message)
		for message != "" {
			text += message
			message, _ = connReader.ReadString('\n')
			message = strings.TrimSpace(message)
		}
	} else {
		status = false
	}
	if !status {
		return status
	}
	roomKey, _ := strconv.Atoi(roomID)
	sendMessageToChatroom(roomKey, clientName, text)
	return status
}

func handleDisconnect(message string, conn net.Conn) bool {
	status := true
	clientIP := strings.TrimPrefix(message, "DISCONNECT:")
	clientIP = strings.TrimSpace(clientIP)
	connReader := bufio.NewReader(conn)
	message, _ = connReader.ReadString('\n')
	clientPort := ""
	if strings.HasPrefix(message, "PORT:") && status {
		clientPort = strings.TrimPrefix(message, "PORT:")
		clientPort = strings.TrimSpace(clientPort)
	} else {
		status = false
	}
	message, _ = connReader.ReadString('\n')
	clientName := ""
	if strings.HasPrefix(message, "CLIENT_NAME:") && status {
		clientName = strings.TrimPrefix(message, "CLIENT_NAME:")
		clientName = strings.TrimSpace(clientName)
	} else {
		status = false
	}
	if !status {
		return status
	}
	clientID := getClientID(clientName, conn)
	messageToSend := clientName + " has left this chatroom."
	chatRooms := getClientChatrooms(clientID)
	for i := range chatRooms {
		roomID := chatRooms[i]
		sendMessageToChatroom(roomID, clientName, messageToSend)
		removeClientFromRoom(clientID, roomID)
	}
	Clients[clientID].Conn.Close()
	delete(Clients, clientID)
	return status
}

func handleDefault() bool {
	fmt.Println("No such command")
	return false
}
