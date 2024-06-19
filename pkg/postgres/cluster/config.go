package cluster

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	v1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

const (
	PostgresVersion       = "14"
	PostgresContainerPort = 5432
	// TeamId required by zalando operator
	TeamId                    = "db"
	EnvVarStunnelSidecarImage = "STUNNEL_SIDECAR_IMAGE"
	StunnelContainerName      = englishword.EnglishWord_stunnel
	StunnelContainerPort      = 15432
	StunnelCertMountPath      = "/server/ca.pem"
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
