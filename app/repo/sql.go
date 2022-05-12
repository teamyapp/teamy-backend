package repo

import (
	"strconv"
	"strings"
)

func toIDsString(ids []uint64) string {
	idStrings := make([]string, 0)
	for _, singleID := range ids {
		idStrings = append(idStrings, strconv.Itoa(int(singleID)))
	}
	return strings.Join(idStrings, ",")
}
