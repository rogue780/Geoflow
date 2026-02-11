// Package index provides spatial indexing structures for the GeoFlow language.
package index

import (
	"fmt"
	"math"
	"sort"

	"github.com/rogue780/geoflow/internal/object"
)

// geomBounds extracts the bounding box from any geometry object.
// Returns (minX, minY, maxX, maxY, ok).
func geomBounds(obj object.Object) (minX, minY, maxX, maxY float64, ok bool) {
	switch g := obj.(type) {
	case *object.Point:
		return g.Coord.X, g.Coord.Y, g.Coord.X, g.Coord.Y, true
	case *object.BBox:
		return g.MinX, g.MinY, g.MaxX, g.MaxY, true
	case object.Geometry:
		coords := g.Coordinates()
		if len(coords) == 0 {
			return 0, 0, 0, 0, false
		}
		minX, minY = coords[0].X, coords[0].Y
		maxX, maxY = coords[0].X, coords[0].Y
		for _, c := range coords[1:] {
			if c.X < minX {
				minX = c.X
			}
			if c.Y < minY {
				minY = c.Y
			}
			if c.X > maxX {
				maxX = c.X
			}
			if c.Y > maxY {
				maxY = c.Y
			}
		}
		return minX, minY, maxX, maxY, true
	default:
		return 0, 0, 0, 0, false
	}
}

// bboxIntersects tests whether two bounding boxes overlap.
func bboxIntersects(a, b object.RTreeEntry) bool {
	return a.MinX <= b.MaxX && a.MaxX >= b.MinX &&
		a.MinY <= b.MaxY && a.MaxY >= b.MinY
}

// toFloat extracts a float64 from an Integer or Float object.
func toFloat(obj object.Object) (float64, bool) {
	switch v := obj.(type) {
	case *object.Integer:
		return float64(v.Value), true
	case *object.Float:
		return v.Value, true
	default:
		return 0, false
	}
}

// GetExports returns all exported functions for the spatial index module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"new":     &object.Builtin{Name: "index.new", Fn: rtreeNew},
		"insert":  &object.Builtin{Name: "index.insert", Fn: rtreeInsert},
		"query":   &object.Builtin{Name: "index.query", Fn: rtreeQuery},
		"nearest": &object.Builtin{Name: "index.nearest", Fn: rtreeNearest},
		"size":    &object.Builtin{Name: "index.size", Fn: rtreeSize},
	}
}

// rtreeNew creates a new empty RTree.
func rtreeNew(args ...object.Object) object.Object {
	if len(args) != 0 {
		return &object.Error{Message: "index.new expects 0 arguments"}
	}
	return &object.RTree{Entries: []object.RTreeEntry{}}
}

// rtreeInsert returns a new RTree with the given entry appended.
// Arguments: rtree, geometry, item
func rtreeInsert(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "index.insert expects 3 arguments (rtree, geometry, item)"}
	}
	rt, ok := args[0].(*object.RTree)
	if !ok {
		return &object.Error{Message: "index.insert: first argument must be an RTree"}
	}
	minX, minY, maxX, maxY, ok := geomBounds(args[1])
	if !ok {
		return &object.Error{Message: "index.insert: second argument must be a geometry"}
	}
	entry := object.RTreeEntry{
		MinX: minX,
		MinY: minY,
		MaxX: maxX,
		MaxY: maxY,
		Item: args[2],
	}
	// Create a new slice to preserve immutability of the original RTree.
	newEntries := make([]object.RTreeEntry, len(rt.Entries)+1)
	copy(newEntries, rt.Entries)
	newEntries[len(rt.Entries)] = entry
	return &object.RTree{Entries: newEntries}
}

// rtreeQuery returns a List of items whose bounding boxes intersect the query bbox.
// Arguments: rtree, bbox
func rtreeQuery(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "index.query expects 2 arguments (rtree, bbox)"}
	}
	rt, ok := args[0].(*object.RTree)
	if !ok {
		return &object.Error{Message: "index.query: first argument must be an RTree"}
	}
	qMinX, qMinY, qMaxX, qMaxY, ok := geomBounds(args[1])
	if !ok {
		return &object.Error{Message: "index.query: second argument must be a geometry or BBox"}
	}
	queryEntry := object.RTreeEntry{
		MinX: qMinX,
		MinY: qMinY,
		MaxX: qMaxX,
		MaxY: qMaxY,
	}
	var results []object.Object
	for _, entry := range rt.Entries {
		if bboxIntersects(entry, queryEntry) {
			results = append(results, entry.Item)
		}
	}
	if results == nil {
		results = []object.Object{}
	}
	return &object.List{Elements: results}
}

// rtreeNearest returns a List of the n nearest items by bounding box center distance.
// Arguments: rtree, point, n
func rtreeNearest(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "index.nearest expects 3 arguments (rtree, point, n)"}
	}
	rt, ok := args[0].(*object.RTree)
	if !ok {
		return &object.Error{Message: "index.nearest: first argument must be an RTree"}
	}
	px, py, ok := extractPoint(args[1])
	if !ok {
		return &object.Error{Message: "index.nearest: second argument must be a Point"}
	}
	n, ok2 := args[2].(*object.Integer)
	if !ok2 || n.Value < 0 {
		return &object.Error{Message: "index.nearest: third argument must be a non-negative integer"}
	}
	count := int(n.Value)
	if count == 0 || len(rt.Entries) == 0 {
		return &object.List{Elements: []object.Object{}}
	}

	type distEntry struct {
		dist float64
		item object.Object
	}
	entries := make([]distEntry, len(rt.Entries))
	for i, e := range rt.Entries {
		cx := (e.MinX + e.MaxX) / 2
		cy := (e.MinY + e.MaxY) / 2
		dx := px - cx
		dy := py - cy
		entries[i] = distEntry{dist: math.Sqrt(dx*dx + dy*dy), item: e.Item}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].dist < entries[j].dist
	})
	if count > len(entries) {
		count = len(entries)
	}
	results := make([]object.Object, count)
	for i := 0; i < count; i++ {
		results[i] = entries[i].item
	}
	return &object.List{Elements: results}
}

// rtreeSize returns the number of entries in the RTree.
func rtreeSize(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "index.size expects 1 argument (rtree)"}
	}
	rt, ok := args[0].(*object.RTree)
	if !ok {
		return &object.Error{Message: fmt.Sprintf("index.size: argument must be an RTree, got %s", args[0].Type())}
	}
	return &object.Integer{Value: int64(len(rt.Entries))}
}

// extractPoint extracts x,y coordinates from a Point object.
func extractPoint(obj object.Object) (x, y float64, ok bool) {
	p, ok := obj.(*object.Point)
	if !ok {
		return 0, 0, false
	}
	return p.Coord.X, p.Coord.Y, true
}
