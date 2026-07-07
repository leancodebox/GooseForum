package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/samber/lo"
)

const (
	topicTypeShare = 1
	topicTypeHelp  = 2
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "share", Value: topicTypeShare},
	{Name: "help", Value: topicTypeHelp},
}

var articlesTypeMap = lo.KeyBy(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetArticlesType() *[]datastruct.Option[string, int] {
	return &articlesType
}
