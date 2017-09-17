// reverse-proxy project main.go 客户端
package main

import (
	"flag"
	"fmt"
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
	foreign int `对外服务端口`
}

// 客户端配置
type CConfig struct {
	agencyService string `被代理服务`
	agentAddr     string `代理服务地址`
	localSePport  uint16 `客户端代理服务端口`
}

type Offer struct {
	ClientProxyType string
	ClientPort      int
}

type answer struct {
	Status           bool
	ServiceTurnPort  int
	ServiceProxyType string
	ServicePort      int
}

// 服务端记录被代理的对象
var (
	serviceType uint
	sConfig     = new(SConfig)
)

//隧道计数器

func init() {
	flag.UintVar(&serviceType, "servicetype", 0, "Service type, 0: proxy service host, 1: proxy client")
	flag.StringVar(&sConfig.deal, "stype", "http", "Proxy service protocols, such as HTTP, HTTPS, TCP, UDP, SCOKET")
	flag.IntVar(&sConfig.port, "sport", 8080, "Proxy server startup port, default 8080")
	flag.IntVar(&sConfig.foreign, "svpt", 8081, "Agent service listening port, default 8081")
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
		onceOffer := new(Offer)
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

		//如果判断客户端创建成功 则直接返回已创建隧道
		newTurn:=new(turnFac)
		err=newTurn.CreateTurn(turnPort,servicePort) //一个端口代表一条隧道
		if err!=nil{
			startHttpError(err, onceAnswer, writer)
			return
		}

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
	http.HandleFunc("/close", func(writer http.ResponseWriter, request *http.Request) {

	})
	err := http.ListenAndServe(":"+strconv.Itoa(sConfig.port), nil)//默认创建go
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
