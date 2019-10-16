package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id   int
	Name string
}

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)

	orm.RegisterDataBase("default", "mysql", "root:123456@/socket?charset=utf8")

	orm.RegisterModel(new(User))

	orm.RunSyncdb("default", false, true)
}

func main() {
	service := ":1201"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func handle(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	for {
		request := make([]byte, 128)
		array := make([]string, 128)
		n, err := conn.Read(request)
		if err != nil {
			break
		}
		if n == 0 {
			break
		}
		array = strings.Split(string(request), ";")
		o := orm.NewOrm()
		user := User{Name: array[0]}
		if array[1] == "1" {
			err := o.Read(&user, "Name")
			if err == orm.ErrNoRows {
				_, err = conn.Write([]byte("3"))
				checkError(err)
			} else {
				_, err = conn.Write([]byte("1"))
				checkError(err)
			}
		} else {
			o.Begin()
			if created, _, err := o.ReadOrCreate(&user, "Name"); err == nil {
				if created {
					o.Commit()
					_, err = conn.Write([]byte("1"))
					checkError(err)
				} else {
					_, err = conn.Write([]byte("2"))
					checkError(err)
				}
			} else {
				o.Rollback()
				_, err = conn.Write([]byte("0"))
				checkError(err)
			}
		}
	}
}
