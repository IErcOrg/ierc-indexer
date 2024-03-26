package handler

import (
	"context"
	"errors"
	"time"

	pb "github.com/IErcOrg/IERC_Indexer/api/indexer"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/service"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IndexHandler struct {
	pb.UnimplementedIndexerServer

	ctx    context.Context
	cancel context.CancelFunc

	srv       *service.IndexDomainService
	aggRepo   domain.EventRepository
	fetcher   domain.BlockFetcher
	blockRepo domain.BlockRepository

	logger *log.Helper
}

func NewIndexHandler(
	srv *service.IndexDomainService,
	aggRepo domain.EventRepository,
	fetcher domain.BlockFetcher,
	blockRepo domain.BlockRepository,
	logger log.Logger,
) *IndexHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &IndexHandler{
		UnimplementedIndexerServer: pb.UnimplementedIndexerServer{},
		ctx:                        ctx,
		cancel:                     cancel,
		srv:                        srv,
		aggRepo:                    aggRepo,
		fetcher:                    fetcher,
		blockRepo:                  blockRepo,
		logger:                     log.NewHelper(log.With(logger, "module", "handler")),
	}
}

func (s *IndexHandler) Start(_ context.Context) error {
	return nil
}

func (s *IndexHandler) Stop(_ context.Context) error {
	if s.cancel != nil {
		s.cancel()
	}

	return nil
}

func (s *IndexHandler) SubscribeEvent(req *pb.SubscribeRequest, conn pb.Indexer_SubscribeEventServer) error {

	lastBlockNumber := req.StartBlock
	stream, err := s.aggRepo.SubscribeEvent(conn.Context(), req.StartBlock)
	if err != nil {
		return err
	}

	for {
		select {

		case <-s.ctx.Done():
			s.logger.Info("Handler stop. SubscribeEvent quit")
			return s.ctx.Err()

		case <-conn.Context().Done():
			s.logger.Error("SubscribeEvent stream closed")
			return nil

		case err := <-stream.Err():
			return err

		case data, ok := <-stream.Next():
			if !ok {
				return nil
			}
			if lastBlockNumber > data.BlockNumber {
				continue
			}

			lastBlockNumber = data.BlockNumber

			var reply = pb.SubscribeReply{
				BlockNumber:     data.BlockNumber,
				PrevBlockNumber: data.PreviousBlock(),
				Events:          make([]*pb.Event, 0, len(data.Events)),
			}

			for _, item := range data.Events {
				event := ConvertEventEntityToProtobuf(item)
				if event != nil {
					reply.Events = append(reply.Events, event)
				}
			}

			if err := conn.Send(&reply); err != nil {
				return err
			}
		}
	}
}

func (s *IndexHandler) SubscribeSystemStatus(_ *pb.SubscribeSystemStatusRequest, conn pb.Indexer_SubscribeSystemStatusServer) error {
	var (
		reply  pb.SubscribeSystemStatusReply
		ticker = time.NewTicker(time.Second * 5)
	)
	defer ticker.Stop()

	for {
		select {

		case <-s.ctx.Done():
			s.logger.Info("Handler stop. SubscribeStatus quit")
			return s.ctx.Err()

		case <-conn.Context().Done():
			s.logger.Error("SubscribeStatus stream closed")
			return nil

		case <-ticker.C:
		}

		var needUpdate bool

		latest, err := s.fetcher.GetBlockHeaderByNumber(conn.Context(), 0)
		if err != nil {
			return err
		}

		sync, err := s.blockRepo.QueryLastProcessedBlock(conn.Context(), reply.SyncBlock)
		if err != nil {
			return err
		}

		if latest != nil && reply.LatestBlock < latest.Number {
			reply.LatestBlock = latest.Number
			needUpdate = true
		}

		if sync != nil && reply.SyncBlock < sync.Number {
			reply.SyncBlock = sync.Number
			needUpdate = true
		}

		if !needUpdate {
			continue
		}

		if err := conn.Send(&reply); err != nil {
			return err
		}
	}
}

func (s *IndexHandler) QueryEvents(ctx context.Context, req *pb.QueryEventsRequest) (*pb.QueryEventsReply, error) {

	blocks, err := s.aggRepo.QueryEventsByBlocks(ctx, req.StartBlock, int(req.Size))
	if err != nil {
		return nil, err
	}

	var data = make([]*pb.QueryEventsReply_EventsByBlock, 0, len(blocks))
	for _, block := range blocks {

		var events = make([]*pb.Event, 0, len(block.Events))
		for _, item := range block.Events {
			event := ConvertEventEntityToProtobuf(item)
			if event != nil {
				events = append(events, event)
			}
		}

		data = append(data, &pb.QueryEventsReply_EventsByBlock{
			BlockNumber:     block.BlockNumber,
			PrevBlockNumber: block.PreviousBlock(),
			Events:          events,
		})
	}

	return &pb.QueryEventsReply{EventByBlocks: data}, nil
}

func (s *IndexHandler) QuerySystemStatus(ctx context.Context, req *pb.QuerySystemStatusRequest) (*pb.QuerySystemStatusReply, error) {

	lastBlock, err := s.aggRepo.GetBlockNumberByLastEvent(ctx)

	sync, err := s.blockRepo.QueryLastProcessedBlock(ctx, lastBlock)
	if err != nil {
		return nil, err
	}

	if sync == nil {
		return &pb.QuerySystemStatusReply{SyncBlock: lastBlock}, nil
	}

	return &pb.QuerySystemStatusReply{SyncBlock: sync.Number}, nil
}

func (s *IndexHandler) CheckTransfer(ctx context.Context, req *pb.CheckTransferRequest) (*pb.CheckTransferReply, error) {

	if req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid tx hash")
	}

	if req.PositionIndex < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid position")
	}

	events, err := s.aggRepo.QueryEventsByHash(ctx, req.Hash)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return s.checkTransfer(ctx, req)
	}

	for _, event := range events {
		if int64(event.PosInIERCTxs()) != req.PositionIndex {
			continue
		}

		ee, ok := event.(*domain.IERC20TransferredEvent)
		if !ok || ee.GetOperate() != protocol.OpTransfer {
			return nil, errors.New("not a transfer")
		}

		return &pb.CheckTransferReply{
			Data: &pb.CheckTransferReply_TransferRecord{
				Sender:   ee.Data.From,
				Receiver: ee.Data.To,
				Tick:     ee.Data.Tick,
				Amount:   ee.Data.Amount.String(),
				Status:   ee.ErrCode == 0,
			},
		}, nil
	}

	return nil, status.Error(codes.NotFound, "not found")
}

func (s *IndexHandler) checkTransfer(ctx context.Context, req *pb.CheckTransferRequest) (*pb.CheckTransferReply, error) {

	tx, err := s.blockRepo.QueryTransactionByHash(ctx, req.GetHash())
	if err != nil {
		return nil, err
	}

	if !tx.IsProcessed {
		return nil, status.Error(codes.NotFound, "not found")
	}

	if tx.IERCTransaction == nil {
		return nil, errors.New(tx.Remark)
	}

	tt, ok := tx.IERCTransaction.(*protocol.TransferCommand)
	if !ok || tt.Operate != protocol.OpTransfer {
		return nil, errors.New("not a transfer")
	}

	if req.PositionIndex > int64(len(tt.Records)) {
		return nil, status.Error(codes.NotFound, "not found")
	}

	record := tt.Records[req.PositionIndex]

	return &pb.CheckTransferReply{
		Data: &pb.CheckTransferReply_TransferRecord{
			Sender:   record.From,
			Receiver: record.Recv,
			Tick:     record.Tick,
			Amount:   record.Amount.String(),
			Status:   tx.Code == 0,
		},
	}, nil
}
