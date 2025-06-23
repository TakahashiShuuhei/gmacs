package domain

import (
	"github.com/TakahashiShuuhei/gmacs/events"
)

// ConfigLoader interface for Lua configuration loading
type ConfigLoader interface {
	LoadConfig(configPath string) error
	Close()
}

// HookManager interface for event hooks
type HookManager interface {
	AddHook(event string, fn func(...interface{}) error)
	TriggerHook(event string, args ...interface{})
}

// PluginManagerInterface for plugin management
type PluginManagerInterface interface {
	LoadPlugin(name string) error
	UnloadPlugin(name string) error
	GetPlugin(name string) (PluginInterface, bool)
	ListPlugins() []PluginInfo
	Shutdown() error
}

// PluginInterface represents a loaded plugin
type PluginInterface interface {
	Name() string
	Version() string
	Description() string
	GetCommands() []CommandSpec
	GetMajorModes() []MajorModeSpec
	GetMinorModes() []MinorModeSpec
	GetKeyBindings() []KeyBindingSpec
}

// Plugin-related data structures
type CommandSpec struct {
	Name        string
	Description string
	Interactive bool
	Handler     string
}

type MajorModeSpec struct {
	Name         string
	Extensions   []string
	Description  string
	KeyBindings  []KeyBindingSpec
}

type MinorModeSpec struct {
	Name        string
	Description string
	Global      bool
	KeyBindings []KeyBindingSpec
}

type KeyBindingSpec struct {
	Sequence string
	Command  string
	Mode     string
}

type PluginInfo struct {
	Name        string
	Version     string
	Description string
	State       int
	Enabled     bool
}

type Editor struct {
	buffers         []*Buffer
	layout          *WindowLayout
	eventQueue      *events.EventQueue
	running         bool
	minibuffer      *Minibuffer
	commandRegistry *CommandRegistry
	keyBindings     *KeyBindingMap
	metaPressed     bool
	modeManager     *ModeManager
	configLoader    ConfigLoader
	hookManager     HookManager
	options         map[string]interface{}
	pluginManager   PluginManagerInterface // プラグインマネージャー
}

// EditorConfig holds configuration options for editor initialization
type EditorConfig struct {
	ConfigLoader  ConfigLoader            // Optional Lua config loader
	HookManager   HookManager             // Optional hook manager
	PluginManager PluginManagerInterface  // Optional plugin manager
}

// NewEditor creates a new editor with default configuration
func NewEditor() *Editor {
	editor := newEditorWithConfig(EditorConfig{})

	// Register built-in commands directly for backward compatibility
	editor.RegisterBuiltinCommands()

	return editor
}

// NewEditorWithConfig creates a new editor with the specified configuration
func NewEditorWithConfig(configLoader ConfigLoader, hookManager HookManager) *Editor {
	return newEditorWithConfig(EditorConfig{
		ConfigLoader: configLoader,
		HookManager:  hookManager,
	})
}

// newEditorWithConfig is the internal constructor that handles configuration
func newEditorWithConfig(config EditorConfig) *Editor {
	buffer := NewBuffer("*scratch*")
	window := NewWindow(buffer, 80, 22)       // 24-2 for mode line and minibuffer
	layout := NewWindowLayout(window, 80, 24) // Total terminal size

	editor := &Editor{
		buffers:         []*Buffer{buffer},
		layout:          layout,
		eventQueue:      events.NewEventQueue(100),
		running:         true,
		minibuffer:      NewMinibuffer(),
		commandRegistry: NewCommandRegistry(),
		keyBindings:     NewKeyBindingMap(),
		metaPressed:     false,
		modeManager:     NewModeManager(),
		configLoader:    config.ConfigLoader,
		hookManager:     config.HookManager,
		options:         make(map[string]interface{}),
		pluginManager:   config.PluginManager,
	}

	// Built-in commands are now registered via Lua configuration

	// Initialize buffer with fundamental mode
	err := editor.modeManager.SetMajorMode(buffer, "fundamental-mode")
	if err != nil {
		// TODO: Handle error appropriately
	}

	return editor
}

// RegisterBuiltinCommands registers all built-in editor commands using automatic discovery
func (e *Editor) RegisterBuiltinCommands() {
	// Use automatic function discovery - no need to manually call each registration
	registrationFunctions := GetAllRegistrationFunctions()
	for _, registerFunc := range registrationFunctions {
		registerFunc(e.commandRegistry)
	}
	
	// Special case for minor mode commands (not yet refactored)
	e.registerMinorModeCommands()
}


// Cleanup closes any resources when the editor is shutting down
func (e *Editor) Cleanup() {
	if e.configLoader != nil {
		e.configLoader.Close()
		e.configLoader = nil
	}
	
	if e.pluginManager != nil {
		e.pluginManager.Shutdown()
		e.pluginManager = nil
	}
}

// EditorInterface implementation for Lua API

// BindKey implements global key binding
func (e *Editor) BindKey(sequence, command string) error {
	cmd, exists := e.commandRegistry.Get(command)
	if !exists {
		return &ConfigError{Message: "Unknown command: " + command}
	}

	// Convert command to function
	cmdFunc := func(editor *Editor) error {
		return cmd.Execute(editor)
	}

	e.keyBindings.BindKeySequence(sequence, cmdFunc)
	return nil
}

// LocalBindKey implements mode-specific key binding
func (e *Editor) LocalBindKey(modeName, sequence, command string) error {
	// Check if command exists
	cmd, exists := e.commandRegistry.Get(command)
	if !exists {
		return &ConfigError{Message: "Unknown command: " + command}
	}

	// Convert command to function
	cmdFunc := func(editor *Editor) error {
		return cmd.Execute(editor)
	}

	// Try to find major mode first
	if majorMode, exists := e.modeManager.GetMajorModeByName(modeName); exists {
		keyBindings := majorMode.KeyBindings()
		if keyBindings != nil {
			keyBindings.BindKeySequence(sequence, cmdFunc)
			return nil
		}
	}

	// Try to find minor mode
	if minorMode, exists := e.modeManager.GetMinorModeByName(modeName); exists {
		keyBindings := minorMode.KeyBindings()
		if keyBindings != nil {
			keyBindings.BindKeySequence(sequence, cmdFunc)
			return nil
		}
	}

	return &ConfigError{Message: "Unknown mode: " + modeName}
}

// RegisterCommand implements custom command registration
func (e *Editor) RegisterCommand(name string, fn func() error) error {
	// Convert to CommandFunc
	cmdFunc := func(editor *Editor) error {
		return fn()
	}

	e.commandRegistry.RegisterFunc(name, cmdFunc)
	return nil
}

// SetOption implements option setting
func (e *Editor) SetOption(name string, value interface{}) error {
	e.options[name] = value
	return nil
}

// GetOption implements option getting
func (e *Editor) GetOption(name string) (interface{}, error) {
	value, exists := e.options[name]
	if !exists {
		return nil, &ConfigError{Message: "Unknown option: " + name}
	}
	return value, nil
}

// RegisterMajorMode implements major mode registration
func (e *Editor) RegisterMajorMode(name string, config map[string]interface{}) error {
	// TODO: Implement dynamic major mode registration
	return nil
}

// RegisterMinorMode implements minor mode registration
func (e *Editor) RegisterMinorMode(name string, config map[string]interface{}) error {
	// TODO: Implement dynamic minor mode registration
	return nil
}

// AddHook implements hook registration
func (e *Editor) AddHook(event string, fn func(...interface{}) error) error {
	if e.hookManager != nil {
		e.hookManager.AddHook(event, fn)
	}
	return nil
}

// TriggerHook triggers hooks for an event
func (e *Editor) TriggerHook(event string, args ...interface{}) {
	if e.hookManager != nil {
		e.hookManager.TriggerHook(event, args...)
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

func (e *Editor) CurrentBuffer() *Buffer {
	window := e.CurrentWindow()
	if window != nil {
		return window.Buffer()
	}
	return nil
}

func (e *Editor) CurrentWindow() *Window {
	if e.layout != nil {
		return e.layout.CurrentWindow()
	}
	return nil
}

func (e *Editor) Layout() *WindowLayout {
	return e.layout
}

func (e *Editor) EventQueue() *events.EventQueue {
	return e.eventQueue
}

func (e *Editor) HandleEvent(event events.Event) {
	switch ev := event.(type) {
	case events.KeyEventData:
		e.handleKeyEvent(ev)
	case events.ResizeEventData:
		e.handleResizeEvent(ev)
	case events.QuitEventData:
		e.running = false
	default:
		// Unknown event type
	}
}

func (e *Editor) handleKeyEvent(event events.KeyEventData) {

	// Always process key sequences first to handle multi-key sequences correctly
	cmd, matched, continuing := e.keyBindings.ProcessKeyPress(event.Key, event.Ctrl, event.Meta)

	// If we have a continuing sequence, always handle it first
	if continuing {
		return
	}

	// Handle minibuffer input first if active (except for special cases)
	if e.minibuffer.IsActive() {
		// Special commands that should always execute immediately
		if event.Ctrl && event.Key == "g" { // C-g (KeyboardQuit)
			if matched {
				cmd(e)
			}
			return
		}

		// Try minibuffer handling first
		handled := e.minibuffer.HandleInput(event, e)
		if handled {
			return
		}
	}

	// If not handled by minibuffer, check for matched global commands
	if matched {
		cmd(e)
		return
	}

	// Handle Meta key press detection
	if event.Key == "\x1b" || event.Key == "Escape" {
		e.metaPressed = true
		// Reset key sequence on Escape
		e.keyBindings.ResetSequence()
		return
	}

	// Handle M-x command
	if e.metaPressed && event.Key == "x" {
		e.minibuffer.StartCommandInput()
		e.metaPressed = false
		return
	}

	// Handle other Meta key combinations
	if e.metaPressed {
		metaSequence := "M-" + event.Key
		if cmd, found := e.keyBindings.LookupSequence(metaSequence); found {
			cmd(e)
			e.metaPressed = false
			return
		}
		// Reset meta state for unbound keys
		e.metaPressed = false
	}

	// Check for any remaining key bindings through the unified system
	// (single keys and raw sequences that weren't caught by sequence processing)
	if cmd, found := e.keyBindings.LookupSequence(event.Key); found {
		cmd(e)
		return
	}

	// Regular text input
	buffer := e.CurrentBuffer()
	if buffer == nil {
		return
	}

	if event.Rune != 0 && !event.Ctrl && !event.Meta {
		if event.Key == "Enter" || event.Key == "Return" {
			buffer.InsertChar('\n')

			// Check for auto-a-mode and add 'a' if enabled
			e.processMinorModeHooks(buffer, "newline")

			EnsureCursorVisible(e)
		} else {
			buffer.InsertChar(event.Rune)
			EnsureCursorVisible(e)
		}
	}
}

func (e *Editor) handleResizeEvent(event events.ResizeEventData) {
	if e.layout != nil {
		e.layout.Resize(event.Width, event.Height)
	}
}

func (e *Editor) IsRunning() bool {
	return e.running
}

func (e *Editor) Quit() {
	e.running = false
}

func (e *Editor) Minibuffer() *Minibuffer {
	return e.minibuffer
}

func (e *Editor) SetMinibufferMessage(message string) {
	e.minibuffer.SetMessage(message)
}

// GetKeySequenceInProgress returns the current key sequence in progress, if any
func (e *Editor) GetKeySequenceInProgress() string {
	sequence := e.keyBindings.GetCurrentSequence()
	return FormatSequence(sequence)
}


// AddBuffer adds a new buffer to the editor
func (e *Editor) AddBuffer(buffer *Buffer) {
	e.buffers = append(e.buffers, buffer)

	// Auto-detect and set major mode for new buffer
	if buffer.MajorMode() == nil {
		mode, err := e.modeManager.AutoDetectMajorMode(buffer)
		if err == nil {
			e.modeManager.SetMajorMode(buffer, mode.Name())
		}
	}
}

// SwitchToBuffer switches the current window to the specified buffer
func (e *Editor) SwitchToBuffer(buffer *Buffer) {
	window := e.CurrentWindow()
	if window != nil {
		window.SetBuffer(buffer)
	}
}

// FindBuffer finds a buffer by name
func (e *Editor) FindBuffer(name string) *Buffer {
	for _, buffer := range e.buffers {
		if buffer.Name() == name {
			return buffer
		}
	}
	return nil
}

// ModeManager returns the mode manager
func (e *Editor) ModeManager() *ModeManager {
	return e.modeManager
}

// PluginManager returns the plugin manager
func (e *Editor) PluginManager() PluginManagerInterface {
	return e.pluginManager
}

// SetPluginManager sets the plugin manager
func (e *Editor) SetPluginManager(pm PluginManagerInterface) {
	e.pluginManager = pm
}

// RegisterPluginCommands registers all commands from a loaded plugin
func (e *Editor) RegisterPluginCommands(plugin PluginInterface) error {
	commands := plugin.GetCommands()
	
	for _, cmdSpec := range commands {
		// Capture cmdSpec in closure to avoid loop variable issues
		cmdName := cmdSpec.Name
		
		// Convert plugin command handler string to actual command function
		cmdFunc := func(editor *Editor) error {
			// Try to cast to CommandPlugin interface for direct execution
			if cmdPlugin, ok := plugin.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
				// Direct execution via CommandPlugin interface
				return cmdPlugin.ExecuteCommand(cmdName)
			}
			
			// For regular plugins, show actual execution message
			message := "Hello from " + plugin.Name() + "! Command '" + cmdName + "' executed successfully."
			editor.SetMinibufferMessage(message)
			
			return nil
		}
		
		// Register the command in the editor's command registry
		e.commandRegistry.RegisterFunc(cmdSpec.Name, cmdFunc)
	}
	
	return nil
}

// UnregisterPluginCommands removes all commands from an unloaded plugin
func (e *Editor) UnregisterPluginCommands(plugin PluginInterface) error {
	commands := plugin.GetCommands()
	
	for _, cmdSpec := range commands {
		// Remove the command from the registry
		delete(e.commandRegistry.commands, cmdSpec.Name)
	}
	
	return nil
}

// CommandRegistry returns the command registry (for plugin integration)
func (e *Editor) CommandRegistry() *CommandRegistry {
	return e.commandRegistry
}

// RegisterPluginModes registers all modes from a loaded plugin
func (e *Editor) RegisterPluginModes(plugin PluginInterface) error {
	// Register major modes
	majorModes := plugin.GetMajorModes()
	for _, modeSpec := range majorModes {
		pluginMajorMode := NewPluginMajorMode(modeSpec, plugin)
		e.modeManager.RegisterMajorMode(pluginMajorMode)
	}
	
	// Register minor modes
	minorModes := plugin.GetMinorModes()
	for _, modeSpec := range minorModes {
		pluginMinorMode := NewPluginMinorMode(modeSpec, plugin)
		e.modeManager.RegisterMinorMode(pluginMinorMode)
	}
	
	return nil
}

// UnregisterPluginModes removes all modes from an unloaded plugin
func (e *Editor) UnregisterPluginModes(plugin PluginInterface) error {
	// Remove major modes
	majorModes := plugin.GetMajorModes()
	for _, modeSpec := range majorModes {
		delete(e.modeManager.majorModes, modeSpec.Name)
	}
	
	// Remove minor modes  
	minorModes := plugin.GetMinorModes()
	for _, modeSpec := range minorModes {
		delete(e.modeManager.minorModes, modeSpec.Name)
	}
	
	return nil
}

// RegisterPluginKeyBindings registers global key bindings from a loaded plugin
func (e *Editor) RegisterPluginKeyBindings(plugin PluginInterface) error {
	keyBindings := plugin.GetKeyBindings()
	
	for _, kbSpec := range keyBindings {
		// グローバルキーバインドのみを登録（Mode が空またはglobal）
		if kbSpec.Mode == "" || kbSpec.Mode == "global" {
			// プラグインコマンドハンドラーを作成
			cmdFunc := func(editor *Editor) error {
				// TODO: 実際のプラグインコマンドハンドラーを呼び出す実装
				// 現在はプレースホルダー
				message := "Plugin global keybinding: " + kbSpec.Command + " from " + plugin.Name()
				editor.SetMinibufferMessage(message)
				return nil
			}
			
			// エディタのグローバルキーバインドに登録
			e.keyBindings.BindKeySequence(kbSpec.Sequence, cmdFunc)
		}
	}
	
	return nil
}

// UnregisterPluginKeyBindings removes global key bindings from an unloaded plugin
func (e *Editor) UnregisterPluginKeyBindings(plugin PluginInterface) error {
	keyBindings := plugin.GetKeyBindings()
	
	// 現在のキーバインドから削除
	// 注：この実装はシンプルなため、プラグイン固有のバインディングを追跡する
	// より高度な実装では、プラグインごとのバインディング管理が必要
	
	for _, kbSpec := range keyBindings {
		if kbSpec.Mode == "" || kbSpec.Mode == "global" {
			// グローバルキーバインドマップから削除
			e.removeKeyBinding(kbSpec.Sequence)
		}
	}
	
	return nil
}

// removeKeyBinding はキーバインドマップから指定されたキーシーケンスを削除する
func (e *Editor) removeKeyBinding(sequence string) {
	e.keyBindings.RemoveSequence(sequence)
}

// KeyBindings returns the key binding map (for plugin integration and testing)
func (e *Editor) KeyBindings() *KeyBindingMap {
	return e.keyBindings
}

func (e *Editor) registerMinorModeCommands() {
	// Register auto-a-mode minor mode
	autoAMode := NewAutoAMode()
	e.modeManager.RegisterMinorMode(autoAMode)
	
	// Register the command
	e.commandRegistry.RegisterFunc("auto-a-mode", func(editor *Editor) error {
		buffer := editor.CurrentBuffer()
		if buffer == nil {
			return &ModeError{Message: "No current buffer"}
		}
		
		return editor.ModeManager().ToggleMinorMode(buffer, "auto-a-mode")
	})
}

func (e *Editor) processMinorModeHooks(buffer *Buffer, event string) {
	// Process minor mode hooks for specific events
	minorModes := buffer.MinorModes()

	for _, mode := range minorModes {
		if autoAMode, ok := mode.(*AutoAMode); ok && event == "newline" {
			autoAMode.ProcessNewline(buffer)
		}
	}
}
