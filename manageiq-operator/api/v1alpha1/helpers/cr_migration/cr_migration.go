package cr_migration

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Migrate(cr *miqv1alpha1.ManageIQ, client client.Client, scheme *runtime.Scheme) (*miqv1alpha1.ManageIQ, controllerutil.MutateFn) {
	f := func() error {
		cr = migrate20210503163000(cr)
		cr = migrate20210504113000(cr)
		cr = migrate20240508124600(cr, client, scheme)

		return nil
	}

	return cr, f
}
