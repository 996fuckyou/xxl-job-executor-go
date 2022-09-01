package task

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLog(t *testing.T) {
	logrus.New()
	logrus.Info("hello world")
}
