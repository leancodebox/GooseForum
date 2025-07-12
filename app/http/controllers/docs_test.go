package controllers

import (
	"fmt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"testing"
)

func TestJsonDecodeItem(t *testing.T) {
	data := `[
    {
        "slug": "getting-started",
        "title": "快速开始",
        "description": "快速安装和配置 GooseForum",
        "children": [
            {
                "slug": "installation",
                "title": "安装指南"
            },
            {
                "slug": "configuration",
                "title": "基础配置"
            }
        ]
    },
    {
        "slug": "user-guide",
        "title": "用户指南",
        "description": "面向普通用户的使用说明",
        "children": [
            {
                "slug": "posting",
                "title": "发帖指南"
            },
            {
                "slug": "profile",
                "title": "个人资料"
            }
        ]
    },
    {
        "slug": "admin-guide",
        "title": "管理员指南",
        "description": "论坛管理和维护指南"
    }
]`
	res := jsonopt.Decode[[]DirectoryItem](data)
	fmt.Println(res)
}
