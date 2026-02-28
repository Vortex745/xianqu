package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ResendEmailRequest Resend API 请求体
type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

// SendVerificationEmail 通过 Resend API 发送验证码邮件
func SendVerificationEmail(toEmail string, code string) error {
	apiKey := "re_bht3H8RT_KFwSLHfxru9ZoWmbhroS5b63"
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY 未配置")
	}

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="margin:0;padding:0;background:#f4f4f7;font-family:'Segoe UI',Arial,sans-serif;">
  <div style="max-width:480px;margin:40px auto;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,0.08);">
    <div style="background:linear-gradient(135deg,#ffda44 0%%,#ffc107 100%%);padding:32px 24px;text-align:center;">
      <h1 style="margin:0;font-size:24px;color:#1a1a1a;font-weight:900;letter-spacing:2px;">XIANQU</h1>
      <p style="margin:6px 0 0;font-size:13px;color:rgba(26,26,26,0.7);font-weight:600;">闲趣社区二手</p>
    </div>
    <div style="padding:36px 32px;">
      <h2 style="margin:0 0 12px;font-size:20px;color:#333;">您好 👋</h2>
      <p style="margin:0 0 24px;font-size:15px;color:#555;line-height:1.6;">
        您正在进行邮箱验证操作，以下是您的安全验证码：
      </p>
      <div style="background:#f8f9fa;border:2px dashed #ffda44;border-radius:12px;padding:20px;text-align:center;margin:0 0 24px;">
        <span style="font-size:36px;font-weight:900;letter-spacing:8px;color:#1a1a1a;">%s</span>
      </div>
      <div style="background:#fffbeb;border-radius:8px;padding:14px 16px;margin:0 0 24px;">
        <p style="margin:0;font-size:13px;color:#92400e;line-height:1.5;">
          ⏱ 验证码有效期为 <strong>10 分钟</strong>，请尽快使用。<br>
          🔒 如果您没有请求此验证码，请忽略此邮件。
        </p>
      </div>
      <hr style="border:none;border-top:1px solid #eee;margin:24px 0;">
      <p style="margin:0;font-size:12px;color:#aaa;text-align:center;">
        此邮件由系统自动发出，请勿直接回复。
      </p>
    </div>
  </div>
</body>
</html>`, code)

	// 使用 Resend API 发送
	reqBody := ResendEmailRequest{
		From:    "闲趣社区 <onboarding@315279.xyz>",
		To:      []string{toEmail},
		Subject: fmt.Sprintf("【闲趣】您的验证码：%s", code),
		Html:    htmlContent,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend API 返回错误 (%d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("✅ 验证码邮件已发送至 %s\n", toEmail)
	return nil
}
