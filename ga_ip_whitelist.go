// Package traefik_github_actions_ip_whitelist_plugin Traefik Github Actions IP Whitelist Plugin.
package traefik_github_actions_ip_whitelist_plugin

import (
	"context"
	"encoding/json"
	"log"
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
	next        TcpHandler
	name        string
	additionalCIDRs []string
}

// New creates a new Github Action IP Whitelist plugin.
func New(ctx context.Context, next TcpHandler, config *Config, name string) (TcpHandler, error) {
	return &GithubActionWhitelist{}, nil
}

// TcpHandler is the TCP Handlers interface.
type TcpHandler interface {
	ServeTCP(conn TcpWriteCloser)
}

// TcpWriteCloser describes a net.Conn with a CloseWrite method.
type TcpWriteCloser interface {
	net.Conn
	// CloseWrite on a network connection, indicates that the issuer of the call
	// has terminated sending on that connection.
	// It corresponds to sending a FIN packet.
	CloseWrite() error
}

type GithubMetaResponse struct {
	Actions []string
}

func (wl *GithubActionWhitelist) ServeTCP(conn TcpWriteCloser) {
	addr := conn.RemoteAddr().String()

	req, err := http.NewRequest("GET", "https://api.github.com/meta", nil)
	if err != nil {
		log.Print("Error preparing request to https://api.github.com/meta")
		conn.Close()
		return
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Print("Error connecting to https://api.github.com/meta")
		conn.Close()
		return
	}

	var meta GithubMetaResponse
	err = json.NewDecoder(resp.Body).Decode(&meta)
	if err != nil {
		log.Print("Error parsing response from https://api.github.com/meta")
		conn.Close()
		return
	}

	for _, cidr := range meta.Actions {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Skipping CIDR %s", cidr)
		} else if ipnet.Contains(net.ParseIP(addr)) {
			log.Printf("Source address originates Github Actions %s", addr)
			wl.next.ServeTCP(conn)
		}
	}

	for _, cidr := range wl.additionalCIDRs {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("Skipping CIDR %s", cidr)
		} else if ipnet.Contains(net.ParseIP(addr)) {
			log.Printf("Source address originates from allowed CIDRs %s", addr)
			wl.next.ServeTCP(conn)
		}
	}

	log.Printf("Source address is not allowed %s", addr)
	conn.Close()
	return
}
