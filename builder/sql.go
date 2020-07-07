package builder

import (
	"scm.wcs.fortna.com/lngo/sqlbundle"
)

type Sql struct {
}

func (b Sql) Clean(ctx BuildContext) error {
	sqlbundle.SetLogWriter(ctx.LogWriter)
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		Version: ctx.Version,
		WorkDir: ctx.WorkDir,
	})
	if err != nil {
		return err
	}
	return bundle.Clean()
}

func (b Sql) PreBuild(ctx BuildContext) error {
	return nil
}

func (b Sql) PostFail(ctx BuildContext) error {
	return b.Clean(ctx)
}

func (b Sql) Build(ctx BuildContext) error {
	sqlbundle.SetLogWriter(ctx.LogWriter)
	bundle, err := sqlbundle.NewSQLBundle(sqlbundle.Argument{
		Version: ctx.Version,
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
