package generator

import (
	"go.jetify.com/typeid"
	"strings"
)

func GenerateId(prefix string) string {
	tid, _ := typeid.WithPrefix(prefix)
	combined := tid.Prefix() + strings.ToUpper(tid.Suffix())

	return combined
}
