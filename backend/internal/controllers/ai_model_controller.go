package controllers

import (
	"encoding/json"
	"fmt"
	"gotest/internal/models"
	"gotest/internal/services"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AIModelController struct{}

// ────────────── 管理端 CRUD ──────────────

// Create 新增模型配置
func (ctrl *AIModelController) Create(c *gin.Context) {
	var input struct {
		Provider    string  `json:"provider" binding:"required"`
		ModelName   string  `json:"model_name" binding:"required"`
		APIKey      string  `json:"api_key" binding:"required"`
		BaseURL     string  `json:"base_url"`
		PricePerK   float64 `json:"price_per_k"`
		Status      int     `json:"status"`
		Priority    int     `json:"priority"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	m := models.AIModel{
		Provider:    input.Provider,
		ModelName:   input.ModelName,
		BaseURL:     input.BaseURL,
		PricePerK:   input.PricePerK,
		Status:      input.Status,
		Priority:    input.Priority,
		Description: input.Description,
	}

	if err := services.CreateAIModel(&m, input.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "创建成功", "data": m})
}

// Update 编辑模型配置
func (ctrl *AIModelController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}

	var input struct {
		Provider    *string  `json:"provider"`
		ModelName   *string  `json:"model_name"`
		APIKey      string   `json:"api_key"` // 空则不修改
		BaseURL     *string  `json:"base_url"`
		PricePerK   *float64 `json:"price_per_k"`
		Status      *int     `json:"status"`
		Priority    *int     `json:"priority"`
		Description *string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	updates := make(map[string]interface{})
	if input.Provider != nil {
		updates["provider"] = *input.Provider
	}
	if input.ModelName != nil {
		updates["model_name"] = *input.ModelName
	}
	if input.BaseURL != nil {
		updates["base_url"] = *input.BaseURL
	}
	if input.PricePerK != nil {
		updates["price_per_k"] = *input.PricePerK
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.Priority != nil {
		updates["priority"] = *input.Priority
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if err := services.UpdateAIModel(uint(id), updates, input.APIKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// Delete 删除模型配置
func (ctrl *AIModelController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}
	if err := services.DeleteAIModel(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// List 管理端列表
func (ctrl *AIModelController) List(c *gin.Context) {
	list, err := services.ListAIModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetOne 获取单个配置详情
func (ctrl *AIModelController) GetOne(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}
	m, err := services.GetAIModelByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

// ToggleStatus 启用/停用模型
func (ctrl *AIModelController) ToggleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}
	var input struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	updates := map[string]interface{}{"status": input.Status}
	if err := services.UpdateAIModel(uint(id), updates, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "操作成功"})
}

// ────────────── 前台公共接口 ──────────────

// ActiveList 获取可用模型列表（前台调用）
func (ctrl *AIModelController) ActiveList(c *gin.Context) {
	list, err := services.ListActiveAIModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetModelSecret 内部接口：获取模型解密信息（仅供AI Service调用，需鉴权）
func (ctrl *AIModelController) GetModelSecret(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效ID"})
		return
	}
	m, err := services.GetDecryptedModel(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模型不可用"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         m.ID,
		"provider":   m.Provider,
		"model_name": m.ModelName,
		"api_key":    m.APIKey,
		"base_url":   m.BaseURL,
	})
}

// ────────────── 用量上报 ──────────────

// ReportUsage 上报Tokens用量
func (ctrl *AIModelController) ReportUsage(c *gin.Context) {
	var input struct {
		ModelID      uint   `json:"model_id" binding:"required"`
		AppType      string `json:"app_type" binding:"required"`
		PromptTokens int    `json:"prompt_tokens"`
		OutputTokens int    `json:"output_tokens"`
		TotalTokens  int    `json:"total_tokens"`
		SessionID    string `json:"session_id"`
		UserID       uint   `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 如果未提供 total，自动计算
	total := input.TotalTokens
	if total == 0 {
		total = input.PromptTokens + input.OutputTokens
	}

	log := &models.AIUsageLog{
		ModelID:      input.ModelID,
		AppType:      input.AppType,
		PromptTokens: input.PromptTokens,
		OutputTokens: input.OutputTokens,
		TotalTokens:  total,
		SessionID:    input.SessionID,
		UserID:       input.UserID,
	}

	if err := services.ReportUsage(log); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上报失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "上报成功", "cost": log.Cost})
}

// ────────────── 数据看板 ──────────────

// Dashboard 看板数据
func (ctrl *AIModelController) Dashboard(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	provider := c.Query("provider")
	appType := c.Query("app_type")

	query := services.UsageDashboardQuery{
		Days:     days,
		Provider: provider,
		AppType:  appType,
	}

	items, err := services.GetUsageDashboard(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	summary, err := services.GetUsageSummary(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":   items,
		"summary": summary,
	})
}

// DetectModels 根据 API Key 和 Base URL 自动探测可用模型
func (ctrl *AIModelController) DetectModels(c *gin.Context) {
	var input struct {
		Provider string `json:"provider" binding:"required"`
		APIKey   string `json:"api_key" binding:"required"`
		BaseURL  string `json:"base_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	baseURL := input.BaseURL
	if baseURL == "" {
		switch input.Provider {
		case "deepseek":
			baseURL = "https://api.deepseek.com"
		case "aliyun":
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		case "minimax":
			baseURL = "https://api.minimax.chat/v1"
		case "zhipu":
			baseURL = "https://open.bigmodel.cn/api/paas/v4"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先选择提供商或填入 Base URL"})
			return
		}
	}

	// 统一探测逻辑：调用 /models 接口
	detectURL := strings.TrimRight(baseURL, "/") + "/models"
	// 特殊处理：有些厂商的 /models 必须在 /v1 下，或者是 DeepSeek 的特殊处理
	if input.Provider == "deepseek" && !strings.Contains(baseURL, "/v1") {
		// DeepSeek 允许直接请求根路径下的 /models
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", detectURL, nil)
	req.Header.Set("Authorization", "Bearer "+input.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "连接模型商失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errMsg := string(body)
		if len(errMsg) > 200 {
			errMsg = errMsg[:200] + "..."
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("API 验证失败 (%d): %s", resp.StatusCode, errMsg)})
		return
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "响应数据解析失败"})
		return
	}

	var modelNames []string
	for _, m := range result.Data {
		modelNames = append(modelNames, m.ID)
	}

	if len(modelNames) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "未探测到可用模型", "models": []string{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "探测成功", "models": modelNames})
}
