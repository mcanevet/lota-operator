package controller

import "github.com/mcanevet/lota-operator/pkg/controller/lotaresource"

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, lotaresource.Add)
}
