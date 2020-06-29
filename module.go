package buildpack

type Module struct {
	Id   int
	Name string
	Path string
}

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
