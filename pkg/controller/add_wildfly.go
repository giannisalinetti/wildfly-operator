package controller

import (
	"github.com/giannisalinetti/wildfly-operator/pkg/controller/wildfly"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, wildfly.Add)
}
