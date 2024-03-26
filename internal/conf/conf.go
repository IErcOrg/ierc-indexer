package conf

import (
	"encoding/json"
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewConfigFromPath)

type Config struct {
	Config config.Config

	*Bootstrap

	InvalidTxHash map[string]struct{}
}

func NewConfigFromPath(path string, logger log.Logger) (*Config, func(), error) {
	helper := log.NewHelper(logger)
	helper.Infof("load config file. path: %s", path)

	c := config.New(
		config.WithSource(
			file.NewSource(path),
		),
	)
	cleanup := func() {
		_ = c.Close()
	}

	if err := c.Load(); err != nil {
		cleanup()
		return nil, nil, err
	}

	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		cleanup()
		return nil, nil, err
	}

	helper.Infof("load invalid tx hash. path: %s", bc.Runtime.InvalidTxHashPath)
	invalidHash, err := LoadInvalidHashMap(bc.Runtime.InvalidTxHashPath)
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	return &Config{
		Config:        c,
		Bootstrap:     &bc,
		InvalidTxHash: invalidHash,
	}, cleanup, nil
}

func LoadInvalidHashMap(path string) (map[string]struct{}, error) {

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var records map[string][]string
	err = json.Unmarshal(bytes, &records)
	if err != nil {
		return nil, err
	}

	var invalidHashMap = make(map[string]struct{})
	for _, record := range records {
		for _, r := range record {
			if _, existed := invalidHashMap[r]; existed {
				log.Infof("repeat hash: %s", r)
			}
			invalidHashMap[r] = struct{}{}
		}
	}

	if len(invalidHashMap) == 0 {
		log.Info("invalid hash list is empty")
		return invalidHashMap, nil
	}

	return invalidHashMap, nil
}
