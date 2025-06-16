package command

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// InteractiveFunc represents an interactive function (like Emacs interactive functions)
type InteractiveFunc struct {
	Name        string
	Description string
	Handler     func(args ...interface{}) error
	ArgSpec     string // Argument specification (like Emacs interactive spec)
}

// Registry manages all registered interactive functions
type Registry struct {
	mu        sync.RWMutex
	functions map[string]*InteractiveFunc
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]*InteractiveFunc),
	}
}

// Register registers a new interactive function
func (r *Registry) Register(name, description, argSpec string, handler func(args ...interface{}) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if name == "" {
		return errors.New("command name cannot be empty")
	}
	
	if handler == nil {
		return errors.New("command handler cannot be nil")
	}
	
	if _, exists := r.functions[name]; exists {
		return fmt.Errorf("command %s already exists", name)
	}
	
	r.functions[name] = &InteractiveFunc{
		Name:        name,
		Description: description,
		Handler:     handler,
		ArgSpec:     argSpec,
	}
	
	return nil
}

// Unregister removes an interactive function
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.functions[name]; !exists {
		return fmt.Errorf("command %s not found", name)
	}
	
	delete(r.functions, name)
	return nil
}

// Get retrieves an interactive function by name
func (r *Registry) Get(name string) (*InteractiveFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	fn, exists := r.functions[name]
	return fn, exists
}

// Execute executes an interactive function by name
func (r *Registry) Execute(name string, args ...interface{}) error {
	fn, exists := r.Get(name)
	if !exists {
		return fmt.Errorf("command %s not found", name)
	}
	
	return fn.Handler(args...)
}

// List returns all registered command names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.functions))
	for name := range r.functions {
		names = append(names, name)
	}
	
	sort.Strings(names)
	return names
}

// ListWithPrefix returns all command names that start with the given prefix
func (r *Registry) ListWithPrefix(prefix string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var matches []string
	for name := range r.functions {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}
	
	sort.Strings(matches)
	return matches
}

// GetAll returns all registered functions
func (r *Registry) GetAll() map[string]*InteractiveFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make(map[string]*InteractiveFunc)
	for name, fn := range r.functions {
		result[name] = fn
	}
	
	return result
}

// Count returns the number of registered functions
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return len(r.functions)
}

// Global registry instance
var globalRegistry = NewRegistry()

// Register registers a function in the global registry
func Register(name, description, argSpec string, handler func(args ...interface{}) error) error {
	return globalRegistry.Register(name, description, argSpec, handler)
}

// Execute executes a function from the global registry
func Execute(name string, args ...interface{}) error {
	return globalRegistry.Execute(name, args...)
}

// Get retrieves a function from the global registry
func Get(name string) (*InteractiveFunc, bool) {
	return globalRegistry.Get(name)
}

// List returns all registered command names from the global registry
func List() []string {
	return globalRegistry.List()
}

// ListWithPrefix returns command names with prefix from the global registry
func ListWithPrefix(prefix string) []string {
	return globalRegistry.ListWithPrefix(prefix)
}

// GetGlobalRegistry returns the global registry instance
func GetGlobalRegistry() *Registry {
	return globalRegistry
}