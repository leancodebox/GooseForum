package cmd

import (
	"bufio"
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/goose/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/goose/preferences"
	timeopt "github.com/leancodebox/GooseForum/app/bundles/timopt"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/service/pointservice"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "hexo:tool",
		Short: "hexo迁移工具",
		Run:   runHexoTool,
	})
}

type Blog struct {
	TitleConfig TitleConfig
	Title       string
	Content     string
}

func runHexoTool(_ *cobra.Command, _ []string) {

	var basePath = preferences.Get("path.hexo", "")

	userId := 1
	if len(basePath) == 0 {
		fmt.Println("请填写有效目标路径")
		return
	}
	blogs, err := traverse(basePath)
	if err != nil {
		fmt.Println("Error traversing directory:", err)
		return
	}
	slices.SortFunc(blogs, func(a, b Blog) int {
		aDate := timeopt.Str2Time(a.TitleConfig.Date)
		bDate := timeopt.Str2Time(b.TitleConfig.Date)
		if aDate.After(bDate) {
			return 1
		}
		return -1
	})
	var tags []string
	var categories []string
	for _, data := range blogs {
		tags = append(tags, data.TitleConfig.Tags...)
		categories = append(categories, data.TitleConfig.Categories...)
		fmt.Println(data.Title)
		fmt.Println(data.TitleConfig)
		old := articles.GetByUserAndTitle(userId, data.Title)
		if old.Id > 0 {
			fmt.Println(userId, data.Title, "已经存在")
			continue
		}
		writeDate := timeopt.Str2Time(data.TitleConfig.Date)
		art := articles.Entity{
			UserId:        1,
			Content:       data.Content,
			ArticleStatus: 1,
			Title:         data.Title,
			UpdatedAt:     writeDate,
			CreatedAt:     writeDate,
		}
		if false {
			articles.Save(&art)
			pointservice.RewardPoints(uint64(userId), 10, pointservice.RewardPoints4WriteArticles)
		}
	}
	fmt.Println(len(blogs))
	fmt.Println(array.RemoveDuplicates(tags))
	fmt.Println(array.RemoveDuplicates(categories))
}

func traverse(path string) ([]Blog, error) {
	blogs := make([]Blog, 0)

	files, err := os.ReadDir(path)
	if err != nil {
		return blogs, err
	}

	for _, file := range files {
		if file.IsDir() {
			subPath := filepath.Join(path, file.Name())
			subBlogs, err := traverse(subPath)
			if err != nil {
				return blogs, err
			}
			blogs = append(blogs, subBlogs...)
		} else {
			if strings.HasSuffix(file.Name(), ".md") {
				filePath := filepath.Join(path, file.Name())
				data, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Println("Error reading file:", err)
					continue
				}

				title := ""
				content := ""

				// Parse front-matter and content
				scanner := bufio.NewScanner(strings.NewReader(string(data)))
				isFrontMatter := true
				headerB := strings.Builder{}
				var tc TitleConfig
				for scanner.Scan() {
					line := scanner.Text()
					fmt.Println(line)
					if line == "---" {
						if isFrontMatter {
							isFrontMatter = false
						} else {
							break
						}
						continue
					}
					headerB.WriteString(line + "\n")
					if strings.HasPrefix(line, "title:") {
						title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
						title = strings.Trim(title, `'`)
					}
				}
				yaml.Unmarshal([]byte(headerB.String()), &tc)
				for scanner.Scan() {
					content += scanner.Text() + "\n"
				}
				blogs = append(blogs, Blog{
					TitleConfig: tc,
					Title:       title,
					Content:     content,
				})
			}
		}
	}

	return blogs, nil
}

type TitleConfig struct {
	Title      string   `json:"title"`
	Toc        bool     `json:"toc"`
	Date       string   `json:"date"`
	Tags       []string `json:"tags"`
	Categories []string `json:"categories"`
}
