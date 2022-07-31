package service

import (
	"context"
	"log"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	cloudClientRegistry *cloudAPI.ClientRegistry
	threadDao           dao.Thread
}

func (t Thread) createThread(ct context.Context) (uint64, error) {
	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	threadID := genThreadIDRes.UniqueNumber
	return threadID, t.threadDao.CreateThread(threadID)
}

func NewThread(cloudClientRegistry *cloudAPI.ClientRegistry, threadDao dao.Thread) Thread {
	return Thread{
		cloudClientRegistry: cloudClientRegistry,
		threadDao:           threadDao,
	}
}
