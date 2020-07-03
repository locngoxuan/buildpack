package builder

type YarnLib struct {
}

func (b YarnLib) Clean(ctx BuildContext) error {
	return nil
}

func (b YarnLib) PreBuild(ctx BuildContext) error {
	return nil
}

func (b YarnLib) Build(ctx BuildContext) error {
	return nil
}

func (b YarnLib) PostBuild(ctx BuildContext) error {
	return nil
}

func init() {
	registries["yarn_lib"] = &YarnLib{}
}
