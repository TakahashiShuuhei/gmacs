package plugin

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// StringError matches the plugin's StringError type for gob serialization
type StringError struct {
	Message string
}

func (se StringError) Error() string {
	return se.Message
}

func init() {
	// Register StringError with gob for RPC serialization
	// Use the same name as the plugin side to avoid name conflicts
	gob.RegisterName("main.StringError", StringError{})
}

// GRPCPluginはHashiCorp go-pluginのGRPCプラグイン実装
// 後でprotobufが利用可能になったら、gRPCに変更する予定
type GRPCPlugin struct {
	Impl Plugin
}

func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// TODO: gRPCサーバー実装（protobuf生成後）
	return nil
}

func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	// TODO: gRPCクライアント実装（protobuf生成後）
	return nil, nil
}

// 現在はRPCベースで実装（後でgRPCに移行）
type RPCPlugin struct {
	Impl Plugin
}

func (p *RPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (p *RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// RPCServer はプラグイン側のRPCサーバー
type RPCServer struct {
	Impl Plugin
}

// RPCClient はホスト側のRPCクライアント
type RPCClient struct {
	client *rpc.Client
}

// Plugin インターフェースの実装（RPCClient）
func (c *RPCClient) Name() string {
	var resp string
	err := c.client.Call("Plugin.Name", new(interface{}), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Version() string {
	var resp string
	err := c.client.Call("Plugin.Version", new(interface{}), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Description() string {
	var resp string
	err := c.client.Call("Plugin.Description", new(interface{}), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Initialize(ctx context.Context, host HostInterface) error {
	// TODO: HostInterfaceの適切な渡し方を実装
	var resp error
	err := c.client.Call("Plugin.Initialize", map[string]interface{}{}, &resp)
	return err
}

func (c *RPCClient) Cleanup() error {
	var resp error
	err := c.client.Call("Plugin.Cleanup", new(interface{}), &resp)
	return err
}

func (c *RPCClient) GetCommands() []CommandSpec {
	var resp []CommandSpec
	err := c.client.Call("Plugin.GetCommands", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetMajorModes() []MajorModeSpec {
	var resp []MajorModeSpec
	err := c.client.Call("Plugin.GetMajorModes", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetMinorModes() []MinorModeSpec {
	var resp []MinorModeSpec
	err := c.client.Call("Plugin.GetMinorModes", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetKeyBindings() []KeyBindingSpec {
	var resp []KeyBindingSpec
	err := c.client.Call("Plugin.GetKeyBindings", new(interface{}), &resp)
	if err != nil {
		return nil
	}
	return resp
}

// CommandPlugin interface implementation
func (c *RPCClient) ExecuteCommand(name string, args ...interface{}) error {
	// Convert args to a simpler structure for gob encoding
	argsStrings := make([]string, len(args))
	for i, arg := range args {
		argsStrings[i] = fmt.Sprintf("%v", arg)
	}
	
	request := map[string]interface{}{
		"name": name,
		"args": argsStrings, // Use string slice instead of []interface{}
	}
	var resp error
	err := c.client.Call("Plugin.ExecuteCommand", request, &resp)
	if err != nil {
		// Check if it's a "method not found" error
		if err.Error() == "rpc: can't find method Plugin.ExecuteCommand" {
			return fmt.Errorf("plugin does not support ExecuteCommand (plugin needs to be updated to implement CommandPlugin interface)")
		}
		return fmt.Errorf("RPC call failed: %v", err)
	}
	return resp
}

func (c *RPCClient) GetCompletions(command string, prefix string) []string {
	request := map[string]interface{}{
		"command": command,
		"prefix":  prefix,
	}
	var resp []string
	err := c.client.Call("Plugin.GetCompletions", request, &resp)
	if err != nil {
		return nil
	}
	return resp
}

// RPCサーバー側のメソッド実装
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
	// TODO: 適切なHostInterface実装
	*resp = s.Impl.Initialize(context.Background(), nil)
	return nil
}

func (s *RPCServer) Cleanup(args interface{}, resp *error) error {
	*resp = s.Impl.Cleanup()
	return nil
}

func (s *RPCServer) GetCommands(args interface{}, resp *[]CommandSpec) error {
	*resp = s.Impl.GetCommands()
	return nil
}

func (s *RPCServer) GetMajorModes(args interface{}, resp *[]MajorModeSpec) error {
	*resp = s.Impl.GetMajorModes()
	return nil
}

func (s *RPCServer) GetMinorModes(args interface{}, resp *[]MinorModeSpec) error {
	*resp = s.Impl.GetMinorModes()
	return nil
}

func (s *RPCServer) GetKeyBindings(args interface{}, resp *[]KeyBindingSpec) error {
	*resp = s.Impl.GetKeyBindings()
	return nil
}

// CommandPlugin RPC server methods
func (s *RPCServer) ExecuteCommand(args map[string]interface{}, resp *error) error {
	if cmdPlugin, ok := s.Impl.(CommandPlugin); ok {
		name, _ := args["name"].(string)
		argsStrings, _ := args["args"].([]string)
		
		// Convert string slice back to []interface{}
		pluginArgs := make([]interface{}, len(argsStrings))
		for i, arg := range argsStrings {
			pluginArgs[i] = arg
		}
		
		*resp = cmdPlugin.ExecuteCommand(name, pluginArgs...)
	} else {
		*resp = fmt.Errorf("plugin does not implement CommandPlugin interface")
	}
	return nil
}

func (s *RPCServer) GetCompletions(args map[string]interface{}, resp *[]string) error {
	if cmdPlugin, ok := s.Impl.(CommandPlugin); ok {
		command, _ := args["command"].(string)
		prefix, _ := args["prefix"].(string)
		*resp = cmdPlugin.GetCompletions(command, prefix)
	} else {
		*resp = []string{}
	}
	return nil
}

// プラグインマップ定義
var PluginMap = map[string]plugin.Plugin{
	"gmacs-plugin": &RPCPlugin{},
}

// Handshake はプラグインとホスト間のハンドシェイク設定
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GMACS_PLUGIN",
	MagicCookieValue: "gmacs-plugin-magic-cookie",
}