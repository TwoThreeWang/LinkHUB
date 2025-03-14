package utils

import (
	"regexp"
	"strings"
)

// ExtractUsernameFromEmail 从email中提取用户名
func ExtractUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		return parts[0]
	}
	return "" // 如果没有 "@" 符号，或者电子邮件格式不正确，返回空字符串
}

// IsValidEmailByRegexp 电子邮件格式校验
func IsValidEmailByRegexp(email string) bool {
	// 电子邮件地址正则表达式
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+(?:\.[a-zA-Z]{2,})+$`)
	return emailRegex.MatchString(email)
}
