package builder

type MvnLib struct {
}

func (b MvnLib) PreBuild(ctx BuildContext) error {
	return nil
}

func (b MvnLib) Build(ctx BuildContext) error {
	return nil
}

func (b MvnLib) PostBuild(ctx BuildContext) error {
	return nil
}

func init() {
	registries["mvn_lib"] = &MvnLib{}
}
