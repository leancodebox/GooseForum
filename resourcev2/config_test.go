package resourcev2

import (
	"fmt"
	"testing"
)

func TestGetMetaList(t *testing.T) {
	fmt.Println(GetRealFilePath("src/main.js"))
	fmt.Println(GetImportInfoPath("src/main.js"))
}
