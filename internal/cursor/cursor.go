package cursor

// Point represents a position in a buffer (line, column)
type Point struct {
	Line int // 0-indexed line number
	Col  int // 0-indexed column number
}

// Cursor represents the cursor state in a buffer
type Cursor struct {
	point       Point
	mark        *Point // for region selection
	goalColumn  int    // preferred column for vertical movement
	hasGoalCol  bool
}

// New creates a new cursor at the beginning of the buffer
func New() *Cursor {
	return &Cursor{
		point: Point{Line: 0, Col: 0},
	}
}

// Point returns the current cursor position
func (c *Cursor) Point() Point {
	return c.point
}

// SetPoint sets the cursor position
func (c *Cursor) SetPoint(line, col int) {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}
	c.point = Point{Line: line, Col: col}
	c.hasGoalCol = false
}

// Line returns the current line number
func (c *Cursor) Line() int {
	return c.point.Line
}

// Col returns the current column number
func (c *Cursor) Col() int {
	return c.point.Col
}

// SetLine sets the line number
func (c *Cursor) SetLine(line int) {
	if line < 0 {
		line = 0
	}
	c.point.Line = line
	c.hasGoalCol = false
}

// SetCol sets the column number
func (c *Cursor) SetCol(col int) {
	if col < 0 {
		col = 0
	}
	c.point.Col = col
	c.hasGoalCol = false
}

// MoveTo moves the cursor to the specified position
func (c *Cursor) MoveTo(line, col int) {
	c.SetPoint(line, col)
}

// MoveLeft moves the cursor left by one character
func (c *Cursor) MoveLeft() {
	if c.point.Col > 0 {
		c.point.Col--
	}
	c.hasGoalCol = false
}

// MoveRight moves the cursor right by one character
func (c *Cursor) MoveRight() {
	c.point.Col++
	c.hasGoalCol = false
}

// MoveUp moves the cursor up by one line
func (c *Cursor) MoveUp() {
	if c.point.Line > 0 {
		c.point.Line--
		c.restoreGoalColumn()
	}
}

// MoveDown moves the cursor down by one line
func (c *Cursor) MoveDown() {
	c.point.Line++
	c.restoreGoalColumn()
}

// SetGoalColumn sets the goal column for vertical movement
func (c *Cursor) SetGoalColumn(col int) {
	c.goalColumn = col
	c.hasGoalCol = true
}

// restoreGoalColumn restores the goal column after vertical movement
func (c *Cursor) restoreGoalColumn() {
	if c.hasGoalCol {
		c.point.Col = c.goalColumn
	}
}

// BeginningOfLine moves cursor to the beginning of the current line
func (c *Cursor) BeginningOfLine() {
	c.point.Col = 0
	c.hasGoalCol = false
}

// EndOfLine moves cursor to the end of the current line
func (c *Cursor) EndOfLine(lineLength int) {
	c.point.Col = lineLength
	c.hasGoalCol = false
}

// Mark returns the mark position (for region selection)
func (c *Cursor) Mark() *Point {
	return c.mark
}

// SetMark sets the mark at the current cursor position
func (c *Cursor) SetMark() {
	c.mark = &Point{Line: c.point.Line, Col: c.point.Col}
}

// ClearMark clears the mark
func (c *Cursor) ClearMark() {
	c.mark = nil
}

// HasMark returns whether a mark is set
func (c *Cursor) HasMark() bool {
	return c.mark != nil
}

// GetRegion returns the region between cursor and mark
func (c *Cursor) GetRegion() (start, end Point, hasRegion bool) {
	if c.mark == nil {
		return Point{}, Point{}, false
	}
	
	start = c.point
	end = *c.mark
	
	// Ensure start comes before end
	if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
		start, end = end, start
	}
	
	return start, end, true
}