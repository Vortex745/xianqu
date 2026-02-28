package controllers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

type FileController struct{}

// Upload 处理文件上传
func (fc *FileController) Upload(c *gin.Context) {
	// 1. 防崩溃保护
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

	// 2. ★★★ 动态获取当前运行目录 ★★★
	// 无论你在 C盘、D盘 还是 E盘，它都能找到当前程序所在的位置
	workDir, _ := os.Getwd()
	uploadDir := filepath.Join(workDir, "uploads")

	// 3. 自动创建文件夹
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	// 4. 生成文件名
	ext := filepath.Ext(file.Filename)
	// 使用纳秒时间戳，确保唯一性
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, filename)

	// 5. 判断是否需要压缩 (如果大于 2MB = 2 * 1024 * 1024 bytes)
	const maxSize = 2 * 1024 * 1024
	if file.Size > maxSize {
		// 尝试压缩
		fmt.Printf(">>> 文件大小超过 2MB (%.2f MB)，正在进行压缩处理...\n", float64(file.Size)/1024/1024)
		if err := compressAndSave(file, dst); err != nil {
			// 如果压缩失败，尝试原样保存
			fmt.Println("压缩失败，尝试直接保存:", err)
			if err := c.SaveUploadedFile(file, dst); err != nil {
				fmt.Println("文件保存失败:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
				return
			}
		} else {
			fmt.Println(">>> 压缩成功并已保存")
		}
	} else {
		// 小于 2MB 直接保存
		if err := c.SaveUploadedFile(file, dst); err != nil {
			fmt.Println("文件保存失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
			return
		}
	}

	fmt.Println(">>> 图片已保存至:", dst)

	// 6. ★★★ 动态生成访问 URL ★★★
	// c.Request.Host 会自动获取当前访问的 IP:端口 (如 localhost:8081 或 192.168.1.5:8081)
	// 这样换了电脑、换了 IP 也能正常显示
	protocol := "http://"
	if c.Request.TLS != nil {
		protocol = "https://"
	}
	fullURL := fmt.Sprintf("%s%s/uploads/%s", protocol, c.Request.Host, filename)

	c.JSON(http.StatusOK, gin.H{
		"url": fullURL,
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
