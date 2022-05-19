package rpc

import (
	"context"
	"log"

	"github.com/teamyapp/cloud/app/api/rpc/proto"
	"github.com/teamyapp/cloud/app/gen"
	"google.golang.org/grpc"
)

type GeneratorService struct {
	proto.UnimplementedGeneratorServer
	uniqueNumberGeneratorFactory gen.UniqueNumberFactory
	uniqueNumberGenerators       map[string]*gen.UniqueNumber
	uniqueStringGenerators       map[string]*gen.UniqueString
}

var _ proto.GeneratorServer = (*GeneratorService)(nil)
var _ Service = (*GeneratorService)(nil)

func (g GeneratorService) GenerateUniqueNumber(
	ctx context.Context,
	request *proto.GenerateUniqueNumberRequest,
) (*proto.GenerateUniqueNumberResponse, error) {
	uniqueNumGen, ok := g.uniqueNumberGenerators[request.SequenceName]
	if !ok {
		var err error
		uniqueNumGen, err = g.uniqueNumberGeneratorFactory.MakeUniqueNumber(request.SequenceName)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		g.uniqueNumberGenerators[request.SequenceName] = uniqueNumGen
	}

	uniqueNum, err := uniqueNumGen.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &proto.GenerateUniqueNumberResponse{UniqueNumber: uniqueNum}, nil
}

func (g GeneratorService) GenerateUniqueString(
	ctx context.Context,
	request *proto.GenerateUniqueStringRequest,
) (*proto.GenerateUniqueStringResponse, error) {
	uniqueStringGen, ok := g.uniqueStringGenerators[request.SequenceName]
	if !ok {
		strGen, err := gen.NewUniqueString(request.SequenceName, int(request.StringLength), request.Alphabet, g.uniqueNumberGeneratorFactory)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		uniqueStringGen = &strGen
		g.uniqueStringGenerators[request.SequenceName] = uniqueStringGen
	}

	uniqueStr, err := uniqueStringGen.GenerateUniqueString()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &proto.GenerateUniqueStringResponse{UniqueString: uniqueStr}, nil
}

func (g GeneratorService) registerServer(server *grpc.Server) {
	proto.RegisterGeneratorServer(server, g)
}

func NewGeneratorService(uniqueNumberGeneratorFactory gen.UniqueNumberFactory) GeneratorService {
	return GeneratorService{
		uniqueNumberGeneratorFactory: uniqueNumberGeneratorFactory,
		uniqueNumberGenerators:       make(map[string]*gen.UniqueNumber),
		uniqueStringGenerators:       make(map[string]*gen.UniqueString),
	}
}
