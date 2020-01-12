package datastore

import (
	"errors"
	"strings"
)

type Handler interface {
	SaveFile(key, filepath string) error
	LoadFile(key, filepath string) error
}

func CreateStore(connection string) (Store, error) {
	if strings.HasPrefix(connection, "redis://") {
		return createRedisStore(connection)
	}
	return nil, errors.New("ReportsStore: Connection doesn't supported: " + connection)
}
