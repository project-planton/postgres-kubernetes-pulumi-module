// Package service adds a kubernetes service which is required to forward traffic from istio pods to stunnel sidecar containers running alongside postgres pods.
package stunnel

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/english/enums/englishword"
	postgrescluster "github.com/plantoncloud/postgres-kubernetes-pulumi-blueprint/pkg/postgres/cluster"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	pulk8scv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	StunnelServiceName = "stunnel"
)

type Input struct {
	Namespace *pulk8scv1.Namespace
}

func Resources(ctx *pulumi.Context, input *Input) error {
	if _, err := addService(ctx, input); err != nil {
		return errors.Wrap(err, "failed to add stunnel service")
	}
	return nil
}

/*
apiVersion: v1
kind: Service
metadata:

	name: stunnel
	namespace: planton-pcs-dev-postgres-apr

spec:

	type: ClusterIP
	ports:
	- name: postgresql
	  port: 5432
	  protocol: TCP
	  targetPort: 15432
	selector:
	  application: spilo
	  cluster-name: pcs-apr
	  team: pcs
*/
func addService(ctx *pulumi.Context, input *Input) (*corev1.Service, error) {
	svc, err := corev1.NewService(ctx, StunnelServiceName, &corev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(StunnelServiceName),
			Namespace: input.Namespace.Metadata.Name(),
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Selector: pulumi.StringMap{
				//postgres-operator generated labels for the postgres pod
				englishword.EnglishWord_application.String(): pulumi.String(englishword.EnglishWord_spilo.String()),
				englishword.EnglishWord_team.String():        pulumi.String(postgrescluster.TeamId),
				"cluster-name":                               pulumi.String(postgrescluster.GetDatabaseName()),
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:       pulumi.String(englishword.EnglishWord_postgres.String()),
					Protocol:   pulumi.String("TCP"),
					Port:       pulumi.Int(postgrescluster.PostgresContainerPort),
					TargetPort: pulumi.Int(postgrescluster.StunnelContainerPort),
				},
			},
		},
	}, pulumi.Parent(input.Namespace))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to add service")
	}
	return svc, nil
}
