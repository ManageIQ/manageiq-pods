package cr_migration

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	miqtool "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/helpers/miq-components"
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


	cr.Spec.MigrationsRan = append(cr.Spec.MigrationsRan, migrationId)

	return cr
}
