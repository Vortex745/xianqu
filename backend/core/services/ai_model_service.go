package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"gotest/config"
	"gotest/core/models"

	"gorm.io/gorm"
)

// ────────────── 加密工具 ──────────────
// 使用 AES-GCM 加密 API Key，密钥可从环境变量读取，此处写死一个 32 字节密钥
var aesKey = []byte("xianqu_ai_model_enc_key_2024!!!!") // 32 字节 = AES-256

func encryptAPIKey(plainText string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func decryptAPIKey(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}

// maskAPIKey 返回尾部4位的掩码展示
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

// ────────────── CRUD 操作 ──────────────

// CreateAIModel 新建模型配置
func CreateAIModel(m *models.AIModel, rawAPIKey string) error {
	encrypted, err := encryptAPIKey(rawAPIKey)
	if err != nil {
		return fmt.Errorf("加密API Key失败: %w", err)
	}
	m.APIKey = encrypted
	return config.DB.Create(m).Error
}

// UpdateAIModel 更新模型配置
func UpdateAIModel(id uint, updates map[string]interface{}, newAPIKey string) error {
	if newAPIKey != "" {
		encrypted, err := encryptAPIKey(newAPIKey)
		if err != nil {
			return fmt.Errorf("加密API Key失败: %w", err)
		}
		updates["api_key"] = encrypted
	}
	return config.DB.Model(&models.AIModel{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAIModel 删除模型配置
func DeleteAIModel(id uint) error {
	return config.DB.Delete(&models.AIModel{}, id).Error
}

// GetAIModelByID 获取单个模型（管理端，含掩码Key）
func GetAIModelByID(id uint) (*models.AIModel, error) {
	var m models.AIModel
	if err := config.DB.First(&m, id).Error; err != nil {
		return nil, err
	}
	// 解密后仅展示尾4位
	if plain, err := decryptAPIKey(m.APIKey); err == nil {
		m.APIKeyHint = maskAPIKey(plain)
	}
	m.APIKey = "" // 不返回密文
	return &m, nil
}

// ListAIModels 管理端列表（全部，含掩码Key）
func ListAIModels() ([]models.AIModel, error) {
	var list []models.AIModel
	if err := config.DB.Order("priority desc, id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		if plain, err := decryptAPIKey(list[i].APIKey); err == nil {
			list[i].APIKeyHint = maskAPIKey(plain)
		}
		list[i].APIKey = ""
	}
	return list, nil
}

// ListActiveAIModels 前台接口：仅返回启用的模型（不含敏感信息）
func ListActiveAIModels() ([]map[string]interface{}, error) {
	var list []models.AIModel
	if err := config.DB.Where("status = 1").Order("priority desc, id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	result := make([]map[string]interface{}, 0, len(list))
	for _, m := range list {
		result = append(result, map[string]interface{}{
			"id":          m.ID,
			"provider":    m.Provider,
			"model_name":  m.ModelName,
			"base_url":    m.BaseURL,
			"price_per_k": m.PricePerK,
			"priority":    m.Priority,
		})
	}
	return result, nil
}

// GetDecryptedModel 内部使用：获取解密后的完整模型信息（供AI Service调用）
func GetDecryptedModel(id uint) (*models.AIModel, error) {
	var m models.AIModel
	if err := config.DB.Where("id = ? AND status = 1", id).First(&m).Error; err != nil {
		return nil, err
	}
	plain, err := decryptAPIKey(m.APIKey)
	if err != nil {
		return nil, fmt.Errorf("解密API Key失败: %w", err)
	}
	m.APIKey = plain
	return &m, nil
}

// ────────────── 用量上报 ──────────────

// ReportUsage 上报一次AI调用用量
func ReportUsage(log *models.AIUsageLog) error {
	// 自动计算成本
	var m models.AIModel
	if err := config.DB.First(&m, log.ModelID).Error; err != nil {
		return fmt.Errorf("模型不存在: %w", err)
	}
	log.Cost = float64(log.TotalTokens) / 1000.0 * m.PricePerK
	log.Cost = math.Round(log.Cost*1000000) / 1000000 // 保留6位小数
	log.Provider = m.Provider
	log.ModelName = m.ModelName

	if err := config.DB.Create(log).Error; err != nil {
		return err
	}

	// 异步更新每日汇总表
	go updateDailyStat(log)
	return nil
}

// updateDailyStat 更新当日汇总
func updateDailyStat(log *models.AIUsageLog) {
	today := time.Now().Truncate(24 * time.Hour)
	var stat models.AIUsageDailyStat

	result := config.DB.Where(
		"date = ? AND model_id = ? AND app_type = ?",
		today, log.ModelID, log.AppType,
	).First(&stat)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		stat = models.AIUsageDailyStat{
			Date:        today,
			ModelID:     log.ModelID,
			AppType:     log.AppType,
			TotalTokens: int64(log.TotalTokens),
			TotalCost:   log.Cost,
			CallCount:   1,
			Provider:    log.Provider,
			ModelName:   log.ModelName,
		}
		config.DB.Create(&stat)
		return
	}

	config.DB.Model(&stat).Updates(map[string]interface{}{
		"total_tokens": gorm.Expr("total_tokens + ?", log.TotalTokens),
		"total_cost":   gorm.Expr("total_cost + ?", log.Cost),
		"call_count":   gorm.Expr("call_count + 1"),
	})
}

// ────────────── 数据看板 ──────────────

// UsageDashboardQuery 看板查询参数
type UsageDashboardQuery struct {
	Days     int    // 7 或 30
	Provider string // 可选筛选
	AppType  string // 可选筛选
}

// DashboardDataItem 看板数据项
type DashboardDataItem struct {
	Date        string  `json:"date"`
	ModelID     uint    `json:"model_id"`
	Provider    string  `json:"provider"`
	ModelName   string  `json:"model_name"`
	AppType     string  `json:"app_type"`
	TotalTokens int64   `json:"total_tokens"`
	TotalCost   float64 `json:"total_cost"`
	CallCount   int64   `json:"call_count"`
}

// GetUsageDashboard 获取看板数据
func GetUsageDashboard(q UsageDashboardQuery) ([]DashboardDataItem, error) {
	if q.Days <= 0 {
		q.Days = 7
	}
	since := time.Now().AddDate(0, 0, -q.Days).Truncate(24 * time.Hour)

	db := config.DB.Model(&models.AIUsageDailyStat{}).Where("date >= ?", since)
	if q.Provider != "" {
		db = db.Where("provider = ?", q.Provider)
	}
	if q.AppType != "" {
		db = db.Where("app_type = ?", q.AppType)
	}

	var items []DashboardDataItem
	err := db.Select(
		"date, model_id, provider, model_name, app_type, total_tokens, total_cost, call_count",
	).Order("date asc, model_id asc").Find(&items).Error
	return items, err
}

// GetUsageSummary 获取汇总统计
func GetUsageSummary(q UsageDashboardQuery) (map[string]interface{}, error) {
	if q.Days <= 0 {
		q.Days = 7
	}
	since := time.Now().AddDate(0, 0, -q.Days).Truncate(24 * time.Hour)

	db := config.DB.Model(&models.AIUsageDailyStat{}).Where("date >= ?", since)
	if q.Provider != "" {
		db = db.Where("provider = ?", q.Provider)
	}
	if q.AppType != "" {
		db = db.Where("app_type = ?", q.AppType)
	}

	type SumResult struct {
		TotalTokens int64
		TotalCost   float64
		TotalCalls  int64
	}
	var res SumResult
	err := db.Select(
		"COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(total_cost), 0) as total_cost, COALESCE(SUM(call_count), 0) as total_calls",
	).Scan(&res).Error

	return map[string]interface{}{
		"total_tokens": res.TotalTokens,
		"total_cost":   res.TotalCost,
		"total_calls":  res.TotalCalls,
		"days":         q.Days,
	}, err
}
