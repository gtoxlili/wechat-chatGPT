package signature

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/cespare/xxhash"
	"io"
	"os"
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

// 计算文件的hash值
func GetFileHash(file *os.File) ([]byte, error) {
	h := xxhash.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return nil, err
	}
	defer file.Seek(0, 0)
	return h.Sum(nil), nil
}
