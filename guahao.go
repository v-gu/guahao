package main

import (
	glog "github.com/golang/glog"

	provider "github.com/v-gu/guahao/provider"
	_ "github.com/v-gu/guahao/provider/driver/zjol"
)

func main() {
	var err error

	// // catch panic, continue then
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		glog.Warningf("$v, retrying... ", err)
	// 	}
	// }()

	for {
		err = provider.Book()
		if err != nil {
			glog.Fatalln(err)
		}
	}
	glog.Infof("done\n")
}
