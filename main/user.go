package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建user
func NewUser(conn net.Conn, server *Server) (user *User) {
	userAddr := conn.RemoteAddr().String()
	user = &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动监听当前user的chan消息的goroutine
	go user.ListenMessage()
	return
}

// 用户的上线业务
func (u *User) Online() {
	fmt.Printf("%s 连接建立成功\n", u.Name)
	// 用户上线，加入OnlineMap
	u.server.RWLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.RWLock.Unlock()
	// 广播当前用户上线消息
	u.server.BroadCast(u, "已上线")
}

// 用户的下线业务
func (u *User) Offline() {
	// 用户下线，将用户从OnlineMap中删除
	u.server.RWLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.RWLock.Unlock()
	// 广播当前用户下线消息
	u.server.BroadCast(u, "已下线")
}

// 给当前User对应的Socker发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// 用户处理消息的业务
func (u *User) DoMessage(msg string) {
	fmt.Printf("msg的内容是：%s。\n", msg)
	if msg == "who" {
		// 查询返回当前在线用户
		u.server.RWLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[%s] %s 在线...\n", user.Addr, user.Name)
			u.SendMsg(onlineMsg)
		}
		u.server.RWLock.Unlock()
	} else if strings.HasPrefix(msg, "rename|") {
		// rename|xxx
		renameStrSlice := strings.Split(msg, "|")
		newName := strings.Join(renameStrSlice[1:], "|")
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.SendMsg("用户名" + newName + "已被使用\n")
		} else {
			u.server.RWLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.RWLock.Unlock()
			u.Name = newName
			u.SendMsg("您已经更新用户名：" + u.Name + "\n")
		}
	} else if strings.HasPrefix(msg, "to|") {
		// 获取对方用户名
		toSomeOneStrSlice := strings.Split(msg, "|")
		someOnestr := strings.Join(toSomeOneStrSlice[1:2], "|")
		// 判断用户名是否存在
		u.server.RWLock.RLock()
		someOneObj, ok := u.server.OnlineMap[someOnestr]
		u.server.RWLock.RUnlock()
		if ok {
			// 获取消息内容
			msg := strings.Join(toSomeOneStrSlice[2:], "|")
			someOneObj.SendMsg(u.Name + "悄悄对你说：:" + msg)
		} else {
			u.SendMsg("你想对话的用户:" + someOnestr + " 不存在\r\n")
		}
	} else {
		u.server.BroadCast(u, msg)
	}

}

// 监听当前User channel，有消息就发送给客户端
func (u *User) ListenMessage() {
	// for {
	// 	msg := <-u.C
	// 	// // fmt.Println(msg)
	// 	u.conn.Write([]byte(msg + "\n"))
	// }

	for msg := range u.C {
		u.conn.Write([]byte(msg + "\n"))
	}

}
