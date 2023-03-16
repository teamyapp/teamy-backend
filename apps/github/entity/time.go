package entity

import (
	"strconv"
	"strings"
	"time"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/inject"
)

type githubTime struct {
	time.Time
}

func (g *githubTime) UnmarshalJSON(buf []byte) (err error) {
	dataCollector := inject.Injector.Get(new(telemetry.DataCollector)).(telemetry.DataCollector)
	str := strings.Trim(string(buf), `"`)
	unixTimestamp, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		g.Time = time.Unix(unixTimestamp, 0)
		return nil
	}

	tm, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	g.Time = tm
	return nil
}

func (g githubTime) String() string {
	return g.Time.String()
}
