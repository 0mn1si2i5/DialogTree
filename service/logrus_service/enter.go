// Path: ./service/logrus_service/enter.go

package logrus_service

import "github.com/sirupsen/logrus"

func myDebug(format string, args ...interface{}) {
	logrus.Debugf()
}
