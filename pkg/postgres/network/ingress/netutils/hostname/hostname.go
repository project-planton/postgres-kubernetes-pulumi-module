package hostname

import (
	"fmt"
)

func GetInternalHostname(postgresKubernetesId, environmentName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s-internal.%s", postgresKubernetesId, environmentName, endpointDomainName)
}

func GetExternalHostname(postgresKubernetesId, environmentName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s.%s", postgresKubernetesId, environmentName, endpointDomainName)
}
