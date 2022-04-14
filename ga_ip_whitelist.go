// Package traefik_github_actions_ip_whitelist_plugin Traefik Github Actions IP Whitelist Plugin.
package traefik_github_actions_ip_whitelist_plugin

import (
	"context"
	"encoding/json"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tcp"
	"net"
	"net/http"
)

const (
	typeName = "GithubActionsIpWhitelist"
)

// Config the plugin configuration.
type Config struct {
	AdditionalCIDRs []string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// GithubActionWhitelist a Github Action IP Whitelist plugin.
type GithubActionWhitelist struct {
	next        tcp.Handler
	name        string
	additionalCIDRs []string
}

// New creates a new Github Action IP Whitelist plugin.
func New(ctx context.Context, next tcp.Handler, config *Config, name string) (tcp.Handler, error) {
	return &GithubActionWhitelist{}, nil
}

type GithubMetaResponse struct {
	Actions []string
}

func (wl *GithubActionWhitelist) ServeTCP(conn tcp.WriteCloser) {
	ctx := middlewares.GetLoggerCtx(context.Background(), wl.name, typeName)
	logger := log.FromContext(ctx)

	addr := conn.RemoteAddr().String()

	req, err := http.NewRequest("GET", "https://api.github.com/meta", nil)
	if err != nil {
		logger.Error("Error preparing request to https://api.github.com/meta")
		conn.Close()
		return
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error connecting to https://api.github.com/meta")
		conn.Close()
		return
	}

	var meta GithubMetaResponse
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		logger.Error("Error parsing response from https://api.github.com/meta")
		conn.Close()
		return
	}

	for _, cidr := range meta.Actions {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Warnf("Skipping CIDR %s", cidr)
		} else if ipnet.Contains(net.ParseIP(addr)) {
			logger.Debugf("Source address originates Github Actions", addr)
			wl.next.ServeTCP(conn)
		}
	}

	for _, cidr := range wl.additionalCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Warnf("Skipping CIDR %s", cidr)
		} else if ipnet.Contains(net.ParseIP(addr)) {
			logger.Debugf("Source address originates from allowed CIDRs", addr)
			wl.next.ServeTCP(conn)
		}
	}

	logger.Warnf("Source address is not allowed %s", addr)
	conn.Close()
	return
}
