package priv

func EnsureOwner(uid, gid int, paths ...string) error {
	return nil
}

func IsPrivileged() bool {
	return false
}

func RunAs(uid, gid int) error {
	return nil
}

func RunAsEffective(uid, gid int) error {
	return nil
}

func SetEnvironmentForUser(uid int) error {
	return nil
}
