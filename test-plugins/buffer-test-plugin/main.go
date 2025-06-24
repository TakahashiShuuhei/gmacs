package main

import (
	"context"
	"fmt"
	"log"
	"net/rpc"
	"os"

	"github.com/hashicorp/go-plugin"
	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

// TestRPCServer implements a working RPC server for our test plugin
type TestRPCServer struct {
	Impl   pluginsdk.Plugin
	broker *plugin.MuxBroker
}

func (s *TestRPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *TestRPCServer) Version(args interface{}, resp *string) error {
	*resp = s.Impl.Version()
	return nil
}

func (s *TestRPCServer) Description(args interface{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *TestRPCServer) Initialize(args map[string]interface{}, resp *error) error {
	log.Printf("[RPC-Server] Initialize called with args: %+v", args)
	
	// Extract the host broker ID from args
	hostBrokerID, ok := args["hostBrokerID"].(uint32)
	if !ok {
		log.Printf("[RPC-Server] hostBrokerID not provided or wrong type")
		*resp = fmt.Errorf("hostBrokerID not provided")
		return nil
	}
	
	log.Printf("[RPC-Server] Connecting to host broker ID: %d", hostBrokerID)
	
	// Connect to the host's RPC server using MuxBroker
	conn, err := s.broker.Dial(hostBrokerID)
	if err != nil {
		log.Printf("[RPC-Server] Failed to connect to host broker: %v", err)
		*resp = fmt.Errorf("failed to connect to host broker: %v", err)
		return nil
	}
	
	log.Printf("[RPC-Server] Successfully connected to host broker")
	
	// Create RPC client for host interface
	hostClient := &TestRPCHostClient{client: rpc.NewClient(conn)}
	
	log.Printf("[RPC-Server] Created host RPC client, initializing plugin")
	
	// Initialize the plugin with the host interface
	*resp = s.Impl.Initialize(context.Background(), hostClient)
	if *resp != nil {
		log.Printf("[RPC-Server] Plugin initialization failed: %v", *resp)
	} else {
		log.Printf("[RPC-Server] Plugin initialization succeeded")
	}
	return nil
}

func (s *TestRPCServer) Cleanup(args interface{}, resp *error) error {
	*resp = s.Impl.Cleanup()
	return nil
}

func (s *TestRPCServer) GetCommands(args interface{}, resp *[]pluginsdk.CommandSpec) error {
	*resp = s.Impl.GetCommands()
	return nil
}

func (s *TestRPCServer) GetMajorModes(args interface{}, resp *[]pluginsdk.MajorModeSpec) error {
	*resp = s.Impl.GetMajorModes()
	return nil
}

func (s *TestRPCServer) GetMinorModes(args interface{}, resp *[]pluginsdk.MinorModeSpec) error {
	*resp = s.Impl.GetMinorModes()
	return nil
}

func (s *TestRPCServer) GetKeyBindings(args interface{}, resp *[]pluginsdk.KeyBindingSpec) error {
	*resp = s.Impl.GetKeyBindings()
	return nil
}

// CommandPlugin RPC server methods
func (s *TestRPCServer) ExecuteCommand(args map[string]interface{}, resp *error) error {
	log.Printf("[RPC-Server] ExecuteCommand called with args: %+v", args)
	
	if cmdPlugin, ok := s.Impl.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
		name, _ := args["name"].(string)
		argsStrings, _ := args["args"].([]string)
		
		// Convert string slice back to []interface{}
		pluginArgs := make([]interface{}, len(argsStrings))
		for i, arg := range argsStrings {
			pluginArgs[i] = arg
		}
		
		log.Printf("[RPC-Server] Calling plugin ExecuteCommand(%s, %v)", name, pluginArgs)
		*resp = cmdPlugin.ExecuteCommand(name, pluginArgs...)
		
		if *resp != nil {
			log.Printf("[RPC-Server] ExecuteCommand failed: %v", *resp)
		} else {
			log.Printf("[RPC-Server] ExecuteCommand succeeded")
		}
	} else {
		log.Printf("[RPC-Server] Plugin does not implement CommandPlugin interface")
		*resp = fmt.Errorf("plugin does not implement CommandPlugin interface")
	}
	return nil
}

func (s *TestRPCServer) GetCompletions(args map[string]interface{}, resp *[]string) error {
	if cmdPlugin, ok := s.Impl.(interface{ GetCompletions(string, string) []string }); ok {
		command, _ := args["command"].(string)
		prefix, _ := args["prefix"].(string)
		*resp = cmdPlugin.GetCompletions(command, prefix)
	} else {
		*resp = []string{}
	}
	return nil
}

// TestBufferInfo represents buffer state for RPC transmission
type TestBufferInfo struct {
	Name     string
	Content  string
	Position int
	IsDirty  bool
	Filename string
}

// TestBufferProxy provides a client-side proxy for buffer operations via RPC
type TestBufferProxy struct {
	info TestBufferInfo
}

func (b *TestBufferProxy) Name() string           { return b.info.Name }
func (b *TestBufferProxy) Content() string        { return b.info.Content }
func (b *TestBufferProxy) CursorPosition() int    { return b.info.Position }
func (b *TestBufferProxy) IsDirty() bool          { return b.info.IsDirty }
func (b *TestBufferProxy) Filename() string       { return b.info.Filename }

func (b *TestBufferProxy) SetContent(content string) {
	b.info.Content = content
	// TODO: Implement RPC call to sync content to host
}

func (b *TestBufferProxy) InsertAt(pos int, text string) {
	// TODO: Implement RPC call to insert text at position
}

func (b *TestBufferProxy) DeleteRange(start, end int) {
	// TODO: Implement RPC call to delete text range
}

func (b *TestBufferProxy) SetCursorPosition(pos int) {
	b.info.Position = pos
	// TODO: Implement RPC call to sync cursor position to host
}

func (b *TestBufferProxy) MarkDirty() {
	b.info.IsDirty = true
	// TODO: Implement RPC call to mark buffer dirty on host
}

// TestRPCHostClient provides a simple host client for test plugins
type TestRPCHostClient struct {
	client *rpc.Client
}

func (h *TestRPCHostClient) GetCurrentBuffer() pluginsdk.BufferInterface {
	log.Printf("[RPC] GetCurrentBuffer called")
	
	// Make RPC call to host
	var resp TestBufferInfo
	err := h.client.Call("Host.GetCurrentBuffer", struct{}{}, &resp)
	if err != nil {
		log.Printf("[RPC] GetCurrentBuffer call failed: %v", err)
		return nil
	}
	
	// Check if we got an empty buffer info (indicates no current buffer)
	if resp.Name == "" {
		log.Printf("[RPC] GetCurrentBuffer returned empty buffer (no current buffer)")
		return nil
	}
	
	log.Printf("[RPC] GetCurrentBuffer succeeded: %+v", resp)
	return &TestBufferProxy{info: resp}
}
func (h *TestRPCHostClient) GetCurrentWindow() pluginsdk.WindowInterface  { return nil }
func (h *TestRPCHostClient) SetStatus(message string)                      {}
func (h *TestRPCHostClient) ShowMessage(message string)                    {}
func (h *TestRPCHostClient) ExecuteCommand(name string, args ...interface{}) error { return nil }
func (h *TestRPCHostClient) SetMajorMode(bufferName, modeName string) error { return nil }
func (h *TestRPCHostClient) ToggleMinorMode(bufferName, modeName string) error { return nil }
func (h *TestRPCHostClient) AddHook(event string, handler func(...interface{}) error) {}
func (h *TestRPCHostClient) TriggerHook(event string, args ...interface{}) {}

func (h *TestRPCHostClient) CreateBuffer(name string) pluginsdk.BufferInterface {
	log.Printf("[RPC] CreateBuffer called with name: %s", name)
	
	// Make RPC call to host
	var resp TestBufferInfo
	err := h.client.Call("Host.CreateBuffer", name, &resp)
	if err != nil {
		log.Printf("[RPC] CreateBuffer call failed: %v", err)
		return nil
	}
	
	log.Printf("[RPC] CreateBuffer succeeded: %+v", resp)
	return &TestBufferProxy{info: resp}
}

func (h *TestRPCHostClient) FindBuffer(name string) pluginsdk.BufferInterface { return nil }
func (h *TestRPCHostClient) SwitchToBuffer(name string) error {
	log.Printf("[RPC] SwitchToBuffer called with name: %s", name)
	
	// Make RPC call to host
	var resp error
	err := h.client.Call("Host.SwitchToBuffer", name, &resp)
	if err != nil {
		log.Printf("[RPC] SwitchToBuffer call failed: %v", err)
		return err
	}
	
	log.Printf("[RPC] SwitchToBuffer succeeded")
	return resp
}
func (h *TestRPCHostClient) OpenFile(path string) error { return nil }
func (h *TestRPCHostClient) SaveBuffer(bufferName string) error { return nil }
func (h *TestRPCHostClient) GetOption(name string) (interface{}, error) { return nil, nil }
func (h *TestRPCHostClient) SetOption(name string, value interface{}) error { return nil }

// TestRPCPlugin implements the go-plugin interface
type TestRPCPlugin struct{}

func (p *TestRPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &TestRPCServer{Impl: pluginInstance, broker: broker}, nil
}

func (p *TestRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return nil, nil
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[BUFFER-TEST-PLUGIN] ")
	log.Printf("Starting buffer test plugin...")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GMACS_PLUGIN",
			MagicCookieValue: "gmacs-plugin-magic-cookie",
		},
		Plugins: map[string]plugin.Plugin{
			"gmacs-plugin": &TestRPCPlugin{},
		},
	})
}