package main

import (
	"fmt"

	"github.com/chroblert/JC-GOUtils/log"
)

func main() {
	fmt.Println("this is a test")
	log.InitLogs("logs/app.log", 20000, 2, 3)
	log.Info("xxx")
}
