package object

import "fmt"

const RASTER_OBJ Type = "Raster"

// Raster represents a 2D grid of float64 values with spatial metadata.
type Raster struct {
	Data      [][]float64 // row-major 2D grid
	Width     int
	Height    int
	Bounds    *BBox
	CRSRef    *CRS
	NoData    float64
	HasNoData bool
}

func (r *Raster) Type() Type      { return RASTER_OBJ }
func (r *Raster) Inspect() string { return fmt.Sprintf("Raster(%dx%d)", r.Width, r.Height) }
