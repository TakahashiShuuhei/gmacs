package domain

import (
	"fmt"
	"github.com/TakahashiShuuhei/gmacs/log"
	"github.com/TakahashiShuuhei/gmacs/util"
)

// ScrollUp scrolls the window up by one line
func ScrollUp(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	newScrollTop := window.ScrollTop() - 1
	window.SetScrollTop(newScrollTop)
	return nil
}

// ScrollDown scrolls the window down by one line
func ScrollDown(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	newScrollTop := window.ScrollTop() + 1
	window.SetScrollTop(newScrollTop)
	return nil
}

// ScrollLeft scrolls the window left by one character (only when line wrap is disabled)
func ScrollLeftChar(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	if window.LineWrap() {
		return nil
	}
	
	newScrollLeft := window.ScrollLeft() - 1
	window.SetScrollLeft(newScrollLeft)
	return nil
}

// ScrollRight scrolls the window right by one character (only when line wrap is disabled)
func ScrollRightChar(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	if window.LineWrap() {
		return nil
	}
	
	newScrollLeft := window.ScrollLeft() + 1
	window.SetScrollLeft(newScrollLeft)
	return nil
}

// ToggleLineWrap toggles line wrapping on/off
func ToggleLineWrap(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	newWrap := !window.LineWrap()
	window.SetLineWrap(newWrap)
	
	// Reset horizontal scroll when enabling line wrap
	if newWrap {
		window.SetScrollLeft(0)
	} else {
		// When disabling line wrap, ensure cursor is visible
		EnsureCursorVisible(editor)
	}
	
	wrapStatus := "disabled"
	if newWrap {
		wrapStatus = "enabled"
	}
	
	editor.SetMinibufferMessage("Line wrap " + wrapStatus)
	log.Info("Line wrap toggled: %s", wrapStatus)
	return nil
}

// PageUp scrolls up by one screen height
func PageUp(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	_, windowHeight := window.Size()
	
	if window.LineWrap() {
		// In line wrap mode, we need to count actual screen lines
		// For simplicity, scroll by window height in buffer lines, but ensure bounds
		currentScrollTop := window.ScrollTop()
		newScrollTop := currentScrollTop - windowHeight
		if newScrollTop < 0 {
			newScrollTop = 0
		}
		window.SetScrollTop(newScrollTop)
	} else {
		// In no-wrap mode, use traditional calculation
		newScrollTop := window.ScrollTop() - windowHeight
		window.SetScrollTop(newScrollTop)
	}
	return nil
}

// PageDown scrolls down by one screen height
func PageDown(editor *Editor) error {
	window := editor.CurrentWindow()
	if window == nil {
		return nil
	}
	
	_, windowHeight := window.Size()
	
	if window.LineWrap() {
		// In line wrap mode, we need to count actual screen lines
		// For simplicity, scroll by window height in buffer lines, but ensure bounds
		currentScrollTop := window.ScrollTop()
		buffer := editor.CurrentBuffer()
		if buffer == nil {
			return nil
		}
		
		maxScroll := len(buffer.Content()) - 1
		if maxScroll < 0 {
			maxScroll = 0
		}
		
		newScrollTop := currentScrollTop + windowHeight
		if newScrollTop > maxScroll {
			newScrollTop = maxScroll
		}
		window.SetScrollTop(newScrollTop)
	} else {
		// In no-wrap mode, use traditional calculation
		newScrollTop := window.ScrollTop() + windowHeight
		window.SetScrollTop(newScrollTop)
	}
	return nil
}

// EnsureCursorVisible adjusts scrolling to make sure cursor is visible
func EnsureCursorVisible(editor *Editor) error {
	window := editor.CurrentWindow()
	buffer := editor.CurrentBuffer()
	if window == nil || buffer == nil {
		return nil
	}
	
	// Get the actual screen position of the cursor
	log.Info("SCROLL_TIMING: EnsureCursorVisible called - getting cursor screen position")
	screenRow, screenCol := window.CursorPosition()
	log.Info("SCROLL_TIMING: EnsureCursorVisible - cursor screen position: (%d,%d)", screenRow, screenCol)
	
	if window.LineWrap() {
		// In line wrap mode, work with screen rows properly
		_, windowHeight := window.Size()
		
		if screenRow < 0 {
			// Cursor is above visible area - need to scroll up
			// Find a scroll position that makes the cursor visible
			cursorPos := buffer.Cursor()
			newScrollTop := cursorPos.Row
			
			// Make sure this doesn't scroll too far up
			if newScrollTop < 0 {
				newScrollTop = 0
			}
			
			log.Info("SCROLL_TIMING: EnsureCursorVisible (wrap mode) scrolling UP from %d to %d (cursor above visible)", window.ScrollTop(), newScrollTop)
			window.SetScrollTop(newScrollTop)
		} else if screenRow >= windowHeight {
			// Cursor is below visible area - need to scroll down
			// We need to find a scroll position where the cursor row will be visible
			cursorPos := buffer.Cursor()
			
			// Start from current scroll position and increment until cursor is visible
			oldScrollTop := window.ScrollTop()
			maxScrollTop := len(buffer.Content()) - 1
			if maxScrollTop < 0 {
				maxScrollTop = 0
			}
			
			// Try increasing scroll position until cursor is in visible area
			for newScrollTop := oldScrollTop + 1; newScrollTop <= maxScrollTop && newScrollTop <= cursorPos.Row; newScrollTop++ {
				window.SetScrollTop(newScrollTop)
				newScreenRow, _ := window.CursorPosition()
				if newScreenRow >= 0 && newScreenRow < windowHeight {  // Just make cursor visible
					log.Info("SCROLL_TIMING: EnsureCursorVisible (wrap mode) scrolling DOWN from %d to %d (cursor below visible)", oldScrollTop, newScrollTop)
					break
				}
			}
		}
	} else {
		// No line wrap mode - simpler logic
		cursorPos := buffer.Cursor()
		
		// Vertical scrolling with improved strategy
		if cursorPos.Row < window.ScrollTop() {
			log.Info("SCROLL_TIMING: EnsureCursorVisible (no wrap) scrolling UP from %d to %d (cursor above visible)", window.ScrollTop(), cursorPos.Row)
			window.SetScrollTop(cursorPos.Row)
		} else if cursorPos.Row >= window.ScrollTop()+window.height {
			// Cursor is beyond visible area - minimum scroll to make it visible
			newScrollTop := window.ScrollTop() + 1  // Scroll just enough to make cursor visible
			log.Info("SCROLL_TIMING: EnsureCursorVisible (no wrap) minimal scroll from %d to %d (cursor beyond visible)", window.ScrollTop(), newScrollTop)
			window.SetScrollTop(newScrollTop)
		}
		
		// Horizontal scrolling
		windowWidth, _ := window.Size()
		cursorPos = buffer.Cursor()
		
		// Calculate the actual display column for the cursor
		if cursorPos.Row < len(buffer.Content()) {
			line := buffer.Content()[cursorPos.Row]
			if cursorPos.Col <= len(line) {
				displayCol := util.StringWidthUpTo(line, cursorPos.Col)
				
				// Ensure cursor is visible horizontally
				if displayCol < window.ScrollLeft() {
					// Cursor is left of visible area - scroll left
					newScrollLeft := displayCol
					if newScrollLeft < 0 {
						newScrollLeft = 0
					}
					log.Info("HORIZONTAL_SCROLL: Scrolling left to %d (cursor at displayCol %d)", newScrollLeft, displayCol)
					window.SetScrollLeft(newScrollLeft)
				} else if displayCol >= window.ScrollLeft()+windowWidth {
					// Cursor is right of visible area - scroll right
					newScrollLeft := displayCol - windowWidth + 1
					if newScrollLeft < 0 {
						newScrollLeft = 0
					}
					log.Info("HORIZONTAL_SCROLL: Scrolling right to %d (cursor at displayCol %d)", newScrollLeft, displayCol)
					window.SetScrollLeft(newScrollLeft)
				}
			}
		}
	}
	
	return nil
}

// ShowDebugInfo displays debug information about window and cursor state
func ShowDebugInfo(editor *Editor) error {
	window := editor.CurrentWindow()
	buffer := editor.CurrentBuffer()
	if window == nil || buffer == nil {
		return nil
	}
	
	windowWidth, windowHeight := window.Size()
	cursor := buffer.Cursor()
	scrollTop := window.ScrollTop()
	scrollLeft := window.ScrollLeft()
	lineWrap := window.LineWrap()
	screenRow, screenCol := window.CursorPosition()
	bufferLines := len(buffer.Content())
	
	debugMsg := fmt.Sprintf("Window: %dx%d, Cursor: buf(%d,%d) scr(%d,%d), Scroll: (%d,%d), Lines: %d, Wrap: %t", 
		windowWidth, windowHeight, cursor.Row, cursor.Col, screenRow, screenCol, 
		scrollTop, scrollLeft, bufferLines, lineWrap)
	
	editor.SetMinibufferMessage(debugMsg)
	log.Info("Debug info: %s", debugMsg)
	return nil
}