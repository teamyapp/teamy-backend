package service

import (
	"context"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	dataCollector       obs.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	threadDao           dao.Thread
}

func (t Thread) createThread(ct context.Context) (uint64, error) {
	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	threadID := genThreadIDRes.UniqueNumber
	err = t.threadDao.CreateThread(ct, threadID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return threadID, err
}

func NewThread(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	threadDao dao.Thread,
) Thread {
	return Thread{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		threadDao:           threadDao,
	}
}
