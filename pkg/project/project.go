package project

var (
	description = "Watches and mitigates known issues in security platform components."
	gitSHA      = "n/a"
	name        = "security-pack-helper"
	source      = "https://github.com/giantswarm/security-pack-helper"
	version     = "0.0.1-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
