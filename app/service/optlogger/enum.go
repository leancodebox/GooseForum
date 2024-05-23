package optlogger

import "github.com/spf13/cast"

type OptEnum int

func (receiver OptEnum) TargetTypeEnum() TargetTypeEnum {
	switch receiver {
	case EditUser:
		return User
	}
	return System
}

func (receiver OptEnum) Name() string {
	switch receiver {
	case EditUser:
		return "操作用户"
	}
	return ""
}

func (receiver OptEnum) toInt() int {
	return cast.ToInt(receiver)
}

const (
	EditUser OptEnum = iota
)

type TargetTypeEnum int

func (receiver TargetTypeEnum) Name() string {
	switch receiver {
	case System:
		return "系统"
	case User:
		return "用户"
	}
	return ""
}

func (receiver TargetTypeEnum) toInt() int {
	return cast.ToInt(receiver)
}

const (
	System TargetTypeEnum = iota
	User                  = iota
)
