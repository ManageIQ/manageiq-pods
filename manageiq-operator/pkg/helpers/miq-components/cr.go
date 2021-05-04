package miqtools

import (
	"context"
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func appName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.AppName == "" {
		return "manageiq"
	} else {
		return cr.Spec.AppName
	}
}

func backupLabelName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.BackupLabelName == "" {
		return "manageiq.org/backup"
	} else {
		return cr.Spec.BackupLabelName
	}
}

func databaseRegion(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.DatabaseRegion == "" {
		return "0"
	} else {
		return cr.Spec.DatabaseRegion
	}
}

func databaseSecret(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.DatabaseSecret == "" {
		return "postgresql-secrets"
	} else {
		return cr.Spec.DatabaseSecret
	}
}

func databaseVolumeCapacity(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.DatabaseVolumeCapacity == "" {
		return "15Gi"
	} else {
		return cr.Spec.DatabaseVolumeCapacity
	}
}

func deployMessagingService(cr *miqv1alpha1.ManageIQ) bool {
	if cr.Spec.DeployMessagingService == nil {
		return false
	} else {
		return *cr.Spec.DeployMessagingService
	}
}

func enableApplicationLocalLogin(cr *miqv1alpha1.ManageIQ) bool {
	if cr.Spec.EnableApplicationLocalLogin == nil {
		return true
	} else {
		return *cr.Spec.EnableApplicationLocalLogin
	}
}

func enableSSO(cr *miqv1alpha1.ManageIQ) bool {
	if cr.Spec.EnableSSO == nil {
		return false
	} else {
		return *cr.Spec.EnableSSO
	}
}

func enforceWorkerResourceConstraints(cr *miqv1alpha1.ManageIQ) bool {
	if cr.Spec.EnforceWorkerResourceConstraints == nil {
		return false
	} else {
		return *cr.Spec.EnforceWorkerResourceConstraints
	}
}

func httpdAuthenticationType(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.HttpdAuthenticationType == "" {
		return "internal"
	} else {
		return cr.Spec.HttpdAuthenticationType
	}
}

func httpdImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.HttpdImage != "" {
		return cr.Spec.HttpdImage
	}

	privileged := PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)
	var image string

	if privileged {
		image = "httpd-init"
	} else {
		image = "httpd"
	}

	return httpdImageNamespace(cr) + "/" + image + ":" + httpdImageTag(cr)
}

func httpdImageNamespace(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.HttpdImageNamespace == "" {
		return "manageiq"
	} else {
		return cr.Spec.HttpdImageNamespace
	}
}

func httpdImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.HttpdImageTag == "" {
		return "latest"
	} else {
		return cr.Spec.HttpdImageTag
	}
}

func imagePullSecretName(cr *miqv1alpha1.ManageIQ, client client.Client) string {
	// If the CR does not have the ImagePullSecret defined, set it to 'image-pull-secret' if a secret with that name exists
	if cr.Spec.ImagePullSecret == "" {
		defaultSecretName := "image-pull-secret"
		secretKey := types.NamespacedName{Namespace: cr.Namespace, Name: defaultSecretName}
		secret := &corev1.Secret{}
		client.Get(context.TODO(), secretKey, secret)

		if secret.Name == defaultSecretName {
			return secret.Name
		}
	}

	return cr.Spec.ImagePullSecret
}

func kafkaImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.KafkaImage == "" {
		return cr.Spec.KafkaImageName + ":" + cr.Spec.KafkaImageTag
	} else {
		return cr.Spec.KafkaImage
	}
}

func kafkaImageName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.KafkaImageName == "" {
		return "docker.io/bitnami/kafka"
	} else {
		return cr.Spec.KafkaImageName
	}
}

func kafkaImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.KafkaImageTag == "" {
		return "latest"
	} else {
		return cr.Spec.KafkaImageTag
	}
}

func kafkaVolumeCapacity(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.KafkaVolumeCapacity == "" {
		return "1Gi"
	} else {
		return cr.Spec.KafkaVolumeCapacity
	}
}

func memcachedImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedImage == "" {
		return cr.Spec.MemcachedImageName + ":" + cr.Spec.MemcachedImageTag
	} else {
		return cr.Spec.MemcachedImage
	}
}

func memcachedImageName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedImageName == "" {
		return "manageiq/memcached"
	} else {
		return cr.Spec.MemcachedImageName
	}
}

func memcachedImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedImageTag == "" {
		return "latest"
	} else {
		return cr.Spec.MemcachedImageTag
	}
}

func memcachedMaxConnection(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedMaxConnection == "" {
		return "1024"
	} else {
		return cr.Spec.MemcachedMaxConnection
	}
}

func memcachedMaxMemory(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedMaxMemory == "" {
		return "64"
	} else {
		return cr.Spec.MemcachedMaxMemory
	}
}

func memcachedSlabPageSize(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.MemcachedSlabPageSize == "" {
		return "1m"
	} else {
		return cr.Spec.MemcachedSlabPageSize
	}
}

func orchestratorImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.OrchestratorImage == "" {
		return cr.Spec.OrchestratorImageNamespace + "/" + cr.Spec.OrchestratorImageName + ":" + cr.Spec.OrchestratorImageTag
	} else {
		return cr.Spec.OrchestratorImage
	}
}

func orchestratorImageName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.OrchestratorImageName == "" {
		return "manageiq-orchestrator"
	} else {
		return cr.Spec.OrchestratorImageName
	}
}

func orchestratorImageNamespace(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.OrchestratorImageNamespace == "" {
		return "manageiq"
	} else {
		return cr.Spec.OrchestratorImageNamespace
	}
}

func orchestratorImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.OrchestratorImageTag == "" {
		return "latest"
	} else {
		return cr.Spec.OrchestratorImageTag
	}
}

func orchestratorInitialDelay(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.OrchestratorInitialDelay == "" {
		return "480"
	} else {
		return cr.Spec.OrchestratorInitialDelay
	}
}

func postgresqlImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.PostgresqlImage == "" {
		return cr.Spec.PostgresqlImageName + ":" + cr.Spec.PostgresqlImageTag
	} else {
		return cr.Spec.PostgresqlImage
	}
}

func postgresqlImageName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.PostgresqlImageName == "" {
		return "docker.io/manageiq/postgresql"
	} else {
		return cr.Spec.PostgresqlImageName
	}
}

func postgresqlImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.PostgresqlImageTag == "" {
		return "10"
	} else {
		return cr.Spec.PostgresqlImageTag
	}
}

func postgresqlMaxConnections(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.PostgresqlMaxConnections == "" {
		return "1000"
	} else {
		return cr.Spec.PostgresqlMaxConnections
	}
}

func postgresqlSharedBuffers(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.PostgresqlSharedBuffers == "" {
		return "1GB"
	} else {
		return cr.Spec.PostgresqlSharedBuffers
	}
}

func serverGuid(cr *miqv1alpha1.ManageIQ, c *client.Client) string {
	if cr.Spec.ServerGuid == "" {
		if pod := orchestratorPod(*c); pod != nil {
			for _, env := range pod.Spec.Containers[0].Env {
				if env.Name == "GUID" {
					return env.Value
				}
			}
		}
		return string(cr.GetUID())
	} else {
		return cr.Spec.ServerGuid
	}
}

func zookeeperImage(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.ZookeeperImage == "" {
		return cr.Spec.ZookeeperImageName + ":" + cr.Spec.ZookeeperImageTag
	} else {
		return cr.Spec.ZookeeperImage
	}
}

func zookeeperImageName(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.ZookeeperImageName == "" {
		return "docker.io/bitnami/zookeeper"
	} else {
		return cr.Spec.ZookeeperImageName
	}
}

func zookeeperImageTag(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.ZookeeperImageTag == "" {
		return "latest"
	} else {
		return cr.Spec.ZookeeperImageTag
	}
}

func zookeeperVolumeCapacity(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.ZookeeperVolumeCapacity == "" {
		return "1Gi"
	} else {
		return cr.Spec.ZookeeperVolumeCapacity
	}
}

func ManageCR(cr *miqv1alpha1.ManageIQ, c *client.Client) (*miqv1alpha1.ManageIQ, controllerutil.MutateFn) {
	f := func() error {
		varDeployMessagingService := deployMessagingService(cr)
		varEnableApplicationLocalLogin := enableApplicationLocalLogin(cr)
		varEnableSSO := enableSSO(cr)
		varEnforceWorkerResourceConstraints := enforceWorkerResourceConstraints(cr)

		cr.Spec.AppName = appName(cr)
		cr.Spec.BackupLabelName = backupLabelName(cr)
		cr.Spec.DatabaseRegion = databaseRegion(cr)
		cr.Spec.DatabaseSecret = databaseSecret(cr)
		cr.Spec.DatabaseVolumeCapacity = databaseVolumeCapacity(cr)
		cr.Spec.DeployMessagingService = &varDeployMessagingService
		cr.Spec.EnableApplicationLocalLogin = &varEnableApplicationLocalLogin
		cr.Spec.EnableSSO = &varEnableSSO
		cr.Spec.EnforceWorkerResourceConstraints = &varEnforceWorkerResourceConstraints
		cr.Spec.HttpdAuthenticationType = httpdAuthenticationType(cr)
		cr.Spec.HttpdImage = httpdImage(cr)
		cr.Spec.ImagePullSecret = imagePullSecretName(cr, *c)
		cr.Spec.KafkaImage = kafkaImage(cr)
		cr.Spec.KafkaImageName = kafkaImageName(cr)
		cr.Spec.KafkaImageTag = kafkaImageTag(cr)
		cr.Spec.KafkaVolumeCapacity = kafkaVolumeCapacity(cr)
		cr.Spec.MemcachedImage = memcachedImage(cr)
		cr.Spec.MemcachedImageName = memcachedImageName(cr)
		cr.Spec.MemcachedImageTag = memcachedImageTag(cr)
		cr.Spec.MemcachedMaxConnection = memcachedMaxConnection(cr)
		cr.Spec.MemcachedMaxMemory = memcachedMaxMemory(cr)
		cr.Spec.MemcachedSlabPageSize = memcachedSlabPageSize(cr)
		cr.Spec.OrchestratorImage = orchestratorImage(cr)
		cr.Spec.OrchestratorImageName = orchestratorImageName(cr)
		cr.Spec.OrchestratorImageNamespace = orchestratorImageNamespace(cr)
		cr.Spec.OrchestratorImageTag = orchestratorImageTag(cr)
		cr.Spec.OrchestratorInitialDelay = orchestratorInitialDelay(cr)
		cr.Spec.PostgresqlImage = postgresqlImage(cr)
		cr.Spec.PostgresqlImageName = postgresqlImageName(cr)
		cr.Spec.PostgresqlImageTag = postgresqlImageTag(cr)
		cr.Spec.PostgresqlMaxConnections = postgresqlMaxConnections(cr)
		cr.Spec.PostgresqlSharedBuffers = postgresqlSharedBuffers(cr)
		cr.Spec.ServerGuid = serverGuid(cr, c)
		cr.Spec.ZookeeperImage = zookeeperImage(cr)
		cr.Spec.ZookeeperImageName = zookeeperImageName(cr)
		cr.Spec.ZookeeperImageTag = zookeeperImageTag(cr)
		cr.Spec.ZookeeperVolumeCapacity = zookeeperVolumeCapacity(cr)

		addBackupLabel(backupLabelName(cr), &cr.ObjectMeta)

		return nil
	}

	return cr, f
}
