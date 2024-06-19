package hostname

import (
	"fmt"
)

func GetInternalHostname(mongodbClusterId, environmentName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s-internal.%s", mongodbClusterId, environmentName, endpointDomainName)
}

func GetExternalHostname(mongodbClusterId, environmentName, endpointDomainName string) string {
	return fmt.Sprintf("%s.%s.%s", mongodbClusterId, environmentName, endpointDomainName)
}
