package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-plugin"
	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type StringError struct {
	Message string
}

func (se StringError) Error() string {
	return se.Message
}

func NewStringError(message string) error {
	return StringError{Message: message}
}

func init() {
	gob.Register(StringError{})
}

type FileTestPlugin struct {
	host pluginsdk.HostInterface
}

func (p *FileTestPlugin) Name() string {
	return "file-test-plugin"
}

func (p *FileTestPlugin) Version() string {
	return "1.0.0"
}

func (p *FileTestPlugin) Description() string {
	return "Plugin for testing file operations"
}

func (p *FileTestPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
	fmt.Printf("[PLUGIN] Initialize called with host: %T\n", host)
	p.host = host
	fmt.Printf("[PLUGIN] Initialize completed, host stored: %v\n", p.host != nil)
	return nil
}

func (p *FileTestPlugin) Cleanup() error {
	return nil
}

func (p *FileTestPlugin) GetCommands() []pluginsdk.CommandSpec {
	fmt.Printf("[PLUGIN] GetCommands called, returning 4 commands\n")
	commands := []pluginsdk.CommandSpec{
		{
			Name:        "file-test-create",
			Description: "Test file creation and opening",
			Interactive: true,
			Handler:     "HandleFileCreate",
		},
		{
			Name:        "file-test-save",
			Description: "Test file saving",
			Interactive: true,
			Handler:     "HandleFileSave",
		},
		{
			Name:        "file-test-content",
			Description: "Test file content operations",
			Interactive: true,
			Handler:     "HandleFileContent",
		},
		{
			Name:        "file-test-all",
			Description: "Test all file operations in sequence",
			Interactive: true,
			Handler:     "HandleFileAll",
		},
	}
	fmt.Printf("[PLUGIN] GetCommands returning %d commands: ", len(commands))
	for _, cmd := range commands {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	return commands
}

func (p *FileTestPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
	return []pluginsdk.MajorModeSpec{}
}

func (p *FileTestPlugin) GetMinorModes() []pluginsdk.MinorModeSpec {
	return []pluginsdk.MinorModeSpec{}
}

func (p *FileTestPlugin) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	return []pluginsdk.KeyBindingSpec{}
}

func (p *FileTestPlugin) HandleFileCreate() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	// テスト用ファイルを作成
	testFile := filepath.Join(os.TempDir(), "gmacs-plugin-test.txt")
	testContent := "Test file created by plugin\nLine 2\nLine 3"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File creation FAILED: %v", err))
	}

	// ファイルを開く
	err = p.host.OpenFile(testFile)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File open FAILED: %v", err))
	}

	return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File create/open test PASSED: %s", testFile))
}

func (p *FileTestPlugin) HandleFileSave() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// バッファの内容を変更
	originalContent := buffer.Content()
	modifiedContent := originalContent + "\nModified by plugin"
	buffer.SetContent(modifiedContent)
	buffer.MarkDirty()

	// バッファを保存
	bufferName := buffer.Name()
	err := p.host.SaveBuffer(bufferName)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File save FAILED: %v", err))
	}

	// バッファの状態を再取得してダーティフラグをチェック
	refreshedBuffer := p.host.GetCurrentBuffer()
	if refreshedBuffer == nil {
		return NewStringError("PLUGIN_MESSAGE:File save test FAILED: buffer disappeared after save")
	}
	
	if refreshedBuffer.IsDirty() {
		return NewStringError("PLUGIN_MESSAGE:File save test PARTIAL: file saved but dirty flag still set")
	}

	return NewStringError("PLUGIN_MESSAGE:File save test PASSED")
}

func (p *FileTestPlugin) HandleFileContent() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError("PLUGIN_MESSAGE:ERROR: No current buffer")
	}

	// ファイル情報をテスト
	name := buffer.Name()
	filename := buffer.Filename()
	content := buffer.Content()
	isDirty := buffer.IsDirty()

	if filename == "" {
		return NewStringError("PLUGIN_MESSAGE:File content test PARTIAL: buffer has no associated file")
	}

	// ファイルが実際に存在するかチェック
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File content test FAILED: file does not exist: %s", filename))
	}

	result := fmt.Sprintf("File: %s, Buffer: %s, Content length: %d, Dirty: %t", 
		filename, name, len(content), isDirty)
	return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File content test PASSED - %s", result))
}

func (p *FileTestPlugin) HandleFileAll() error {
	if p.host == nil {
		return NewStringError("ERROR: host is nil")
	}

	// Step 1: Create and open file
	testFile := filepath.Join(os.TempDir(), "gmacs-plugin-test.txt")
	testContent := "Test file created by plugin\nLine 2\nLine 3"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File all test FAILED: create failed: %v", err))
	}

	err = p.host.OpenFile(testFile)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File all test FAILED: open failed: %v", err))
	}

	// Step 2: Test file content operations
	buffer := p.host.GetCurrentBuffer()
	if buffer == nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File all test FAILED: No current buffer after open - host type: %T", p.host))
	}

	filename := buffer.Filename()
	if filename == "" {
		return NewStringError("PLUGIN_MESSAGE:File all test FAILED: buffer has no associated file")
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File all test FAILED: file does not exist: %s", filename))
	}

	// Step 3: Test file saving
	originalContent := buffer.Content()
	modifiedContent := originalContent + "\nModified by plugin"
	buffer.SetContent(modifiedContent)
	buffer.MarkDirty()

	bufferName := buffer.Name()
	err = p.host.SaveBuffer(bufferName)
	if err != nil {
		return NewStringError(fmt.Sprintf("PLUGIN_MESSAGE:File all test FAILED: save failed: %v", err))
	}

	// バッファの状態を再取得してダーティフラグをチェック
	refreshedBuffer := p.host.GetCurrentBuffer()
	if refreshedBuffer == nil {
		return NewStringError("PLUGIN_MESSAGE:File all test FAILED: buffer disappeared after save")
	}
	
	if refreshedBuffer.IsDirty() {
		return NewStringError("PLUGIN_MESSAGE:File all test FAILED: file saved but dirty flag still set")
	}

	return NewStringError("PLUGIN_MESSAGE:File all test PASSED: create, open, content check, and save all successful")
}

// CommandPlugin インターフェース実装
func (p *FileTestPlugin) ExecuteCommand(name string, args ...interface{}) error {
	// 初期化確認
	if p.host == nil {
		// For debugging - use SimpleHostInterface but return error message about it
		hostImpl := &SimpleHostInterface{}
		p.Initialize(context.Background(), hostImpl)
		if name == "file-test-all" {
			return NewStringError("PLUGIN_MESSAGE:File all test FAILED: host was nil, using SimpleHostInterface - RPC initialization issue")
		}
	}
	fmt.Printf("[PLUGIN] ExecuteCommand called: %s with host: %T\n", name, p.host)

	switch name {
	case "file-test-create":
		return p.HandleFileCreate()
	case "file-test-save":
		return p.HandleFileSave()
	case "file-test-content":
		return p.HandleFileContent()
	case "file-test-all":
		return p.HandleFileAll()
	default:
		return fmt.Errorf("unknown command: %s", name)
	}
}

func (p *FileTestPlugin) GetCompletions(command string, prefix string) []string {
	return []string{}
}

// SimpleHostInterface for testing
type SimpleHostInterface struct{}

func (h *SimpleHostInterface) GetCurrentBuffer() pluginsdk.BufferInterface { return nil }
func (h *SimpleHostInterface) GetCurrentWindow() pluginsdk.WindowInterface { return nil }
func (h *SimpleHostInterface) SetStatus(message string)                     {}
func (h *SimpleHostInterface) ShowMessage(message string)                   {}
func (h *SimpleHostInterface) ExecuteCommand(name string, args ...interface{}) error {
	return nil
}
func (h *SimpleHostInterface) SetMajorMode(bufferName, modeName string) error { return nil }
func (h *SimpleHostInterface) ToggleMinorMode(bufferName, modeName string) error {
	return nil
}
func (h *SimpleHostInterface) AddHook(event string, handler func(...interface{}) error) {}
func (h *SimpleHostInterface) TriggerHook(event string, args ...interface{})             {}
func (h *SimpleHostInterface) CreateBuffer(name string) pluginsdk.BufferInterface       { return nil }
func (h *SimpleHostInterface) FindBuffer(name string) pluginsdk.BufferInterface         { return nil }
func (h *SimpleHostInterface) SwitchToBuffer(name string) error                         { return nil }
func (h *SimpleHostInterface) OpenFile(path string) error                               { return nil }
func (h *SimpleHostInterface) SaveBuffer(bufferName string) error                       { return nil }
func (h *SimpleHostInterface) GetOption(name string) (interface{}, error)               { return nil, nil }
func (h *SimpleHostInterface) SetOption(name string, value interface{}) error           { return nil }

var pluginInstance = &FileTestPlugin{}

// RPCPlugin は標準的なgmacs RPCプラグイン実装
type RPCPlugin struct {
	Impl pluginsdk.Plugin
	broker *plugin.MuxBroker
}

func (p *RPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl, broker: broker}, nil
}

func (p *RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c, broker: b}, nil
}

// RPCServer はプラグイン側のRPCサーバー
type RPCServer struct {
	Impl   pluginsdk.Plugin
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
	err := c.client.Call("Plugin.Name", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Version() string {
	var resp string
	err := c.client.Call("Plugin.Version", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Description() string {
	var resp string
	err := c.client.Call("Plugin.Description", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
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
	err := c.client.Call("Plugin.Cleanup", interface{}(nil), &resp)
	return err
}

func (c *RPCClient) GetCommands() []pluginsdk.CommandSpec {
	fmt.Printf("[RPC-Client] GetCommands called\n")
	var resp []pluginsdk.CommandSpec
	err := c.client.Call("Plugin.GetCommands", interface{}(nil), &resp)
	if err != nil {
		fmt.Printf("[RPC-Client] GetCommands RPC call failed: %v\n", err)
		return nil
	}
	fmt.Printf("[RPC-Client] GetCommands received %d commands: ", len(resp))
	for _, cmd := range resp {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	return resp
}

func (c *RPCClient) GetMajorModes() []pluginsdk.MajorModeSpec {
	var resp []pluginsdk.MajorModeSpec
	err := c.client.Call("Plugin.GetMajorModes", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetMinorModes() []pluginsdk.MinorModeSpec {
	var resp []pluginsdk.MinorModeSpec
	err := c.client.Call("Plugin.GetMinorModes", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	var resp []pluginsdk.KeyBindingSpec
	err := c.client.Call("Plugin.GetKeyBindings", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

// RPCServer Plugin インターフェースの実装
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

func (s *RPCServer) GetCommands(args interface{}, resp *[]pluginsdk.CommandSpec) error {
	fmt.Printf("[RPC-Server] GetCommands called\n")
	commands := s.Impl.GetCommands()
	fmt.Printf("[RPC-Server] Got %d commands from Impl: ", len(commands))
	for _, cmd := range commands {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	*resp = commands
	fmt.Printf("[RPC-Server] GetCommands setting resp to %d commands\n", len(*resp))
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

func (s *RPCServer) ExecuteCommand(args map[string]interface{}, resp *error) error {
	name, _ := args["name"].(string)
	argsSlice, _ := args["args"].([]interface{})

	fmt.Printf("[RPC-Server] ExecuteCommand called: %s\n", name)

	if cmdPlugin, ok := s.Impl.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
		*resp = cmdPlugin.ExecuteCommand(name, argsSlice...)
		if *resp != nil {
			fmt.Printf("[RPC-Server] ExecuteCommand failed: %v\n", *resp)
		}
	} else {
		*resp = fmt.Errorf("plugin does not support command execution")
	}
	return nil
}

// RPCHostClient はプラグイン側でホストの機能をRPC経由で呼び出すクライアント
type RPCHostClient struct {
	client *rpc.Client
}

// HostInterface implementation for RPC client
func (h *RPCHostClient) GetCurrentBuffer() pluginsdk.BufferInterface {
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

func (h *RPCHostClient) GetCurrentWindow() pluginsdk.WindowInterface {
	// TODO: Implement RPC call to host
	return nil
}

func (h *RPCHostClient) SetStatus(message string) {
	var resp error
	h.client.Call("Host.SetStatus", message, &resp)
}

func (h *RPCHostClient) ShowMessage(message string) {
	var resp error
	h.client.Call("Host.ShowMessage", message, &resp)
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

func (h *RPCHostClient) CreateBuffer(name string) pluginsdk.BufferInterface {
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

func (h *RPCHostClient) FindBuffer(name string) pluginsdk.BufferInterface {
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

// RPCHostServer はgmacs側でホスト機能をRPC経由で提供するサーバー
type RPCHostServer struct {
	Impl pluginsdk.HostInterface
}

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
func (h *RPCHostServer) GetCurrentBuffer(args interface{}, resp *BufferInfo) error {
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