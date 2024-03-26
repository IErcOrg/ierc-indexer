//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/IErcOrg/IERC_Indexer/internal/conf"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/service"
	"github.com/IErcOrg/IERC_Indexer/internal/facade"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(string, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		conf.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		facade.ProviderSet,
		newApp,
	))
}
