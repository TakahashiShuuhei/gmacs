package main

import (
	"log"
	"os"

	"github.com/hashicorp/go-plugin"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[FILE-TEST-PLUGIN] ")
	log.Printf("Starting file test plugin...")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GMACS_PLUGIN",
			MagicCookieValue: "gmacs-plugin-magic-cookie",
		},
		Plugins: map[string]plugin.Plugin{
			"gmacs-plugin": &RPCPlugin{Impl: pluginInstance},
		},
	})
}