package handler

import (
	"context"
	"testing"

	pb "github.com/IErcOrg/IERC_Indexer/api/indexer"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func TestIndexerService(t *testing.T) {
	suite.Run(t, new(TestIndexerServiceSuite))
}

type TestIndexerServiceSuite struct {
	suite.Suite
	client pb.IndexerClient
}

func (s *TestIndexerServiceSuite) SetupSuite() {
	conn, err := grpc.Dial("localhost:12301", grpc.WithInsecure())
	s.NoError(err)

	s.client = pb.NewIndexerClient(conn)
}

func (s *TestIndexerServiceSuite) TestSubscribe() {
	stream, err := s.client.SubscribeEvent(context.Background(), &pb.SubscribeRequest{StartBlock: 4998000})
	s.NoError(err)

	for {
		reply, err := stream.Recv()
		if err != nil {
			s.Nil(err)
			return
		}
		_ = reply

		//for _, e := range reply.Events {
		//
		//	switch e.Kind {
		//	case pb.Event_IERC20TickCreated:
		//		var data = &pb.IERC20TickCreated{}
		//		err := proto.Unmarshal(e.Data, data)
		//		s.NoError(err)
		//		spew.Dump(data)
		//
		//	case pb.Event_IERC20Transferred:
		//		var data = &pb.IERC20TickTransferred{}
		//		err := proto.Unmarshal(e.Data, data)
		//		s.NoError(err)
		//		spew.Dump(data)
		//	}
		//}

	}
}

func (s *TestIndexerServiceSuite) TestSubscribeStatus() {
	stream, err := s.client.SubscribeSystemStatus(context.Background(), &pb.SubscribeSystemStatusRequest{})
	s.NoError(err)

	for {
		reply, err := stream.Recv()
		if err != nil {
			s.Nil(err)
			return
		}

		spew.Dump(reply)

	}
}
