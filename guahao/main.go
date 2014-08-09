package main

import (
	"flag"
	"os"

	glog "github.com/golang/glog"

	zjol "v-io.co/guahao/zjol"
)

var (
	storageDPath = "storage"
	configDPath  = map[string]string{"zjol": "storage/zjol"}
)

func init() {
	// make runtime directories
	err := os.MkdirAll(storageDPath, os.ModeDir|os.ModePerm)
	if err != nil {
		glog.Fatalf("unable to create storage directory: %s\n", err)
	}
	for _, path := range configDPath {
		os.MkdirAll(path, os.ModeDir|os.ModePerm)
		if err != nil {
			glog.Fatalf("unable to create session directory: %s\n", err)
		}
	}
}

func parseArgs() {
	flag.Parse()
}

func loadZjol() {
	// // check session-id file
	// sessionFile, err := os.Open(sessionFPath)
	// if err != nil {
	// 	sessionFile, err = os.Create(sessionFPath)
	// 	if err != nil {
	// 		glog.Fatalln(err)
	// 	}
	// 	defer sessionFile.Close()
	// 	return realLogin()
	// }
	// defer sessionFile.Close()

	// reader := bufio.NewReader(sessionFile)
	// session.SessionId, err = reader.ReadString('\n')
	// if err == nil {
	// 	session.SessionId = session.SessionId[:len(session.SessionId)-1]
	// } else if err != nil {
	// 	glog.Warningf("only session id stored in cache file, redirect to realLogin()\n")
	// 	return realLogin()
	// }
	// session.UserId, err = reader.ReadString('\n')
	// if err == nil {
	// 	session.UserId = session.UserId[:len(session.UserId)-1]
	// } else if err != nil && err != io.EOF {
	// 	glog.Infof("error occured, redirect to realLogin()\n")
	// 	return realLogin()
	// }

	// client := &http.Client{Jar: jar}
	// sidCookie := &http.Cookie{
	// 	Domain: "guahao.zjol.com.cn", Path: "/",
	// 	Name: "ASP.NET_SessionId", Value: session.SessionId,
	// 	HttpOnly: true, MaxAge: 0}
	// uidCookie := &http.Cookie{
	// 	Domain: "guahao.zjol.com.cn", Path: "/",
	// 	Name: "UserId", Value: session.UserId,
	// 	HttpOnly: true, MaxAge: 0}
	// u, err := url.Parse(domUrl)
	// client.Jar.SetCookies(u, []*http.Cookie{sidCookie, uidCookie})

}

func main() {
	parseArgs()
	// startServer()

	// book zjol
	zjolUser := "330602198503130017" // login username
	zjolPass := "Td135128887"        // login password
	var session *zjol.Session
	var err error
	for {
		session, err = zjol.RealLogin(zjolUser, zjolPass)
		if err != nil {
			glog.Error(err)
		} else {
			break
		}
	}
	glog.V(1).Infof("sid -> '%s'\n", session.SessionId)
	glog.V(1).Infof("uid -> '%s'\n", session.UserId)
	session.Dept = "9291"
	err = session.Book(9)
	if err != nil {
		panic(err)
	}

	glog.Infof("done\n")
}
