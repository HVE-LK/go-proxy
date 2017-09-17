package main

import (
	"net"
	"strconv"
	"fmt"
	"sync"
)
/*
隧道
*/
//隧道连接池
type Turn struct {
	Uid       string
	Tid       int
	Ln        net.Listener //隧道
	Tcp       net.Conn
	OnlyRead  chan *DataPipe
	OnlyWrite chan *DataPipe
}

//数据数据格式
type DataPipe struct {
	Headers []byte //头部信息
	Data    []byte // body
	Tid     int    //隧道id
	Uid     string
}

//心跳
type heartbeat struct {
	version string
	status  bool
	time    int //心跳时间 毫秒
}

//数据统计
type ProxyObjects struct {
	Tid               int `隧道id`
	Uid               int     `用户id`
	State             bool    `用户状态`
	UpstreamBandWidth float32 `上行流量`
	DownlinkBandWidth float32 `下行流量`
	Count             int     `重连计数`
}

type turnFac struct {
	/*创建隧道，创建对应服务 */
	turn *Turn
}
var (
	tid int=0
	TurnLink    = make(map[int]*Turn)
	TurnState=make(map[int]*ProxyObjects)
	lock sync.Mutex
)
//隧道初始化
func Start()  {

}
//发送隧道数据
func SendData(data interface{}){

}
//创建新隧道
func (this *turnFac)CreateTurn(ServicePort int,PrivatePort int)error{
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(ServicePort))
	if err != nil {
		fmt.Errorf("start fail in %s %s", ServicePort, err)
		return err
	}
	for {
		conn, err := ln.Accept() //一条隧道 多个信道
		if err != nil {
			//对象写入错误信息
			continue
		}
		//生成客户端uuid 保存客户端连接
		turnId:=editTid()
		currTurn := &Turn{
			GetUuid(),
			turnId,
			ln,
			conn,
			make(chan *DataPipe),
			make(chan *DataPipe),
		}
		private:=new(ProxyService)
		private.CreateClientService(PrivatePort,"http",this.turn.Tid)
		TurnLink[this.turn.Tid],this.turn = currTurn,currTurn
		go this.handleConnection()
	}
}
func (this *turnFac)CloseTurn()(err error){
	err=TurnLink[this.turn.Tid].Ln.Close()
	if err!=nil {
		return err
	}
	return nil
}
func (this *turnFac)GetTurnState()(*ProxyObjects){
	return  TurnState[this.turn.Tid]
}
func (this *turnFac)handleConnection()  {
	conn:=this.turn.Tcp
	buffer:=make([]byte,65535)
	go func() {
		onceData:=new(DataPipe)
		for{
			n,err:=conn.Read(buffer)
			if err==nil{
				onceData.Data=buffer[:n]
				this.turn.OnlyRead<-onceData
			}else {
				close(this.turn.OnlyRead)
			}
		}
	}()
	go func() {
		conn.Write((<-this.turn.OnlyWrite).Data)
	}()
}

func editTid()  (tid int){
	lock.Lock()
	tid++
	defer lock.Unlock()
	return tid
}