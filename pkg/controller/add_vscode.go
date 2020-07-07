package controller

import (
	"tmax.io/l2c-operator/pkg/controller/vscode"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, vscode.Add)
}
