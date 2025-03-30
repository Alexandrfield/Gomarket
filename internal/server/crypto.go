package server

import (
	"crypto/sha256"
	b64 "encoding/base64"
)

func ComplicatedPasswd(passwd string) string {
	h := sha256.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	sEnc := b64.StdEncoding.EncodeToString(bs)
	return sEnc
}
