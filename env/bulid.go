package env

var BuildEnvWhitelist = []string{
	"CNB_STACK_ID",
	"HOSTNAME",
}

func NewBuildEnv(environ []string) *Env {
	return &Env{
		RootDirMap: POSIXBuildEnv,
		vars:       varsFromEnviron(environ, isNotWhitelisted),
	}
}

func isNotWhitelisted(k string) bool {
	for _, wk := range BuildEnvWhitelist {
		if wk == k {
			return false
		}
	}
	for _, wks := range POSIXBuildEnv {
		for _, wk := range wks {
			if wk == k {
				return false
			}
		}
	}
	return true
}

var POSIXBuildEnv = map[string][]string{
	"bin": {
		"PATH",
	},
	"lib": {
		"LD_LIBRARY_PATH",
		"LIBRARY_PATH",
	},
	"include": {
		"CPATH",
	},
	"pkgconfig": {
		"PKG_CONFIG_PATH",
	},
}