// reverse-proxy project main.go 客户端
package Server

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

var serviceConfig bool

// 后期加入证书认证
// 服务端配置

type SConfig struct {
	deal    string `代理服务类型`
	port    int    `代理服务监听端口`
	foreign uint16 `对外服务端口`
}

// 客户端配置
type CConfig struct {
	agencyService string `被代理服务`
	agentAddr     string `代理服务地址`
	localSePport  uint16 `客户端代理服务端口`
}

type offer struct {
	ClientProxyType string
	ClientPort      int
}

type answer struct {
	Status           bool
	ServiceTurnPort  int
	ServiceProxyType string
	ServicePort      int
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

//心跳
type heartbeat struct {
	version string
	status  bool
	time    int //心跳时间 毫秒
}

//隧道连接池
type Turn struct {
	Uid       string
	Tid       int
	Ln        net.Listener //隧道
	Tcp       net.Conn
	OnlyRead  chan []byte
	OnlyWrite chan []byte
}

//数据管道格式
type DataPipe struct {
	Headers string //头部信息
	Data    []byte // body
	Tid     int    //信道id
	Uid     string
}

// 服务端记录被代理的对象
var (
	serviceType uint
	tid         = 0
	sConfig     = new(SConfig)
	turnLink    = make(map[string]*Turn)
)

//隧道计数器

func init() {
	flag.UintVar(&serviceType, "servicetype", 0, "Service type, 0: proxy service host, 1: proxy client")
	flag.StringVar(&sConfig.deal, "stype", "http", "Proxy service protocols, such as HTTP, HTTPS, TCP, UDP, SCOKET")
	flag.IntVar(&sConfig.port, "sport", 8080, "Agent service listening port, default 8080")
	flag.Parse()
}

func main() {
	/*
		定义command 启动服务
		默认服务端和客户端使用tcp连接，同一网段使用同一条隧道，避免造成不必要的资源浪费。
		默认心跳时间：30s 后续根据情况调整
		断线重连机制：如果重连时间超过1分钟 则放弃重新连接，否则服务端发起请求连接信号
		默认只记录系统产生的日志
		可代理的协议：http，https，tcp，udp，socket
		-h 帮助
		-c 客户端 作为client必须被指定
		-cl 被代理服务地址
		-cp 客户端服务启动端口，不配置系统默认自己选择

		-s service 作为service必须被指定
		-cp service 启动端口，作为服务端必须被指定
		-st 代理协议 作为服务端必须被指定
	*/
	//  思考问题：代理实现方案，如何防止粘包情况(自定义包)，并发量，
	fmt.Println("start")
	fmt.Println(sConfig.deal, sConfig.port)
	if serviceType == 0 {
		start() //服务端初始化
	} else {
		//客户端初始化
	}
	//服务端开始实现
}

func start() {
	http.HandleFunc("/proxy", func(writer http.ResponseWriter, request *http.Request) {
		/*
		offer{
			type:"http"  判断代理协议
			clientPort:"8899" 客户端代理请求端口
		}
		answer{
			status:true|false
			port:"1234"
			type:"http"
			uuid:""
		}
		处理客户端 offer 分配端口
		*/
		//读取body
		defer request.Body.Close()
		body, err := ioutil.ReadAll(request.Body)
		onceOffer := new(offer)
		onceAnswer := new(answer)
		if err != nil {
			startHttpError(err, onceAnswer, writer)
			return
		}
		//解析 body
		err = json.Unmarshal(body, &onceOffer)
		if err != nil {
			startHttpError(err, onceAnswer, writer)
			return
		}
		if onceOffer.ClientProxyType != sConfig.deal {
			startHttpError(err, onceAnswer, writer)
			return
		}
		//生成隧道连接服务 生成uuid  监听服务类型端口并和隧道服务绑定关系
		turnPort, servicePort := GetPort(), GetPort()
		fmt.Println(turnPort, servicePort)
		listenTurnTcp(turnPort) //一个端口代表一条隧道
		onceAnswer.Status = true
		onceAnswer.ServiceProxyType = sConfig.deal
		onceAnswer.ServiceTurnPort = turnPort
		onceAnswer.ServicePort = servicePort
		toClientData, err := json.Marshal(onceAnswer)
		if err != nil {
			startHttpError(err, onceAnswer, writer)
			return
		}
		writer.Write(toClientData)
	})
	err := http.ListenAndServe(":"+strconv.Itoa(sConfig.port), nil)
	if err != nil {
		fmt.Println("ListenAndServe fail in %s", sConfig.port)
	}
}
func startHttpError(err error, onceAnswer *answer, res http.ResponseWriter) {
	onceAnswer.Status = false
	fmt.Println("client data is abnormal %s", err)
	toClientData, err := json.Marshal(onceAnswer)
	res.Write(toClientData)
}
func listenTurnTcp(port int) net.Listener {
	// 启动本地tcp服务
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Errorf("start fail in %s %s", port, err)
	}
	//tid 开始计数
	tid++
	for {
		conn, err := ln.Accept() //一条信道
		if err != nil {
			//对象写入错误信息
			continue
		}
		//生成客户端uuid 保存客户端连接
		currTurn := &Turn{
			GetUuid(),
			tid,
			ln,
			conn,
			make(chan []byte),
			make(chan []byte),
		}
		turnLink[currTurn.Uid] = currTurn
		go handleConnection(conn, currTurn)
	}
	return ln
}
func handleConnection(conn net.Conn, turn *Turn) {
	//处理客户端建立的连接 成功后生成uuid 作为唯一标识  返回一个answer
	// 判断是否初始话 是  建立服务端口监听 否 数据通道
	//需要在这里做数据收发
	go func() { //读取数据
		data := make([]byte, 65534)
		for {
			i, err := conn.Read(data)
			if err != nil {
				continue
			}
			turn.OnlyRead<-data[0:i]//写入tcp数据 在代理客户端处理
		}
	}()
	go func() { //写入数据
		conn.Write(<-turn.OnlyWrite)//写入请求数据 在http中处理
	}()
}
