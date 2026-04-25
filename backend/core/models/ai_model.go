package models

import (
	"time"

	"gorm.io/gorm"
)

// AIModel AI模型配置表
type AIModel struct {
	gorm.Model
	Provider    string  `gorm:"type:varchar(50);not null;comment:模型提供商(deepseek/aliyun/minimax/zhipu)" json:"provider"`
	ModelName   string  `gorm:"type:varchar(100);not null;comment:模型标识(如deepseek-chat)" json:"model_name"`
	APIKey      string  `gorm:"type:varchar(512);not null;comment:API密钥(加密存储)" json:"-"`
	APIKeyHint  string  `gorm:"-" json:"api_key_hint,omitempty"` // 仅返回尾部4位
	BaseURL     string  `gorm:"type:varchar(255);comment:自定义API地址" json:"base_url"`
	PricePerK   float64 `gorm:"type:decimal(10,4);not null;default:0;comment:单价(元/千Tokens)" json:"price_per_k"`
	Status      int     `gorm:"default:1;comment:状态(1:启用 0:停用)" json:"status"`
	Priority    int     `gorm:"default:0;comment:优先级(数值越大优先级越高)" json:"priority"`
	Description string  `gorm:"type:varchar(255);comment:备注说明" json:"description"`
}

func (AIModel) TableName() string {
	return "ai_models"
}

// AIUsageLog AI用量记录表
type AIUsageLog struct {
	gorm.Model
	ModelID       uint    `gorm:"not null;index;comment:关联模型ID" json:"model_id"`
	AppType       string  `gorm:"type:varchar(20);not null;index;comment:应用类型(customer_service/agent)" json:"app_type"`
	PromptTokens  int     `gorm:"default:0;comment:输入Tokens" json:"prompt_tokens"`
	OutputTokens  int     `gorm:"default:0;comment:输出Tokens" json:"output_tokens"`
	TotalTokens   int     `gorm:"default:0;comment:总Tokens" json:"total_tokens"`
	Cost          float64 `gorm:"type:decimal(10,6);default:0;comment:本次成本(元)" json:"cost"`
	SessionID     string  `gorm:"type:varchar(64);comment:会话ID" json:"session_id"`
	UserID        uint    `gorm:"default:0;comment:触发用户ID" json:"user_id"`

	// 冗余字段，快速查询
	Provider  string `gorm:"type:varchar(50);comment:提供商快照" json:"provider"`
	ModelName string `gorm:"type:varchar(100);comment:模型名快照" json:"model_name"`

	// 关联
	AIModel AIModel `gorm:"foreignKey:ModelID" json:"ai_model,omitempty"`
}

func (AIUsageLog) TableName() string {
	return "ai_usage_logs"
}

// AIUsageDailyStat 每日汇总统计（可选，加速看板查询）
type AIUsageDailyStat struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"type:date;not null;index;comment:日期" json:"date"`
	ModelID     uint      `gorm:"not null;index;comment:模型ID" json:"model_id"`
	AppType     string    `gorm:"type:varchar(20);not null;index;comment:应用类型" json:"app_type"`
	TotalTokens int64     `gorm:"default:0;comment:当日总Tokens" json:"total_tokens"`
	TotalCost   float64   `gorm:"type:decimal(12,4);default:0;comment:当日总成本" json:"total_cost"`
	CallCount   int64     `gorm:"default:0;comment:调用次数" json:"call_count"`

	// 冗余
	Provider  string `gorm:"type:varchar(50)" json:"provider"`
	ModelName string `gorm:"type:varchar(100)" json:"model_name"`
}

func (AIUsageDailyStat) TableName() string {
	return "ai_usage_daily_stats"
}
