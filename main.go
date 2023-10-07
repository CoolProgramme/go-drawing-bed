package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"github.com/joho/godotenv"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

// MaxFileSize 允许上传的最大文件大小
const MaxFileSize = 10 << 20 // 10 MB

// AllowOrigins 允许域
var AllowOrigins []string

// Port 端口
var Port string

// Url 返回的图片Url前缀
var Url string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8080"
	}
	AllowOrigins = strings.Split(os.Getenv("AllowOrigins"), ",")
	Url = os.Getenv("URL")
	if Url == "" {
		Url = "http://127.0.0.1:" + Port
	}
}

func main() {
	router := gin.Default()

	// CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     AllowOrigins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/static", "./static")

	// 为 multipart forms 设置较低的内存限制 (默认是 32 MiB)
	router.MaxMultipartMemory = MaxFileSize // 10 MiB

	// 上传接口，仅允许上传图片
	router.POST("/upload", uploadHandler)

	err := router.Run(":" + Port)
	if err != nil {
		panic(err)
	}
}

func uploadHandler(context *gin.Context) {

	upload, err := context.FormFile("file")

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if upload.Size > MaxFileSize {
		context.JSON(http.StatusBadRequest, gin.H{"error": "请将图片大小压缩至不超过10MB！"})
		return
	}

	file, err := upload.Open()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	head := make([]byte, 261)
	_, err = file.Read(head)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !filetype.IsImage(head) {
		context.JSON(http.StatusBadRequest, gin.H{"error": "仅允许上传图片类型！"})
		return
	}

	fileName := upload.Filename

	now := time.Now()

	dst := fmt.Sprintf("./static/%d/%d/%d/%s", now.Year(), int(now.Month()), now.Day(), fileName)

	err = context.SaveUploadedFile(upload, dst)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK,
		gin.H{
			"message": "图片上传成功！",
			"data": gin.H{
				"name": fileName,
				"url":  Url + dst[1:],
			},
		},
	)
}
