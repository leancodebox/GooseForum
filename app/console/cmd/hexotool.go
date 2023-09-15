package cmd

import (
	"bufio"
	"fmt"
	Articles2 "github.com/leancodebox/GooseForum/app/models/bbs/Articles"
	"github.com/leancodebox/goose/preferences"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	appendCommand(&cobra.Command{
		Use:   "hexo:tool",
		Short: "hexo tool",
		Run:   runHexoTool,
	})
}

type Blog struct {
	Title   string
	Content string
}

func runHexoTool(_ *cobra.Command, _ []string) {

	var basePath = preferences.Get("path.hexo", "")

	if len(basePath) == 0 {
		fmt.Println("请填写有效目标路径")
		return
	}
	blogs, err := traverse(basePath)
	if err != nil {
		fmt.Println("Error traversing directory:", err)
		return
	}
	for _, data := range blogs {
		art := Articles2.Entity{UserId: 1, Content: data.Content, Title: data.Title}
		Articles2.Save(&art)
	}
	fmt.Println(len(blogs))
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
				for scanner.Scan() {
					line := scanner.Text()
					if line == "---" {
						if isFrontMatter {
							isFrontMatter = false
						} else {
							break
						}
					}
					if strings.HasPrefix(line, "title:") {
						title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
					}
				}
				for scanner.Scan() {
					content += scanner.Text() + "\n"
				}

				blogs = append(blogs, Blog{
					Title:   title,
					Content: content,
				})
			}
		}
	}

	return blogs, nil
}
