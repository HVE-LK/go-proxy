package main

import (
	"net"
	"net/http"
	"strconv"
	"net/http/httputil"
	"log"
)

//监听服务
type ProxyService struct {
	link net.Conn
}

func (this *ProxyService)CreateClientService(port int,deal string,Tid int) error{
	onceTrun:=TurnLink[Tid]
	if deal=="http"{
		http.ListenAndServe(":"+strconv.Itoa(port),nil)
		http.HandleFunc("*", func(writer http.ResponseWriter, request *http.Request) {
			//转发数据到隧道 阻塞 直到数据
			//响应客户端
			dataFormat:=new(DataPipe)
			oncedata,err:=httputil.DumpRequest(request,false)
			if err!=nil{
				http.Error(writer, "Header parsing fails", http.StatusInternalServerError)
				return
			}
			dataFormat.Headers=oncedata
			if request.Body!=nil{
				//发送头部 数据 body数据
				buffer:=make([]byte,65535)
				for{
					n,err:=request.Body.Read(buffer)
					if err==nil{
						dataFormat.Data=buffer[:n]
						onceTrun.OnlyWrite<-dataFormat
					}else {
						close(onceTrun.OnlyWrite)//数据写入完成
					}
				}
			}else {
				onceTrun.OnlyWrite<-dataFormat
			}
			resData:=<-onceTrun.OnlyRead
			// 解析resData 响应数据 由隧道端关闭
			writer.Write(resData.Data)
			log.Println("数据发送完成...")
		})
	}else {

	}
	return nil
}
