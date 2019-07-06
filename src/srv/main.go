package main

import (
	"dlserver/src/srv/utils"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"regexp"
	"strings"

	"github.com/emicklei/go-restful"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

//AppInfo ...
type AppInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

//FeedBack 用户发送的POST请求数据
type FeedBack struct {
	Name    string `json:"name"`
	Email   string `json:"post"`
	Tel     string `json:"tel"`
	Address string `json:"address"`
	Guest   string `json:"guest"`
}

//MailInfo ...
type MailInfo struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
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
		}
	}
	log.Println("file not found")
	rsp.WriteEntity(info)
}

func sendMail(req *restful.Request, rsp *restful.Response) {
	mail := new(MailInfo)
	err := req.ReadEntity(mail)
	if err != nil {
		log.Println(err)
		rsp.WriteErrorString(http.StatusInternalServerError, "read post data error")
		return
	}

	log.Println(mail)
	err = utils.SendMail(utils.MAIL_USER, utils.MAIL_PASSWD, utils.MAIL_SERVER, mail.To, mail.Subject, mail.Body, "html")
	if err != nil {
		log.Println(err)
		rsp.WriteError(http.StatusInternalServerError, err)
		return
	}
	return
}

func ResetPasswordToDefault(host, email string) error {
	var dataSource string
	datasourceFormat := "root:poersmart2015@tcp(%s:3306)/poerSmart?charset=utf8"
	dataSource = fmt.Sprintf(datasourceFormat, host)

	email = strings.ToLower(email)

	e, err := xorm.NewEngine("mysql", dataSource)
	if err != nil {
		return err
	}

	defer e.Close()

	err = e.Ping()
	if err != nil {
		return err
	}
	authenticate := utils.BasicAuthenticate(email, "88888888")
	AuthenticateMd5 := utils.BasicAuthenticateMd5(authenticate)
	log.Println("Update User Set AuthenticateMd5='" + AuthenticateMd5 + "' where Email='" + email + "'")
	_, err = e.Table("User").Exec("Update User Set AuthenticateMd5='" + AuthenticateMd5 + "' where Email='" + email + "'")
	return err
}

type ResetStatus struct {
	Msg string `json:"message"`
}

func resetUserPassword(req *restful.Request, rsp *restful.Response) {

	var stat ResetStatus
	stat.Msg = "OK"
	server := req.QueryParameter("server")
	if server == "" {
		stat.Msg = "no server given"
		rsp.WriteEntity(stat)
		return
	}
	email := req.QueryParameter("email")
	if email == "" {
		stat.Msg = "no email given"
		rsp.WriteEntity(stat)
		return
	}
	err := ResetPasswordToDefault(server, email)
	if err != nil {
		stat.Msg = err.Error()
		rsp.WriteEntity(stat)
		return
	}
	rsp.WriteEntity(stat)
	return
}

func sendFeedBackMail(req *restful.Request, rsp *restful.Response) {
	err := req.Request.ParseForm()
	if err != nil {
		log.Println(err)
		rsp.WriteErrorString(http.StatusInternalServerError, "parse form error")
		return
	}

	feedback := new(FeedBack)
	//log.Println(req.Request.FormValue("name"))
	feedback.Name = req.Request.FormValue("name")
	feedback.Email = req.Request.FormValue("post")
	feedback.Tel = req.Request.FormValue("tel")
	feedback.Address = req.Request.FormValue("address")
	feedback.Guest = req.Request.FormValue("guest")

	log.Println(feedback)

	subject := "FeedBack from " + feedback.Name + "(" + feedback.Email + ") (tel:" + feedback.Tel + ")"

	body := `<br/><bold>来自(` + feedback.Address +
		`)的反馈信息:</bold><br/><p>` + feedback.Guest + "</p><br/><br/>"

	err = utils.SendMail(utils.MAIL_USER, utils.MAIL_PASSWD, utils.MAIL_SERVER, utils.MAIL_RECIVER, subject, body, "html")
	log.Println(err)
}

func main() {

	ws := new(restful.WebService)
	ws.Path("/cgi").
		Produces(restful.MIME_JSON).
		Consumes(restful.MIME_JSON)
	ws.Route(ws.GET("/appinfo").To(getAppInfo).Writes(AppInfo{}))
	ws.Route(ws.POST("/poer/mail").Consumes("application/x-www-form-urlencoded").To(sendFeedBackMail))
	ws.Route(ws.POST("/sendmail").To(sendMail).Reads(&MailInfo{}))
	ws.Route(ws.GET("/resetpwd").To(resetUserPassword).Writes(ResetStatus{}))
	restful.Add(ws)

	listener, err := net.Listen("tcp", "127.0.0.1:9142")
	if err != nil {
		panic(err)
	}
	//fcgi.Serve(listener, srv)
	fcgi.Serve(listener, restful.DefaultContainer)
}
