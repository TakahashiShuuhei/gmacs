package domain

// SplitType represents the type of window split
type SplitType int

const (
	SplitNone SplitType = iota
	SplitHorizontal // 上下分割 (split-window-below)
	SplitVertical   // 左右分割 (split-window-right)
)

// WindowLayoutNode represents a node in the window layout tree
type WindowLayoutNode struct {
	// Split information
	SplitType  SplitType
	SplitRatio float64 // 分割比率 (0.0-1.0)
	
	// Child nodes (if split)
	Left  *WindowLayoutNode
	Right *WindowLayoutNode
	
	// Leaf node (actual window)
	Window *Window
	
	// Display position and size
	X      int
	Y      int
	Width  int
	Height int
}

// WindowLayout manages the overall window layout
type WindowLayout struct {
	root        *WindowLayoutNode
	activeNode  *WindowLayoutNode
	totalWidth  int
	totalHeight int
}

// NewWindowLayout creates a new window layout with a single window
func NewWindowLayout(window *Window, width, height int) *WindowLayout {
	node := &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    window,
		X:         0,
		Y:         0,
		Width:     width,
		Height:    height,
	}
	
	return &WindowLayout{
		root:        node,
		activeNode:  node,
		totalWidth:  width,
		totalHeight: height,
	}
}

// CurrentWindow returns the currently active window
func (wl *WindowLayout) CurrentWindow() *Window {
	if wl.activeNode != nil && wl.activeNode.Window != nil {
		return wl.activeNode.Window
	}
	return nil
}

// SetActiveWindow sets the active window by finding the node containing the window
func (wl *WindowLayout) SetActiveWindow(window *Window) {
	wl.activeNode = wl.findNodeByWindow(wl.root, window)
}

// findNodeByWindow recursively searches for a node containing the specified window
func (wl *WindowLayout) findNodeByWindow(node *WindowLayoutNode, window *Window) *WindowLayoutNode {
	if node == nil {
		return nil
	}
	
	if node.Window == window {
		return node
	}
	
	if left := wl.findNodeByWindow(node.Left, window); left != nil {
		return left
	}
	
	if right := wl.findNodeByWindow(node.Right, window); right != nil {
		return right
	}
	
	return nil
}

// IsLeaf returns true if the node is a leaf (contains a window)
func (node *WindowLayoutNode) IsLeaf() bool {
	return node.Window != nil
}

// IsSplit returns true if the node is split (has child nodes)
func (node *WindowLayoutNode) IsSplit() bool {
	return node.Left != nil && node.Right != nil
}

// Resize updates the total size and recalculates layout
func (wl *WindowLayout) Resize(width, height int) {
	wl.totalWidth = width
	wl.totalHeight = height
	// Recalculate layout (will be implemented later)
	wl.calculateLayout()
}

// calculateLayout recursively calculates position and size for all nodes
func (wl *WindowLayout) calculateLayout() {
	if wl.root != nil {
		wl.calculateNodeLayout(wl.root, 0, 0, wl.totalWidth, wl.totalHeight)
	}
}

// calculateNodeLayout recursively calculates layout for a node and its children
func (wl *WindowLayout) calculateNodeLayout(node *WindowLayoutNode, x, y, width, height int) {
	node.X = x
	node.Y = y
	node.Width = width
	node.Height = height
	
	if node.IsLeaf() {
		// Leaf node: resize the window
		// Reserve 2 lines for mode line and minibuffer
		contentHeight := height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}
		if node.Window != nil {
			node.Window.Resize(width, contentHeight)
		}
		return
	}
	
	// Split node: calculate child sizes
	if node.SplitType == SplitVertical {
		// Left-right split
		leftWidth := int(float64(width) * node.SplitRatio)
		rightWidth := width - leftWidth
		
		if node.Left != nil {
			wl.calculateNodeLayout(node.Left, x, y, leftWidth, height)
		}
		if node.Right != nil {
			wl.calculateNodeLayout(node.Right, x+leftWidth, y, rightWidth, height)
		}
	} else if node.SplitType == SplitHorizontal {
		// Top-bottom split
		topHeight := int(float64(height) * node.SplitRatio)
		bottomHeight := height - topHeight
		
		if node.Left != nil {
			wl.calculateNodeLayout(node.Left, x, y, width, topHeight)
		}
		if node.Right != nil {
			wl.calculateNodeLayout(node.Right, x, y+topHeight, width, bottomHeight)
		}
	}
}

// GetAllWindows returns all windows in the layout
func (wl *WindowLayout) GetAllWindows() []*Window {
	var windows []*Window
	wl.collectWindows(wl.root, &windows)
	return windows
}

// collectWindows recursively collects all windows from the layout tree
func (wl *WindowLayout) collectWindows(node *WindowLayoutNode, windows *[]*Window) {
	if node == nil {
		return
	}
	
	if node.IsLeaf() && node.Window != nil {
		*windows = append(*windows, node.Window)
		return
	}
	
	wl.collectWindows(node.Left, windows)
	wl.collectWindows(node.Right, windows)
}

// SplitWindowRight splits the current window vertically (left-right)
func (wl *WindowLayout) SplitWindowRight() *Window {
	if wl.activeNode == nil || !wl.activeNode.IsLeaf() {
		return nil
	}
	
	// Create new buffer and window
	newBuffer := NewBuffer("*scratch*")
	newWindow := NewWindow(newBuffer, 0, 0) // Size will be calculated later
	
	// Create new node for the new window
	newNode := &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    newWindow,
	}
	
	// Convert current leaf node to split node
	oldWindow := wl.activeNode.Window
	oldNode := &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    oldWindow,
	}
	
	// Update current node to become a split node
	wl.activeNode.SplitType = SplitVertical
	wl.activeNode.SplitRatio = 0.5 // 50-50 split
	wl.activeNode.Left = oldNode
	wl.activeNode.Right = newNode
	wl.activeNode.Window = nil // No longer a leaf
	
	// Set new window as active
	wl.activeNode = newNode
	
	// Recalculate layout
	wl.calculateLayout()
	
	return newWindow
}

// SplitWindowBelow splits the current window horizontally (top-bottom)
func (wl *WindowLayout) SplitWindowBelow() *Window {
	if wl.activeNode == nil || !wl.activeNode.IsLeaf() {
		return nil
	}
	
	// Create new buffer and window
	newBuffer := NewBuffer("*scratch*")
	newWindow := NewWindow(newBuffer, 0, 0) // Size will be calculated later
	
	// Create new node for the new window
	newNode := &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    newWindow,
	}
	
	// Convert current leaf node to split node
	oldWindow := wl.activeNode.Window
	oldNode := &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    oldWindow,
	}
	
	// Update current node to become a split node
	wl.activeNode.SplitType = SplitHorizontal
	wl.activeNode.SplitRatio = 0.5 // 50-50 split
	wl.activeNode.Left = oldNode  // Top window
	wl.activeNode.Right = newNode // Bottom window
	wl.activeNode.Window = nil    // No longer a leaf
	
	// Set new window as active
	wl.activeNode = newNode
	
	// Recalculate layout
	wl.calculateLayout()
	
	return newWindow
}

// NextWindow switches to the next window in the layout
func (wl *WindowLayout) NextWindow() {
	windows := wl.GetAllWindows()
	if len(windows) <= 1 {
		return // Only one window
	}
	
	// Find current window index
	currentWindow := wl.CurrentWindow()
	currentIndex := -1
	for i, window := range windows {
		if window == currentWindow {
			currentIndex = i
			break
		}
	}
	
	// Move to next window
	nextIndex := (currentIndex + 1) % len(windows)
	wl.SetActiveWindow(windows[nextIndex])
}

// DeleteCurrentWindow deletes the current window and returns true if successful
func (wl *WindowLayout) DeleteCurrentWindow() bool {
	if wl.activeNode == nil || wl.root == wl.activeNode {
		return false // Cannot delete the root/only window
	}
	
	// Find parent of current node
	parent := wl.findParent(wl.root, wl.activeNode)
	if parent == nil {
		return false
	}
	
	// Determine which child to keep
	var keepChild *WindowLayoutNode
	if parent.Left == wl.activeNode {
		keepChild = parent.Right
	} else {
		keepChild = parent.Left
	}
	
	// Replace parent with the kept child
	parent.SplitType = keepChild.SplitType
	parent.SplitRatio = keepChild.SplitRatio
	parent.Left = keepChild.Left
	parent.Right = keepChild.Right
	parent.Window = keepChild.Window
	
	// Set active node to the kept child
	if keepChild.IsLeaf() {
		wl.activeNode = parent
	} else {
		// Find first leaf in the kept subtree
		wl.activeNode = wl.findFirstLeaf(parent)
	}
	
	// Recalculate layout
	wl.calculateLayout()
	
	return true
}

// DeleteOtherWindows deletes all windows except the current one
func (wl *WindowLayout) DeleteOtherWindows() {
	if wl.activeNode == nil {
		return
	}
	
	currentWindow := wl.CurrentWindow()
	if currentWindow == nil {
		return
	}
	
	// Replace root with a single window node
	wl.root = &WindowLayoutNode{
		SplitType: SplitNone,
		Window:    currentWindow,
		X:         0,
		Y:         0,
		Width:     wl.totalWidth,
		Height:    wl.totalHeight,
	}
	
	wl.activeNode = wl.root
	
	// Recalculate layout
	wl.calculateLayout()
}

// findParent finds the parent node of the target node
func (wl *WindowLayout) findParent(node, target *WindowLayoutNode) *WindowLayoutNode {
	if node == nil || node == target {
		return nil
	}
	
	if node.Left == target || node.Right == target {
		return node
	}
	
	if parent := wl.findParent(node.Left, target); parent != nil {
		return parent
	}
	
	return wl.findParent(node.Right, target)
}

// findFirstLeaf finds the first leaf node in the subtree
func (wl *WindowLayout) findFirstLeaf(node *WindowLayoutNode) *WindowLayoutNode {
	if node == nil {
		return nil
	}
	
	if node.IsLeaf() {
		return node
	}
	
	if left := wl.findFirstLeaf(node.Left); left != nil {
		return left
	}
	
	return wl.findFirstLeaf(node.Right)
}

// GetAllWindowNodes returns all window nodes (leaf nodes) for rendering
func (wl *WindowLayout) GetAllWindowNodes() []*WindowLayoutNode {
	var nodes []*WindowLayoutNode
	wl.collectWindowNodes(wl.root, &nodes)
	return nodes
}

// collectWindowNodes recursively collects all leaf nodes
func (wl *WindowLayout) collectWindowNodes(node *WindowLayoutNode, nodes *[]*WindowLayoutNode) {
	if node == nil {
		return
	}
	
	if node.IsLeaf() {
		*nodes = append(*nodes, node)
		return
	}
	
	wl.collectWindowNodes(node.Left, nodes)
	wl.collectWindowNodes(node.Right, nodes)
}