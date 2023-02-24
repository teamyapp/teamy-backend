package implementation

import (
	"strconv"
	"strings"
)

func toIDsString(ids []uint64) string {
	idStrings := make([]string, 0)
	for _, singleID := range ids {
		idStrings = append(idStrings, strconv.FormatUint(singleID, 10))
	}

	return strings.Join(idStrings, ",")
}
