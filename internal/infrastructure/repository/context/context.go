package context

import (
	"context"

	"gorm.io/gorm"
)

type transactionDBKey struct{}

func WithTransactionDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionDBKey{}, db)
}

func TransactionDBFromContext(ctx context.Context) *gorm.DB {
	db, ok := ctx.Value(transactionDBKey{}).(*gorm.DB)
	if !ok {
		return nil
	}

	return db
}

type updateKindKey struct{}

type UpdateKind uint8

const (
	None UpdateKind = iota + 1
	UpdateCache
	UpdateDB
)

func WithUpdateKind(ctx context.Context, kind UpdateKind) context.Context {
	return context.WithValue(ctx, updateKindKey{}, kind)
}

func UpdateKindFromContext(ctx context.Context) UpdateKind {
	value, ok := ctx.Value(updateKindKey{}).(UpdateKind)
	if !ok {
		return None
	}

	return value
}
