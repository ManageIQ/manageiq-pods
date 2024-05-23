package cr_migration

import (
	"context"

	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	miqutilsv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1/miqutils"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func migrate20240508124600(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) *miqv1alpha1.ManageIQ {
	migrationId := "20240508124600"
	for _, migration := range cr.Spec.MigrationsRan {
		if migration == migrationId {
			return cr
		}
	}

	if topic := miqutilsv1alpha1.FindKafkaTopic(client, scheme, cr.Namespace, "manageiq.liveness-check", "kafka.strimzi.io"); topic != nil {
		client.Delete(context.TODO(), topic)
	}

	cr.Spec.MigrationsRan = append(cr.Spec.MigrationsRan, migrationId)

	return cr
}
