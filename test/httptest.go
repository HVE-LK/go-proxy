package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"net/http/httputil"
	"log"
)

const (
	upload_path string = "./upload/"
)

func load_success(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "上传成功!")

}

//上传
func uploadHandle(w http.ResponseWriter, r *http.Request) {
	//从请求当中判断方法

	if r.Method == "GET" {
		io.WriteString(w, "<html><head><title>上传</title></head>"+
			"<body><form action='#' method=\"post\" enctype=\"multipart/form-data\">"+
			"<label>上传图片</label>"+":"+
			"<input type=\"file\" name='file'  /><br/><br/>    "+
			"<label><input type=\"submit\" value=\"上传图片\"/></label></form></body></html>")
	} else {
		fmt.Println("开始上传")
		//获取文件内容 要这样获取
		fmt.Println("开始获取数据")
		reqdata,err:=httputil.DumpRequest(r,false)
		buff:=make([]byte,65535)
		for{
			n,err:=r.Body.Read(buff)
			if err != nil {
				break
			}
			log.Println(n)
		}
		if err!=nil{
			log.Println("reqdata",string(reqdata))
		}else{
			log.Println("DumpRequestOut:",string(reqdata))
		}



		file, head, err := r.FormFile("file")
		fmt.Println("开始获取数据")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Println("穿件文件")
		fW, err := os.Create("./" + head.Filename)
		if err != nil {
			fmt.Println("文件创建失败")
			return
		}
		defer fW.Close()
		_, err = io.Copy(fW, file)
		fmt.Println("拷贝数据")
		if err != nil {
			fmt.Println("文件保存失败")
			return
		}
		//io.WriteString(w, head.Filename+" 保存成功")
		http.Redirect(w, r, "/success", http.StatusFound)
		//io.WriteString(w, head.Filename)
	}
}

func main() {
	fmt.Println("OK!")
	//启动一个http 服务器
	http.HandleFunc("/success", load_success)
	//上传
	http.HandleFunc("/upload", uploadHandle)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("服务器启动失败")
		return
	}
	fmt.Println("服务器启动成功")
}
