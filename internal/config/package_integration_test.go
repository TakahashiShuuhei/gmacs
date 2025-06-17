package config

import (
	"testing"
	
	"github.com/yuin/gopher-lua"
	"github.com/TakahashiShuuhei/gmacs/internal/command"
	pkg "github.com/TakahashiShuuhei/gmacs/internal/package"
)

// Mock package for testing
type testPackage struct {
	info      pkg.PackageInfo
	enabled   bool
	namespace string
}

func newTestPackage(name, url, namespace string) *testPackage {
	return &testPackage{
		info: pkg.PackageInfo{
			Name: name,
			URL:  url,
		},
		namespace: namespace,
	}
}

func (tp *testPackage) GetInfo() pkg.PackageInfo {
	return tp.info
}

func (tp *testPackage) Initialize() error {
	return nil
}

func (tp *testPackage) Cleanup() error {
	return nil
}

func (tp *testPackage) IsEnabled() bool {
	return tp.enabled
}

func (tp *testPackage) Enable() error {
	tp.enabled = true
	return nil
}

func (tp *testPackage) Disable() error {
	tp.enabled = false
	return nil
}

func (tp *testPackage) ExtendLuaAPI(luaTable *lua.LTable, vm *lua.LState) error {
	// Add test function to Lua
	luaTable.RawSetString("test_function", vm.NewFunction(tp.luaTestFunction))
	return nil
}

func (tp *testPackage) GetNamespace() string {
	return tp.namespace
}

func (tp *testPackage) luaTestFunction(L *lua.LState) int {
	message := L.CheckString(1)
	L.Push(lua.LString("test: " + message))
	return 1
}


func TestLuaConfig_APIExtension(t *testing.T) {
	// Create mock editor and Lua config
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)
	
	// Create test package
	testPkg := newTestPackage("test-package", "github.com/test/test-package", "testpkg")
	
	// Register API extension
	err := luaConfig.RegisterAPIExtension(testPkg)
	if err != nil {
		t.Errorf("Failed to register API extension: %v", err)
	}
	
	// Initialize Lua VM and expose API
	luaConfig.vm = lua.NewState()
	defer luaConfig.vm.Close()
	luaConfig.exposeGmacsAPI()
	
	// Test that the extension API is available
	err = luaConfig.vm.DoString(`
		local result = gmacs.testpkg.test_function("hello")
		if result ~= "test: hello" then
			error("Expected 'test: hello', got '" .. result .. "'")
		end
	`)
	
	if err != nil {
		t.Errorf("Failed to execute Lua code with extension API: %v", err)
	}
}

func TestLuaConfig_AfterPackagesLoaded(t *testing.T) {
	// Create mock editor and Lua config
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)
	
	// Initialize Lua VM and expose API
	luaConfig.vm = lua.NewState()
	defer luaConfig.vm.Close()
	luaConfig.exposeGmacsAPI()
	
	// Test after_packages_loaded functionality
	err := luaConfig.vm.DoString(`
		gmacs.after_packages_loaded(function()
			gmacs.message("packages loaded callback executed")
		end)
	`)
	
	if err != nil {
		t.Errorf("Failed to register after_packages_loaded callback: %v", err)
	}
	
	// Check that callback was registered
	if len(luaConfig.packageLoadedCallbacks) != 1 {
		t.Errorf("Expected 1 callback, got %d", len(luaConfig.packageLoadedCallbacks))
	}
	
	// Execute callbacks
	err = luaConfig.executePackageLoadedCallbacks()
	if err != nil {
		t.Errorf("Failed to execute package loaded callbacks: %v", err)
	}
	
	// Check that message was sent
	if len(mockEditor.minibuffer.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(mockEditor.minibuffer.messages))
	}
	
	if mockEditor.minibuffer.messages[0] != "packages loaded callback executed" {
		t.Errorf("Expected 'packages loaded callback executed', got '%s'", mockEditor.minibuffer.messages[0])
	}
}

func TestLuaConfig_MultipleAPIExtensions(t *testing.T) {
	// Create mock editor and Lua config
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)
	
	// Create multiple test packages
	testPkg1 := newTestPackage("test-package-1", "github.com/test/test-package-1", "testpkg")
	testPkg2 := newTestPackage("test-package-2", "github.com/test/test-package-2", "testpkg2")
	
	// Register both API extensions
	err := luaConfig.RegisterAPIExtension(testPkg1)
	if err != nil {
		t.Errorf("Failed to register first API extension: %v", err)
	}
	
	err = luaConfig.RegisterAPIExtension(testPkg2)
	if err != nil {
		t.Errorf("Failed to register second API extension: %v", err)
	}
	
	// Initialize Lua VM and expose API
	luaConfig.vm = lua.NewState()
	defer luaConfig.vm.Close()
	luaConfig.exposeGmacsAPI()
	
	// Test that both extension APIs are available
	err = luaConfig.vm.DoString(`
		local result1 = gmacs.testpkg.test_function("from1")
		local result2 = gmacs.testpkg2.test_function("from2")
		
		if result1 ~= "test: from1" then
			error("Expected 'test: from1', got '" .. result1 .. "'")
		end
		
		if result2 ~= "test: from2" then
			error("Expected 'test: from2', got '" .. result2 .. "'")
		end
	`)
	
	if err != nil {
		t.Errorf("Failed to execute Lua code with multiple extension APIs: %v", err)
	}
}

func TestLuaConfig_UsePackage(t *testing.T) {
	// Create mock editor and Lua config
	mockEditor := &mockEditor{
		registry:   command.NewRegistry(),
		minibuffer: &mockMinibuffer{},
	}
	
	luaConfig := NewLuaConfig(mockEditor)
	
	// Initialize Lua VM and expose API
	luaConfig.vm = lua.NewState()
	defer luaConfig.vm.Close()
	luaConfig.exposeGmacsAPI()
	
	// Test use_package functionality
	err := luaConfig.vm.DoString(`
		gmacs.use_package("github.com/user/test-package", "v1.0.0")
		gmacs.use_package("github.com/user/another-package")  -- without version
	`)
	
	if err != nil {
		t.Errorf("Failed to execute use_package: %v", err)
	}
	
	// For now, use_package just logs, so no specific assertions
	// In the future, this would integrate with the package manager
}