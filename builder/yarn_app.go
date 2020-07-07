package builder

type YarnApp struct {
	Yarn
}

func init() {
	registries["yarn_app"] = &YarnApp{}
}
