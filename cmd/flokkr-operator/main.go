package main

import (
	"runtime"
	"github.com/sirupsen/logrus"
	"github.com/flokkr/flokkr-operator/pkg/operator"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

}


func main() {
	printVersion()

	flokkroperator.Run()
}
