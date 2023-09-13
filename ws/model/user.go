package model

import (
	"github.com/dgrijalva/jwt-go"
)

type MyClaims struct {
	Account string `json:"account"`
	jwt.StandardClaims
}

/*type Client struct {
	//UUID     string
	Account  string
	UserName any
	Socket   *websocket.Conn
	Send     chan Message //消息
}*/

const (
	Online  int = 1
	Offline int = 0
)

type DBMessage struct {
	Message Message `json:"message"`
	Status  int
}

const (
	Text    int = 1
	Picture int = 2
	Video   int = 3
)

type Message struct {
	SenderName    any    `json:"sendername"`
	SenderAccount string `json:"senderaccount"`
	//SenderUUID    string `json:"senderuuid,omitempty"`
	RecipientAccount string `json:"recipientAccount,omitempty"`
	Content          string `json:"content,omitempty"`
	Type             int    `json:"type"`
	Time             string `json:"time"`
}
