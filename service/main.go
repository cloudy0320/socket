package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id      int
	Name    string
	message *message `orm:"reverse(one)"`
}

type user struct {
	User
	conn net.Conn
}

type message struct {
	Id      int
	From    string
	To      *User `orm:"rel(one)"`
	Message string
	Stats   int `orm:"default(1)"`
}

var users []user

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	orm.RegisterDataBase("default", "mysql", "root:123456@/socket?charset=utf8")

	orm.RegisterModel(new(User), new(message))

	orm.RunSyncdb("default", false, true)
}

func main() {
	service := ":1200"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	get()
	for {
	A:
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		request := make([]byte, 128)
		_, err = conn.Read(request)
		checkError(err)
		array := make([]string, 128)
		array = strings.Split(string(request), ";")
		if array[2] == "2" {
			newUser := new(user)
			newUser.conn = conn
			newUser.Name = array[0]
			users = append(users, *newUser)
		} else if array[2] == "1" {
			fmt.Println(users)
			for i, v := range users {
				if v.Name == array[0] && v.conn == nil {
					users[i].conn = conn
				} else if v.Name == array[0] && v.conn != nil {
					_, err = conn.Write([]byte("这个账号已经被登录了！"))
					checkError(err)
					conn.Close()
					goto A
				}
			}
		} else {
			continue
		}
		if array[1] == "1" {
			check(array[0], conn)
			go handleClient(conn, array[0])
		} else {
			check(array[0], conn)
			go chat(conn, array[0])
		}
	}
}

func handleClient(conn net.Conn, name string) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	request := make([]byte, 128)                          // set maxium request length to 128B to prevent flood attack                                // close connection before exit
	for {
		read_len, err := conn.Read(request)
		//fmt.Printf("长度：%d", read_len)
		if err != nil {
			fmt.Println()
			del(conn)
			conn.Close()
			break
		}
		if read_len == 0 {
			del(conn)
			conn.Close()
			break // connection already closed by client
		} else {
			index := strings.Index(string(request), "说")
			if string(request[index+3:index+8]) == "/list" {
				var online []string
				for _, v := range users {
					if name == v.Name {
						online = append(online, v.Name+"(你自己)")
					} else if v.conn != nil {
						online = append(online, v.Name+"(在线)")
					} else {
						online = append(online, v.Name+"(不在线)")
					}
				}
				list := strings.Join(online, "\n")
				_, err = conn.Write([]byte(list))
				checkError(err)
			} else if string(request[index+1:index+8]) == "/change" {
				go chat(conn, name)
				return
			} else {
				for _, v := range users {
					if v.conn != nil {
						_, err = v.conn.Write(request)
						checkError(err)
					}
				}
			}
		}

		request = make([]byte, 128) // clear last read content
	}
}

func chat(conn net.Conn, name string) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // set 2 minutes timeout
	flag := 1
Again:
	toName := make([]byte, 128)
	request := make([]byte, 128)
	var online []string
	me := new(user)
	for _, v := range users {
		if name == v.Name {
			online = append(online, v.Name+"(你自己)")
			*me = v
		} else if v.conn != nil {
			online = append(online, v.Name+"(在线)")
		} else {
			online = append(online, v.Name+"(不在线)")
		}
	}
	response := strings.Join(online, "\n")
	conn.Write([]byte(response))

	read_len, err := conn.Read(toName)
	if err != nil {
		fmt.Println(err)
		conn.Close()
		del(conn)
		return
	}
	if read_len == 0 {
		conn.Close()
		del(conn)
		return // connection already closed by client
	}
	for _, v := range users {
		if strings.Split(string(toName), ";")[0] == v.Name && v.conn != nil {
			for {
				read_len, err := conn.Read(request)
				if err != nil {
					fmt.Println(err)
					flag = 0
					break
				}
				if read_len == 0 {
					flag = 0
					break // connection already closed by client
				} else if string(request[:5]) == "/quit" {
					goto Again
				} else if string(request[:7]) == "/change" {
					go handleClient(conn, name)
					return
				} else {
					_, err = v.conn.Write([]byte(name + "对你说" + string(request)))
					checkError(err)
				}
				request = make([]byte, 128) // clear last read content
			}
		} else if strings.Split(string(toName), ";")[0] == v.Name && v.conn == nil {
			_, err := conn.Write([]byte("此人不在线，你可以发送离线消息"))
			checkError(err)
			for {
				newMessage := new(message)
				read_len, err := conn.Read(request)
				if err != nil {
					fmt.Println(err)
					flag = 0
					break
				}
				if read_len == 0 {
					flag = 0
					break // connection already closed by client
				} else if string(request[:5]) == "/quit" {
					goto Again
				} else if string(request[:7]) == "/change" {
					go handleClient(conn, name)
					return
				} else {
					o := orm.NewOrm()
					o.Begin()
					newMessage.Message = string(request)
					newMessage.From = me.Name
					newMessage.To = &v.User
					_, err := o.Insert(newMessage)
					fmt.Println(err)
					if err == nil {
						o.Commit()
					}
				}
				request = make([]byte, 128) // clear last read content
			}
		}
		if flag == 0 {
			break
		}
	}
	if flag == 1 {
		_, err := conn.Write([]byte("没有这个人，请重新输入\n"))
		checkError(err)
		goto Again
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func del(conn net.Conn) {
	for i, v := range users {
		if v.conn == conn {
			users[i].conn = nil
		}
	}
}

func get() {
	var usersData []User
	userTemp := new(user)
	o := orm.NewOrm()
	_, err := o.QueryTable("user").All(&usersData)
	for _, v := range usersData {
		userTemp.User = v
		users = append(users, *userTemp)
	}
	if err != nil {
		fmt.Println("获取数据库失败")
		os.Exit(1)
	}
}

func check(name string, conn net.Conn) {
	var response []string
	o := orm.NewOrm()
	userTemp := User{Name: name}
	err := o.Read(&userTemp, "Name")
	if err == orm.ErrNoRows {
		fmt.Println("查询不到")
	}
	var messages []message
	n, err := o.QueryTable("message").Filter("To", userTemp.Id).Filter("stats", 0).All(&messages)
	if err == nil && n != 0 {
		for _, v := range messages {
			response = append(response, v.From+"对你说"+v.Message)
			v.Stats = 1
			o.Update(&v)
		}
		response = append(response, "以上是离线信息")
		s := strings.Join(response, "\n")
		conn.Write([]byte(s))
	}
	fmt.Println(err)
}
