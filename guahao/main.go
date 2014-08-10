package main

import (
	"flag"
	"os"

	glog "github.com/golang/glog"

	zjol "github.com/v-gu/guahao/zjol"
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
