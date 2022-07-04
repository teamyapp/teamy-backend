package github

import (
	"log"
	"strconv"
	"strings"
	"time"
)

type githubTime struct {
	time.Time
}

func (c *githubTime) UnmarshalJSON(buf []byte) (err error) {
	str := strings.Trim(string(buf), `"`)
	unixTimestamp, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		c.Time = time.Unix(unixTimestamp, 0)
		return nil
	}

	tm, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		log.Println(err)
		return err
	}

	c.Time = tm
	return nil
}

func (c githubTime) String() string {
	return c.Time.String()
}
