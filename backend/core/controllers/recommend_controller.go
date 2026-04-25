package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gotest/config"
	"gotest/core/middleware"
	"gotest/core/models"

	"github.com/gin-gonic/gin"
)

type RecommendController struct{}

type aiRecommendUserItemScore struct {
	ProductID uint    `json:"product_id"`
	Score     float64 `json:"score"`
}

type aiRecommendBehaviorRow struct {
	UserID    uint    `json:"user_id"`
	ProductID uint    `json:"product_id"`
	Weight    float64 `json:"weight"`
}

type aiRecommendCandidateProduct struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Price      float64 `json:"price"`
	ViewCount  int     `json:"view_count"`
	Popularity float64 `json:"popularity"`
}

type aiRecommendPayload struct {
	UserID         uint                          `json:"user_id"`
	UserItemScores []aiRecommendUserItemScore    `json:"user_item_scores"`
	BehaviorRows   []aiRecommendBehaviorRow      `json:"behavior_rows"`
	CandidateItems []aiRecommendCandidateProduct `json:"candidate_products"`
	TopK           int                           `json:"top_k"`
}

type aiRecommendResponse struct {
	ProductIDs []uint `json:"product_ids"`
	Source     string `json:"source"`
}

// List 获取推荐商品列表，接口：GET /api/products/recommend
func (rc *RecommendController) List(c *gin.Context) {
	limit := 20
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 50 {
		limit = 50
	}

	userID := rc.resolveOptionalUserID(c.GetHeader("Authorization"))
	shouldUsePersonalized := rc.personalizedReady() && userID > 0
	if !shouldUsePersonalized {
		rc.respondWithHot(c, limit, "hot")
		return
	}

	payload := rc.buildAIPayload(userID, limit)
	if len(payload.UserItemScores) == 0 || len(payload.CandidateItems) == 0 {
		rc.respondWithHot(c, limit, "hot")
		return
	}

	aiResult, err := rc.callAIRecommend(payload)
	if err != nil || len(aiResult.ProductIDs) == 0 {
		rc.respondWithHot(c, limit, "hot")
		return
	}

	products := rc.loadProductsByIDs(aiResult.ProductIDs, limit)
	if len(products) == 0 {
		rc.respondWithHot(c, limit, "hot")
		return
	}

	if len(products) < limit {
		exists := make(map[uint]bool, len(products))
		for _, p := range products {
			exists[p.ID] = true
		}
		hotIDs := rc.getHotProductIDs(limit*2, exists)
		hotProducts := rc.loadProductsByIDs(hotIDs, limit-len(products))
		products = append(products, hotProducts...)
	}

	source := strings.TrimSpace(aiResult.Source)
	if source == "" {
		source = "personalized"
	}

	c.JSON(http.StatusOK, gin.H{
		"list":   products,
		"total":  len(products),
		"page":   1,
		"source": source,
	})
}

func (rc *RecommendController) resolveOptionalUserID(authHeader string) uint {
	tokenString := strings.TrimSpace(authHeader)
	if tokenString == "" {
		return 0
	}
	if len(tokenString) > 7 && strings.EqualFold(tokenString[:7], "Bearer ") {
		tokenString = strings.TrimSpace(tokenString[7:])
	}
	claims, err := middleware.ParseToken(tokenString)
	if err != nil {
		return 0
	}
	return claims.UserID
}

func (rc *RecommendController) personalizedReady() bool {
	var behaviorCount int64
	config.DB.Model(&models.UserBehavior{}).Count(&behaviorCount)
	if behaviorCount < 500 {
		return false
	}

	since := time.Now().AddDate(0, 0, -30)
	var activeUsers int64
	config.DB.Model(&models.UserBehavior{}).
		Where("created_at >= ?", since).
		Distinct("user_id").
		Count(&activeUsers)

	return activeUsers >= 50
}

func (rc *RecommendController) buildAIPayload(userID uint, limit int) aiRecommendPayload {
	type userScoreRow struct {
		ProductID uint    `gorm:"column:product_id"`
		Score     float64 `gorm:"column:score"`
	}

	var userScores []userScoreRow
	_ = config.DB.Raw(`
		SELECT product_id, COALESCE(SUM(weight), 0) AS score
		FROM user_behaviors
		WHERE user_id = ?
		GROUP BY product_id
		ORDER BY score DESC
		LIMIT 300
	`, userID).Scan(&userScores).Error

	type behaviorRow struct {
		UserID    uint    `gorm:"column:user_id"`
		ProductID uint    `gorm:"column:product_id"`
		Weight    float64 `gorm:"column:weight"`
	}
	behaviorSince := time.Now().AddDate(0, 0, -90)
	var behaviorRows []behaviorRow
	_ = config.DB.Model(&models.UserBehavior{}).
		Select("user_id", "product_id", "weight").
		Where("created_at >= ?", behaviorSince).
		Order("created_at DESC").
		Limit(20000).
		Find(&behaviorRows).Error

	var products []models.Product
	_ = config.DB.Model(&models.Product{}).
		Select("id", "name", "category", "price", "view_count", "created_at", "status").
		Where("status = ?", 1).
		Order("created_at DESC").
		Limit(300).
		Find(&products).Error

	payload := aiRecommendPayload{
		UserID:         userID,
		UserItemScores: make([]aiRecommendUserItemScore, 0, len(userScores)),
		BehaviorRows:   make([]aiRecommendBehaviorRow, 0, len(behaviorRows)),
		CandidateItems: make([]aiRecommendCandidateProduct, 0, len(products)),
		TopK:           limit,
	}

	for _, item := range userScores {
		payload.UserItemScores = append(payload.UserItemScores, aiRecommendUserItemScore{
			ProductID: item.ProductID,
			Score:     item.Score,
		})
	}
	for _, row := range behaviorRows {
		payload.BehaviorRows = append(payload.BehaviorRows, aiRecommendBehaviorRow{
			UserID:    row.UserID,
			ProductID: row.ProductID,
			Weight:    row.Weight,
		})
	}
	for _, p := range products {
		payload.CandidateItems = append(payload.CandidateItems, aiRecommendCandidateProduct{
			ID:         p.ID,
			Name:       p.Name,
			Category:   p.Category,
			Price:      p.Price,
			ViewCount:  p.ViewCount,
			Popularity: float64(p.ViewCount),
		})
	}

	return payload
}

func (rc *RecommendController) callAIRecommend(payload aiRecommendPayload) (aiRecommendResponse, error) {
	aiURL := strings.TrimSpace(os.Getenv("AI_RECOMMEND_URL"))
	if aiURL == "" {
		baseURL := strings.TrimSpace(os.Getenv("AI_SERVICE_BASE_URL"))
		if baseURL == "" {
			baseURL = "http://localhost:8008"
		}
		aiURL = strings.TrimRight(baseURL, "/") + "/ai/recommend"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return aiRecommendResponse{}, err
	}

	req, err := http.NewRequest(http.MethodPost, aiURL, bytes.NewReader(body))
	if err != nil {
		return aiRecommendResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return aiRecommendResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return aiRecommendResponse{}, fmt.Errorf("ai recommend failed with status %d", resp.StatusCode)
	}

	var result aiRecommendResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return aiRecommendResponse{}, err
	}
	return result, nil
}

func (rc *RecommendController) respondWithHot(c *gin.Context, limit int, source string) {
	hotIDs := rc.getHotProductIDs(limit, nil)
	products := rc.loadProductsByIDs(hotIDs, limit)
	c.JSON(http.StatusOK, gin.H{
		"list":   products,
		"total":  len(products),
		"page":   1,
		"source": source,
	})
}

func (rc *RecommendController) getHotProductIDs(limit int, exclude map[uint]bool) []uint {
	type hotIDRow struct {
		ProductID uint `gorm:"column:product_id"`
	}
	since := time.Now().AddDate(0, 0, -30)
	var rows []hotIDRow
	_ = config.DB.Raw(`
		SELECT
			p.id AS product_id
		FROM products p
		LEFT JOIN user_behaviors ub ON ub.product_id = p.id AND ub.created_at >= ?
		WHERE p.status = 1
		GROUP BY p.id, p.created_at
		ORDER BY
			COALESCE(SUM(CASE
				WHEN ub.action = 'view' THEN 1
				WHEN ub.action = 'like' THEN 2
				WHEN ub.action = 'buy' THEN 5
				ELSE 0 END), 0) DESC,
			p.created_at DESC
		LIMIT ?
	`, since, limit*3).Scan(&rows).Error

	ids := make([]uint, 0, limit)
	for _, row := range rows {
		if row.ProductID == 0 {
			continue
		}
		if exclude != nil && exclude[row.ProductID] {
			continue
		}
		ids = append(ids, row.ProductID)
		if len(ids) >= limit {
			return ids
		}
	}
	return ids
}

func (rc *RecommendController) loadProductsByIDs(ids []uint, limit int) []models.Product {
	if len(ids) == 0 || limit <= 0 {
		return []models.Product{}
	}

	orderedIDs := make([]uint, 0, len(ids))
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		orderedIDs = append(orderedIDs, id)
		if len(orderedIDs) >= limit {
			break
		}
	}
	if len(orderedIDs) == 0 {
		return []models.Product{}
	}

	var products []models.Product
	if err := config.DB.Preload("User").
		Where("id IN ? AND status = ?", orderedIDs, 1).
		Find(&products).Error; err != nil {
		return []models.Product{}
	}

	productMap := make(map[uint]models.Product, len(products))
	for _, p := range products {
		if p.Image == "" {
			p.Image = "/uploads/default_product.png"
		}
		productMap[p.ID] = p
	}

	result := make([]models.Product, 0, len(orderedIDs))
	for _, id := range orderedIDs {
		p, ok := productMap[id]
		if !ok {
			continue
		}
		result = append(result, p)
	}
	return result
}
