package pkg

import (
	"fmt"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/postgreskubernetes"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/apiresource/enums/apiresourcekind"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-module/pkg/outputs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname string
	IngressInternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	PostgresKubernetes      *postgreskubernetes.PostgresKubernetes
	PostgresPodSectorLabels map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *postgreskubernetes.PostgresKubernetesStackInput) *Locals {
	locals := &Locals{}
	//assign value for the local variable to make it available across the module.
	locals.PostgresKubernetes = stackInput.ApiResource

	postgresKubernetes := stackInput.ApiResource

	//decide on the namespace
	locals.Namespace = postgresKubernetes.Metadata.Id

	ctx.Export(outputs.Namespace, pulumi.String(locals.Namespace))

	locals.PostgresPodSectorLabels = map[string]string{
		"planton.cloud/resource-kind": apiresourcekind.ApiResourceKind_postgres_kubernetes.String(),
		"planton.cloud/resource-id":   postgresKubernetes.Metadata.Id,
	}

	ctx.Export(outputs.PostgresUserCredentialsSecretName, pulumi.Sprintf("postgres.db-%s.credentials.postgresql.acid.zalan.do",
		postgresKubernetes.Metadata.Id))
	ctx.Export(outputs.PostgresUsernameSecretKey, pulumi.String("username"))
	ctx.Export(outputs.PostgresPasswordSecretKey, pulumi.String("password"))

	locals.KubeServiceName = fmt.Sprintf("%s-master", postgresKubernetes.Metadata.Name)

	//export kubernetes service name
	ctx.Export(outputs.Service, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(outputs.KubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(outputs.KubePortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if postgresKubernetes.Spec.Ingress == nil ||
		!postgresKubernetes.Spec.Ingress.IsEnabled ||
		postgresKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return locals
	}

	locals.IngressExternalHostname = fmt.Sprintf("%s.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", postgresKubernetes.Metadata.Id,
		postgresKubernetes.Spec.Ingress.EndpointDomainName)

	//export ingress hostnames
	ctx.Export(outputs.IngressExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(outputs.IngressInternalHostname, pulumi.String(locals.IngressInternalHostname))

	return locals
}
