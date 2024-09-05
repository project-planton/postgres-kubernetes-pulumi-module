package outputs

import (
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/postgreskubernetes"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/kubernetes"
	"github.com/plantoncloud/stack-job-runner-golang-sdk/pkg/automationapi/autoapistackoutput"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

const (
	Namespace                         = "namespace"
	Service                           = "service"
	KubePortForwardCommand            = "kube-port-forward-command"
	KubeEndpoint                      = "kube-endpoint"
	IngressExternalHostname           = "ingress-external-hostname"
	IngressInternalHostname           = "ingress-internal-hostname"
	PostgresUserCredentialsSecretName = "postgres-user-credentials-secret-name"
	PostgresUsernameSecretKey         = "postgres-username-secret-key"
	PostgresPasswordSecretKey         = "postgres-password-secret-key"
)

func PulumiOutputsToStackOutputsConverter(pulumiOutputs auto.OutputMap,
	input *postgreskubernetes.PostgresKubernetesStackInput) *postgreskubernetes.PostgresKubernetesStackOutputs {
	return &postgreskubernetes.PostgresKubernetesStackOutputs{
		Namespace:          autoapistackoutput.GetVal(pulumiOutputs, Namespace),
		Service:            autoapistackoutput.GetVal(pulumiOutputs, Service),
		PortForwardCommand: autoapistackoutput.GetVal(pulumiOutputs, KubePortForwardCommand),
		KubeEndpoint:       autoapistackoutput.GetVal(pulumiOutputs, KubeEndpoint),
		ExternalHostname:   autoapistackoutput.GetVal(pulumiOutputs, IngressExternalHostname),
		InternalHostname:   autoapistackoutput.GetVal(pulumiOutputs, IngressInternalHostname),
		UsernameSecret: &kubernetes.KubernernetesSecretKey{
			Name: autoapistackoutput.GetVal(pulumiOutputs, PostgresUserCredentialsSecretName),
			Key:  autoapistackoutput.GetVal(pulumiOutputs, PostgresUsernameSecretKey),
		},
		PasswordSecret: &kubernetes.KubernernetesSecretKey{
			Name: autoapistackoutput.GetVal(pulumiOutputs, PostgresUserCredentialsSecretName),
			Key:  autoapistackoutput.GetVal(pulumiOutputs, PostgresPasswordSecretKey),
		},
	}
}
