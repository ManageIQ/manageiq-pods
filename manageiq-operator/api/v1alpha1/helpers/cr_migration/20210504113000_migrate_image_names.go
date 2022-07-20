package cr_migration

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/helpers/miq-components"
)

func migrate20210504113000(cr *miqv1alpha1.ManageIQ) *miqv1alpha1.ManageIQ {
	migrationId := "20210504113000"
	for _, migration := range cr.Spec.MigrationsRan {
		if migration == migrationId {
			return cr
		}
	}

	// Prefer HttpdImage rather than HttpdImageNamespace and HttpdImageTag
	if cr.Spec.HttpdImage == "" && cr.Spec.HttpdImageNamespace != "" && cr.Spec.HttpdImageTag != "" {
		privileged := miqtool.PrivilegedHttpd(cr.Spec.HttpdAuthenticationType)
		var image string

		if privileged {
			image = "httpd-init"
		} else {
			image = "httpd"
		}

		cr.Spec.HttpdImage = cr.Spec.HttpdImageNamespace + "/" + image + ":" + cr.Spec.HttpdImageTag
	}

	cr.Spec.HttpdImageNamespace = ""
	cr.Spec.HttpdImageTag = ""

	// Prefer KafkaImage rather than KafkaImageName and KafkaImageTag
	if cr.Spec.KafkaImage == "" && cr.Spec.KafkaImageName != "" && cr.Spec.KafkaImageTag != "" {
		cr.Spec.KafkaImage = cr.Spec.KafkaImageName + ":" + cr.Spec.KafkaImageTag
	}
	cr.Spec.KafkaImageName = ""
	cr.Spec.KafkaImageTag = ""

	// Prefer MemcachedImage rather than MemcachedImageName and MemcachedImageTag
	if cr.Spec.MemcachedImage == "" && cr.Spec.MemcachedImageName != "" && cr.Spec.MemcachedImageTag != "" {
		cr.Spec.MemcachedImage = cr.Spec.MemcachedImageName + ":" + cr.Spec.MemcachedImageTag
	}
	cr.Spec.MemcachedImageName = ""
	cr.Spec.MemcachedImageTag = ""

	// Prefer OrchestratorImage rather than OrchestratorImageNamespace, OrchestratorImageName and OrchestratorImageTag
	if cr.Spec.OrchestratorImage == "" && cr.Spec.OrchestratorImageNamespace != "" && cr.Spec.OrchestratorImageName != "" && cr.Spec.OrchestratorImageTag != "" {
		cr.Spec.OrchestratorImage = cr.Spec.OrchestratorImageNamespace + "/" + cr.Spec.OrchestratorImageName + ":" + cr.Spec.OrchestratorImageTag
	}
	cr.Spec.OrchestratorImageName = ""
	cr.Spec.OrchestratorImageNamespace = ""
	cr.Spec.OrchestratorImageTag = ""

	// Prefer PostgresqlImage rather than PostgresqlImageName and PostgresqlImageTag
	if cr.Spec.PostgresqlImage == "" && cr.Spec.PostgresqlImageName != "" && cr.Spec.PostgresqlImageTag != "" {
		cr.Spec.PostgresqlImage = cr.Spec.PostgresqlImageName + ":" + cr.Spec.PostgresqlImageTag
	}
	cr.Spec.PostgresqlImageName = ""
	cr.Spec.PostgresqlImageTag = ""

	// Prefer ZookeeperImage rather than ZookeeperImageName and ZookeeperImageTag
	if cr.Spec.ZookeeperImage == "" && cr.Spec.ZookeeperImageName != "" && cr.Spec.ZookeeperImageTag != "" {
		cr.Spec.ZookeeperImage = cr.Spec.ZookeeperImageName + ":" + cr.Spec.ZookeeperImageTag
	}
	cr.Spec.ZookeeperImageName = ""
	cr.Spec.ZookeeperImageTag = ""

	cr.Spec.MigrationsRan = append(cr.Spec.MigrationsRan, migrationId)

	return cr
}
