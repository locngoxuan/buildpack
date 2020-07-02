package builder

import (
	"scm.wcs.fortna.com/lngo/sqlbundle"
)

type Sql struct {
}

func (b Sql) PreBuild(ctx BuildContext) error {
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		WorkDir: ctx.WorkDir,
	})
	if err != nil {
		return err
	}
	return bundle.Clean()
}

func (b Sql) Build(ctx BuildContext) error {
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		WorkDir: ctx.WorkDir,
	})
	if err != nil {
		return err
	}
	return bundle.Pack()
}

func (b Sql) PostBuild(ctx BuildContext) error {
	return nil
}

func init() {
	registries["sql"] = &Sql{}
}