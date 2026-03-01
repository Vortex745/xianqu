package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

type FileController struct{}

// Upload 处理文件上传 (修改为返回 Base64 以支持 Vercel Serverless)
func (fc *FileController) Upload(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("上传服务发生异常:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务内部错误"})
		}
	}()

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}

	// 读取上传的文件到内存中进行处理
	srcFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传的文件"})
		return
	}
	defer srcFile.Close()

	// 解码图片
	img, format, err := image.Decode(srcFile)
	if err != nil {
		// 如果解码失败，直接将原始文件转为 base64
		srcFile.Seek(0, 0)
		fileBytes, _ := io.ReadAll(srcFile)
		mimeType := http.DetectContentType(fileBytes)
		encoded := base64.StdEncoding.EncodeToString(fileBytes)
		c.JSON(http.StatusOK, gin.H{"url": fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)})
		return
	}

	// 计算新尺寸 (最大边长 1920)
	const maxDim = 1920
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newImg image.Image = img
	if width > maxDim || height > maxDim {
		newImg = resize.Thumbnail(maxDim, maxDim, img, resize.Lanczos3)
	}

	// 将压缩后的图片写入缓冲区
	var buf bytes.Buffer
	format = strings.ToLower(format)
	mimeType := "image/jpeg"

	switch format {
	case "png":
		mimeType = "image/png"
		err = png.Encode(&buf, newImg)
	default:
		err = jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图片处理失败"})
		return
	}

	// 转换为 Base64
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	c.JSON(http.StatusOK, gin.H{
		"url": dataURI,
	})
}

// 辅助函数：压缩并保存
func compressAndSave(fileHeader *multipart.FileHeader, dst string) error {
	// 1. 打开源文件
	srcFiles, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer srcFiles.Close()

	// 2. 解码图片
	img, format, err := image.Decode(srcFiles)
	if err != nil {
		return fmt.Errorf("decode error: %v", err) // 可能是非图片文件
	}

	// 3. 计算新尺寸 (如果过大则等比缩放)
	// 设定最大边长 1920 (兼顾清晰度和大小)
	const maxDim = 1920
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	var newImg image.Image = img
	if width > maxDim || height > maxDim {
		// resize.Thumbnail 会在保持比例的前提下，将图片缩放到 bounding box 内
		newImg = resize.Thumbnail(maxDim, maxDim, img, resize.Lanczos3)
	}

	// 4. 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 5. 根据格式编码保存
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// JPEG 质量设为 80 (通常能大幅减小体积且肉眼难辨与原图区别)
		return jpeg.Encode(out, newImg, &jpeg.Options{Quality: 80})
	case "png":
		// PNG 压缩 (默认级别)
		return png.Encode(out, newImg)
	default:
		// 其他格式转为 JPEG
		return jpeg.Encode(out, newImg, &jpeg.Options{Quality: 80})
	}
}
