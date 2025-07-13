package optlogger

import "github.com/spf13/cast"

type OptEnum int

func (receiver OptEnum) TargetTypeEnum() TargetTypeEnum {
	switch receiver {
	case EditUser:
		return User
	case EditArticle:
		return User
	case CreateDocProject, UpdateDocProject, DeleteDocProject:
		return DocProject
	case CreateDocVersion, UpdateDocVersion, DeleteDocVersion:
		return DocVersion
	case CreateDocContent, UpdateDocContent, DeleteDocContent, PublishDocContent, DraftDocContent:
		return DocContent
	}
	return System
}

func (receiver OptEnum) Name() string {
	switch receiver {
	case EditUser:
		return "操作用户"
	case EditArticle:
		return "编辑文章"
	case CreateDocProject:
		return "创建文档项目"
	case UpdateDocProject:
		return "更新文档项目"
	case DeleteDocProject:
		return "删除文档项目"
	case CreateDocVersion:
		return "创建文档版本"
	case UpdateDocVersion:
		return "更新文档版本"
	case DeleteDocVersion:
		return "删除文档版本"
	case CreateDocContent:
		return "创建文档内容"
	case UpdateDocContent:
		return "更新文档内容"
	case DeleteDocContent:
		return "删除文档内容"
	case PublishDocContent:
		return "发布文档内容"
	case DraftDocContent:
		return "设为草稿"
	}
	return ""
}

func (receiver OptEnum) toInt() int {
	return cast.ToInt(receiver)
}

const (
	EditUser OptEnum = iota
	EditArticle
	CreateDocProject
	UpdateDocProject
	DeleteDocProject
	CreateDocVersion
	UpdateDocVersion
	DeleteDocVersion
	CreateDocContent
	UpdateDocContent
	DeleteDocContent
	PublishDocContent
	DraftDocContent
)

type TargetTypeEnum int

func (receiver TargetTypeEnum) Name() string {
	switch receiver {
	case System:
		return "系统"
	case User:
		return "用户"
	case DocProject:
		return "文档项目"
	case DocVersion:
		return "文档版本"
	case DocContent:
		return "文档内容"
	}
	return ""
}

func (receiver TargetTypeEnum) toInt() int {
	return cast.ToInt(receiver)
}

const (
	System     TargetTypeEnum = iota
	User                      = iota
	Article                   = iota
	DocProject                = iota
	DocVersion                = iota
	DocContent                = iota
)
