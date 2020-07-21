### Introduction

Buildpack is created in order to build up an independent and persistent environment for building and publishing application by using standard containers.

Buildpack executes your build as a series of build steps, where each build step is run in a Docker container. A build step can do anything that can be done from a container irrespective of the environment. To perform your tasks, you can either [use the supported build steps](#) provided by Buildpack or [write your own build steps](#).

### Install

For install it from source, following these steps:

```shell
$ git clone git@github.com:locngoxuan/buildpack.git

$ make dev && sudo make install

$ mkdir -p /etc/buildpack
$ mkdir -p /etc/buildpack/plugins
$ mkdir -p /etc/buildpack/plugins/builder
$ mkdir -p /etc/buildpack/plugins/publisher
```



Or download release then install:

```shell
$ wget 'https://github.com/locngoxuan/buildpack/releases/download/2.0.0/buildpack-2.0.0-1.el7.x86_64.rpm'

$ rpm -iUvh buildpack-2.0.0-1.el7.x86_64.rpm
```



### Usage

- ##### Structure

At the beginning, `buildpack` loads structure of project from `Buildpackfile`. After understanding how many modules are there and which one is executed, it walk through them one by one for executing the build. If there are any module with same id, they will be executed in parallel. 

For each module, `Buildpackfile.build` and `Buildpackfile.publish` are needed. They include information of builder and publisher tool that will be used for that module.

``` shell
# mvn example
application/
├── module1/
├── src
├── pom.xml
├── Buildpackfile.build
├── Buildpackfile.publish
└── Buildpackfile

# mvn with parent pom example
application/
├── module1/
│   ├── src
│   ├── pom.xml
│   ├── Buildpackfile.build
│   └── Buildpackfile.publish
├── pom.xml
├── Buildpackfile.build
├── Buildpackfile.publish
└── Buildpackfile

# mix project mvn and nodejs
application/
├── module1/
│   ├── src
│   ├── pom.xml
│   ├── Buildpackfile.build
│   └── Buildpackfile.publish
├── module2/
│   ├── src
│   ├── index.js
│   ├── package.json
│   ├── Buildpackfile.build
│   └── Buildpackfile.publish
├── pom.xml
├── Buildpackfile.build
├── Buildpackfile.publish
└── Buildpackfile
```



- ##### Command

```shell
Usage: buildpack COMMAND [OPTIONS]
COMMAND:
  clean         Clean build folder		
  build         Run build and publish to repository
  version       Show version of buildpack
  help          Show usage

Examples:
  buildpack clean
  buildpack version
  buildpack build --dev-mode
  buildpack build --release
  buildpack build --path --skip-progress

Options:
  -config string
    	specific path to config file
  -dev-mode
    	enable local mode to disable container build
  -increase-version
    	force to increase version after build
  -log-dir string
    	log directory
  -module string
    	list of module
  -no-backward
    	if true, then major version will be increased
  -no-git-tag
    	skip tagging source code
  -patch
    	build for patching
  -release
    	build for releasing
  -share-data string
    	sharing directory
  -skip-clean
    	skip clean everything after build complete
  -skip-container
    	skip container build
  -skip-progress
    	use text plain instead of progress ui
  -skip-publish
    	skip publish build to repository
  -verbose
    	show more detail in console
  -version string
    	version number
```

