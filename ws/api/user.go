package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"ws/api/middleware"
	"ws/dao"
	"ws/model"
	"ws/service"
	"ws/utils"
)

type Client struct {
	//UUID     string
	Account  string
	UserName any
	Socket   *websocket.Conn
	//Send     chan model.Message //用mq后舍弃
}

type ClientManager struct { //连接池，可以管理所有的终端连接，并提供注册、注销、续期功能。
	Clients    map[string]*Client //当前在线用户
	Unregister chan *Client       //下线用户
}

var Manager = ClientManager{
	Clients:    make(map[string]*Client),
	Unregister: make(chan *Client),
}

func (c *Client) ClientSendMessage() { //客户向服务器发送信息
	defer func() {
		Manager.Unregister <- c
		_ = c.Socket.Close()
	}()

	for {
		_, msg, err := c.Socket.ReadMessage()
		if err != nil {
			log.Println("消息异常 ", err)
			break
			//_ = c.Socket.Close()
		}

		if msg != nil {
			wsmsg := new(model.Message)

			json.Unmarshal(msg, &wsmsg) //将json传给结构体
			//SenderUID := c.UUID
			message := model.Message{
				SenderName:       c.UserName,
				SenderAccount:    c.Account,
				RecipientAccount: wsmsg.RecipientAccount,
				Content:          wsmsg.Content,
				Type:             wsmsg.Type,
				Time:             time.Now().Format(time.RFC3339),
			} //格式化为发送给目标的信息

			if dao.IsFriend(message.SenderAccount, message.RecipientAccount) { //判断是否为好友
				MQmsg, _ := json.Marshal(message)
				service.SendMessageToMQ(MQmsg) //把消息放进消息队列

			} else { //不是好友就把消息转回发送者，且不保存进mongo
				MQmsg, _ := json.Marshal(model.Message{
					SenderName:       message.SenderName,
					SenderAccount:    message.SenderAccount,
					RecipientAccount: message.SenderAccount,
					Content:          "对方不是你的好友，发送失败！",
					Type:             model.Text,
					Time:             time.Now().Format(time.RFC3339),
				})
				service.SendMessageToMQ(MQmsg)
			}

		}

	}
}

func (c *Client) SeverSendMessage() { //服务器向用户发送信息
	defer func() {
		_ = c.Socket.Close()
		dao.RedisClose()
	}()
	//msg, _ := json.Marshal(c)

	c.Socket.WriteJSON(gin.H{
		"account":  c.Account,
		"username": c.UserName,
	})

	msg := dao.SendOfflineMessage(c.Account)
	for _, value := range msg {
		c.Socket.WriteMessage(websocket.TextMessage, value)
		fmt.Println(c.UserName, "离线信息已发送", string(value))
	} //用户上线先发送离线信息

	for {

		msgs := service.ConsumeMessage()

		for MQmsg := range msgs { //从队列中拿消息

			var mqmessage model.Message

			//message := string(MQmsg.Body)
			//log.Printf("从队列中接收的消息为: %s\n", message)

			json.Unmarshal(MQmsg.Body, &mqmessage) //将数据反格式化

			switch mqmessage.Type {
			case model.Text:
				break

			case model.Picture:
				base64String := mqmessage.Content
				//data := strings.SplitN(base64String, ",", 2)
				decodedImage, err := base64.StdEncoding.DecodeString(base64String)
				if err != nil {
					fmt.Println("解码失败:", err)
					return
				}

				a := time.Now().Unix()

				filePath := fmt.Sprintf("%s%d%s", "./images/", a, ".png") //图片存放路径
				fp := strings.TrimPrefix(filePath, ".")
				err = os.WriteFile(filePath, decodedImage, 0644) //把文件写进路径
				if err != nil {
					log.Println("保存图片失败:", err)
					return
				}

				fmt.Println("图片保存成功:", filePath)
				mqmessage.Content = "http://47.113.220.87:8080" + fp

				break

			case model.Video:
				base64String := mqmessage.Content
				videoData, err := base64.StdEncoding.DecodeString(base64String)
				if err != nil {
					log.Println("视频解码失败：", err)
				}
				filePath := fmt.Sprintf("%s%d%s", "./video/", time.Now().Unix(), ".mp4") //图片存放路径

				err = ioutil.WriteFile(filePath, videoData, 0644)
				if err != nil {
					log.Println("视频保存失败：", err)
				}
				log.Println("视频保存成功")
				fp := strings.TrimPrefix(filePath, ".")
				mqmessage.Content = "http://47.113.220.87:8080" + fp
				break
			}

			flag := false //假设不在线

			for account, rec := range Manager.Clients {

				//fmt.Println(account, message.RecipientAccount)
				if strings.Compare(mqmessage.SenderAccount, mqmessage.RecipientAccount) == 0 { //信息发送给自己，不记录进数控库
					mesg, _ := json.Marshal(mqmessage.Content)

					rec.Socket.WriteMessage(websocket.TextMessage, mesg)

				} else if strings.Compare(account, mqmessage.RecipientAccount) == 0 { //如果该用户在线
					mesg, _ := json.Marshal(mqmessage)
					rec.Socket.WriteMessage(websocket.TextMessage, mesg) //将信息发送给接收者所在连接

					dbmessage := model.DBMessage{
						mqmessage,
						model.Online,
					}
					//dbmsg, _ := json.Marshal(mqmessage)
					dao.SaveMessage(dbmessage) //将信息保存至数据库
					log.Println(c.UserName, "已发送信息：", mqmessage.Content)
					flag = true
				}
			}

			if !flag { //用户不在线
				dbmessage := model.DBMessage{
					mqmessage,
					model.Offline,
				}
				dao.SaveMessage(dbmessage)
				fmt.Println("用户不在线，消息已经保存", mqmessage.Content)
			}

		}

	}

}

func StartSocket(c *gin.Context) {
	conn, _ := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}).Upgrade(c.Writer, c.Request, nil) //升级成ws协议

	//Uuid := service.UUID()
	account, _ := c.Get("account")
	a := fmt.Sprintf("%s", account)
	//c.Set("uuid", Uuid)

	NewClient := &Client{
		Account:  a,
		UserName: dao.FindUsernameFromAccount(a),
		Socket:   conn,
		//Send:     make(chan model.Message),
	}

	Manager.Clients[NewClient.Account] = NewClient //进行连接的在线用户
	log.Println(NewClient.UserName, "上线")

	go NewClient.UnRegister()
	go NewClient.SeverSendMessage()
	go NewClient.ClientSendMessage()

}

func (c *Client) UnRegister() {
	for {
		select {
		case unre := <-Manager.Unregister:
			delete(Manager.Clients, unre.Account)
			//close(unre.Send)
			log.Println(c.UserName, "下线")
		}
	}
}

func GetUUIDAndContent(c *gin.Context) []string {

	uid := c.Query("uuid")
	content := c.Query("content")
	var a = []string{uid, content}
	return a
}

func Register(c *gin.Context) {
	/*if err := c.ShouldBind(&model.User{}); err != nil {
		utils.RespSuccess(c, "verification failed")
		return
	}*/

	account := c.PostForm("account")
	password := c.PostForm("password")
	username := c.PostForm("username")

	if dao.AddUser(account, username, []byte(password)) { //添加用户
		utils.RespSuccess(c, "注册成功")
	} else {
		utils.RespFail(c, "注册失败")
	}
	dao.RegisterFriendQueue(account)
}

func Login(c *gin.Context) {
	/*if err := c.ShouldBind(&model.UserLogin{}); err != nil {
		utils.RespFail(c, "verification failed")
		return
	}*/

	account := c.PostForm("account")
	password := c.PostForm("password")

	selectPassword := dao.SelectPasswordFromAccount(account)

	//err := bcrypt.CompareHashAndPassword(selectPassword, []byte(password))
	if strings.Compare(password, string(selectPassword)) != 0 {
		print(selectPassword)
		utils.RespFail(c, "密码或账户名错误")
		return
	}

	claim := model.MyClaims{
		Account: account, // 自定义字段
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 2).Unix(), // 过期时间
			Issuer:    "cxk",                                // 签发人
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, _ := token.SignedString(middleware.Secret)
	//uid, _ := c.Get("uuid")
	c.JSON(http.StatusOK, gin.H{
		"message": "欢迎," + dao.FindUsernameFromAccount(account),
		"token":   tokenString,
	})
}

func SearchForFriend(c *gin.Context) {
	c.JSON(200, dao.GetAllUser())
}

func AddFriend(c *gin.Context) {
	account := c.PostForm("friend")
	myaccount, _ := c.Get("account")
	dao.RedisAddFriend(myaccount, account)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "添加好友成功",
	})
}

func DeleteFriend(c *gin.Context) {
	faccount := c.PostForm("friend")
	myaccount, _ := c.Get("account")
	dao.RedisDeleteFriend(faccount, myaccount)
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除好友成功",
	})
}
