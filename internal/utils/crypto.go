package utils

import (
	"crypto/sha1"
	"fmt"
	"encoding/hex"
)

func StableSeed(parts ...any) int64 {
	// 1. 创建 SHA1 哈希对象
	h := sha1.New()

	// 2. 将所有 parts 依次写入哈希
	for _, part := range parts {
		// 把每个 part 转成字符串后写入
		_, _ = h.Write([]byte(fmt.Sprint(part)))
		// 写入一个空字节作为分隔符（避免 "ab"+"c" 和 "a"+"bc" 产生相同哈希）
		_, _ = h.Write([]byte{0})
	}

	// 3. 计算哈希值（20 字节）
	sum := h.Sum(nil)

	// 4. 把前 8 字节转成 int64
	var seed int64
	for idx := 0; idx < 8; idx++ {
		// 每次左移 8 位，然后或上当前字节
		seed = (seed << 8) | int64(sum[idx])
	}

	// 5. 确保种子是正数
	if seed < 0 {
		seed = -seed
	}

	return seed
}


func ShortSHA(parts ...any) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(fmt.Sprint(part)))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}
