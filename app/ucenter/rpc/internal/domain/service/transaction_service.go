package service

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/repository"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type TransactionService struct {
	repo *repository.TransactionRepository
}

func NewTransactionService(repo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) FindTransaction(
	ctx context.Context,
	memberID int64,
	pageNo int64,
	pageSize int64,
	symbol string,
	startTime string,
	endTime string,
	transactionType string,
) ([]*assetpb.MemberTransaction, int64, error) {
	list, total, err := s.repo.FindTransaction(ctx, memberID, pageNo, pageSize, symbol, startTime, endTime, transactionType)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*assetpb.MemberTransaction, 0, len(list))
	for _, transaction := range list {
		resp = append(resp, transaction.ToProto())
	}
	return resp, total, nil
}
