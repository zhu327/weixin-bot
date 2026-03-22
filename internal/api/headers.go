package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

// RandomWechatUin matches Node randomWechatUin: base64(String(randomUint32BE)).
func RandomWechatUin() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	v := binary.BigEndian.Uint32(b[:])
	s := fmt.Sprintf("%d", v)
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

// BuildHeaders returns POST headers for authenticated bot calls.
func BuildHeaders(token string) (map[string]string, error) {
	uin, err := RandomWechatUin()
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"Content-Type":      "application/json",
		"AuthorizationType": "ilink_bot_token",
		"Authorization":     "Bearer " + token,
		"X-WECHAT-UIN":      uin,
	}, nil
}
