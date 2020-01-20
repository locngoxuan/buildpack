package main

type BuilderMvn struct {
	RunFnc  Run
	MvnOpts []string
}

type Run func(arg ...string) error

func (b *BuilderMvn) LoadConfig() error {
	return nil
}

func (b *BuilderMvn) Clean() error {
	arg := make([]string, 0)
	arg = append(arg, "clean")
	arg = append(arg, b.MvnOpts...)
	return b.RunFnc(arg...)
}

func (b *BuilderMvn) Build() error {
	arg := make([]string, 0)
	arg = append(arg, "install")
	arg = append(arg, b.MvnOpts...)
	return b.RunFnc(arg...)
}

func runLocal(arg ...string) error {
	return nil
}

func runIncontainer(arg ...string) error {
	return nil
}
