package locals

import (
	"fmt"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/postgreskubernetes/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/outputs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	IngressExternalHostname string
	IngressInternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	PostgresKubernetes      *model.PostgresKubernetes
	PostgresPodSectorLabels map[string]string
)

func Initializer(ctx *pulumi.Context, stackInput *model.PostgresKubernetesStackInput) {
	//assign value for the local variable to make it available across the module.
	PostgresKubernetes = stackInput.ApiResource

	postgresKubernetes := stackInput.ApiResource

	//decide on the namespace
	Namespace = postgresKubernetes.Metadata.Id

	PostgresPodSectorLabels = map[string]string{
		"app.kubernetes.io/component": "master",
		"app.kubernetes.io/instance":  postgresKubernetes.Metadata.Id,
		"app.kubernetes.io/name":      "postgres",
	}

	KubeServiceName = fmt.Sprintf("%s-master", postgresKubernetes.Metadata.Name)

	//export kubernetes service name
	ctx.Export(outputs.Service, pulumi.String(KubeServiceName))

	KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local.", KubeServiceName, Namespace)

	//export kubernetes endpoint
	ctx.Export(outputs.KubeEndpoint, pulumi.String(KubeServiceFqdn))

	KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		Namespace, KubeServiceName)

	//export kube-port-forward command
	ctx.Export(outputs.KubePortForwardCommand, pulumi.String(KubePortForwardCommand))

	if postgresKubernetes.Spec.Ingress == nil ||
		!postgresKubernetes.Spec.Ingress.IsEnabled ||
		postgresKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return
	}

	IngressExternalHostname = fmt.Sprintf("%s.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	IngressInternalHostname = fmt.Sprintf("%s-internal.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	//export ingress hostnames
	ctx.Export(outputs.IngressExternalHostname, pulumi.String(IngressExternalHostname))
	ctx.Export(outputs.IngressInternalHostname, pulumi.String(IngressInternalHostname))
}
