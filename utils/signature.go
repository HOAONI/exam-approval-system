package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GeneratePaperSignature 生成试卷签名
func GeneratePaperSignature(paperID uint, title string, content string, questions string, timestamp time.Time) string {
	// 构建签名字符串
	data := fmt.Sprintf("%d|%s|%s|%s|%d", paperID, title, content, questions, timestamp.Unix())

	// 使用HMAC-SHA256算法生成签名
	// 注意：在实际应用中，应该使用环境变量或配置文件来存储密钥
	key := []byte("your-secret-key-here")
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPaperSignature 验证试卷签名
func VerifyPaperSignature(paperID uint, title string, content string, questions string, timestamp time.Time, signature string) bool {
	// 重新生成签名
	expectedSignature := GeneratePaperSignature(paperID, title, content, questions, timestamp)

	// 比较签名
	return signature == expectedSignature
}
