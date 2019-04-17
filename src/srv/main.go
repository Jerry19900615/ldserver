package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/emicklei/go-restful"
)

type AppInfo struct {
	Name   string `json:"name"`
	Status string `json: status`
}

func getFilesDir() string {
	dir := os.Getenv("POER_CDN_FILE_PATH")
	if dir == "" {
		dir = "/home/html/files"
	}
	return dir
}

func main() {
	ws := new(restful.WebService)
	ws.Path("/app")
	ws.Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.GET("/info").To(getAppInfo).Writes(AppInfo{}))
	restful.Add(ws)

	ws = new(restful.WebService)
	ws.Path("/")
	ws.Route(ws.GET("/{subpath:*}").To(serveStaticFiles))
	restful.Add(ws)
	log.Println("start server on http://localhost:9140")
	http.ListenAndServe(":9140", nil)
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
		}
	}
	rsp.WriteEntity(info)
}

func serveStaticFiles(req *restful.Request, rsp *restful.Response) {
	actual := path.Join("./", req.PathParameter("subpath"))
	http.ServeFile(
		rsp.ResponseWriter,
		req.Request,
		actual,
	)
}
