package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

const maxImageLength = 1200

func ginFileServer(ginApp *gin.Engine) {
	r := ginApp.Group("file")
	r.POST("/img-upload", func(c *gin.Context) {
		// 判断上传令牌是否合法
		token := c.PostForm("token")
		if token != "your_token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 判断文件大小是否合法
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
			return
		}
		if file.Size > 5*1024*1024 { // 限制文件大小为5MB
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "File size too large"})
			return
		}

		// 打开上传的文件
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "File open failed"})
			return
		}
		defer src.Close()

		fileExt := path.Ext(file.Filename)
		fileNameInt := time.Now().Unix()
		fileNameStr := cast.ToString(fileNameInt)
		fileName := fileNameStr + fileExt
		folderName := time.Now().Format("2006/01/02")
		folderPath := filepath.Join("./storage/uploads", folderName)
		if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "File create failed"})
			return
		}

		// 创建目标文件
		filePath := filepath.Join(folderPath, fileName)
		dst, err := os.Create(filePath)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "File create failed"})
			return
		}
		defer dst.Close()

		// 缩放图片
		err = scaleImage(src, dst)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Image scaling failed"})
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully", fileName))
	})
}

func scaleImage(src io.Reader, dst io.Writer) error {
	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > maxImageLength || height > maxImageLength {
		ratio := float64(maxImageLength) / math.Max(float64(width), float64(height))

		newWidth := int(float64(width) * ratio)
		newHeight := int(float64(height) * ratio)

		resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, img.Bounds(), draw.Over, nil)

		err := jpeg.Encode(dst, resizedImg, nil)
		if err != nil {
			return err
		}
	} else {
		_, err := io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	return nil
}
