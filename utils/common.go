package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"strings"
)

type CheckPoint struct {
	Height uint64 `json:"height"`
	Root   string `json:"root"`
}

func (c *CheckPoint) String() string {
	return fmt.Sprintf("height:%v", c.Height, "root:", c.Root)
}
func (c *CheckPoint) Equal(ck *CheckPoint) bool {
	if ck == nil || c == nil {
		return false
	}
	if c.Root == ck.Root && c.Height == ck.Height {
		return true
	}
	return false
}

func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func IsDuplicateError(err string) bool {
	return strings.Contains(err, "Duplicate entry")
}

func JSON(v interface{}) string {
	bs, _ := json.Marshal(v)
	return string(bs)
}

func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("caught goroutine exception", "err", r)
			}
		}()

		fn()
	}()
}
