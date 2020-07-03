package builder

type YarnApp struct {
}

func (b YarnApp) Clean(ctx BuildContext) error {
	return nil
}

func (b YarnApp) PreBuild(ctx BuildContext) error {
	return nil
}

func (b YarnApp) Build(ctx BuildContext) error {
	return nil
}

func (b YarnApp) PostBuild(ctx BuildContext) error {
	return nil
}

func init() {
	registries["yarn_app"] = &YarnApp{}
}
