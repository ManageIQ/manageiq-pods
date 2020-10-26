package miqtools

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
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

func databaseRegion(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.DatabaseRegion == "" {
		return "0"
	} else {
		return cr.Spec.DatabaseRegion
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

func httpdAuthenticationType(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.HttpdAuthenticationType == "" {
		return "internal"
	} else {
		return cr.Spec.HttpdAuthenticationType
	}
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
		return "latest-kasparov"
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

func serverGuid(cr *miqv1alpha1.ManageIQ) string {
	if cr.Spec.ServerGuid == "" {
		return string(cr.GetUID())
	} else {
		return cr.Spec.ServerGuid
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

func ManageCR(cr *miqv1alpha1.ManageIQ) (*miqv1alpha1.ManageIQ, controllerutil.MutateFn) {
	f := func() error {
		varDeployMessagingService := deployMessagingService(cr)
		varEnableApplicationLocalLogin := enableApplicationLocalLogin(cr)
		varEnableSSO := enableSSO(cr)
		varEnforceWorkerResourceConstraints := enforceWorkerResourceConstraints(cr)

		cr.Spec.AppName = appName(cr)
		cr.Spec.BackupLabelName = backupLabelName(cr)
		cr.Spec.DatabaseRegion = databaseRegion(cr)
		cr.Spec.DatabaseVolumeCapacity = databaseVolumeCapacity(cr)
		cr.Spec.DeployMessagingService = &varDeployMessagingService
		cr.Spec.EnableApplicationLocalLogin = &varEnableApplicationLocalLogin
		cr.Spec.EnableSSO = &varEnableSSO
		cr.Spec.EnforceWorkerResourceConstraints = &varEnforceWorkerResourceConstraints
		cr.Spec.HttpdAuthenticationType = httpdAuthenticationType(cr)
		cr.Spec.HttpdImageNamespace = httpdImageNamespace(cr)
		cr.Spec.HttpdImageTag = httpdImageTag(cr)
		cr.Spec.KafkaImageName = kafkaImageName(cr)
		cr.Spec.KafkaImageTag = kafkaImageTag(cr)
		cr.Spec.KafkaVolumeCapacity = kafkaVolumeCapacity(cr)
		cr.Spec.MemcachedImageName = memcachedImageName(cr)
		cr.Spec.MemcachedImageTag = memcachedImageTag(cr)
		cr.Spec.MemcachedMaxConnection = memcachedMaxConnection(cr)
		cr.Spec.MemcachedMaxMemory = memcachedMaxMemory(cr)
		cr.Spec.MemcachedSlabPageSize = memcachedSlabPageSize(cr)
		cr.Spec.OrchestratorImageName = orchestratorImageName(cr)
		cr.Spec.OrchestratorImageNamespace = orchestratorImageNamespace(cr)
		cr.Spec.OrchestratorImageTag = orchestratorImageTag(cr)
		cr.Spec.OrchestratorInitialDelay = orchestratorInitialDelay(cr)
		cr.Spec.PostgresqlImageName = postgresqlImageName(cr)
		cr.Spec.PostgresqlImageTag = postgresqlImageTag(cr)
		cr.Spec.PostgresqlMaxConnections = postgresqlMaxConnections(cr)
		cr.Spec.PostgresqlSharedBuffers = postgresqlSharedBuffers(cr)
		cr.Spec.ServerGuid = serverGuid(cr)
		cr.Spec.ZookeeperImageName = zookeeperImageName(cr)
		cr.Spec.ZookeeperImageTag = zookeeperImageTag(cr)
		cr.Spec.ZookeeperVolumeCapacity = zookeeperVolumeCapacity(cr)

		addBackupLabel(backupLabelName(cr), &cr.ObjectMeta)

		return nil
	}

	return cr, f
}
