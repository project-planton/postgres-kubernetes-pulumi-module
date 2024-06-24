package service

import (
	"fmt"
	"github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
)

func GetKubeServiceNameFqdn(postgresKubernetesName, namespace string) string {
	return fmt.Sprintf("%s.%s.%s", GetKubeServiceName(postgresKubernetesName), namespace, dns.DefaultDomain)
}

func GetKubeServiceName(postgresKubernetesName string) string {
	return fmt.Sprintf(postgresKubernetesName)
}
