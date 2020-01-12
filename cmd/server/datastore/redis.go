package reports

import (
	"github.com/garyburd/redigo/redis"
)

//redisAddr = "redis://user:@localhost:6379/0"

type RedisStore struct {
	conn redis.Conn
}

func createRedisStore(connection string) (Store, error) {
	s := RedisStore{}
	var err error
	s.conn, err = redis.DialURL(connection)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveReport - send report binary file into store
func (s *RedisStore) SaveReport(key, filepath string) error {

	s.conn.Do(key)

	return nil
}
