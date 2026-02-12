// Package analysis provides spatial analysis functions for the GeoFlow language.
package analysis

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/rogue780/geoflow/internal/object"
)

// GetExports returns all exported functions for the geo.analysis module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		"voronoi":            &object.Builtin{Name: "analysis.voronoi", Fn: voronoi},
		"delaunay":           &object.Builtin{Name: "analysis.delaunay", Fn: delaunay},
		"nearestNeighbor":    &object.Builtin{Name: "analysis.nearestNeighbor", Fn: nearestNeighbor},
		"kNearestNeighbors":  &object.Builtin{Name: "analysis.kNearestNeighbors", Fn: kNearestNeighbors},
		"idw":                &object.Builtin{Name: "analysis.idw", Fn: idw},
		"dbscan":             &object.Builtin{Name: "analysis.dbscan", Fn: dbscan},
		"kmeans":             &object.Builtin{Name: "analysis.kmeans", Fn: kmeans},
		"concaveHull":        &object.Builtin{Name: "analysis.concaveHull", Fn: concaveHull},
		"alphaShape":         &object.Builtin{Name: "analysis.alphaShape", Fn: alphaShape},
	}
}

// ── Helper functions ──

// extractPoints converts a List of Point objects to a slice of coordinates.
func extractPoints(list *object.List) ([]*object.Point, error) {
	points := make([]*object.Point, len(list.Elements))
	for i, elem := range list.Elements {
		pt, ok := elem.(*object.Point)
		if !ok {
			return nil, fmt.Errorf("element at index %d is %s, not a Point", i, elem.Type())
		}
		points[i] = pt
	}
	return points, nil
}

// euclideanDist computes the Euclidean distance between two points.
func euclideanDist(a, b *object.Point) float64 {
	dx := a.Coord.X - b.Coord.X
	dy := a.Coord.Y - b.Coord.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// pointsToList converts a slice of Points to a List object.
func pointsToList(pts []*object.Point) *object.List {
	elems := make([]object.Object, len(pts))
	for i, p := range pts {
		elems[i] = p
	}
	return &object.List{Elements: elems}
}

// ── nearestNeighbor: brute-force nearest neighbor ──

func nearestNeighbor(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "analysis.nearestNeighbor expects 2 arguments (point, candidates)"}
	}
	pt, ok := args[0].(*object.Point)
	if !ok {
		return &object.Error{Message: "analysis.nearestNeighbor: first argument must be a Point"}
	}
	list, ok := args[1].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.nearestNeighbor: second argument must be a List of Points"}
	}

	candidates, err := extractPoints(list)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.nearestNeighbor: %s", err)}
	}
	if len(candidates) == 0 {
		return &object.Error{Message: "analysis.nearestNeighbor: candidates list is empty"}
	}

	best := candidates[0]
	bestDist := euclideanDist(pt, best)
	for _, c := range candidates[1:] {
		d := euclideanDist(pt, c)
		if d < bestDist {
			bestDist = d
			best = c
		}
	}
	return best
}

// ── kNearestNeighbors: brute-force k-NN ──

func kNearestNeighbors(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "analysis.kNearestNeighbors expects 3 arguments (point, candidates, k)"}
	}
	pt, ok := args[0].(*object.Point)
	if !ok {
		return &object.Error{Message: "analysis.kNearestNeighbors: first argument must be a Point"}
	}
	list, ok := args[1].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.kNearestNeighbors: second argument must be a List of Points"}
	}
	kObj, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "analysis.kNearestNeighbors: third argument must be an integer"}
	}
	k := int(kObj.Value)

	candidates, err := extractPoints(list)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.kNearestNeighbors: %s", err)}
	}
	if len(candidates) == 0 {
		return &object.List{Elements: []object.Object{}}
	}
	if k <= 0 {
		return &object.List{Elements: []object.Object{}}
	}
	if k > len(candidates) {
		k = len(candidates)
	}

	// Sort by distance from pt
	type distEntry struct {
		point *object.Point
		dist  float64
	}
	entries := make([]distEntry, len(candidates))
	for i, c := range candidates {
		entries[i] = distEntry{point: c, dist: euclideanDist(pt, c)}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].dist < entries[j].dist
	})

	result := make([]*object.Point, k)
	for i := 0; i < k; i++ {
		result[i] = entries[i].point
	}
	return pointsToList(result)
}

// ── IDW: Inverse Distance Weighting interpolation ──

func idw(args ...object.Object) object.Object {
	if len(args) < 3 || len(args) > 4 {
		return &object.Error{Message: "analysis.idw expects 3-4 arguments (points, values, target, [power=2.0])"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.idw: first argument must be a List of Points"}
	}
	valuesList, ok := args[1].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.idw: second argument must be a List of numbers"}
	}
	target, ok := args[2].(*object.Point)
	if !ok {
		return &object.Error{Message: "analysis.idw: third argument must be a Point"}
	}

	power := 2.0
	if len(args) == 4 {
		p, err := toFloat64(args[3])
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("analysis.idw: power: %s", err)}
		}
		power = p
	}

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.idw: points: %s", err)}
	}

	if len(points) != len(valuesList.Elements) {
		return &object.Error{Message: "analysis.idw: points and values lists must have the same length"}
	}
	if len(points) == 0 {
		return &object.Error{Message: "analysis.idw: points list is empty"}
	}

	values := make([]float64, len(valuesList.Elements))
	for i, v := range valuesList.Elements {
		val, err := toFloat64(v)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("analysis.idw: values[%d]: %s", i, err)}
		}
		values[i] = val
	}

	// IDW formula: sum(w_i * v_i) / sum(w_i) where w_i = 1/d_i^p
	var weightSum, valueSum float64
	for i, pt := range points {
		d := euclideanDist(target, pt)
		if d == 0 {
			// Target is exactly on a known point
			return &object.Float{Value: values[i]}
		}
		w := 1.0 / math.Pow(d, power)
		weightSum += w
		valueSum += w * values[i]
	}

	if weightSum == 0 {
		return &object.Float{Value: 0}
	}
	return &object.Float{Value: valueSum / weightSum}
}

// ── DBSCAN clustering ──

func dbscan(args ...object.Object) object.Object {
	if len(args) != 3 {
		return &object.Error{Message: "analysis.dbscan expects 3 arguments (points, eps, minPts)"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.dbscan: first argument must be a List of Points"}
	}
	epsVal, err := toFloat64(args[1])
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.dbscan: eps: %s", err)}
	}
	minPtsObj, ok := args[2].(*object.Integer)
	if !ok {
		return &object.Error{Message: "analysis.dbscan: minPts must be an integer"}
	}
	minPts := int(minPtsObj.Value)

	points, pErr := extractPoints(pointsList)
	if pErr != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.dbscan: %s", pErr)}
	}

	n := len(points)
	if n == 0 {
		return &object.List{Elements: []object.Object{}}
	}

	const (
		undefined = 0
		noise     = -1
	)
	labels := make([]int, n)
	clusterID := 0

	// regionQuery returns indices of points within eps distance of point at index p
	regionQuery := func(p int) []int {
		var neighbors []int
		for i := 0; i < n; i++ {
			if euclideanDist(points[p], points[i]) <= epsVal {
				neighbors = append(neighbors, i)
			}
		}
		return neighbors
	}

	for i := 0; i < n; i++ {
		if labels[i] != undefined {
			continue
		}

		neighbors := regionQuery(i)
		if len(neighbors) < minPts {
			labels[i] = noise
			continue
		}

		clusterID++
		labels[i] = clusterID

		// Seed set (excluding i itself)
		seedSet := make([]int, 0, len(neighbors))
		for _, nb := range neighbors {
			if nb != i {
				seedSet = append(seedSet, nb)
			}
		}

		for j := 0; j < len(seedSet); j++ {
			q := seedSet[j]
			if labels[q] == noise {
				labels[q] = clusterID
			}
			if labels[q] != undefined {
				continue
			}
			labels[q] = clusterID

			qNeighbors := regionQuery(q)
			if len(qNeighbors) >= minPts {
				for _, nb := range qNeighbors {
					// Add to seed set if not already processed
					if labels[nb] == undefined || labels[nb] == noise {
						found := false
						for _, s := range seedSet {
							if s == nb {
								found = true
								break
							}
						}
						if !found {
							seedSet = append(seedSet, nb)
						}
					}
				}
			}
		}
	}

	// Group points by cluster
	clusters := make(map[int][]*object.Point)
	for i, label := range labels {
		if label > 0 { // skip noise points
			clusters[label] = append(clusters[label], points[i])
		}
	}

	// Convert to list of lists
	// Sort cluster IDs for deterministic output
	clusterIDs := make([]int, 0, len(clusters))
	for id := range clusters {
		clusterIDs = append(clusterIDs, id)
	}
	sort.Ints(clusterIDs)

	result := make([]object.Object, len(clusterIDs))
	for i, id := range clusterIDs {
		result[i] = pointsToList(clusters[id])
	}
	return &object.List{Elements: result}
}

// ── K-Means clustering (Lloyd's algorithm) ──

func kmeans(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "analysis.kmeans expects 2 arguments (points, k)"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.kmeans: first argument must be a List of Points"}
	}
	kObj, ok := args[1].(*object.Integer)
	if !ok {
		return &object.Error{Message: "analysis.kmeans: second argument must be an integer"}
	}
	k := int(kObj.Value)

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.kmeans: %s", err)}
	}

	n := len(points)
	if n == 0 {
		return &object.List{Elements: []object.Object{}}
	}
	if k <= 0 {
		return &object.Error{Message: "analysis.kmeans: k must be positive"}
	}
	if k > n {
		k = n
	}

	// Initialize centroids by selecting k random points
	rng := rand.New(rand.NewSource(42)) // deterministic seed for reproducibility
	perm := rng.Perm(n)
	centroids := make([][2]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = [2]float64{points[perm[i]].Coord.X, points[perm[i]].Coord.Y}
	}

	assignments := make([]int, n)
	maxIter := 100

	for iter := 0; iter < maxIter; iter++ {
		changed := false

		// Assign each point to the nearest centroid
		for i, pt := range points {
			bestCluster := 0
			bestDist := math.Inf(1)
			for c := 0; c < k; c++ {
				dx := pt.Coord.X - centroids[c][0]
				dy := pt.Coord.Y - centroids[c][1]
				d := dx*dx + dy*dy
				if d < bestDist {
					bestDist = d
					bestCluster = c
				}
			}
			if assignments[i] != bestCluster {
				assignments[i] = bestCluster
				changed = true
			}
		}

		if !changed {
			break
		}

		// Recompute centroids
		for c := 0; c < k; c++ {
			var sumX, sumY float64
			count := 0
			for i, pt := range points {
				if assignments[i] == c {
					sumX += pt.Coord.X
					sumY += pt.Coord.Y
					count++
				}
			}
			if count > 0 {
				centroids[c] = [2]float64{sumX / float64(count), sumY / float64(count)}
			}
		}
	}

	// Group points by assignment
	clusters := make([][]*object.Point, k)
	for i := 0; i < k; i++ {
		clusters[i] = []*object.Point{}
	}
	for i, pt := range points {
		clusters[assignments[i]] = append(clusters[assignments[i]], pt)
	}

	// Convert to list of lists (skip empty clusters)
	result := make([]object.Object, 0, k)
	for _, cluster := range clusters {
		if len(cluster) > 0 {
			result = append(result, pointsToList(cluster))
		}
	}
	return &object.List{Elements: result}
}

// ── Voronoi diagram (simplified) ──
// Returns a list of polygons, one per input point, approximating Voronoi cells
// within the given bounds using a grid-based approach.

func voronoi(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "analysis.voronoi expects 2 arguments (points, bounds)"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.voronoi: first argument must be a List of Points"}
	}
	bounds, bErr := parseBounds(args[1])
	if bErr != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.voronoi: %s", bErr)}
	}

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.voronoi: %s", err)}
	}

	n := len(points)
	if n == 0 {
		return &object.List{Elements: []object.Object{}}
	}
	if n == 1 {
		// Single point: the entire bounds is the cell
		poly := boundsToPolygon(bounds)
		return &object.List{Elements: []object.Object{poly}}
	}

	// Simplified Voronoi: for each pair of adjacent points, compute the
	// perpendicular bisector and clip the bounds. As a simpler approach,
	// we build approximate Voronoi cells using nearest-point assignment on
	// a grid, then compute the convex hull of grid cells assigned to each site.
	gridRes := 20 // grid resolution per axis
	dx := (bounds.MaxX - bounds.MinX) / float64(gridRes)
	dy := (bounds.MaxY - bounds.MinY) / float64(gridRes)

	// Assign grid cells to nearest site
	cellPoints := make([][]*object.Point, n)
	for i := range cellPoints {
		cellPoints[i] = []*object.Point{}
	}

	for gi := 0; gi <= gridRes; gi++ {
		for gj := 0; gj <= gridRes; gj++ {
			gx := bounds.MinX + float64(gj)*dx
			gy := bounds.MinY + float64(gi)*dy
			gridPt := &object.Point{Coord: object.Coordinate{X: gx, Y: gy}}

			bestIdx := 0
			bestDist := euclideanDist(gridPt, points[0])
			for idx := 1; idx < n; idx++ {
				d := euclideanDist(gridPt, points[idx])
				if d < bestDist {
					bestDist = d
					bestIdx = idx
				}
			}
			cellPoints[bestIdx] = append(cellPoints[bestIdx], gridPt)
		}
	}

	// Compute convex hull for each cell
	result := make([]object.Object, 0, n)
	for i := 0; i < n; i++ {
		if len(cellPoints[i]) < 3 {
			// Degenerate: return a tiny polygon around the site
			p := points[i]
			eps := 0.0001
			poly := &object.Polygon{
				ExteriorRing: []object.Coordinate{
					{X: p.Coord.X - eps, Y: p.Coord.Y - eps},
					{X: p.Coord.X + eps, Y: p.Coord.Y - eps},
					{X: p.Coord.X + eps, Y: p.Coord.Y + eps},
					{X: p.Coord.X - eps, Y: p.Coord.Y + eps},
					{X: p.Coord.X - eps, Y: p.Coord.Y - eps},
				},
			}
			result = append(result, poly)
		} else {
			hull := convexHullFromPoints(cellPoints[i])
			result = append(result, hull)
		}
	}
	return &object.List{Elements: result}
}

// ── Delaunay triangulation (simplified) ──
// Uses Bowyer-Watson incremental algorithm.

func delaunay(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "analysis.delaunay expects 1 argument (points)"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.delaunay: argument must be a List of Points"}
	}

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.delaunay: %s", err)}
	}

	n := len(points)
	if n < 3 {
		return &object.List{Elements: []object.Object{}}
	}

	// Bowyer-Watson algorithm
	type triangle struct {
		a, b, c int // indices into an extended points array
	}

	// Find bounding box
	minX, minY := points[0].Coord.X, points[0].Coord.Y
	maxX, maxY := points[0].Coord.X, points[0].Coord.Y
	for _, p := range points[1:] {
		if p.Coord.X < minX {
			minX = p.Coord.X
		}
		if p.Coord.Y < minY {
			minY = p.Coord.Y
		}
		if p.Coord.X > maxX {
			maxX = p.Coord.X
		}
		if p.Coord.Y > maxY {
			maxY = p.Coord.Y
		}
	}

	dx := maxX - minX
	dy := maxY - minY
	dmax := dx
	if dy > dmax {
		dmax = dy
	}
	midX := (minX + maxX) / 2
	midY := (minY + maxY) / 2

	// Create super-triangle vertices (indices n, n+1, n+2)
	superA := &object.Point{Coord: object.Coordinate{X: midX - 20*dmax, Y: midY - dmax}}
	superB := &object.Point{Coord: object.Coordinate{X: midX, Y: midY + 20*dmax}}
	superC := &object.Point{Coord: object.Coordinate{X: midX + 20*dmax, Y: midY - dmax}}

	allPts := make([]*object.Point, n+3)
	copy(allPts, points)
	allPts[n] = superA
	allPts[n+1] = superB
	allPts[n+2] = superC

	triangles := []triangle{{a: n, b: n + 1, c: n + 2}}

	// circumcircle test
	inCircumcircle := func(t triangle, px, py float64) bool {
		ax := allPts[t.a].Coord.X
		ay := allPts[t.a].Coord.Y
		bx := allPts[t.b].Coord.X
		by := allPts[t.b].Coord.Y
		cx := allPts[t.c].Coord.X
		cy := allPts[t.c].Coord.Y

		d := 2 * (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by))
		if math.Abs(d) < 1e-12 {
			return false
		}
		ux := ((ax*ax+ay*ay)*(by-cy) + (bx*bx+by*by)*(cy-ay) + (cx*cx+cy*cy)*(ay-by)) / d
		uy := ((ax*ax+ay*ay)*(cx-bx) + (bx*bx+by*by)*(ax-cx) + (cx*cx+cy*cy)*(bx-ax)) / d
		r2 := (ax-ux)*(ax-ux) + (ay-uy)*(ay-uy)
		dist2 := (px-ux)*(px-ux) + (py-uy)*(py-uy)
		return dist2 <= r2
	}

	type edge struct {
		a, b int
	}

	for i := 0; i < n; i++ {
		px := allPts[i].Coord.X
		py := allPts[i].Coord.Y

		var badTriangles []int
		for ti, t := range triangles {
			if inCircumcircle(t, px, py) {
				badTriangles = append(badTriangles, ti)
			}
		}

		// Find boundary edges of the polygonal hole
		edgeCount := make(map[edge]int)
		for _, ti := range badTriangles {
			t := triangles[ti]
			edges := []edge{{t.a, t.b}, {t.b, t.c}, {t.c, t.a}}
			for _, e := range edges {
				// Normalize edge direction
				ne := e
				if ne.a > ne.b {
					ne.a, ne.b = ne.b, ne.a
				}
				edgeCount[ne]++
			}
		}

		var boundaryEdges []edge
		for e, count := range edgeCount {
			if count == 1 {
				boundaryEdges = append(boundaryEdges, e)
			}
		}

		// Remove bad triangles (in reverse order to preserve indices)
		sort.Sort(sort.Reverse(sort.IntSlice(badTriangles)))
		for _, ti := range badTriangles {
			triangles = append(triangles[:ti], triangles[ti+1:]...)
		}

		// Create new triangles from boundary edges to the new point
		for _, e := range boundaryEdges {
			triangles = append(triangles, triangle{a: e.a, b: e.b, c: i})
		}
	}

	// Remove triangles that share vertices with the super-triangle
	var resultTriangles []triangle
	for _, t := range triangles {
		if t.a >= n || t.b >= n || t.c >= n {
			continue
		}
		resultTriangles = append(resultTriangles, t)
	}

	// Convert triangles to polygons
	result := make([]object.Object, len(resultTriangles))
	for i, t := range resultTriangles {
		ax := allPts[t.a].Coord.X
		ay := allPts[t.a].Coord.Y
		bx := allPts[t.b].Coord.X
		by := allPts[t.b].Coord.Y
		cx := allPts[t.c].Coord.X
		cy := allPts[t.c].Coord.Y
		result[i] = &object.Polygon{
			ExteriorRing: []object.Coordinate{
				{X: ax, Y: ay},
				{X: bx, Y: by},
				{X: cx, Y: cy},
				{X: ax, Y: ay},
			},
		}
	}
	return &object.List{Elements: result}
}

// ── Concave hull (placeholder: returns convex hull) ──

func concaveHull(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "analysis.concaveHull expects 1 argument (points)"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.concaveHull: argument must be a List of Points"}
	}

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.concaveHull: %s", err)}
	}

	if len(points) < 3 {
		return &object.Error{Message: "analysis.concaveHull: need at least 3 points"}
	}

	return convexHullFromPoints(points)
}

// ── Alpha shape (placeholder: returns convex hull) ──

func alphaShape(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "analysis.alphaShape expects 1-2 arguments (points, [alpha])"}
	}

	pointsList, ok := args[0].(*object.List)
	if !ok {
		return &object.Error{Message: "analysis.alphaShape: first argument must be a List of Points"}
	}

	points, err := extractPoints(pointsList)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("analysis.alphaShape: %s", err)}
	}

	if len(points) < 3 {
		return &object.Error{Message: "analysis.alphaShape: need at least 3 points"}
	}

	// Placeholder: return convex hull regardless of alpha
	return convexHullFromPoints(points)
}

// ── Convex hull (Andrew's monotone chain) ──

func convexHullFromPoints(points []*object.Point) *object.Polygon {
	n := len(points)
	if n < 3 {
		coords := make([]object.Coordinate, n+1)
		for i, p := range points {
			coords[i] = p.Coord
		}
		if n > 0 {
			coords[n] = points[0].Coord
		}
		return &object.Polygon{ExteriorRing: coords}
	}

	// Sort points lexicographically by (X, Y)
	sorted := make([]*object.Point, n)
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Coord.X != sorted[j].Coord.X {
			return sorted[i].Coord.X < sorted[j].Coord.X
		}
		return sorted[i].Coord.Y < sorted[j].Coord.Y
	})

	// Cross product of vectors OA and OB where O is origin
	cross := func(o, a, b object.Coordinate) float64 {
		return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
	}

	// Build lower hull
	var lower []*object.Point
	for _, p := range sorted {
		for len(lower) >= 2 && cross(lower[len(lower)-2].Coord, lower[len(lower)-1].Coord, p.Coord) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}

	// Build upper hull
	var upper []*object.Point
	for i := n - 1; i >= 0; i-- {
		p := sorted[i]
		for len(upper) >= 2 && cross(upper[len(upper)-2].Coord, upper[len(upper)-1].Coord, p.Coord) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}

	// Concatenate lower and upper hulls, removing last point of each (duplicate)
	hull := make([]object.Coordinate, 0, len(lower)+len(upper))
	for _, p := range lower[:len(lower)-1] {
		hull = append(hull, p.Coord)
	}
	for _, p := range upper[:len(upper)-1] {
		hull = append(hull, p.Coord)
	}

	// Close the ring
	if len(hull) > 0 {
		hull = append(hull, hull[0])
	}

	return &object.Polygon{ExteriorRing: hull}
}

// ── Utility helpers ──

// parseBounds extracts a BBox from a BBox object or a Map.
func parseBounds(obj object.Object) (*object.BBox, error) {
	switch b := obj.(type) {
	case *object.BBox:
		return b, nil
	case *object.Map:
		minX, err := getMapFloat(b, "minX")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		minY, err := getMapFloat(b, "minY")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		maxX, err := getMapFloat(b, "maxX")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		maxY, err := getMapFloat(b, "maxY")
		if err != nil {
			return nil, fmt.Errorf("bounds: %s", err)
		}
		return &object.BBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}, nil
	default:
		return nil, fmt.Errorf("bounds must be a BBox or Map with minX, minY, maxX, maxY keys")
	}
}

func getMapFloat(m *object.Map, key string) (float64, error) {
	v, ok := m.Get(key)
	if !ok {
		return 0, fmt.Errorf("missing key %q", key)
	}
	return toFloat64(v)
}

func boundsToPolygon(b *object.BBox) *object.Polygon {
	return &object.Polygon{
		ExteriorRing: []object.Coordinate{
			{X: b.MinX, Y: b.MinY},
			{X: b.MaxX, Y: b.MinY},
			{X: b.MaxX, Y: b.MaxY},
			{X: b.MinX, Y: b.MaxY},
			{X: b.MinX, Y: b.MinY},
		},
	}
}

// toFloat64 converts an Integer or Float to float64.
func toFloat64(obj object.Object) (float64, error) {
	switch v := obj.(type) {
	case *object.Float:
		return v.Value, nil
	case *object.Integer:
		return float64(v.Value), nil
	default:
		return 0, fmt.Errorf("expected a number, got %s", obj.Type())
	}
}
