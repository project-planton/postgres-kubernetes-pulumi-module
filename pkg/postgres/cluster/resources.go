package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/enums/kubernetesworkloadingresstype"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"

	"github.com/pkg/errors"
	"github.com/plantoncloud-inc/go-commons/kubernetes/manifest"
	plantoncloudk8sv1model "github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubecluster/model"
	"github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/network/ingress/istio/cert"
	pulumik8syaml "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	zalandov1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8sapimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {

}

func Resources(ctx *pulumi.Context) error {
	if err := addPostgresKubernetes(ctx); err != nil {
		return errors.Wrap(err, "failed to add postgres cluster")
	}
	return nil
}

func addPostgresKubernetes(ctx *pulumi.Context) error {
	i := extractInput(ctx)
	postgresKubernetesObject := buildPostgresKubernetesObject(i)
	resourceName := fmt.Sprintf("postgres-cluster-%s", postgresKubernetesObject.Name)
	manifestPath := filepath.Join(i.WorkspaceDir, fmt.Sprintf("%s.yaml", resourceName))
	if err := manifest.Create(manifestPath, postgresKubernetesObject); err != nil {
		return errors.Wrapf(err, "failed to create %s manifest file", manifestPath)
	}
	_, err := pulumik8syaml.NewConfigFile(ctx, resourceName, &pulumik8syaml.ConfigFileArgs{
		File: manifestPath,
	}, pulumi.DependsOn([]pulumi.Resource{i.Namespace}), pulumi.Parent(i.Namespace))
	if err != nil {
		return errors.Wrap(err, "failed to add virtual-service manifest")
	}
	return nil
}

func buildPostgresKubernetesObject(i *input) *zalandov1.Postgresql {
	postgresql := &zalandov1.Postgresql{
		TypeMeta: k8sapimachineryv1.TypeMeta{
			APIVersion: "acid.zalan.do/zalandov1",
			Kind:       "postgresql",
		},
		ObjectMeta: k8sapimachineryv1.ObjectMeta{
			Name:      GetDatabaseName(),
			Namespace: i.NamespaceName,
			Labels:    i.Labels,
		},
		Spec: zalandov1.PostgresSpec{
			PodAnnotations: map[string]string{"postgres-cluster-id": i.ResourceId},
			PostgresqlParam: zalandov1.PostgresqlParam{
				PgVersion: PostgresVersion,
			},
			Volume: zalandov1.Volume{
				Size: i.ContainerSpec.DiskSize,
			},
			Patroni:           zalandov1.Patroni{},
			Resources:         getResources(i.ContainerSpec.Resources),
			TeamID:            TeamId,
			NumberOfInstances: i.ContainerSpec.Replicas,
			AdditionalVolumes: []zalandov1.AdditionalVolume{
				{
					Name:      "stunnel-ca",
					MountPath: StunnelCertMountPath,
					SubPath:   "tls-combined.pem",
					TargetContainers: []string{
						StunnelContainerName.String(),
					},
					VolumeSource: k8scorev1.VolumeSource{
						Secret: &k8scorev1.SecretVolumeSource{
							SecretName: cert.GetCertSecretName(cert.Name),
						},
					},
				},
			},
		},
	}

	addSidecars(i, postgresql)

	return postgresql
}

func addSidecars(i *input, postgresql *zalandov1.Postgresql) error {
	stunnelSidecarImage, isEnvVarSet := os.LookupEnv(EnvVarStunnelSidecarImage)
	if !isEnvVarSet {
		return errors.Errorf("%s environment variables is not set", EnvVarStunnelSidecarImage)
	}
	if !i.IsIngressEnabled ||
		i.IngressType != kubernetesworkloadingresstype.KubernetesWorkloadIngressType_ingress_controller {
		return nil
	}
	postgresql.Spec.Sidecars = []zalandov1.Sidecar{
		{
			Name:        StunnelContainerName.String(),
			Resources:   getStunnelSidecarResources(),
			DockerImage: stunnelSidecarImage,
			Ports: []k8scorev1.ContainerPort{
				{
					Name:          englishword.EnglishWord_postgres.String(),
					ContainerPort: StunnelContainerPort,
					Protocol:      "TCP",
				},
			},
			Env: []k8scorev1.EnvVar{
				{
					Name:  "STUNNEL_MODE",
					Value: englishword.EnglishWord_server.String(),
				}, {
					Name:  "STUNNEL_LOG_LEVEL",
					Value: englishword.EnglishWord_debug.String(),
				}, {
					Name:  "STUNNEL_ACCEPT_PORT",
					Value: strconv.Itoa(StunnelContainerPort),
				}, {
					Name:  "STUNNEL_FORWARD_HOST",
					Value: englishword.EnglishWord_localhost.String(),
				}, {
					Name:  "STUNNEL_FORWARD_PORT",
					Value: strconv.Itoa(PostgresContainerPort),
				},
			},
		},
	}
	return nil
}

func getStunnelSidecarResources() *zalandov1.Resources {
	return &zalandov1.Resources{
		ResourceRequests: zalandov1.ResourceDescription{
			CPU:    "100m",
			Memory: "100Mi",
		},
		ResourceLimits: zalandov1.ResourceDescription{
			CPU:    "500m",
			Memory: "1Gi",
		},
	}
}

func getResources(inputResources *plantoncloudk8sv1model.ContainerResources) *zalandov1.Resources {
	return &zalandov1.Resources{
		ResourceRequests: zalandov1.ResourceDescription{
			CPU:    inputResources.Requests.Cpu,
			Memory: inputResources.Requests.Memory,
		},
		ResourceLimits: zalandov1.ResourceDescription{
			CPU:    inputResources.Limits.Cpu,
			Memory: inputResources.Limits.Memory,
		},
	}
}

// GetDatabaseName returns input for zalando operator for name of the database.
// for zolando operator the name is required to be always prefixed by teamId
// a kubernetes service with the same name is created by the operator
func GetDatabaseName() string {
	return fmt.Sprintf("%s-server", TeamId)
}
