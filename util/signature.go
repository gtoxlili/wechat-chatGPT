package util

import (
	"crypto/sha1"
	"encoding/hex"
	"sort"
)

func CheckSignature(signature, timestamp, nonce, token string) bool {
	// 1）将token、timestamp、nonce三个参数进行字典序排序
	sl := []string{token, timestamp, nonce}
	sort.Strings(sl)
	// 2）将三个参数字符串拼接成一个字符串进行sha1加密
	sum := sha1.Sum([]byte(sl[0] + sl[1] + sl[2]))
	// 3）开发者获得加密后的字符串可与 signature 对比，标识该请求来源于微信
	return signature == hex.EncodeToString(sum[:])
}
