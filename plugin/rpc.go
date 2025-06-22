package plugin

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

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