package http

import "strings"

// joinStrings 将字符串切片用逗号连接
func joinStrings(strs []string) string {
	return strings.Join(strs, ", ")
}
