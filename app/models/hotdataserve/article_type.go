package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/samber/lo"
)

const (
	topicTypeShare = 1
	topicTypeHelp  = 2
)

var topicTypes = []datastruct.Option[string, int]{
	{Name: "share", Value: topicTypeShare},
	{Name: "help", Value: topicTypeHelp},
}

var topicTypeMap = lo.KeyBy(topicTypes, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetTopicTypes() *[]datastruct.Option[string, int] {
	return &topicTypes
}
