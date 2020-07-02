package buildpack

type Module struct {
	Id   int
	Name string
	Path string
}

type SortedById []Module

func (a SortedById) Len() int           { return len(a) }
func (a SortedById) Less(i, j int) bool { return a[i].Id < a[j].Id }
func (a SortedById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (m *Module) start() {
	/**
	1. Read configuration
		- Read Buildpackfile.build
		- Read Buildoackfile.publish
	2. Clean
		- Clean result of build
		- Clean result of publish
		- Clean .buildpack/{module-name}
	3. Build
		- Pre build
		- Build
		- Post build
	4. Publish
		- Pre publish
		- Publish
		- Post publish
	5. Clean (Allow skip)
		- Clean result of build
		- Clean result of publish
		- Clean .buildpack/{module-name}
	 */
}
