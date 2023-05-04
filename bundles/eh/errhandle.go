package eh

import (
	"fmt"
	"github.com/leancodebox/GooseForum/bundles/logger"
)

func init() {
	bundleInit()
}

func bundleInit() {
	fmt.Println("init eh")
}

func IfErr(err error) bool {
	if err != nil {
		logger.Error(err)
		return true
	}
	return false
}
