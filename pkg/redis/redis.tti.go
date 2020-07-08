// Copyright © 2020 The Things Industries B.V.

package redis

// ReadOnlyConfig represents Redis read-only configuration.
type ReadOnlyConfig struct {
	Address  string `name:"address" description:"Address of the Redis server"`
	Password string `name:"password" description:"Password of the Redis server"`
	Database int    `name:"database" description:"Redis database to use"`
	PoolSize int    `name:"pool-size" description:"The maximum number of database connections"`
}
