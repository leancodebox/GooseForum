package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/samber/lo"
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "分享", Value: int(articles.Share)},
	{Name: "求助", Value: int(articles.Help)},
}

var articlesTypeMap = lo.KeyBy(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetArticlesType() *[]datastruct.Option[string, int] {
	return &articlesType
}
