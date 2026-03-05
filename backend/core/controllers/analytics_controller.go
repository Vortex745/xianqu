package controllers

import (
	"net/http"
	"time"

	"gotest/config"
	"gotest/core/models"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct{}

var supportedAnalyticsRanges = map[int]bool{
	1:  true,
	7:  true,
	30: true,
}

func resolveAnalyticsRange(c *gin.Context) int {
	rangeDays := 7
	switch c.Query("range") {
	case "1":
		rangeDays = 1
	case "7":
		rangeDays = 7
	case "30":
		rangeDays = 30
	}
	if !supportedAnalyticsRanges[rangeDays] {
		return 7
	}
	return rangeDays
}

func analyticsTimeWindow(rangeDays int) (startUTC time.Time, endUTC time.Time, startLocal time.Time, loc *time.Location) {
	loc, _ = time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(loc)
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startLocal = startOfToday.AddDate(0, 0, -(rangeDays - 1))
	endLocal := startOfToday.AddDate(0, 0, 1) // [start, end)
	return startLocal.UTC(), endLocal.UTC(), startLocal, loc
}

// Overview 核心指标总览
func (ac *AnalyticsController) Overview(c *gin.Context) {
	rangeDays := resolveAnalyticsRange(c)
	startUTC, endUTC, _, _ := analyticsTimeWindow(rangeDays)

	var dau int64
	var onSaleCount int64
	var viewCount int64
	var buyCount int64
	var gmv float64

	config.DB.Model(&models.UserBehavior{}).
		Where("created_at >= ? AND created_at < ?", startUTC, endUTC).
		Distinct("user_id").
		Count(&dau)

	config.DB.Model(&models.Product{}).
		Where("status = ?", 1).
		Count(&onSaleCount)

	config.DB.Model(&models.UserBehavior{}).
		Where("created_at >= ? AND created_at < ? AND action = ?", startUTC, endUTC, "view").
		Count(&viewCount)
	config.DB.Model(&models.UserBehavior{}).
		Where("created_at >= ? AND created_at < ? AND action = ?", startUTC, endUTC, "buy").
		Count(&buyCount)

	config.DB.Model(&models.Order{}).
		Select("COALESCE(SUM(price), 0)").
		Where("status >= ? AND created_at >= ? AND created_at < ?", 2, startUTC, endUTC).
		Scan(&gmv)

	conversionRate := 0.0
	if viewCount > 0 {
		conversionRate = float64(buyCount) / float64(viewCount) * 100
	}
	avgDailyViews := float64(viewCount) / float64(rangeDays)

	c.JSON(http.StatusOK, gin.H{
		"range":            rangeDays,
		"dau":              dau,
		"on_sale_count":    onSaleCount,
		"gmv":              gmv,
		"conversion_rate":  conversionRate,
		"avg_daily_views":  avgDailyViews,
		"view_count":       viewCount,
		"buy_count":        buyCount,
		"window_start_utc": startUTC,
		"window_end_utc":   endUTC,
	})
}

// BehaviorTrend 行为趋势（view/like/cart/buy）
func (ac *AnalyticsController) BehaviorTrend(c *gin.Context) {
	rangeDays := resolveAnalyticsRange(c)
	startUTC, endUTC, startLocal, loc := analyticsTimeWindow(rangeDays)

	type trendRow struct {
		Day    string `json:"day"`
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}

	var rows []trendRow
	if err := config.DB.Raw(`
		SELECT
			TO_CHAR((created_at AT TIME ZONE 'Asia/Shanghai')::date, 'YYYY-MM-DD') AS day,
			action,
			COUNT(*) AS count
		FROM user_behaviors
		WHERE created_at >= ? AND created_at < ? AND action IN ('view', 'like', 'cart', 'buy')
		GROUP BY day, action
		ORDER BY day ASC
	`, startUTC, endUTC).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询行为趋势失败"})
		return
	}

	dates := make([]string, 0, rangeDays)
	dateToIndex := make(map[string]int, rangeDays)
	for i := 0; i < rangeDays; i++ {
		day := startLocal.AddDate(0, 0, i).In(loc).Format("2006-01-02")
		dateToIndex[day] = i
		dates = append(dates, day)
	}

	series := map[string][]int64{
		"view": make([]int64, rangeDays),
		"like": make([]int64, rangeDays),
		"cart": make([]int64, rangeDays),
		"buy":  make([]int64, rangeDays),
	}

	for _, row := range rows {
		idx, ok := dateToIndex[row.Day]
		if !ok {
			continue
		}
		if _, exists := series[row.Action]; !exists {
			continue
		}
		series[row.Action][idx] = row.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"range":  rangeDays,
		"dates":  dates,
		"series": series,
	})
}

// HotProducts 热门商品 Top10
func (ac *AnalyticsController) HotProducts(c *gin.Context) {
	rangeDays := resolveAnalyticsRange(c)
	startUTC, endUTC, _, _ := analyticsTimeWindow(rangeDays)

	type hotProductRow struct {
		ProductID uint   `json:"product_id" gorm:"column:product_id"`
		Name      string `json:"name" gorm:"column:name"`
		Category  string `json:"category" gorm:"column:category"`
		ViewCount int64  `json:"view_count" gorm:"column:view_count"`
		LikeCount int64  `json:"like_count" gorm:"column:like_count"`
		BuyCount  int64  `json:"buy_count" gorm:"column:buy_count"`
	}

	var rows []hotProductRow
	if err := config.DB.Raw(`
		SELECT
			p.id AS product_id,
			p.name,
			p.category,
			SUM(CASE WHEN ub.action = 'view' THEN 1 ELSE 0 END) AS view_count,
			SUM(CASE WHEN ub.action = 'like' THEN 1 ELSE 0 END) AS like_count,
			SUM(CASE WHEN ub.action = 'buy' THEN 1 ELSE 0 END) AS buy_count
		FROM user_behaviors ub
		JOIN products p ON p.id = ub.product_id
		WHERE ub.created_at >= ? AND ub.created_at < ?
		GROUP BY p.id, p.name, p.category
		ORDER BY
			(SUM(CASE WHEN ub.action = 'view' THEN 1 ELSE 0 END) * 1
			+ SUM(CASE WHEN ub.action = 'like' THEN 1 ELSE 0 END) * 2
			+ SUM(CASE WHEN ub.action = 'buy' THEN 1 ELSE 0 END) * 5) DESC,
			p.id DESC
		LIMIT 10
	`, startUTC, endUTC).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询热门商品失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"range": rangeDays,
		"list":  rows,
	})
}

// CategoryStats 分类成交占比
func (ac *AnalyticsController) CategoryStats(c *gin.Context) {
	rangeDays := resolveAnalyticsRange(c)
	startUTC, endUTC, _, _ := analyticsTimeWindow(rangeDays)

	type categoryRow struct {
		Category string `json:"category" gorm:"column:category"`
		BuyCount int64  `json:"buy_count" gorm:"column:buy_count"`
	}

	var rows []categoryRow
	if err := config.DB.Raw(`
		SELECT
			COALESCE(NULLIF(p.category, ''), '其他') AS category,
			COUNT(*) AS buy_count
		FROM user_behaviors ub
		JOIN products p ON p.id = ub.product_id
		WHERE ub.action = 'buy' AND ub.created_at >= ? AND ub.created_at < ?
		GROUP BY COALESCE(NULLIF(p.category, ''), '其他')
		ORDER BY buy_count DESC
	`, startUTC, endUTC).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询分类占比失败"})
		return
	}

	total := int64(0)
	for _, row := range rows {
		total += row.BuyCount
	}

	type categoryResp struct {
		Category string  `json:"category"`
		BuyCount int64   `json:"buy_count"`
		Ratio    float64 `json:"ratio"`
	}
	resp := make([]categoryResp, 0, len(rows))
	for _, row := range rows {
		ratio := 0.0
		if total > 0 {
			ratio = float64(row.BuyCount) / float64(total) * 100
		}
		resp = append(resp, categoryResp{
			Category: row.Category,
			BuyCount: row.BuyCount,
			Ratio:    ratio,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"range": rangeDays,
		"total": total,
		"list":  resp,
	})
}

// UserActivity 用户活跃时段热力图（7x24）
func (ac *AnalyticsController) UserActivity(c *gin.Context) {
	rangeDays := resolveAnalyticsRange(c)
	startUTC, endUTC, _, _ := analyticsTimeWindow(rangeDays)

	type heatmapRow struct {
		WeekDay int   `json:"weekday" gorm:"column:weekday"`
		Hour    int   `json:"hour" gorm:"column:hour"`
		Count   int64 `json:"count" gorm:"column:count"`
	}

	var rows []heatmapRow
	if err := config.DB.Raw(`
		SELECT
			EXTRACT(ISODOW FROM (created_at AT TIME ZONE 'Asia/Shanghai'))::int AS weekday,
			EXTRACT(HOUR FROM (created_at AT TIME ZONE 'Asia/Shanghai'))::int AS hour,
			COUNT(*) AS count
		FROM user_behaviors
		WHERE created_at >= ? AND created_at < ?
		GROUP BY weekday, hour
	`, startUTC, endUTC).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询活跃时段失败"})
		return
	}

	dayLabels := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	hours := make([]int, 24)
	for i := 0; i < 24; i++ {
		hours[i] = i
	}

	grid := make([][]int64, 7)
	for day := 0; day < 7; day++ {
		grid[day] = make([]int64, 24)
	}

	heatData := make([][]int64, 0, len(rows))
	maxCount := int64(0)
	for _, row := range rows {
		if row.WeekDay < 1 || row.WeekDay > 7 || row.Hour < 0 || row.Hour > 23 {
			continue
		}
		dayIndex := row.WeekDay - 1
		grid[dayIndex][row.Hour] = row.Count
		heatData = append(heatData, []int64{int64(row.Hour), int64(dayIndex), row.Count})
		if row.Count > maxCount {
			maxCount = row.Count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"range":     rangeDays,
		"days":      dayLabels,
		"hours":     hours,
		"matrix":    grid,
		"data":      heatData,
		"max_count": maxCount,
	})
}
