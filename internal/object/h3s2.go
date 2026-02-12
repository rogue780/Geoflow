package object

import "fmt"

const H3INDEX_OBJ Type = "H3Index"

// H3Index represents an H3 hierarchical hexagonal grid index.
type H3Index struct {
	Value      uint64
	Resolution int
}

func (h *H3Index) Type() Type      { return H3INDEX_OBJ }
func (h *H3Index) Inspect() string { return fmt.Sprintf("H3(%d, res=%d)", h.Value, h.Resolution) }

const S2CELLID_OBJ Type = "S2CellId"

// S2CellId represents an S2 geometry cell identifier.
type S2CellId struct {
	Value uint64
	Level int
}

func (s *S2CellId) Type() Type      { return S2CELLID_OBJ }
func (s *S2CellId) Inspect() string { return fmt.Sprintf("S2(%d, level=%d)", s.Value, s.Level) }
