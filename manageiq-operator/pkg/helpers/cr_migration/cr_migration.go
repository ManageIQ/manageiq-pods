package cr_migration

import (
	miqv1alpha1 "github.com/ManageIQ/manageiq-pods/manageiq-operator/pkg/apis/manageiq/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Migrate(cr *miqv1alpha1.ManageIQ) (*miqv1alpha1.ManageIQ, controllerutil.MutateFn) {
	f := func() error {
		cr = migrate20210503163000(cr)

		return nil
	}

	return cr, f
}
