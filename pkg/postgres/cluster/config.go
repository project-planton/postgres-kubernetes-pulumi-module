package cluster

import (
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

const (
	PostgresVersion       = "14"
	PostgresContainerPort = 5432
)

var (
	//not sure what the user flags are useful for
	defaultUserFlags = v1.UserFlags{
		"superuser",
		"createdb",
	}

	defaultPostgresParameters = map[string]string{
		"max_connections": "2000",
	}
)
