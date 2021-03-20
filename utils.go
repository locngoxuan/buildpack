package buildpack

import (
	"github.com/locngoxuan/buildpack/config"
	"github.com/locngoxuan/buildpack/core"
)

func aggregateDockerConfigInfo(global config.DockerGlobalConfig) ([]string, []config.DockerRegistry) {
	if arg.BuildLocal {
		return []string{}, []config.DockerRegistry{}
	}
	hostSet := make(map[string]struct{})
	hostSet[core.DefaultDockerUnixSock] = struct{}{}
	hostSet[core.DefaultDockerTCPSock] = struct{}{}
	if len(global.Hosts) > 0 {
		for _, host := range global.Hosts {
			hostSet[host] = struct{}{}
		}
	}
	if len(cfg.DockerConfig.Hosts) > 0 {
		for _, host := range cfg.DockerConfig.Hosts {
			hostSet[host] = struct{}{}
		}
	}

	registryMap := make(map[string]config.DockerRegistry)

	if len(global.Registries) > 0 {
		for _, registry := range global.Registries {
			registryMap[registry.Address] = registry
		}
	}
	if len(cfg.DockerConfig.Registries) > 0 {
		for _, registry := range cfg.DockerConfig.Registries {
			registryMap[registry.Address] = registry
		}
	}

	hosts := make([]string, 0)
	for host := range hostSet {
		hosts = append(hosts, host)
	}

	registries := make([]config.DockerRegistry, 0)
	registries = append(registries, core.DefaultDockerHubRegistry)
	for _, registry := range registryMap {
		registries = append(registries, registry)
	}
	return hosts, registries
}
