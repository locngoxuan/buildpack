package builder

type MvnApp struct {
}

func (b MvnApp) PreBuild(ctx BuildContext) error {
	return nil
}

func (b MvnApp) Build(ctx BuildContext) error {
	return nil
}

func (b MvnApp) PostBuild(ctx BuildContext) error {
	return nil
}

func init() {
	registries["mvn_app"] = &MvnApp{}
}
