package main

import (
	"context"
	"log"
	"net/rpc"
	"os"

	"github.com/hashicorp/go-plugin"
	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

// RPCServer implements the RPC server for the plugin
type RPCServer struct {
	Impl pluginsdk.Plugin
}

func (s *RPCServer) ExecuteCommand(args map[string]interface{}, resp *error) error {
	name, _ := args["name"].(string)
	argsSlice, _ := args["args"].([]interface{})

	log.Printf("[PLUGIN] ExecuteCommand called: %s", name)

	// 初期化確認
	if s.Impl == nil {
		*resp = NewStringError("Plugin not initialized")
		return nil
	}

	// 一時的なホストインターフェース設定
	hostImpl := &SimpleHostInterface{}
	s.Impl.Initialize(context.Background(), hostImpl)

	// コマンド実行
	if cmdPlugin, ok := s.Impl.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
		*resp = cmdPlugin.ExecuteCommand(name, argsSlice...)
	} else {
		*resp = NewStringError("Plugin does not support command execution")
	}

	return nil
}

func (s *RPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *RPCServer) Version(args interface{}, resp *string) error {
	*resp = s.Impl.Version()
	return nil
}

func (s *RPCServer) Description(args interface{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *RPCServer) Initialize(args map[string]interface{}, resp *error) error {
	hostImpl := &SimpleHostInterface{}
	*resp = s.Impl.Initialize(context.Background(), hostImpl)
	return nil
}

func (s *RPCServer) Cleanup(args interface{}, resp *error) error {
	*resp = s.Impl.Cleanup()
	return nil
}

func (s *RPCServer) GetCommands(args interface{}, resp *[]pluginsdk.CommandSpec) error {
	*resp = s.Impl.GetCommands()
	return nil
}

func (s *RPCServer) GetMajorModes(args interface{}, resp *[]pluginsdk.MajorModeSpec) error {
	*resp = s.Impl.GetMajorModes()
	return nil
}

func (s *RPCServer) GetMinorModes(args interface{}, resp *[]pluginsdk.MinorModeSpec) error {
	*resp = s.Impl.GetMinorModes()
	return nil
}

func (s *RPCServer) GetKeyBindings(args interface{}, resp *[]pluginsdk.KeyBindingSpec) error {
	*resp = s.Impl.GetKeyBindings()
	return nil
}

func (s *RPCServer) GetCompletions(args map[string]interface{}, resp *[]string) error {
	command, _ := args["command"].(string)
	prefix, _ := args["prefix"].(string)
	if cmdPlugin, ok := s.Impl.(interface{ GetCompletions(string, string) []string }); ok {
		*resp = cmdPlugin.GetCompletions(command, prefix)
	}
	return nil
}

// CustomRPCPlugin implements the go-plugin interface
type CustomRPCPlugin struct{}

func (p *CustomRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: pluginInstance}, nil
}

func (p *CustomRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return nil, nil
}

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
			"gmacs-plugin": &CustomRPCPlugin{},
		},
	})
}