package builder

const (
	noBuildTool = "nobuild"
)

type NoBuild struct {
}

func (n *NoBuild) Name() string {
	return noBuildTool
}
func (n *NoBuild) GenerateConfig(ctx BuildContext) error {
	return nil
}
func (n *NoBuild) LoadConfig(ctx BuildContext) error {
	return nil
}
func (n *NoBuild) Clean(ctx BuildContext) error {
	return nil
}
func (n *NoBuild) PreBuild(ctx BuildContext) error {
	return nil
}
func (n *NoBuild) Build(ctx BuildContext) error {
	return nil
}
func (n *NoBuild) PostBuild(ctx BuildContext) error {
	return nil
}
