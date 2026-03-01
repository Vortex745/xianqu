package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

// Upload 处理文件上传 (支持直传 Superbed 图床)
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

	// 读取上传的文件
	srcFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取上传的文件"})
		return
	}
	defer srcFile.Close()

	// 尝试解码图片以进行可能的压缩/处理
	img, format, err := image.Decode(srcFile)
	var finalData []byte
	var mimeType string

	if err != nil {
		// 非图片或解码失败，获取原始字节
		srcFile.Seek(0, 0)
		finalData, _ = io.ReadAll(srcFile)
		mimeType = http.DetectContentType(finalData)
	} else {
		// 图片处理：计算新尺寸 (最大边长 1920)
		const maxDim = 1920
		bounds := img.Bounds()
		width := uint(bounds.Dx())
		height := uint(bounds.Dy())

		var newImg image.Image = img
		if width > maxDim || height > maxDim {
			newImg = resize.Thumbnail(maxDim, maxDim, img, resize.Lanczos3)
		}

		// 编码处理后的图片
		var buf bytes.Buffer
		format = strings.ToLower(format)
		mimeType = "image/jpeg"
		switch format {
		case "png":
			mimeType = "image/png"
			png.Encode(&buf, newImg)
		default:
			jpeg.Encode(&buf, newImg, &jpeg.Options{Quality: 80})
		}
		finalData = buf.Bytes()
	}

	// 检查是否配置了 ImgBB API Key
	apiKey := os.Getenv("IMGBB_API_KEY")
	if apiKey != "" {
		// 1. 准备向 ImgBB 发送请求
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加文件
		part, err := writer.CreateFormFile("image", file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "构建上传请求失败"})
			return
		}
		part.Write(finalData)
		writer.Close()

		uploadUrl := fmt.Sprintf("https://api.imgbb.com/1/upload?key=%s", apiKey)
		req, err := http.NewRequest("POST", uploadUrl, body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求对象失败"})
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// 2. 执行请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("ImgBB 上传连接失败:", err)
			goto fallback
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		var result struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
			Success bool `json:"success"`
			Status  int  `json:"status"`
		}

		if err := json.Unmarshal(respBody, &result); err == nil && result.Success {
			// 上传成功
			c.JSON(http.StatusOK, gin.H{"url": result.Data.URL})
			return
		} else {
			fmt.Printf("ImgBB 接口返回错误, body: %s\n", string(respBody))
		}
	}

fallback:
	// 备份方案：如果 Token 未配置或 Superbed 上传失败，返回 Base64 Data URI
	encoded := base64.StdEncoding.EncodeToString(finalData)
	c.JSON(http.StatusOK, gin.H{
		"url":  fmt.Sprintf("data:%s;base64,%s", mimeType, encoded),
		"info": "Uploaded as Base64 (Superbed not configured or failed)",
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
