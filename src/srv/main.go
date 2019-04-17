package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http/fcgi"
	"os"
	"regexp"

	"github.com/emicklei/go-restful"
)

type AppInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func getFilesDir() string {
	dir := os.Getenv("POER_CDN_FILE_PATH")
	if dir == "" {
		dir = "./files"
	}
	return dir
}

func getAppInfo(req *restful.Request, rsp *restful.Response) {
	market := req.QueryParameter("market")
	var regStr string
	if market != "" {
		regStr = `^PoerSmart\-` + market + `\-.*\.apk`
	} else {
		regStr = `^PoerSmart\-default\-.*\.apk`
	}

	info := AppInfo{}
	info.Name = ""
	info.Status = "fail"

	reg, err := regexp.Compile(regStr)
	if err != nil {
		log.Println(err)
		rsp.WriteEntity(info)
		return
	}

	fileInfos, err := ioutil.ReadDir(getFilesDir())
	if err != nil {
		log.Println(err)
		rsp.WriteEntity(info)
		return
	}

	for _, f := range fileInfos {
		//log.Println(f.Name())
		if reg.MatchString(f.Name()) {
			info.Name = f.Name()
			info.Status = "success"
			rsp.WriteEntity(info)
			return
		} else {
			//log.Println(f.Name(), regStr)
		}
	}
	log.Println("file not found")
	rsp.WriteEntity(info)
}

func main() {

	ws := new(restful.WebService)
	ws.Path("/cgi").
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON)
	ws.Route(ws.GET("/appinfo").To(getAppInfo).Writes(AppInfo{}))
	restful.Add(ws)

	listener, err := net.Listen("tcp", "127.0.0.1:9142")
	if err != nil {
		panic(err)
	}
	//fcgi.Serve(listener, srv)
	fcgi.Serve(listener, restful.DefaultContainer)
}
