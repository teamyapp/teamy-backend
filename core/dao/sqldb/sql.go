package sqldb

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

func toStrsString(strs []string) string {
	strStrings := make([]string, 0)
	for _, singleStr := range strs {
		strStrings = append(strStrings, singleStr)
	}

	return strings.Join(strStrings, ",")
}
