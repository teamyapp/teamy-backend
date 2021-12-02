package repo

import (
	"strconv"
	"strings"

	"github.com/teamyapp/one/entity"
)

func toIDsString(ids []entity.ID) string {
	idStrings := make([]string, 0)
	for _, singleID := range ids {
		idStrings = append(idStrings, strconv.Itoa(int(singleID)))
	}
	return strings.Join(idStrings, ",")
}
