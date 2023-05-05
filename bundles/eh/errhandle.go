package eh

import (
	"fmt"
	"github.com/leancodebox/GooseForum/bundles/logging"
)

func init() {
	bundleInit()
}

func bundleInit() {
	fmt.Println("init eh")
}

func IfErr(err error) bool {
	if err != nil {
		logging.Error(err)
		return true
	}
	return false
}
