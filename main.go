package main

import (
	"github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		//NoColors:        true,
		HideKeys:    true,
		CallerFirst: true,
	})
}
func main() {
	err := new(transcode).run()
	if err != nil {
		logrus.Fatal(err)
	}
}
