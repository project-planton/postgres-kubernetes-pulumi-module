package service

import (
	"fmt"
	"github.com/plantoncloud-inc/go-commons/kubernetes/network/dns"
)

func GetKubeServiceNameFqdn(mongodbClusterName, namespace string) string {
	return fmt.Sprintf("%s.%s.%s", GetKubeServiceName(mongodbClusterName), namespace, dns.DefaultDomain)
}

func GetKubeServiceName(mongodbClusterName string) string {
	return fmt.Sprintf(mongodbClusterName)
}
