package utils

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"strings"
)

type CheckPoint struct {
	Height *big.Int `json:"height"`
	Root   string   `json:"root"`
}

func (c *CheckPoint) String() string {
	return fmt.Sprintf("height:%v", c.Height.Uint64(), "root:", c.Root)
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
