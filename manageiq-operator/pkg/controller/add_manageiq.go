package controller

import (
	"github.com/manageiq/manageiq-pods/manageiq-operator/pkg/controller/manageiq"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, manageiq.Add)
}
