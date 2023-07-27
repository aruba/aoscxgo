package aoscxgo

import "bitbucket.org/HelgeOlav/utils/version"

const Version = "0.0.2"

func init() {
	version.AddModule(version.ModuleVersion{
		Name:    "aoscxgo",
		Version: Version,
	})
}
