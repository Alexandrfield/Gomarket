package server

import "crypto/sha256"

func ComplicatedPasswd(passwd string) string {
	h := sha256.New()
	h.Write([]byte(passwd))
	bs := h.Sum(nil)
	return string(bs)
}
