package hotdataserve

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/samber/lo"
)

var articlesType = []datastruct.Option[string, int]{
	{Name: "Share", Label: "publish.types.share", Value: int(articles.Share)},
	{Name: "Help", Label: "publish.types.help", Value: int(articles.Help)},
}

var articlesTypeMap = lo.KeyBy(articlesType, func(v datastruct.Option[string, int]) int {
	return v.Value
})

func GetArticlesType() *[]datastruct.Option[string, int] {
	return &articlesType
}
