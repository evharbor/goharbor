package goharbor

import (
	"fmt"
	"testing"
)

func TestAuthKey(t *testing.T) {
	ak := AuthKey{AccessKey: "1111", SecretKey: "2222"}
	key := ak.Key("/a/我/b/?a=b&c=d", "POST", 1000)
	fmt.Println("TestAuthKey: key=" + key)
}
