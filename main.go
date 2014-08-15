package main

import (
	glog "github.com/golang/glog"

	provider "github.com/v-gu/guahao/provider"
	_ "github.com/v-gu/guahao/provider/driver/zjol"
)

func main() {
	var err error

	// book zjol
	for {
		err := provider.Login()
		if err != nil {
			glog.Errorln(err)
		} else {
			break
		}
	}
	err = provider.Book()
	if err != nil {
		glog.Fatalln(err)
	}
	glog.Infof("done\n")
}
