package main

import (
	"flag"

	"github.com/golang/glog"
)

func flagInit() {
	flag.Set("stderrthreshold", "info")
	flag.Set("v", "0")
	flag.Parse()
}
func main() {
	flagInit()

	glog.Info("datanode")
}
