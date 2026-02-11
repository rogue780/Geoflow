package stdlib

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rogue780/geoflow/internal/object"
)

const earthRadiusKm = 6371.0088

func deg2rad(d float64) float64 { return d * math.Pi / 180 }
func rad2deg(r float64) float64 { return r * 180 / math.Pi }

// ════════════════════════════════════════════════════════════
// WKT Parser
// ════════════════════════════════════════════════════════════

// ParseWKT parses a WKT string into a Geometry object.
func ParseWKT(wkt string) (object.Geometry, error) {
	p := &wktParser{input: strings.TrimSpace(wkt), pos: 0}
	return p.parse()
}

type wktParser struct {
	input string
	pos   int
}

func (p *wktParser) parse() (object.Geometry, error) {
	p.skipSpaces()
	typeName := p.readWord()
	upper := strings.ToUpper(typeName)

	hasZ := false
	if p.peekWord() == "Z" || p.peekWord() == "z" {
		p.readWord()
		hasZ = true
	}

	p.skipSpaces()

	// Handle EMPTY geometries
	if p.peekWord() == "EMPTY" {
		p.readWord()
		return p.emptyGeometry(upper)
	}

	switch upper {
	case "POINT":
		return p.parsePoint(hasZ)
	case "LINESTRING":
		return p.parseLineString(hasZ)
	case "POLYGON":
		return p.parsePolygon(hasZ)
	case "MULTIPOINT":
		return p.parseMultiPoint(hasZ)
	case "MULTILINESTRING":
		return p.parseMultiLineString(hasZ)
	case "MULTIPOLYGON":
		return p.parseMultiPolygon(hasZ)
	case "GEOMETRYCOLLECTION":
		return p.parseGeometryCollection()
	default:
		return nil, fmt.Errorf("unknown WKT geometry type: %s", typeName)
	}
}

func (p *wktParser) emptyGeometry(typeName string) (object.Geometry, error) {
	switch typeName {
	case "POINT":
		return &object.Point{Coord: object.Coordinate{X: math.NaN(), Y: math.NaN()}}, nil
	case "LINESTRING":
		return &object.LineString{}, nil
	case "POLYGON":
		return &object.Polygon{}, nil
	case "MULTIPOINT":
		return &object.MultiPoint{}, nil
	case "MULTILINESTRING":
		return &object.MultiLineString{}, nil
	case "MULTIPOLYGON":
		return &object.MultiPolygon{}, nil
	case "GEOMETRYCOLLECTION":
		return &object.GeometryCollection{}, nil
	default:
		return nil, fmt.Errorf("unknown geometry type: %s", typeName)
	}
}

func (p *wktParser) parsePoint(hasZ bool) (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	c, err := p.readCoordinate(hasZ)
	if err != nil {
		return nil, err
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &object.Point{Coord: c}, nil
}

func (p *wktParser) parseLineString(hasZ bool) (object.Geometry, error) {
	coords, err := p.readCoordinateSequence(hasZ)
	if err != nil {
		return nil, err
	}
	return &object.LineString{Coords: coords}, nil
}

func (p *wktParser) parsePolygon(hasZ bool) (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var rings [][]object.Coordinate
	for {
		ring, err := p.readCoordinateSequence(hasZ)
		if err != nil {
			return nil, err
		}
		rings = append(rings, ring)
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	poly := &object.Polygon{ExteriorRing: rings[0]}
	if len(rings) > 1 {
		poly.InteriorRings = rings[1:]
	}
	return poly, nil
}

func (p *wktParser) parseMultiPoint(hasZ bool) (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var points []*object.Point
	for {
		p.skipSpaces()
		hasParen := false
		if p.pos < len(p.input) && p.input[p.pos] == '(' {
			hasParen = true
			p.pos++
		}
		c, err := p.readCoordinate(hasZ)
		if err != nil {
			return nil, err
		}
		if hasParen {
			if err := p.expect(')'); err != nil {
				return nil, err
			}
		}
		points = append(points, &object.Point{Coord: c})
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &object.MultiPoint{Points: points}, nil
}

func (p *wktParser) parseMultiLineString(hasZ bool) (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var lines []*object.LineString
	for {
		coords, err := p.readCoordinateSequence(hasZ)
		if err != nil {
			return nil, err
		}
		lines = append(lines, &object.LineString{Coords: coords})
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &object.MultiLineString{Lines: lines}, nil
}

func (p *wktParser) parseMultiPolygon(hasZ bool) (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var polygons []*object.Polygon
	for {
		geom, err := p.parsePolygon(hasZ)
		if err != nil {
			return nil, err
		}
		polygons = append(polygons, geom.(*object.Polygon))
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &object.MultiPolygon{Polygons: polygons}, nil
}

func (p *wktParser) parseGeometryCollection() (object.Geometry, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var geoms []object.Geometry
	for {
		g, err := p.parse()
		if err != nil {
			return nil, err
		}
		geoms = append(geoms, g)
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &object.GeometryCollection{Geometries: geoms}, nil
}

func (p *wktParser) readCoordinateSequence(hasZ bool) ([]object.Coordinate, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	var coords []object.Coordinate
	for {
		c, err := p.readCoordinate(hasZ)
		if err != nil {
			return nil, err
		}
		coords = append(coords, c)
		p.skipSpaces()
		if p.pos < len(p.input) && p.input[p.pos] == ',' {
			p.pos++
			continue
		}
		break
	}
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return coords, nil
}

func (p *wktParser) readCoordinate(hasZ bool) (object.Coordinate, error) {
	p.skipSpaces()
	x, err := p.readNumber()
	if err != nil {
		return object.Coordinate{}, fmt.Errorf("expected x coordinate: %w", err)
	}
	y, err := p.readNumber()
	if err != nil {
		return object.Coordinate{}, fmt.Errorf("expected y coordinate: %w", err)
	}
	c := object.Coordinate{X: x, Y: y}

	// Try to read Z if expected or if there's another number before delimiter
	if hasZ || p.hasMoreNumbersBeforeDelimiter() {
		z, err := p.readNumber()
		if err == nil {
			c.Z = z
			c.HasZ = true
		}
	}
	return c, nil
}

func (p *wktParser) hasMoreNumbersBeforeDelimiter() bool {
	saved := p.pos
	p.skipSpaces()
	if p.pos >= len(p.input) {
		p.pos = saved
		return false
	}
	ch := p.input[p.pos]
	p.pos = saved
	return ch != ',' && ch != ')' && ch != '('
}

func (p *wktParser) readNumber() (float64, error) {
	p.skipSpaces()
	start := p.pos
	if p.pos < len(p.input) && (p.input[p.pos] == '-' || p.input[p.pos] == '+') {
		p.pos++
	}
	if p.pos >= len(p.input) || (!isDigit(p.input[p.pos]) && p.input[p.pos] != '.') {
		p.pos = start
		return 0, fmt.Errorf("expected number at position %d", start)
	}
	for p.pos < len(p.input) && isDigit(p.input[p.pos]) {
		p.pos++
	}
	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		p.pos++
		for p.pos < len(p.input) && isDigit(p.input[p.pos]) {
			p.pos++
		}
	}
	// Scientific notation
	if p.pos < len(p.input) && (p.input[p.pos] == 'e' || p.input[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.input) && isDigit(p.input[p.pos]) {
			p.pos++
		}
	}
	var v float64
	_, err := fmt.Sscanf(p.input[start:p.pos], "%f", &v)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", p.input[start:p.pos])
	}
	return v, nil
}

func (p *wktParser) readWord() string {
	p.skipSpaces()
	start := p.pos
	for p.pos < len(p.input) && isAlpha(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *wktParser) peekWord() string {
	saved := p.pos
	w := p.readWord()
	p.pos = saved
	return w
}

func (p *wktParser) expect(ch byte) error {
	p.skipSpaces()
	if p.pos >= len(p.input) {
		return fmt.Errorf("expected '%c' but reached end of input", ch)
	}
	if p.input[p.pos] != ch {
		return fmt.Errorf("expected '%c' at position %d, got '%c'", ch, p.pos, p.input[p.pos])
	}
	p.pos++
	return nil
}

func (p *wktParser) skipSpaces() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func isDigit(ch byte) bool  { return ch >= '0' && ch <= '9' }
func isAlpha(ch byte) bool  { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }

// ════════════════════════════════════════════════════════════
// GeoJSON
// ════════════════════════════════════════════════════════════

func geomToGeoJSON(g object.Geometry) interface{} {
	switch geom := g.(type) {
	case *object.Point:
		coords := []float64{geom.Coord.X, geom.Coord.Y}
		if geom.Coord.HasZ {
			coords = append(coords, geom.Coord.Z)
		}
		return map[string]interface{}{"type": "Point", "coordinates": coords}
	case *object.LineString:
		return map[string]interface{}{"type": "LineString", "coordinates": coordsToJSON(geom.Coords)}
	case *object.Polygon:
		rings := []interface{}{coordsToJSON(geom.ExteriorRing)}
		for _, r := range geom.InteriorRings {
			rings = append(rings, coordsToJSON(r))
		}
		return map[string]interface{}{"type": "Polygon", "coordinates": rings}
	case *object.MultiPoint:
		coords := make([]interface{}, len(geom.Points))
		for i, p := range geom.Points {
			c := []float64{p.Coord.X, p.Coord.Y}
			if p.Coord.HasZ {
				c = append(c, p.Coord.Z)
			}
			coords[i] = c
		}
		return map[string]interface{}{"type": "MultiPoint", "coordinates": coords}
	case *object.MultiLineString:
		lines := make([]interface{}, len(geom.Lines))
		for i, l := range geom.Lines {
			lines[i] = coordsToJSON(l.Coords)
		}
		return map[string]interface{}{"type": "MultiLineString", "coordinates": lines}
	case *object.MultiPolygon:
		polys := make([]interface{}, len(geom.Polygons))
		for i, p := range geom.Polygons {
			rings := []interface{}{coordsToJSON(p.ExteriorRing)}
			for _, r := range p.InteriorRings {
				rings = append(rings, coordsToJSON(r))
			}
			polys[i] = rings
		}
		return map[string]interface{}{"type": "MultiPolygon", "coordinates": polys}
	case *object.GeometryCollection:
		geoms := make([]interface{}, len(geom.Geometries))
		for i, g2 := range geom.Geometries {
			geoms[i] = geomToGeoJSON(g2)
		}
		return map[string]interface{}{"type": "GeometryCollection", "geometries": geoms}
	}
	return nil
}

func coordsToJSON(coords []object.Coordinate) []interface{} {
	result := make([]interface{}, len(coords))
	for i, c := range coords {
		arr := []float64{c.X, c.Y}
		if c.HasZ {
			arr = append(arr, c.Z)
		}
		result[i] = arr
	}
	return result
}

func parseGeoJSONGeometry(data map[string]interface{}) (object.Geometry, error) {
	typeName, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("GeoJSON missing 'type' field")
	}
	switch typeName {
	case "Point":
		coords, err := getJSONCoordArray(data, "coordinates")
		if err != nil {
			return nil, err
		}
		c, err := jsonArrayToCoord(coords)
		if err != nil {
			return nil, err
		}
		return &object.Point{Coord: c}, nil
	case "LineString":
		coords, err := getJSONCoordArrayOfArrays(data, "coordinates")
		if err != nil {
			return nil, err
		}
		cs, err := jsonArraysToCoords(coords)
		if err != nil {
			return nil, err
		}
		return &object.LineString{Coords: cs}, nil
	case "Polygon":
		rings, err := getJSONRings(data, "coordinates")
		if err != nil {
			return nil, err
		}
		poly := &object.Polygon{}
		for i, ring := range rings {
			cs, err := jsonArraysToCoords(ring)
			if err != nil {
				return nil, err
			}
			if i == 0 {
				poly.ExteriorRing = cs
			} else {
				poly.InteriorRings = append(poly.InteriorRings, cs)
			}
		}
		return poly, nil
	case "MultiPoint":
		coords, err := getJSONCoordArrayOfArrays(data, "coordinates")
		if err != nil {
			return nil, err
		}
		var points []*object.Point
		for _, ca := range coords {
			c, err := jsonArrayToCoord(ca)
			if err != nil {
				return nil, err
			}
			points = append(points, &object.Point{Coord: c})
		}
		return &object.MultiPoint{Points: points}, nil
	case "MultiLineString":
		rings, err := getJSONRings(data, "coordinates")
		if err != nil {
			return nil, err
		}
		var lines []*object.LineString
		for _, ring := range rings {
			cs, err := jsonArraysToCoords(ring)
			if err != nil {
				return nil, err
			}
			lines = append(lines, &object.LineString{Coords: cs})
		}
		return &object.MultiLineString{Lines: lines}, nil
	case "MultiPolygon":
		raw, ok := data["coordinates"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("MultiPolygon: invalid coordinates")
		}
		var polygons []*object.Polygon
		for _, polyRaw := range raw {
			ringsRaw, ok := polyRaw.([]interface{})
			if !ok {
				return nil, fmt.Errorf("MultiPolygon: invalid polygon")
			}
			poly := &object.Polygon{}
			for i, ringRaw := range ringsRaw {
				ringArr, ok := ringRaw.([]interface{})
				if !ok {
					return nil, fmt.Errorf("MultiPolygon: invalid ring")
				}
				coordArrays := make([][]interface{}, len(ringArr))
				for j, coordRaw := range ringArr {
					ca, ok := coordRaw.([]interface{})
					if !ok {
						return nil, fmt.Errorf("MultiPolygon: invalid coordinate")
					}
					coordArrays[j] = ca
				}
				cs, err := jsonArraysToCoords(coordArrays)
				if err != nil {
					return nil, err
				}
				if i == 0 {
					poly.ExteriorRing = cs
				} else {
					poly.InteriorRings = append(poly.InteriorRings, cs)
				}
			}
			polygons = append(polygons, poly)
		}
		return &object.MultiPolygon{Polygons: polygons}, nil
	case "GeometryCollection":
		geomsRaw, ok := data["geometries"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("GeometryCollection: missing 'geometries'")
		}
		var geoms []object.Geometry
		for _, gr := range geomsRaw {
			gm, ok := gr.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("GeometryCollection: invalid geometry")
			}
			g, err := parseGeoJSONGeometry(gm)
			if err != nil {
				return nil, err
			}
			geoms = append(geoms, g)
		}
		return &object.GeometryCollection{Geometries: geoms}, nil
	case "Feature":
		return parseGeoJSONFeatureGeom(data)
	default:
		return nil, fmt.Errorf("unknown GeoJSON type: %s", typeName)
	}
}

func parseGeoJSONFeatureGeom(data map[string]interface{}) (object.Geometry, error) {
	geomRaw, ok := data["geometry"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Feature: missing 'geometry'")
	}
	return parseGeoJSONGeometry(geomRaw)
}

func getJSONCoordArray(data map[string]interface{}, key string) ([]interface{}, error) {
	raw, ok := data[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid '%s'", key)
	}
	return raw, nil
}

func getJSONCoordArrayOfArrays(data map[string]interface{}, key string) ([][]interface{}, error) {
	raw, ok := data[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid '%s'", key)
	}
	result := make([][]interface{}, len(raw))
	for i, r := range raw {
		arr, ok := r.([]interface{})
		if !ok {
			return nil, fmt.Errorf("%s[%d]: expected array", key, i)
		}
		result[i] = arr
	}
	return result, nil
}

func getJSONRings(data map[string]interface{}, key string) ([][][]interface{}, error) {
	raw, ok := data[key].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid '%s'", key)
	}
	var rings [][][]interface{}
	for _, ringRaw := range raw {
		ring, ok := ringRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid ring in '%s'", key)
		}
		var coordArrays [][]interface{}
		for _, coordRaw := range ring {
			ca, ok := coordRaw.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid coordinate in '%s'", key)
			}
			coordArrays = append(coordArrays, ca)
		}
		rings = append(rings, coordArrays)
	}
	return rings, nil
}

func jsonArrayToCoord(arr []interface{}) (object.Coordinate, error) {
	if len(arr) < 2 {
		return object.Coordinate{}, fmt.Errorf("coordinate needs at least 2 values")
	}
	x, ok1 := jsonToFloat(arr[0])
	y, ok2 := jsonToFloat(arr[1])
	if !ok1 || !ok2 {
		return object.Coordinate{}, fmt.Errorf("coordinate values must be numbers")
	}
	c := object.Coordinate{X: x, Y: y}
	if len(arr) >= 3 {
		z, ok := jsonToFloat(arr[2])
		if ok {
			c.Z = z
			c.HasZ = true
		}
	}
	return c, nil
}

func jsonArraysToCoords(arrays [][]interface{}) ([]object.Coordinate, error) {
	coords := make([]object.Coordinate, len(arrays))
	for i, arr := range arrays {
		c, err := jsonArrayToCoord(arr)
		if err != nil {
			return nil, fmt.Errorf("coordinate %d: %w", i, err)
		}
		coords[i] = c
	}
	return coords, nil
}

func jsonToFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// ════════════════════════════════════════════════════════════
// Spatial Calculations
// ════════════════════════════════════════════════════════════

// haversineDistance computes the great-circle distance in kilometers.
func haversineDistance(c1, c2 object.Coordinate) float64 {
	lat1, lon1 := deg2rad(c1.Y), deg2rad(c1.X)
	lat2, lon2 := deg2rad(c2.Y), deg2rad(c2.X)
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// euclideanDistance computes the Euclidean distance between two coordinates.
func euclideanDistance(c1, c2 object.Coordinate) float64 {
	dx := c1.X - c2.X
	dy := c1.Y - c2.Y
	if c1.HasZ && c2.HasZ {
		dz := c1.Z - c2.Z
		return math.Sqrt(dx*dx + dy*dy + dz*dz)
	}
	return math.Sqrt(dx*dx + dy*dy)
}

// shoelaceArea computes the signed area of a ring using the shoelace formula.
func shoelaceArea(ring []object.Coordinate) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		sum += ring[i].X*ring[j].Y - ring[j].X*ring[i].Y
	}
	return sum / 2
}

// geodesicArea computes the area in square kilometers using the spherical excess formula.
func geodesicArea(ring []object.Coordinate) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		k := (i + 2) % n
		sum += (deg2rad(ring[k].X) - deg2rad(ring[i].X)) * math.Sin(deg2rad(ring[j].Y))
	}
	return math.Abs(sum) * earthRadiusKm * earthRadiusKm / 2
}

// lineLength computes the total Euclidean length of a coordinate sequence.
func lineLength(coords []object.Coordinate) float64 {
	total := 0.0
	for i := 1; i < len(coords); i++ {
		total += euclideanDistance(coords[i-1], coords[i])
	}
	return total
}

// lineLengthGeodesic computes the total geodesic length in km.
func lineLengthGeodesic(coords []object.Coordinate) float64 {
	total := 0.0
	for i := 1; i < len(coords); i++ {
		total += haversineDistance(coords[i-1], coords[i])
	}
	return total
}

// centroidOfRing computes the centroid of a polygon ring.
func centroidOfRing(ring []object.Coordinate) object.Coordinate {
	n := len(ring)
	if n == 0 {
		return object.Coordinate{}
	}
	cx, cy := 0.0, 0.0
	signedArea := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		cross := ring[i].X*ring[j].Y - ring[j].X*ring[i].Y
		signedArea += cross
		cx += (ring[i].X + ring[j].X) * cross
		cy += (ring[i].Y + ring[j].Y) * cross
	}
	signedArea /= 2
	if signedArea == 0 {
		// Degenerate: return average
		for _, c := range ring {
			cx += c.X
			cy += c.Y
		}
		return object.Coordinate{X: cx / float64(n), Y: cy / float64(n)}
	}
	cx /= (6 * signedArea)
	cy /= (6 * signedArea)
	return object.Coordinate{X: cx, Y: cy}
}

// envelope computes the bounding box of a set of coordinates.
func envelope(coords []object.Coordinate) *object.BBox {
	if len(coords) == 0 {
		return &object.BBox{}
	}
	minX, minY := coords[0].X, coords[0].Y
	maxX, maxY := coords[0].X, coords[0].Y
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
	return &object.BBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
}

// pointInRing tests if a point is inside a polygon ring using ray casting.
func pointInRing(pt object.Coordinate, ring []object.Coordinate) bool {
	n := len(ring)
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		yi, yj := ring[i].Y, ring[j].Y
		xi, xj := ring[i].X, ring[j].X
		if ((yi > pt.Y) != (yj > pt.Y)) &&
			(pt.X < (xj-xi)*(pt.Y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

// pointOnSegment tests if a point lies on a line segment within tolerance.
func pointOnSegment(pt, a, b object.Coordinate, tol float64) bool {
	d1 := euclideanDistance(a, pt)
	d2 := euclideanDistance(pt, b)
	d3 := euclideanDistance(a, b)
	return math.Abs(d1+d2-d3) < tol
}

// segmentsIntersect tests if two line segments intersect.
func segmentsIntersect(a1, a2, b1, b2 object.Coordinate) bool {
	d1 := crossProduct2D(b1, b2, a1)
	d2 := crossProduct2D(b1, b2, a2)
	d3 := crossProduct2D(a1, a2, b1)
	d4 := crossProduct2D(a1, a2, b2)

	if ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0)) {
		return true
	}
	if d1 == 0 && onSegment(b1, b2, a1) {
		return true
	}
	if d2 == 0 && onSegment(b1, b2, a2) {
		return true
	}
	if d3 == 0 && onSegment(a1, a2, b1) {
		return true
	}
	if d4 == 0 && onSegment(a1, a2, b2) {
		return true
	}
	return false
}

func crossProduct2D(o, a, b object.Coordinate) float64 {
	return (a.X-o.X)*(b.Y-o.Y) - (a.Y-o.Y)*(b.X-o.X)
}

func onSegment(p, q, r object.Coordinate) bool {
	return r.X <= math.Max(p.X, q.X) && r.X >= math.Min(p.X, q.X) &&
		r.Y <= math.Max(p.Y, q.Y) && r.Y >= math.Min(p.Y, q.Y)
}

// convexHull computes the convex hull of a set of coordinates using Andrew's monotone chain.
func convexHull(coords []object.Coordinate) []object.Coordinate {
	pts := make([]object.Coordinate, len(coords))
	copy(pts, coords)
	sort.Slice(pts, func(i, j int) bool {
		if pts[i].X == pts[j].X {
			return pts[i].Y < pts[j].Y
		}
		return pts[i].X < pts[j].X
	})
	// Remove duplicates
	unique := pts[:0]
	for i, p := range pts {
		if i == 0 || p.X != pts[i-1].X || p.Y != pts[i-1].Y {
			unique = append(unique, p)
		}
	}
	pts = unique
	n := len(pts)
	if n <= 2 {
		return pts
	}
	// Build lower hull
	var lower []object.Coordinate
	for _, p := range pts {
		for len(lower) >= 2 && crossProduct2D(lower[len(lower)-2], lower[len(lower)-1], p) <= 0 {
			lower = lower[:len(lower)-1]
		}
		lower = append(lower, p)
	}
	// Build upper hull
	var upper []object.Coordinate
	for i := len(pts) - 1; i >= 0; i-- {
		p := pts[i]
		for len(upper) >= 2 && crossProduct2D(upper[len(upper)-2], upper[len(upper)-1], p) <= 0 {
			upper = upper[:len(upper)-1]
		}
		upper = append(upper, p)
	}
	// Concatenate, removing last point of each half (it's the first point of the other)
	hull := append(lower[:len(lower)-1], upper[:len(upper)-1]...)
	// Close the ring
	hull = append(hull, hull[0])
	return hull
}

// simplifyDouglasPeucker simplifies a coordinate sequence using the Douglas-Peucker algorithm.
func simplifyDouglasPeucker(coords []object.Coordinate, tolerance float64) []object.Coordinate {
	if len(coords) <= 2 {
		return coords
	}
	// Find the point with maximum distance from the line between first and last
	dmax := 0.0
	index := 0
	end := len(coords) - 1
	for i := 1; i < end; i++ {
		d := perpendicularDistance(coords[i], coords[0], coords[end])
		if d > dmax {
			dmax = d
			index = i
		}
	}
	if dmax > tolerance {
		left := simplifyDouglasPeucker(coords[:index+1], tolerance)
		right := simplifyDouglasPeucker(coords[index:], tolerance)
		return append(left[:len(left)-1], right...)
	}
	return []object.Coordinate{coords[0], coords[end]}
}

func perpendicularDistance(p, a, b object.Coordinate) float64 {
	dx, dy := b.X-a.X, b.Y-a.Y
	if dx == 0 && dy == 0 {
		return euclideanDistance(p, a)
	}
	t := ((p.X-a.X)*dx + (p.Y-a.Y)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	proj := object.Coordinate{X: a.X + t*dx, Y: a.Y + t*dy}
	return euclideanDistance(p, proj)
}

// bufferPoint creates a circular approximation around a point.
func bufferPoint(center object.Coordinate, radius float64, segments int) []object.Coordinate {
	coords := make([]object.Coordinate, segments+1)
	for i := 0; i <= segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		coords[i] = object.Coordinate{
			X: center.X + radius*math.Cos(angle),
			Y: center.Y + radius*math.Sin(angle),
		}
	}
	return coords
}

// ════════════════════════════════════════════════════════════
// CRS / Projection
// ════════════════════════════════════════════════════════════

// Supported projections: EPSG:4326 (WGS84), EPSG:3857 (Web Mercator)

func projectCoord4326To3857(c object.Coordinate) object.Coordinate {
	x := c.X * 20037508.34 / 180.0
	y := math.Log(math.Tan((90+c.Y)*math.Pi/360)) / (math.Pi / 180)
	y = y * 20037508.34 / 180.0
	return object.Coordinate{X: x, Y: y, Z: c.Z, HasZ: c.HasZ}
}

func projectCoord3857To4326(c object.Coordinate) object.Coordinate {
	lon := c.X * 180.0 / 20037508.34
	lat := math.Atan(math.Exp(c.Y*math.Pi/20037508.34)) * 360 / math.Pi - 90
	return object.Coordinate{X: lon, Y: lat, Z: c.Z, HasZ: c.HasZ}
}

func projectCoords(coords []object.Coordinate, proj func(object.Coordinate) object.Coordinate) []object.Coordinate {
	out := make([]object.Coordinate, len(coords))
	for i, c := range coords {
		out[i] = proj(c)
	}
	return out
}

func projectGeometry(g object.Geometry, proj func(object.Coordinate) object.Coordinate) object.Geometry {
	switch geom := g.(type) {
	case *object.Point:
		return &object.Point{Coord: proj(geom.Coord)}
	case *object.LineString:
		return &object.LineString{Coords: projectCoords(geom.Coords, proj)}
	case *object.Polygon:
		poly := &object.Polygon{ExteriorRing: projectCoords(geom.ExteriorRing, proj)}
		for _, r := range geom.InteriorRings {
			poly.InteriorRings = append(poly.InteriorRings, projectCoords(r, proj))
		}
		return poly
	case *object.MultiPoint:
		pts := make([]*object.Point, len(geom.Points))
		for i, p := range geom.Points {
			pts[i] = &object.Point{Coord: proj(p.Coord)}
		}
		return &object.MultiPoint{Points: pts}
	case *object.MultiLineString:
		lines := make([]*object.LineString, len(geom.Lines))
		for i, l := range geom.Lines {
			lines[i] = &object.LineString{Coords: projectCoords(l.Coords, proj)}
		}
		return &object.MultiLineString{Lines: lines}
	case *object.MultiPolygon:
		polys := make([]*object.Polygon, len(geom.Polygons))
		for i, p := range geom.Polygons {
			np := &object.Polygon{ExteriorRing: projectCoords(p.ExteriorRing, proj)}
			for _, r := range p.InteriorRings {
				np.InteriorRings = append(np.InteriorRings, projectCoords(r, proj))
			}
			polys[i] = np
		}
		return &object.MultiPolygon{Polygons: polys}
	case *object.GeometryCollection:
		geoms := make([]object.Geometry, len(geom.Geometries))
		for i, g2 := range geom.Geometries {
			geoms[i] = projectGeometry(g2, proj)
		}
		return &object.GeometryCollection{Geometries: geoms}
	}
	return g
}

// ════════════════════════════════════════════════════════════
// Exported wrappers for use by eval package
// ════════════════════════════════════════════════════════════

func EuclideanDistance(c1, c2 object.Coordinate) float64  { return euclideanDistance(c1, c2) }
func HaversineDistance(c1, c2 object.Coordinate) float64  { return haversineDistance(c1, c2) }
func ShoelaceArea(ring []object.Coordinate) float64       { return shoelaceArea(ring) }
func GeodesicArea(ring []object.Coordinate) float64       { return geodesicArea(ring) }
func LineLength(coords []object.Coordinate) float64       { return lineLength(coords) }
func LineLengthGeodesic(coords []object.Coordinate) float64 { return lineLengthGeodesic(coords) }
func CentroidOfRing(ring []object.Coordinate) object.Coordinate { return centroidOfRing(ring) }
func Envelope(coords []object.Coordinate) *object.BBox    { return envelope(coords) }
func PointInRing(pt object.Coordinate, ring []object.Coordinate) bool { return pointInRing(pt, ring) }
func ConvexHull(coords []object.Coordinate) []object.Coordinate { return convexHull(coords) }
func SimplifyDouglasPeucker(coords []object.Coordinate, tolerance float64) []object.Coordinate {
	return simplifyDouglasPeucker(coords, tolerance)
}
func BufferPoint(center object.Coordinate, radius float64, segments int) []object.Coordinate {
	return bufferPoint(center, radius, segments)
}
func ProjectGeometry(g object.Geometry, proj func(object.Coordinate) object.Coordinate) object.Geometry {
	return projectGeometry(g, proj)
}
func ProjectCoord4326To3857(c object.Coordinate) object.Coordinate { return projectCoord4326To3857(c) }
func ProjectCoord3857To4326(c object.Coordinate) object.Coordinate { return projectCoord3857To4326(c) }
func SpatialContains(outer, inner object.Object) bool     { return spatialContains(outer, inner) }
func SpatialIntersects(a, b object.Object) bool           { return spatialIntersects(a, b) }
func SegmentsIntersect(a1, a2, b1, b2 object.Coordinate) bool { return segmentsIntersect(a1, a2, b1, b2) }

// ════════════════════════════════════════════════════════════
// Builtin Registration
// ════════════════════════════════════════════════════════════

// asGeometry extracts a Geometry from an Object.
func asGeometry(obj object.Object) (object.Geometry, bool) {
	g, ok := obj.(object.Geometry)
	return g, ok
}

// GetGeoBuiltins returns all geospatial built-in functions.
func GetGeoBuiltins() map[string]*object.Builtin {
	builtins := map[string]*object.Builtin{}

	// ── Constructors ──
	builtins["point"] = &object.Builtin{Name: "point", Fn: func(args ...object.Object) object.Object {
		switch len(args) {
		case 2:
			x, ok1 := toFloat(args[0])
			y, ok2 := toFloat(args[1])
			if !ok1 || !ok2 {
				return &object.Error{Message: "point expects numeric arguments"}
			}
			return &object.Point{Coord: object.Coordinate{X: x, Y: y}}
		case 3:
			x, ok1 := toFloat(args[0])
			y, ok2 := toFloat(args[1])
			z, ok3 := toFloat(args[2])
			if !ok1 || !ok2 || !ok3 {
				return &object.Error{Message: "point expects numeric arguments"}
			}
			return &object.Point{Coord: object.Coordinate{X: x, Y: y, Z: z, HasZ: true}}
		default:
			return &object.Error{Message: "point expects 2 or 3 arguments (x, y[, z])"}
		}
	}}

	builtins["linestring"] = &object.Builtin{Name: "linestring", Fn: func(args ...object.Object) object.Object {
		if len(args) < 2 {
			return &object.Error{Message: "linestring expects at least 2 points"}
		}
		coords := make([]object.Coordinate, len(args))
		for i, arg := range args {
			p, ok := arg.(*object.Point)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("linestring argument %d must be a point", i)}
			}
			coords[i] = p.Coord
		}
		return &object.LineString{Coords: coords}
	}}

	builtins["polygon"] = &object.Builtin{Name: "polygon", Fn: func(args ...object.Object) object.Object {
		if len(args) < 1 {
			return &object.Error{Message: "polygon expects at least 1 argument (list of points or list of rings)"}
		}
		// Single list of points → exterior ring
		if list, ok := args[0].(*object.List); ok {
			coords, err := listToCoords(list)
			if err != nil {
				return err
			}
			poly := &object.Polygon{ExteriorRing: coords}
			// Additional args are holes
			for i := 1; i < len(args); i++ {
				holeList, ok := args[i].(*object.List)
				if !ok {
					return &object.Error{Message: fmt.Sprintf("polygon argument %d must be a list of points", i)}
				}
				holeCoords, err := listToCoords(holeList)
				if err != nil {
					return err
				}
				poly.InteriorRings = append(poly.InteriorRings, holeCoords)
			}
			return poly
		}
		// Multiple point args → exterior ring
		coords := make([]object.Coordinate, len(args))
		for i, arg := range args {
			p, ok := arg.(*object.Point)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("polygon argument %d must be a point or list", i)}
			}
			coords[i] = p.Coord
		}
		return &object.Polygon{ExteriorRing: coords}
	}}

	builtins["multipoint"] = &object.Builtin{Name: "multipoint", Fn: func(args ...object.Object) object.Object {
		var points []*object.Point
		for i, arg := range args {
			p, ok := arg.(*object.Point)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("multipoint argument %d must be a point", i)}
			}
			points = append(points, p)
		}
		return &object.MultiPoint{Points: points}
	}}

	builtins["geometryCollection"] = &object.Builtin{Name: "geometryCollection", Fn: func(args ...object.Object) object.Object {
		var geoms []object.Geometry
		for i, arg := range args {
			g, ok := asGeometry(arg)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("geometryCollection argument %d is not a geometry", i)}
			}
			geoms = append(geoms, g)
		}
		return &object.GeometryCollection{Geometries: geoms}
	}}

	builtins["feature"] = &object.Builtin{Name: "feature", Fn: func(args ...object.Object) object.Object {
		if len(args) < 1 || len(args) > 3 {
			return &object.Error{Message: "feature expects 1-3 arguments (geometry[, properties[, id]])"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "first argument to feature must be a geometry"}
		}
		f := &object.Feature{Geom: g, Properties: &object.Map{}}
		if len(args) >= 2 {
			m, ok := args[1].(*object.Map)
			if !ok {
				return &object.Error{Message: "second argument to feature must be a map"}
			}
			f.Properties = m
		}
		if len(args) >= 3 {
			f.ID = args[2]
		}
		return f
	}}

	builtins["featureCollection"] = &object.Builtin{Name: "featureCollection", Fn: func(args ...object.Object) object.Object {
		var features []*object.Feature
		// Accepts either individual features or a list of features
		if len(args) == 1 {
			if list, ok := args[0].(*object.List); ok {
				for i, elem := range list.Elements {
					f, ok := elem.(*object.Feature)
					if !ok {
						return &object.Error{Message: fmt.Sprintf("element %d is not a feature", i)}
					}
					features = append(features, f)
				}
				return &object.FeatureCollection{Features: features}
			}
		}
		for i, arg := range args {
			f, ok := arg.(*object.Feature)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("argument %d is not a feature", i)}
			}
			features = append(features, f)
		}
		return &object.FeatureCollection{Features: features}
	}}

	// ── Parsing / Serialization ──
	builtins["parseWKT"] = &object.Builtin{Name: "parseWKT", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "parseWKT expects 1 argument (string)"}
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return &object.Error{Message: "parseWKT expects a string argument"}
		}
		geom, err := ParseWKT(s.Value)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("WKT parse error: %s", err)}
		}
		return geom
	}}

	builtins["toWKT"] = &object.Builtin{Name: "toWKT", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "toWKT expects 1 argument (geometry)"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "toWKT expects a geometry argument"}
		}
		return &object.String{Value: g.ToWKT()}
	}}

	builtins["parseGeoJSON"] = &object.Builtin{Name: "parseGeoJSON", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "parseGeoJSON expects 1 argument (string)"}
		}
		s, ok := args[0].(*object.String)
		if !ok {
			return &object.Error{Message: "parseGeoJSON expects a string argument"}
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(s.Value), &data); err != nil {
			return &object.Error{Message: fmt.Sprintf("GeoJSON parse error: %s", err)}
		}
		typeName, _ := data["type"].(string)
		if typeName == "FeatureCollection" {
			return parseGeoJSONFeatureCollection(data)
		}
		if typeName == "Feature" {
			return parseGeoJSONFeature(data)
		}
		geom, err := parseGeoJSONGeometry(data)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("GeoJSON parse error: %s", err)}
		}
		return geom
	}}

	builtins["toGeoJSON"] = &object.Builtin{Name: "toGeoJSON", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "toGeoJSON expects 1 argument"}
		}
		var result interface{}
		switch obj := args[0].(type) {
		case object.Geometry:
			result = geomToGeoJSON(obj)
		case *object.Feature:
			props := make(map[string]interface{})
			for _, pair := range obj.Properties.Pairs {
				if s, ok := pair.Key.(*object.String); ok {
					props[s.Value] = objectToJSONValue(pair.Value)
				}
			}
			result = map[string]interface{}{
				"type":       "Feature",
				"geometry":   geomToGeoJSON(obj.Geom),
				"properties": props,
			}
		case *object.FeatureCollection:
			features := make([]interface{}, len(obj.Features))
			for i, f := range obj.Features {
				props := make(map[string]interface{})
				for _, pair := range f.Properties.Pairs {
					if s, ok := pair.Key.(*object.String); ok {
						props[s.Value] = objectToJSONValue(pair.Value)
					}
				}
				features[i] = map[string]interface{}{
					"type":       "Feature",
					"geometry":   geomToGeoJSON(f.Geom),
					"properties": props,
				}
			}
			result = map[string]interface{}{
				"type":     "FeatureCollection",
				"features": features,
			}
		default:
			return &object.Error{Message: "toGeoJSON expects a geometry, feature, or feature collection"}
		}
		b, err := json.Marshal(result)
		if err != nil {
			return &object.Error{Message: fmt.Sprintf("GeoJSON serialize error: %s", err)}
		}
		return &object.String{Value: string(b)}
	}}

	// ── Measurements ──
	builtins["distance"] = &object.Builtin{Name: "distance", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "distance expects 2 arguments (point, point)"}
		}
		p1, ok1 := args[0].(*object.Point)
		p2, ok2 := args[1].(*object.Point)
		if !ok1 || !ok2 {
			return &object.Error{Message: "distance expects two points"}
		}
		return &object.Float{Value: euclideanDistance(p1.Coord, p2.Coord)}
	}}

	builtins["distanceGeodesic"] = &object.Builtin{Name: "distanceGeodesic", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "distanceGeodesic expects 2 arguments (point, point)"}
		}
		p1, ok1 := args[0].(*object.Point)
		p2, ok2 := args[1].(*object.Point)
		if !ok1 || !ok2 {
			return &object.Error{Message: "distanceGeodesic expects two points"}
		}
		return &object.Float{Value: haversineDistance(p1.Coord, p2.Coord)}
	}}

	builtins["area"] = &object.Builtin{Name: "area", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "area expects 1 argument (polygon)"}
		}
		switch g := args[0].(type) {
		case *object.Polygon:
			a := math.Abs(shoelaceArea(g.ExteriorRing))
			for _, ring := range g.InteriorRings {
				a -= math.Abs(shoelaceArea(ring))
			}
			return &object.Float{Value: a}
		case *object.MultiPolygon:
			total := 0.0
			for _, p := range g.Polygons {
				a := math.Abs(shoelaceArea(p.ExteriorRing))
				for _, ring := range p.InteriorRings {
					a -= math.Abs(shoelaceArea(ring))
				}
				total += a
			}
			return &object.Float{Value: total}
		default:
			return &object.Error{Message: "area expects a polygon or multipolygon"}
		}
	}}

	builtins["areaGeodesic"] = &object.Builtin{Name: "areaGeodesic", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "areaGeodesic expects 1 argument (polygon)"}
		}
		switch g := args[0].(type) {
		case *object.Polygon:
			a := geodesicArea(g.ExteriorRing)
			for _, ring := range g.InteriorRings {
				a -= geodesicArea(ring)
			}
			return &object.Float{Value: a}
		case *object.MultiPolygon:
			total := 0.0
			for _, p := range g.Polygons {
				a := geodesicArea(p.ExteriorRing)
				for _, ring := range p.InteriorRings {
					a -= geodesicArea(ring)
				}
				total += a
			}
			return &object.Float{Value: total}
		default:
			return &object.Error{Message: "areaGeodesic expects a polygon or multipolygon"}
		}
	}}

	builtins["perimeter"] = &object.Builtin{Name: "perimeter", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "perimeter expects 1 argument"}
		}
		switch g := args[0].(type) {
		case *object.LineString:
			return &object.Float{Value: lineLength(g.Coords)}
		case *object.Polygon:
			return &object.Float{Value: lineLength(g.ExteriorRing)}
		default:
			return &object.Error{Message: "perimeter expects a linestring or polygon"}
		}
	}}

	builtins["lengthGeodesic"] = &object.Builtin{Name: "lengthGeodesic", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "lengthGeodesic expects 1 argument"}
		}
		switch g := args[0].(type) {
		case *object.LineString:
			return &object.Float{Value: lineLengthGeodesic(g.Coords)}
		case *object.Polygon:
			return &object.Float{Value: lineLengthGeodesic(g.ExteriorRing)}
		default:
			return &object.Error{Message: "lengthGeodesic expects a linestring or polygon"}
		}
	}}

	builtins["centroid"] = &object.Builtin{Name: "centroid", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "centroid expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "centroid expects a geometry"}
		}
		switch geom := g.(type) {
		case *object.Point:
			return geom
		case *object.LineString:
			if len(geom.Coords) == 0 {
				return &object.Error{Message: "centroid of empty linestring"}
			}
			cx, cy := 0.0, 0.0
			for _, c := range geom.Coords {
				cx += c.X
				cy += c.Y
			}
			n := float64(len(geom.Coords))
			return &object.Point{Coord: object.Coordinate{X: cx / n, Y: cy / n}}
		case *object.Polygon:
			if len(geom.ExteriorRing) == 0 {
				return &object.Error{Message: "centroid of empty polygon"}
			}
			c := centroidOfRing(geom.ExteriorRing)
			return &object.Point{Coord: c}
		case *object.MultiPoint:
			if len(geom.Points) == 0 {
				return &object.Error{Message: "centroid of empty multipoint"}
			}
			cx, cy := 0.0, 0.0
			for _, p := range geom.Points {
				cx += p.Coord.X
				cy += p.Coord.Y
			}
			n := float64(len(geom.Points))
			return &object.Point{Coord: object.Coordinate{X: cx / n, Y: cy / n}}
		default:
			coords := g.Coordinates()
			if len(coords) == 0 {
				return &object.Error{Message: "centroid of empty geometry"}
			}
			cx, cy := 0.0, 0.0
			for _, c := range coords {
				cx += c.X
				cy += c.Y
			}
			n := float64(len(coords))
			return &object.Point{Coord: object.Coordinate{X: cx / n, Y: cy / n}}
		}
	}}

	builtins["envelope"] = &object.Builtin{Name: "envelope", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "envelope expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "envelope expects a geometry"}
		}
		coords := g.Coordinates()
		if len(coords) == 0 {
			return &object.Error{Message: "envelope of empty geometry"}
		}
		return envelope(coords)
	}}

	builtins["bbox"] = &object.Builtin{Name: "bbox", Fn: func(args ...object.Object) object.Object {
		if len(args) != 4 {
			return &object.Error{Message: "bbox expects 4 arguments (minX, minY, maxX, maxY)"}
		}
		minX, ok1 := toFloat(args[0])
		minY, ok2 := toFloat(args[1])
		maxX, ok3 := toFloat(args[2])
		maxY, ok4 := toFloat(args[3])
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return &object.Error{Message: "bbox expects numeric arguments"}
		}
		return &object.BBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	}}

	builtins["bboxToPolygon"] = &object.Builtin{Name: "bboxToPolygon", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "bboxToPolygon expects 1 argument (bbox)"}
		}
		bb, ok := args[0].(*object.BBox)
		if !ok {
			return &object.Error{Message: "bboxToPolygon expects a BBox"}
		}
		return &object.Polygon{ExteriorRing: []object.Coordinate{
			{X: bb.MinX, Y: bb.MinY},
			{X: bb.MaxX, Y: bb.MinY},
			{X: bb.MaxX, Y: bb.MaxY},
			{X: bb.MinX, Y: bb.MaxY},
			{X: bb.MinX, Y: bb.MinY},
		}}
	}}

	// ── Spatial Predicates ──
	builtins["contains"] = &object.Builtin{Name: "contains", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "contains expects 2 arguments"}
		}
		return object.NativeBoolToBooleanObject(spatialContains(args[0], args[1]))
	}}

	builtins["within"] = &object.Builtin{Name: "within", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "within expects 2 arguments"}
		}
		// within(a, b) == contains(b, a)
		return object.NativeBoolToBooleanObject(spatialContains(args[1], args[0]))
	}}

	builtins["intersects"] = &object.Builtin{Name: "intersects", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "intersects expects 2 arguments"}
		}
		return object.NativeBoolToBooleanObject(spatialIntersects(args[0], args[1]))
	}}

	builtins["disjoint"] = &object.Builtin{Name: "disjoint", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "disjoint expects 2 arguments"}
		}
		return object.NativeBoolToBooleanObject(!spatialIntersects(args[0], args[1]))
	}}

	builtins["touches"] = &object.Builtin{Name: "touches", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "touches expects 2 arguments"}
		}
		return object.NativeBoolToBooleanObject(spatialTouches(args[0], args[1]))
	}}

	builtins["crosses"] = &object.Builtin{Name: "crosses", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "crosses expects 2 arguments"}
		}
		return object.NativeBoolToBooleanObject(spatialCrosses(args[0], args[1]))
	}}

	builtins["overlaps"] = &object.Builtin{Name: "overlaps", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "overlaps expects 2 arguments"}
		}
		// Simplified: intersects but neither contains the other
		inter := spatialIntersects(args[0], args[1])
		cont1 := spatialContains(args[0], args[1])
		cont2 := spatialContains(args[1], args[0])
		return object.NativeBoolToBooleanObject(inter && !cont1 && !cont2)
	}}

	builtins["equals"] = &object.Builtin{Name: "equals", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "equals expects 2 arguments"}
		}
		g1, ok1 := asGeometry(args[0])
		g2, ok2 := asGeometry(args[1])
		if !ok1 || !ok2 {
			return &object.Error{Message: "equals expects two geometries"}
		}
		return object.NativeBoolToBooleanObject(geometryEquals(g1, g2))
	}}

	// ── Spatial Operations ──
	builtins["buffer"] = &object.Builtin{Name: "buffer", Fn: func(args ...object.Object) object.Object {
		if len(args) < 2 || len(args) > 3 {
			return &object.Error{Message: "buffer expects 2-3 arguments (geometry, distance[, segments])"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "buffer: first argument must be a geometry"}
		}
		radius, ok2 := toFloat(args[1])
		if !ok2 || radius < 0 {
			return &object.Error{Message: "buffer: distance must be a non-negative number"}
		}
		segments := 32
		if len(args) == 3 {
			if s, ok := args[2].(*object.Integer); ok && s.Value > 2 {
				segments = int(s.Value)
			}
		}
		switch pt := g.(type) {
		case *object.Point:
			ring := bufferPoint(pt.Coord, radius, segments)
			return &object.Polygon{ExteriorRing: ring}
		default:
			// For non-points, buffer each coordinate and compute convex hull
			coords := g.Coordinates()
			var allPts []object.Coordinate
			for _, c := range coords {
				allPts = append(allPts, bufferPoint(c, radius, segments)...)
			}
			hull := convexHull(allPts)
			return &object.Polygon{ExteriorRing: hull}
		}
	}}

	builtins["convexHull"] = &object.Builtin{Name: "convexHull", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "convexHull expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "convexHull expects a geometry"}
		}
		coords := g.Coordinates()
		if len(coords) < 3 {
			return &object.Error{Message: "convexHull requires at least 3 points"}
		}
		hull := convexHull(coords)
		return &object.Polygon{ExteriorRing: hull}
	}}

	builtins["simplify"] = &object.Builtin{Name: "simplify", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "simplify expects 2 arguments (geometry, tolerance)"}
		}
		tolerance, ok2 := toFloat(args[1])
		if !ok2 || tolerance < 0 {
			return &object.Error{Message: "simplify: tolerance must be a non-negative number"}
		}
		switch g := args[0].(type) {
		case *object.LineString:
			simplified := simplifyDouglasPeucker(g.Coords, tolerance)
			return &object.LineString{Coords: simplified}
		case *object.Polygon:
			ext := simplifyDouglasPeucker(g.ExteriorRing, tolerance)
			poly := &object.Polygon{ExteriorRing: ext}
			for _, ring := range g.InteriorRings {
				simplified := simplifyDouglasPeucker(ring, tolerance)
				if len(simplified) >= 4 {
					poly.InteriorRings = append(poly.InteriorRings, simplified)
				}
			}
			return poly
		default:
			return &object.Error{Message: "simplify expects a linestring or polygon"}
		}
	}}

	// ── Accessors ──
	builtins["coordinates"] = &object.Builtin{Name: "coordinates", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "coordinates expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "coordinates expects a geometry"}
		}
		coords := g.Coordinates()
		elems := make([]object.Object, len(coords))
		for i, c := range coords {
			elems[i] = &object.Point{Coord: c}
		}
		return &object.List{Elements: elems}
	}}

	builtins["numPoints"] = &object.Builtin{Name: "numPoints", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "numPoints expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "numPoints expects a geometry"}
		}
		return &object.Integer{Value: int64(len(g.Coordinates()))}
	}}

	builtins["geomType"] = &object.Builtin{Name: "geomType", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "geomType expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "geomType expects a geometry"}
		}
		return &object.String{Value: g.GeomType()}
	}}

	builtins["isValid"] = &object.Builtin{Name: "isValid", Fn: func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return &object.Error{Message: "isValid expects 1 argument"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return object.FALSE_OBJ
		}
		switch geom := g.(type) {
		case *object.Point:
			return object.NativeBoolToBooleanObject(!math.IsNaN(geom.Coord.X) && !math.IsNaN(geom.Coord.Y))
		case *object.LineString:
			return object.NativeBoolToBooleanObject(len(geom.Coords) >= 2)
		case *object.Polygon:
			return object.NativeBoolToBooleanObject(len(geom.ExteriorRing) >= 4)
		default:
			return object.TRUE_OBJ
		}
	}}

	// ── CRS / Projection ──
	builtins["project"] = &object.Builtin{Name: "project", Fn: func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return &object.Error{Message: "project expects 3 arguments (geometry, fromSRID, toSRID)"}
		}
		g, ok := asGeometry(args[0])
		if !ok {
			return &object.Error{Message: "project: first argument must be a geometry"}
		}
		from, ok1 := args[1].(*object.Integer)
		to, ok2 := args[2].(*object.Integer)
		if !ok1 || !ok2 {
			return &object.Error{Message: "project: SRID arguments must be integers"}
		}
		fromSRID, toSRID := from.Value, to.Value
		var proj func(object.Coordinate) object.Coordinate
		switch {
		case fromSRID == 4326 && toSRID == 3857:
			proj = projectCoord4326To3857
		case fromSRID == 3857 && toSRID == 4326:
			proj = projectCoord3857To4326
		case fromSRID == toSRID:
			return args[0] // no-op
		default:
			return &object.Error{Message: fmt.Sprintf("project: unsupported projection %d -> %d (supported: 4326 <-> 3857)", fromSRID, toSRID)}
		}
		return projectGeometry(g, proj)
	}}

	// ── Bearing ──
	builtins["bearing"] = &object.Builtin{Name: "bearing", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "bearing expects 2 arguments (point, point)"}
		}
		p1, ok1 := args[0].(*object.Point)
		p2, ok2 := args[1].(*object.Point)
		if !ok1 || !ok2 {
			return &object.Error{Message: "bearing expects two points"}
		}
		lat1 := deg2rad(p1.Coord.Y)
		lon1 := deg2rad(p1.Coord.X)
		lat2 := deg2rad(p2.Coord.Y)
		lon2 := deg2rad(p2.Coord.X)
		dlon := lon2 - lon1
		x := math.Cos(lat2) * math.Sin(dlon)
		y := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dlon)
		bearing := rad2deg(math.Atan2(x, y))
		bearing = math.Mod(bearing+360, 360)
		return &object.Float{Value: bearing}
	}}

	// ── Destination ──
	builtins["destination"] = &object.Builtin{Name: "destination", Fn: func(args ...object.Object) object.Object {
		if len(args) != 3 {
			return &object.Error{Message: "destination expects 3 arguments (point, bearingDeg, distanceKm)"}
		}
		p, ok := args[0].(*object.Point)
		if !ok {
			return &object.Error{Message: "destination: first argument must be a point"}
		}
		bearingDeg, ok2 := toFloat(args[1])
		distKm, ok3 := toFloat(args[2])
		if !ok2 || !ok3 {
			return &object.Error{Message: "destination: bearing and distance must be numbers"}
		}
		lat1 := deg2rad(p.Coord.Y)
		lon1 := deg2rad(p.Coord.X)
		br := deg2rad(bearingDeg)
		angDist := distKm / earthRadiusKm
		lat2 := math.Asin(math.Sin(lat1)*math.Cos(angDist) + math.Cos(lat1)*math.Sin(angDist)*math.Cos(br))
		lon2 := lon1 + math.Atan2(math.Sin(br)*math.Sin(angDist)*math.Cos(lat1), math.Cos(angDist)-math.Sin(lat1)*math.Sin(lat2))
		return &object.Point{Coord: object.Coordinate{X: rad2deg(lon2), Y: rad2deg(lat2)}}
	}}

	// ── Midpoint ──
	builtins["midpoint"] = &object.Builtin{Name: "midpoint", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "midpoint expects 2 arguments (point, point)"}
		}
		p1, ok1 := args[0].(*object.Point)
		p2, ok2 := args[1].(*object.Point)
		if !ok1 || !ok2 {
			return &object.Error{Message: "midpoint expects two points"}
		}
		return &object.Point{Coord: object.Coordinate{
			X: (p1.Coord.X + p2.Coord.X) / 2,
			Y: (p1.Coord.Y + p2.Coord.Y) / 2,
		}}
	}}

	// ── Along ──
	builtins["along"] = &object.Builtin{Name: "along", Fn: func(args ...object.Object) object.Object {
		if len(args) != 2 {
			return &object.Error{Message: "along expects 2 arguments (linestring, fraction)"}
		}
		ls, ok := args[0].(*object.LineString)
		if !ok {
			return &object.Error{Message: "along: first argument must be a linestring"}
		}
		frac, ok2 := toFloat(args[1])
		if !ok2 || frac < 0 || frac > 1 {
			return &object.Error{Message: "along: fraction must be between 0 and 1"}
		}
		if len(ls.Coords) < 2 {
			return &object.Error{Message: "along: linestring must have at least 2 points"}
		}
		totalLen := lineLength(ls.Coords)
		targetLen := totalLen * frac
		accumulated := 0.0
		for i := 1; i < len(ls.Coords); i++ {
			segLen := euclideanDistance(ls.Coords[i-1], ls.Coords[i])
			if accumulated+segLen >= targetLen {
				t := 0.0
				if segLen > 0 {
					t = (targetLen - accumulated) / segLen
				}
				return &object.Point{Coord: object.Coordinate{
					X: ls.Coords[i-1].X + t*(ls.Coords[i].X-ls.Coords[i-1].X),
					Y: ls.Coords[i-1].Y + t*(ls.Coords[i].Y-ls.Coords[i-1].Y),
				}}
			}
			accumulated += segLen
		}
		last := ls.Coords[len(ls.Coords)-1]
		return &object.Point{Coord: last}
	}}

	return builtins
}

// ── Spatial Predicate Implementations ──

func spatialContains(outer, inner object.Object) bool {
	switch o := outer.(type) {
	case *object.Polygon:
		switch i := inner.(type) {
		case *object.Point:
			return pointInRing(i.Coord, o.ExteriorRing) && !pointInAnyHole(i.Coord, o.InteriorRings)
		case *object.LineString:
			for _, c := range i.Coords {
				if !pointInRing(c, o.ExteriorRing) || pointInAnyHole(c, o.InteriorRings) {
					return false
				}
			}
			return true
		case *object.Polygon:
			for _, c := range i.ExteriorRing {
				if !pointInRing(c, o.ExteriorRing) || pointInAnyHole(c, o.InteriorRings) {
					return false
				}
			}
			return true
		case *object.MultiPoint:
			for _, p := range i.Points {
				if !pointInRing(p.Coord, o.ExteriorRing) || pointInAnyHole(p.Coord, o.InteriorRings) {
					return false
				}
			}
			return true
		}
	case *object.BBox:
		switch i := inner.(type) {
		case *object.Point:
			return i.Coord.X >= o.MinX && i.Coord.X <= o.MaxX &&
				i.Coord.Y >= o.MinY && i.Coord.Y <= o.MaxY
		}
	}
	return false
}

func pointInAnyHole(pt object.Coordinate, holes [][]object.Coordinate) bool {
	for _, hole := range holes {
		if pointInRing(pt, hole) {
			return true
		}
	}
	return false
}

func spatialIntersects(a, b object.Object) bool {
	// Quick bbox check
	ga, oka := asGeometry(a)
	gb, okb := asGeometry(b)
	if !oka || !okb {
		return false
	}
	ca := ga.Coordinates()
	cb := gb.Coordinates()
	if len(ca) == 0 || len(cb) == 0 {
		return false
	}
	bba := envelope(ca)
	bbb := envelope(cb)
	if bba.MaxX < bbb.MinX || bba.MinX > bbb.MaxX || bba.MaxY < bbb.MinY || bba.MinY > bbb.MaxY {
		return false
	}

	// containment test
	if spatialContains(a, b) || spatialContains(b, a) {
		return true
	}

	// Edge intersection
	edgesA := extractEdges(ga)
	edgesB := extractEdges(gb)
	for _, ea := range edgesA {
		for _, eb := range edgesB {
			if segmentsIntersect(ea[0], ea[1], eb[0], eb[1]) {
				return true
			}
		}
	}

	// Point-in-polygon checks
	if poly, ok := a.(*object.Polygon); ok {
		for _, c := range cb {
			if pointInRing(c, poly.ExteriorRing) {
				return true
			}
		}
	}
	if poly, ok := b.(*object.Polygon); ok {
		for _, c := range ca {
			if pointInRing(c, poly.ExteriorRing) {
				return true
			}
		}
	}

	return false
}

func spatialTouches(a, b object.Object) bool {
	if !spatialIntersects(a, b) {
		return false
	}
	// Interior points of a should not be inside b and vice versa
	// Simplified: check if geometries share boundary but not interior
	ga, _ := asGeometry(a)
	gb, _ := asGeometry(b)
	edgesA := extractEdges(ga)
	edgesB := extractEdges(gb)
	for _, ea := range edgesA {
		for _, eb := range edgesB {
			if segmentsIntersect(ea[0], ea[1], eb[0], eb[1]) {
				return true
			}
		}
	}
	return false
}

func spatialCrosses(a, b object.Object) bool {
	// A line crosses another geometry if it passes through it
	ls, ok := a.(*object.LineString)
	if !ok {
		return false
	}
	switch g := b.(type) {
	case *object.Polygon:
		insideCount := 0
		outsideCount := 0
		for _, c := range ls.Coords {
			if pointInRing(c, g.ExteriorRing) {
				insideCount++
			} else {
				outsideCount++
			}
		}
		return insideCount > 0 && outsideCount > 0
	case *object.LineString:
		edgesA := extractEdges(ls)
		edgesB := extractEdges(g)
		for _, ea := range edgesA {
			for _, eb := range edgesB {
				if segmentsIntersect(ea[0], ea[1], eb[0], eb[1]) {
					return true
				}
			}
		}
	}
	return false
}

func extractEdges(g object.Geometry) [][2]object.Coordinate {
	var edges [][2]object.Coordinate
	switch geom := g.(type) {
	case *object.LineString:
		for i := 1; i < len(geom.Coords); i++ {
			edges = append(edges, [2]object.Coordinate{geom.Coords[i-1], geom.Coords[i]})
		}
	case *object.Polygon:
		for i := 1; i < len(geom.ExteriorRing); i++ {
			edges = append(edges, [2]object.Coordinate{geom.ExteriorRing[i-1], geom.ExteriorRing[i]})
		}
	case *object.MultiLineString:
		for _, l := range geom.Lines {
			for i := 1; i < len(l.Coords); i++ {
				edges = append(edges, [2]object.Coordinate{l.Coords[i-1], l.Coords[i]})
			}
		}
	case *object.MultiPolygon:
		for _, p := range geom.Polygons {
			for i := 1; i < len(p.ExteriorRing); i++ {
				edges = append(edges, [2]object.Coordinate{p.ExteriorRing[i-1], p.ExteriorRing[i]})
			}
		}
	}
	return edges
}

func geometryEquals(a, b object.Geometry) bool {
	if a.GeomType() != b.GeomType() {
		return false
	}
	ac := a.Coordinates()
	bc := b.Coordinates()
	if len(ac) != len(bc) {
		return false
	}
	for i := range ac {
		if !ac[i].Equals(bc[i]) {
			return false
		}
	}
	return true
}

func listToCoords(list *object.List) ([]object.Coordinate, *object.Error) {
	coords := make([]object.Coordinate, len(list.Elements))
	for i, elem := range list.Elements {
		p, ok := elem.(*object.Point)
		if !ok {
			return nil, &object.Error{Message: fmt.Sprintf("element %d is not a point", i)}
		}
		coords[i] = p.Coord
	}
	return coords, nil
}

func objectToJSONValue(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Integer:
		return o.Value
	case *object.Float:
		return o.Value
	case *object.String:
		return o.Value
	case *object.Boolean:
		return o.Value
	case *object.Nil:
		return nil
	default:
		return o.Inspect()
	}
}

func parseGeoJSONFeature(data map[string]interface{}) object.Object {
	geomRaw, ok := data["geometry"].(map[string]interface{})
	if !ok {
		return &object.Error{Message: "Feature: missing 'geometry'"}
	}
	geom, err := parseGeoJSONGeometry(geomRaw)
	if err != nil {
		return &object.Error{Message: fmt.Sprintf("Feature geometry error: %s", err)}
	}
	props := &object.Map{}
	if propsRaw, ok := data["properties"].(map[string]interface{}); ok {
		for k, v := range propsRaw {
			props.Pairs = append(props.Pairs, object.MapPair{
				Key:   &object.String{Value: k},
				Value: jsonValueToObject(v),
			})
		}
	}
	f := &object.Feature{Geom: geom, Properties: props}
	if id, ok := data["id"]; ok {
		f.ID = jsonValueToObject(id)
	}
	return f
}

func parseGeoJSONFeatureCollection(data map[string]interface{}) object.Object {
	featuresRaw, ok := data["features"].([]interface{})
	if !ok {
		return &object.Error{Message: "FeatureCollection: missing 'features'"}
	}
	var features []*object.Feature
	for i, fr := range featuresRaw {
		fd, ok := fr.(map[string]interface{})
		if !ok {
			return &object.Error{Message: fmt.Sprintf("FeatureCollection: feature %d is invalid", i)}
		}
		fObj := parseGeoJSONFeature(fd)
		if e, ok := fObj.(*object.Error); ok {
			return e
		}
		features = append(features, fObj.(*object.Feature))
	}
	return &object.FeatureCollection{Features: features}
}

func jsonValueToObject(v interface{}) object.Object {
	switch val := v.(type) {
	case float64:
		if val == math.Trunc(val) && val >= -1e15 && val <= 1e15 {
			return &object.Integer{Value: int64(val)}
		}
		return &object.Float{Value: val}
	case string:
		return &object.String{Value: val}
	case bool:
		return object.NativeBoolToBooleanObject(val)
	case nil:
		return object.NIL
	default:
		return &object.String{Value: fmt.Sprintf("%v", val)}
	}
}
