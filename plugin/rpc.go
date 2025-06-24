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

// BufferInfo represents buffer state for RPC transmission
type BufferInfo struct {
	Name     string
	Content  string
	Position int
	IsDirty  bool
	Filename string
}

// RPCBufferProxy provides a client-side proxy for buffer operations via RPC
type RPCBufferProxy struct {
	client *rpc.Client
	info   BufferInfo
}

func (b *RPCBufferProxy) Name() string           { return b.info.Name }
func (b *RPCBufferProxy) Content() string        { return b.info.Content }
func (b *RPCBufferProxy) CursorPosition() int    { return b.info.Position }
func (b *RPCBufferProxy) IsDirty() bool          { return b.info.IsDirty }
func (b *RPCBufferProxy) Filename() string       { return b.info.Filename }

func (b *RPCBufferProxy) SetContent(content string) {
	b.info.Content = content
	// TODO: Implement RPC call to sync content to host
}

func (b *RPCBufferProxy) InsertAt(pos int, text string) {
	// TODO: Implement RPC call to insert text at position
}

func (b *RPCBufferProxy) DeleteRange(start, end int) {
	// TODO: Implement RPC call to delete text range
}

func (b *RPCBufferProxy) SetCursorPosition(pos int) {
	b.info.Position = pos
	// TODO: Implement RPC call to sync cursor position to host
}

func (b *RPCBufferProxy) MarkDirty() {
	b.info.IsDirty = true
	// TODO: Implement RPC call to mark buffer dirty on host
}

func init() {
	// Register StringError with gob for RPC serialization
	// Use the same name as the plugin side to avoid name conflicts
	gob.RegisterName("main.StringError", StringError{})
	
	// Register BufferInfo for RPC serialization
	gob.Register(BufferInfo{})
}

// RPCHostServer はgmacs側でホスト機能をRPC経由で提供するサーバー
type RPCHostServer struct {
	Impl HostInterface
}

// ShowMessage handles RPC calls from plugins to show messages
func (h *RPCHostServer) ShowMessage(message string, resp *error) error {
	h.Impl.ShowMessage(message)
	*resp = nil
	return nil
}

// SetStatus handles RPC calls from plugins to set status
func (h *RPCHostServer) SetStatus(message string, resp *error) error {
	h.Impl.SetStatus(message)
	*resp = nil
	return nil
}

// CreateBuffer handles RPC calls from plugins to create buffers
func (h *RPCHostServer) CreateBuffer(name string, resp *BufferInfo) error {
	buffer := h.Impl.CreateBuffer(name)
	if buffer == nil {
		*resp = BufferInfo{}
		return fmt.Errorf("failed to create buffer")
	}
	
	// Return buffer information via RPC
	*resp = BufferInfo{
		Name:     buffer.Name(),
		Content:  buffer.Content(),
		Position: buffer.CursorPosition(),
		IsDirty:  buffer.IsDirty(),
		Filename: buffer.Filename(),
	}
	return nil
}

// GetCurrentBuffer handles RPC calls from plugins to get current buffer
func (h *RPCHostServer) GetCurrentBuffer(args struct{}, resp *BufferInfo) error {
	fmt.Printf("[RPC-Host] GetCurrentBuffer called\n")
	buffer := h.Impl.GetCurrentBuffer()
	if buffer == nil {
		fmt.Printf("[RPC-Host] GetCurrentBuffer: no current buffer found\n")
		*resp = BufferInfo{}
		return fmt.Errorf("no current buffer")
	}
	
	fmt.Printf("[RPC-Host] GetCurrentBuffer: found buffer '%s'\n", buffer.Name())
	
	// Return buffer information via RPC
	*resp = BufferInfo{
		Name:     buffer.Name(),
		Content:  buffer.Content(),
		Position: buffer.CursorPosition(),
		IsDirty:  buffer.IsDirty(),
		Filename: buffer.Filename(),
	}
	
	fmt.Printf("[RPC-Host] GetCurrentBuffer: returning buffer info: %+v\n", *resp)
	return nil
}

// SwitchToBuffer handles RPC calls from plugins to switch buffers
func (h *RPCHostServer) SwitchToBuffer(name string, resp *error) error {
	*resp = h.Impl.SwitchToBuffer(name)
	return nil
}

// OpenFile handles RPC calls from plugins to open files
func (h *RPCHostServer) OpenFile(path string, resp *error) error {
	fmt.Printf("[RPC-Host] OpenFile called with path: %s\n", path)
	*resp = h.Impl.OpenFile(path)
	if *resp != nil {
		fmt.Printf("[RPC-Host] OpenFile failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Host] OpenFile succeeded for: %s\n", path)
	}
	return nil
}

// SaveBuffer handles RPC calls from plugins to save buffers
func (h *RPCHostServer) SaveBuffer(bufferName string, resp *error) error {
	fmt.Printf("[RPC-Host] SaveBuffer called with buffer: %s\n", bufferName)
	*resp = h.Impl.SaveBuffer(bufferName)
	if *resp != nil {
		fmt.Printf("[RPC-Host] SaveBuffer failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Host] SaveBuffer succeeded for: %s\n", bufferName)
	}
	return nil
}

// RPCHostClient はプラグイン側でホストの機能をRPC経由で呼び出すクライアント
type RPCHostClient struct {
	client *rpc.Client
}

// HostInterface implementation for RPC client
func (h *RPCHostClient) GetCurrentBuffer() BufferInterface {
	var resp BufferInfo
	err := h.client.Call("Host.GetCurrentBuffer", struct{}{}, &resp)
	if err != nil {
		fmt.Printf("[RPC] GetCurrentBuffer call failed: %v\n", err)
		return nil
	}
	
	return &RPCBufferProxy{
		client: h.client,
		info:   resp,
	}
}

func (h *RPCHostClient) GetCurrentWindow() WindowInterface {
	// TODO: Implement RPC call to host
	return nil
}

func (h *RPCHostClient) SetStatus(message string) {
	// TODO: Implement RPC call to host
}

func (h *RPCHostClient) ShowMessage(message string) {
	var resp error
	err := h.client.Call("Host.ShowMessage", message, &resp)
	if err != nil {
		// Log error but don't return it, as ShowMessage shouldn't fail the plugin
		fmt.Printf("[RPC] ShowMessage call failed: %v\n", err)
	}
}

func (h *RPCHostClient) ExecuteCommand(name string, args ...interface{}) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("ExecuteCommand not implemented in RPC client")
}

func (h *RPCHostClient) SetMajorMode(bufferName, modeName string) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("SetMajorMode not implemented in RPC client")
}

func (h *RPCHostClient) ToggleMinorMode(bufferName, modeName string) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("ToggleMinorMode not implemented in RPC client")
}

func (h *RPCHostClient) AddHook(event string, handler func(...interface{}) error) {
	// TODO: Implement RPC call to host
}

func (h *RPCHostClient) TriggerHook(event string, args ...interface{}) {
	// TODO: Implement RPC call to host
}

func (h *RPCHostClient) CreateBuffer(name string) BufferInterface {
	fmt.Printf("[RPC] CreateBuffer called with name: %s\n", name)
	var resp BufferInfo
	err := h.client.Call("Host.CreateBuffer", name, &resp)
	if err != nil {
		fmt.Printf("[RPC] CreateBuffer call failed: %v\n", err)
		return nil
	}
	
	fmt.Printf("[RPC] CreateBuffer succeeded: %+v\n", resp)
	return &RPCBufferProxy{
		client: h.client,
		info:   resp,
	}
}

func (h *RPCHostClient) FindBuffer(name string) BufferInterface {
	// TODO: Implement RPC call to host
	return nil
}

func (h *RPCHostClient) SwitchToBuffer(name string) error {
	var resp error
	err := h.client.Call("Host.SwitchToBuffer", name, &resp)
	if err != nil {
		return fmt.Errorf("RPC call failed: %v", err)
	}
	return resp
}

func (h *RPCHostClient) OpenFile(path string) error {
	fmt.Printf("[RPC] OpenFile called with path: %s\n", path)
	var resp error
	err := h.client.Call("Host.OpenFile", path, &resp)
	if err != nil {
		fmt.Printf("[RPC] OpenFile RPC call failed: %v\n", err)
		return fmt.Errorf("RPC call failed: %v", err)
	}
	if resp != nil {
		fmt.Printf("[RPC] OpenFile failed on host: %v\n", resp)
	} else {
		fmt.Printf("[RPC] OpenFile succeeded on host\n")
	}
	return resp
}

func (h *RPCHostClient) SaveBuffer(bufferName string) error {
	fmt.Printf("[RPC] SaveBuffer called with buffer: %s\n", bufferName)
	var resp error
	err := h.client.Call("Host.SaveBuffer", bufferName, &resp)
	if err != nil {
		fmt.Printf("[RPC] SaveBuffer RPC call failed: %v\n", err)
		return fmt.Errorf("RPC call failed: %v", err)
	}
	if resp != nil {
		fmt.Printf("[RPC] SaveBuffer failed on host: %v\n", resp)
	} else {
		fmt.Printf("[RPC] SaveBuffer succeeded on host\n")
	}
	return resp
}

func (h *RPCHostClient) GetOption(name string) (interface{}, error) {
	// TODO: Implement RPC call to host
	return nil, fmt.Errorf("GetOption not implemented in RPC client")
}

func (h *RPCHostClient) SetOption(name string, value interface{}) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("SetOption not implemented in RPC client")
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

func (p *RPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	// Create RPC server with MuxBroker for bidirectional communication
	return &RPCServer{Impl: p.Impl, broker: broker}, nil
}

func (p *RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c, broker: b}, nil
}

// RPCServer はプラグイン側のRPCサーバー
type RPCServer struct {
	Impl   Plugin
	broker *plugin.MuxBroker
}

// RPCClient はホスト側のRPCクライアント
type RPCClient struct {
	client *rpc.Client
	broker *plugin.MuxBroker
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
	fmt.Printf("[RPC] Initialize called - setting up MuxBroker\n")
	
	// Start RPC server for HostInterface on client side
	hostBrokerID := c.broker.NextId()
	fmt.Printf("[RPC] Starting host RPC server with broker ID: %d\n", hostBrokerID)
	
	// Create a proper RPC server and register the Host service
	go func() {
		// Accept connection from plugin
		conn, err := c.broker.Accept(hostBrokerID)
		if err != nil {
			fmt.Printf("[RPC] Failed to accept connection on broker ID %d: %v\n", hostBrokerID, err)
			return
		}
		
		// Create RPC server and register Host service
		server := rpc.NewServer()
		err = server.RegisterName("Host", &RPCHostServer{Impl: host})
		if err != nil {
			fmt.Printf("[RPC] Failed to register Host service: %v\n", err)
			return
		}
		
		fmt.Printf("[RPC] Host service registered, serving RPC\n")
		server.ServeConn(conn)
	}()
	
	// Send the broker ID to plugin so it can connect back
	args := map[string]interface{}{
		"hostBrokerID": hostBrokerID,
	}
	
	fmt.Printf("[RPC] Calling Plugin.Initialize with args: %+v\n", args)
	var resp error
	err := c.client.Call("Plugin.Initialize", args, &resp)
	if err != nil {
		fmt.Printf("[RPC] Plugin.Initialize failed: %v\n", err)
	} else {
		fmt.Printf("[RPC] Plugin.Initialize succeeded\n")
	}
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
	fmt.Printf("[RPC-Server] Initialize called with args: %+v\n", args)
	
	// Extract the host broker ID from args
	hostBrokerID, ok := args["hostBrokerID"].(uint32)
	if !ok {
		fmt.Printf("[RPC-Server] hostBrokerID not provided or wrong type\n")
		*resp = fmt.Errorf("hostBrokerID not provided")
		return nil
	}
	
	fmt.Printf("[RPC-Server] Connecting to host broker ID: %d\n", hostBrokerID)
	
	// Connect to the host's RPC server using MuxBroker
	conn, err := s.broker.Dial(hostBrokerID)
	if err != nil {
		fmt.Printf("[RPC-Server] Failed to connect to host broker: %v\n", err)
		*resp = fmt.Errorf("failed to connect to host broker: %v", err)
		return nil
	}
	
	fmt.Printf("[RPC-Server] Successfully connected to host broker\n")
	
	// Create RPC client for host interface
	hostClient := &RPCHostClient{client: rpc.NewClient(conn)}
	
	fmt.Printf("[RPC-Server] Created host RPC client, initializing plugin\n")
	
	// Initialize the plugin with the host interface
	*resp = s.Impl.Initialize(context.Background(), hostClient)
	if *resp != nil {
		fmt.Printf("[RPC-Server] Plugin initialization failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Server] Plugin initialization succeeded\n")
	}
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