package domain

import (
	"github.com/TakahashiShuuhei/gmacs/core/log"
)

// SplitWindowRight implements the split-window-right command (C-x 3)
func SplitWindowRight(editor *Editor) error {
	if editor.layout == nil {
		log.Warn("No layout available for split-window-right")
		return nil
	}
	
	newWindow := editor.layout.SplitWindowRight()
	if newWindow == nil {
		log.Warn("Failed to split window right")
		return nil
	}
	
	// Add the new buffer to editor's buffer list
	newBuffer := newWindow.Buffer()
	if newBuffer != nil {
		editor.AddBuffer(newBuffer)
		log.Info("Split window right - created new window with buffer: %s", newBuffer.Name())
	}
	
	return nil
}

// SplitWindowBelow implements the split-window-below command (C-x 2)
func SplitWindowBelow(editor *Editor) error {
	if editor.layout == nil {
		log.Warn("No layout available for split-window-below")
		return nil
	}
	
	newWindow := editor.layout.SplitWindowBelow()
	if newWindow == nil {
		log.Warn("Failed to split window below")
		return nil
	}
	
	// Add the new buffer to editor's buffer list
	newBuffer := newWindow.Buffer()
	if newBuffer != nil {
		editor.AddBuffer(newBuffer)
		log.Info("Split window below - created new window with buffer: %s", newBuffer.Name())
	}
	
	return nil
}

// OtherWindow implements the other-window command (C-x o)
func OtherWindow(editor *Editor) error {
	if editor.layout == nil {
		log.Warn("No layout available for other-window")
		return nil
	}
	
	editor.layout.NextWindow()
	currentWindow := editor.CurrentWindow()
	if currentWindow != nil && currentWindow.Buffer() != nil {
		log.Info("Switched to window with buffer: %s", currentWindow.Buffer().Name())
	}
	
	return nil
}

// DeleteWindow implements the delete-window command (C-x 0)
func DeleteWindow(editor *Editor) error {
	if editor.layout == nil {
		log.Warn("No layout available for delete-window")
		return nil
	}
	
	success := editor.layout.DeleteCurrentWindow()
	if success {
		log.Info("Deleted current window")
	} else {
		log.Warn("Cannot delete the only window")
	}
	
	return nil
}

// DeleteOtherWindows implements the delete-other-windows command (C-x 1)
func DeleteOtherWindows(editor *Editor) error {
	if editor.layout == nil {
		log.Warn("No layout available for delete-other-windows")
		return nil
	}
	
	editor.layout.DeleteOtherWindows()
	log.Info("Deleted all other windows")
	
	return nil
}