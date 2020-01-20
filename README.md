## FORTNA BUILDPACK

#### 1. Usage

```
buildpack [action] [options]

action:
	init : create new buildpack.yml file
	module : add/remove module from buildpack.yml
	snapshot : build then deploy package with label and build number 
	install : build then deploy release package
	
```

#### 2. Options of Action

- ##### INIT
```
options:
	--v : version number of buildpack.yml file
	--m : version number of buildpack.yml file
	
```

- ##### MODULE
```
options:
	--add : add new module from buildpack config
	--remove : remove exist module from buildpack config

```

- ##### SNAPSHOT
```
options:
	--v : version number of buildpack.yml file
    --m : version number of buildpack.yml file
	--container : run build command in container environment
	--phase: list of specific phases will be run
	
	# for git account
	--git-token : access token of git account
	--git-sshpath : path to private key of git account if repo using git instead of http
	--git-sshpass : passphrase of private key of git account
	
	# for jfrog artifactory publisher
	--art-user : username of artifactory account
	--art-pass : pasword of artifactory account
	
	# for docker registry publisher
	--docker-user : username of docker account	
	--docker-pass : pasword of docker account
	
	# for mvn builder
	--m2 : absolute path to .m2 directory. If empty then it will use ${HOME}/.m2/
```

- ##### RELEASE
```
options:
	--v : version number of buildpack.yml file
    --m : version number of buildpack.yml file
	--container : run build command in container environment
	--phase: list of specific phases will be run
	
	# for git account
	--git-token : access token of git account
	--git-sshpath : path to private key of git account if repo using git instead of http
	--git-sshpass : passphrase of private key of git account
	
	# for jfrog artifactory publisher
	--art-user : username of artifactory account
	--art-pass : pasword of artifactory account
	
	# for docker registry publisher
	--docker-user : username of docker account	
	--docker-pass : pasword of docker account
	
	# for mvn builder
	--m2 : absolute path to .m2 directory. If empty then it will use ${HOME}/.m2/
	
```

