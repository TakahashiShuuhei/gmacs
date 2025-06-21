package luaconfig

import (
	"sync"
	lua "github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

// HookFunction represents a function that can be called as a hook
type HookFunction func(args ...interface{}) error

// HookManager manages event hooks for Lua configuration
type HookManager struct {
	hooks map[string][]func(...interface{}) error
	mutex sync.RWMutex
}

// NewHookManager creates a new hook manager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[string][]func(...interface{}) error),
	}
}

// AddHook adds a hook function for the specified event
func (hm *HookManager) AddHook(event string, fn func(...interface{}) error) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	if hm.hooks[event] == nil {
		hm.hooks[event] = make([]func(...interface{}) error, 0)
	}
	
	hm.hooks[event] = append(hm.hooks[event], fn)
	log.Info("Added hook for event: %s (total: %d)", event, len(hm.hooks[event]))
}

// RemoveHook removes a specific hook function (not easily implemented due to function comparison)
// For now, we'll provide RemoveAllHooks for an event
func (hm *HookManager) RemoveAllHooks(event string) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	delete(hm.hooks, event)
	log.Info("Removed all hooks for event: %s", event)
}

// TriggerHook triggers all hooks for the specified event
func (hm *HookManager) TriggerHook(event string, args ...interface{}) {
	hm.mutex.RLock()
	hooks, exists := hm.hooks[event]
	hm.mutex.RUnlock()
	
	if !exists || len(hooks) == 0 {
		return
	}
	
	log.Debug("Triggering %d hooks for event: %s", len(hooks), event)
	
	for i, hook := range hooks {
		err := hook(args...)
		if err != nil {
			log.Error("Hook %d for event %s failed: %v", i, event, err)
		}
	}
}

// GetHookCount returns the number of hooks registered for an event
func (hm *HookManager) GetHookCount(event string) int {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	return len(hm.hooks[event])
}

// ListEvents returns a list of all events that have hooks
func (hm *HookManager) ListEvents() []string {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	events := make([]string, 0, len(hm.hooks))
	for event := range hm.hooks {
		events = append(events, event)
	}
	
	return events
}

// StandardEvents defines the standard hook events available in gmacs
var StandardEvents = []string{
	"before-save",        // Before saving a buffer
	"after-save",         // After saving a buffer
	"after-change",       // After text is changed in a buffer
	"before-change",      // Before text is changed in a buffer
	"key-press",          // When a key is pressed
	"buffer-create",      // When a new buffer is created
	"buffer-switch",      // When switching between buffers
	"mode-activate",      // When a mode is activated
	"mode-deactivate",    // When a mode is deactivated
	"window-create",      // When a new window is created
	"window-resize",      // When a window is resized
	"editor-startup",     // When the editor starts up
	"editor-shutdown",    // When the editor shuts down
}

// LuaHookWrapper creates a hook function that calls a Lua function
func LuaHookWrapper(L *lua.LState, fn lua.LValue) func(...interface{}) error {
	return func(args ...interface{}) error {
		// Convert Go arguments to Lua values
		luaArgs := make([]lua.LValue, len(args))
		for i, arg := range args {
			luaArgs[i] = goValueToLua(L, arg)
		}
		
		// Call the Lua function
		err := L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}, luaArgs...)
		
		if err != nil {
			return &ConfigError{Message: "Lua hook function error: " + err.Error()}
		}
		
		return nil
	}
}