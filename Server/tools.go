package Server

import (
	uuid2 "github.com/satori/go.uuid"
	"math/rand"
)

func GetPort() int  {
	// 产生10000-65535之间端口
	num:=rand.Intn(65535-10000)+10000
	//后续 判断此端口是否被占用
	return num
}

func GetUuid() string {
	uid:=uuid2.NewV4()
	return uid.String()
}