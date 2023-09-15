# IM
简单的IM即时通讯系统

功能：
1.支持登录注册，添加、删除好友
2.支持文本、图片传输

实现：
1.使用Websocket，json格式进行通信
2.使用mysql储存用户信息，redis储存好友信息，mongodb储存聊天记录，rabbmitmq用作消息队列
3.服务部署在云服务器上
