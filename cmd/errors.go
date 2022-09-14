package cmd

import "github.com/giantswarm/microerror"

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfig",
}

var invalidProbeModeError = &microerror.Error{
	Kind: "invalidProbeModeError",
}
