package stdlib_test

import (
	"math"
	"testing"

	"github.com/rogue780/geoflow/internal/object"
)

// ── WKT Parsing ──

func TestWKTPoint(t *testing.T) {
	result := testEval(`parseWKT("POINT (1 2)")`)
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("expected Point, got %T (%s)", result, result.Inspect())
	}
	if pt.Coord.X != 1 || pt.Coord.Y != 2 {
		t.Errorf("expected (1, 2), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestWKTPointZ(t *testing.T) {
	result := testEval(`parseWKT("POINT Z (1 2 3)")`)
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("expected Point, got %T (%s)", result, result.Inspect())
	}
	if !pt.Coord.HasZ || pt.Coord.Z != 3 {
		t.Errorf("expected Z=3, got %v", pt.Coord)
	}
}

func TestWKTLineString(t *testing.T) {
	result := testEval(`parseWKT("LINESTRING (0 0, 1 1, 2 2)")`)
	ls, ok := result.(*object.LineString)
	if !ok {
		t.Fatalf("expected LineString, got %T (%s)", result, result.Inspect())
	}
	if len(ls.Coords) != 3 {
		t.Errorf("expected 3 coords, got %d", len(ls.Coords))
	}
}

func TestWKTPolygon(t *testing.T) {
	result := testEval(`parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")`)
	poly, ok := result.(*object.Polygon)
	if !ok {
		t.Fatalf("expected Polygon, got %T (%s)", result, result.Inspect())
	}
	if len(poly.ExteriorRing) != 5 {
		t.Errorf("expected 5 coords in exterior ring, got %d", len(poly.ExteriorRing))
	}
}

func TestWKTPolygonWithHole(t *testing.T) {
	result := testEval(`parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 8 2, 8 8, 2 8, 2 2))")`)
	poly, ok := result.(*object.Polygon)
	if !ok {
		t.Fatalf("expected Polygon, got %T (%s)", result, result.Inspect())
	}
	if len(poly.InteriorRings) != 1 {
		t.Errorf("expected 1 interior ring, got %d", len(poly.InteriorRings))
	}
}

func TestWKTMultiPoint(t *testing.T) {
	result := testEval(`parseWKT("MULTIPOINT ((0 0), (1 1), (2 2))")`)
	mp, ok := result.(*object.MultiPoint)
	if !ok {
		t.Fatalf("expected MultiPoint, got %T (%s)", result, result.Inspect())
	}
	if len(mp.Points) != 3 {
		t.Errorf("expected 3 points, got %d", len(mp.Points))
	}
}

func TestWKTGeometryCollection(t *testing.T) {
	result := testEval(`parseWKT("GEOMETRYCOLLECTION (POINT (1 2), LINESTRING (0 0, 1 1))")`)
	gc, ok := result.(*object.GeometryCollection)
	if !ok {
		t.Fatalf("expected GeometryCollection, got %T (%s)", result, result.Inspect())
	}
	if len(gc.Geometries) != 2 {
		t.Errorf("expected 2 geometries, got %d", len(gc.Geometries))
	}
}

// ── WKT Literal Syntax ──

func TestWKTLiteralSyntax(t *testing.T) {
	result := testEval(`#POINT(10 20)#`)
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("expected Point, got %T (%s)", result, result.Inspect())
	}
	if pt.Coord.X != 10 || pt.Coord.Y != 20 {
		t.Errorf("expected (10, 20), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestWKTLiteralLineString(t *testing.T) {
	result := testEval(`#LINESTRING(0 0, 5 5, 10 0)#`)
	ls, ok := result.(*object.LineString)
	if !ok {
		t.Fatalf("expected LineString, got %T (%s)", result, result.Inspect())
	}
	if len(ls.Coords) != 3 {
		t.Fatalf("expected 3 coords, got %d", len(ls.Coords))
	}
}

// ── toWKT round-trip ──

func TestToWKT(t *testing.T) {
	result := testEval(`toWKT(point(1, 2))`)
	s, ok := result.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T (%s)", result, result.Inspect())
	}
	if s.Value != "POINT (1 2)" {
		t.Errorf("expected 'POINT (1 2)', got %q", s.Value)
	}
}

// ── Constructors ──

func TestPointConstructor(t *testing.T) {
	result := testEval("point(3.5, 7.2)")
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("expected Point, got %T (%s)", result, result.Inspect())
	}
	if pt.Coord.X != 3.5 || pt.Coord.Y != 7.2 {
		t.Errorf("expected (3.5, 7.2), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestPointConstructor3D(t *testing.T) {
	result := testEval("point(1, 2, 100)")
	pt := result.(*object.Point)
	if !pt.Coord.HasZ || pt.Coord.Z != 100 {
		t.Errorf("expected Z=100, got %v", pt.Coord)
	}
}

func TestLineStringConstructor(t *testing.T) {
	result := testEval("linestring(point(0, 0), point(1, 1), point(2, 0))")
	ls, ok := result.(*object.LineString)
	if !ok {
		t.Fatalf("expected LineString, got %T (%s)", result, result.Inspect())
	}
	if len(ls.Coords) != 3 {
		t.Errorf("expected 3 coords, got %d", len(ls.Coords))
	}
}

func TestPolygonConstructor(t *testing.T) {
	result := testEval("polygon([point(0,0), point(10,0), point(10,10), point(0,10), point(0,0)])")
	poly, ok := result.(*object.Polygon)
	if !ok {
		t.Fatalf("expected Polygon, got %T (%s)", result, result.Inspect())
	}
	if len(poly.ExteriorRing) != 5 {
		t.Errorf("expected 5 coords, got %d", len(poly.ExteriorRing))
	}
}

// ── Measurements ──

func TestEuclideanDistance(t *testing.T) {
	result := testEval("distance(point(0, 0), point(3, 4))")
	assertFloat(t, "distance", result, 5.0)
}

func TestGeodesicDistance(t *testing.T) {
	// NYC to London: ~5570 km
	result := testEval("distanceGeodesic(point(-74.006, 40.7128), point(-0.1278, 51.5074))")
	f, ok := result.(*object.Float)
	if !ok {
		t.Fatalf("expected Float, got %T (%s)", result, result.Inspect())
	}
	// Should be approximately 5570 km
	if f.Value < 5500 || f.Value > 5600 {
		t.Errorf("NYC-London distance: expected ~5570km, got %g km", f.Value)
	}
}

func TestArea(t *testing.T) {
	// 10x10 square = 100
	result := testEval(`area(parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"))`)
	assertFloat(t, "area", result, 100.0)
}

func TestAreaWithHole(t *testing.T) {
	// 10x10 - 6x6 = 100 - 36 = 64
	result := testEval(`area(parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 8 2, 8 8, 2 8, 2 2))"))`)
	assertFloat(t, "area with hole", result, 64.0)
}

func TestPerimeter(t *testing.T) {
	result := testEval(`perimeter(linestring(point(0,0), point(3,0), point(3,4)))`)
	assertFloat(t, "perimeter", result, 7.0) // 3 + 4 = 7
}

func TestCentroid(t *testing.T) {
	result := testEval(`centroid(parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"))`)
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("centroid: expected Point, got %T (%s)", result, result.Inspect())
	}
	if math.Abs(pt.Coord.X-5) > 0.01 || math.Abs(pt.Coord.Y-5) > 0.01 {
		t.Errorf("centroid: expected (5, 5), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestEnvelope(t *testing.T) {
	result := testEval(`envelope(linestring(point(1, 2), point(5, 8), point(3, 3)))`)
	bb, ok := result.(*object.BBox)
	if !ok {
		t.Fatalf("envelope: expected BBox, got %T (%s)", result, result.Inspect())
	}
	if bb.MinX != 1 || bb.MinY != 2 || bb.MaxX != 5 || bb.MaxY != 8 {
		t.Errorf("envelope: expected (1,2)-(5,8), got (%g,%g)-(%g,%g)", bb.MinX, bb.MinY, bb.MaxX, bb.MaxY)
	}
}

// ── Spatial Predicates ──

func TestContains(t *testing.T) {
	assertBool(t, "point in polygon",
		testEval(`contains(parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"), point(5, 5))`), true)
	assertBool(t, "point outside polygon",
		testEval(`contains(parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"), point(15, 5))`), false)
}

func TestWithin(t *testing.T) {
	assertBool(t, "point within polygon",
		testEval(`within(point(5, 5), parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"))`), true)
}

func TestIntersects(t *testing.T) {
	assertBool(t, "overlapping polygons",
		testEval(`intersects(
			parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))"),
			parseWKT("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
		)`), true)
	assertBool(t, "disjoint polygons",
		testEval(`intersects(
			parseWKT("POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"),
			parseWKT("POLYGON ((5 5, 6 5, 6 6, 5 6, 5 5))")
		)`), false)
}

func TestDisjoint(t *testing.T) {
	assertBool(t, "disjoint",
		testEval(`disjoint(
			parseWKT("POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"),
			parseWKT("POLYGON ((5 5, 6 5, 6 6, 5 6, 5 5))")
		)`), true)
}

// ── Spatial Operations ──

func TestConvexHull(t *testing.T) {
	result := testEval(`convexHull(multipoint(point(0,0), point(10,0), point(5,5), point(10,10), point(0,10)))`)
	poly, ok := result.(*object.Polygon)
	if !ok {
		t.Fatalf("convexHull: expected Polygon, got %T (%s)", result, result.Inspect())
	}
	// Should have 5 points (4 corners + closing)
	if len(poly.ExteriorRing) < 4 {
		t.Errorf("convexHull: expected at least 4 points, got %d", len(poly.ExteriorRing))
	}
}

func TestSimplify(t *testing.T) {
	result := testEval(`simplify(linestring(point(0,0), point(1,0.1), point(2,0), point(3,0.1), point(4,0)), 0.5)`)
	ls, ok := result.(*object.LineString)
	if !ok {
		t.Fatalf("simplify: expected LineString, got %T (%s)", result, result.Inspect())
	}
	// With tolerance 0.5, should simplify to just start and end
	if len(ls.Coords) > 3 {
		t.Errorf("simplify: expected at most 3 coords, got %d", len(ls.Coords))
	}
}

func TestBuffer(t *testing.T) {
	result := testEval("buffer(point(0, 0), 5, 16)")
	poly, ok := result.(*object.Polygon)
	if !ok {
		t.Fatalf("buffer: expected Polygon, got %T (%s)", result, result.Inspect())
	}
	if len(poly.ExteriorRing) != 17 { // 16 segments + closing point
		t.Errorf("buffer: expected 17 points, got %d", len(poly.ExteriorRing))
	}
}

// ── GeoJSON ──

func TestParseGeoJSON(t *testing.T) {
	// Round-trip through toGeoJSON/parseGeoJSON
	result := testEval(`parseGeoJSON(toGeoJSON(point(1, 2)))`)
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("parseGeoJSON: expected Point, got %T (%s)", result, result.Inspect())
	}
	if pt.Coord.X != 1 || pt.Coord.Y != 2 {
		t.Errorf("parseGeoJSON: expected (1,2), got (%g,%g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestToGeoJSON(t *testing.T) {
	result := testEval(`toGeoJSON(point(1, 2))`)
	s, ok := result.(*object.String)
	if !ok {
		t.Fatalf("toGeoJSON: expected String, got %T (%s)", result, result.Inspect())
	}
	if s.Value == "" {
		t.Error("toGeoJSON: empty string")
	}
	// Verify it round-trips
	result2 := testEval(`parseGeoJSON(toGeoJSON(point(1, 2)))`)
	pt, ok := result2.(*object.Point)
	if !ok {
		t.Fatalf("round-trip: expected Point, got %T (%s)", result2, result2.Inspect())
	}
	if pt.Coord.X != 1 || pt.Coord.Y != 2 {
		t.Errorf("round-trip: expected (1,2), got (%g,%g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestParseGeoJSONFeature(t *testing.T) {
	// Round-trip through toGeoJSON/parseGeoJSON with a feature
	result := testEval(`parseGeoJSON(toGeoJSON(feature(point(1, 2), {"name": "test"})))`)
	f, ok := result.(*object.Feature)
	if !ok {
		t.Fatalf("parseGeoJSON Feature: expected Feature, got %T (%s)", result, result.Inspect())
	}
	pt := f.Geom.(*object.Point)
	if pt.Coord.X != 1 || pt.Coord.Y != 2 {
		t.Errorf("Feature geometry: expected (1,2), got (%g,%g)", pt.Coord.X, pt.Coord.Y)
	}
	name, found := f.Properties.Get("name")
	if !found {
		t.Error("Feature property 'name' not found")
	} else if s, ok := name.(*object.String); !ok || s.Value != "test" {
		t.Errorf("Feature property 'name': expected 'test', got %s", name.Inspect())
	}
}

// ── Point Methods ──

func TestPointMethods(t *testing.T) {
	assertFloat(t, "point.x", testEval("point(3.5, 7.2).x"), 3.5)
	assertFloat(t, "point.y", testEval("point(3.5, 7.2).y"), 7.2)
	assertFloat(t, "point.lon", testEval("point(3.5, 7.2).lon"), 3.5)
	assertFloat(t, "point.lat", testEval("point(3.5, 7.2).lat"), 7.2)
}

func TestLineStringMethods(t *testing.T) {
	assertInt(t, "numPoints", testEval("linestring(point(0,0), point(1,1), point(2,0)).numPoints"), 3)
}

func TestBBoxMethods(t *testing.T) {
	assertFloat(t, "bbox.minX", testEval("bbox(1, 2, 3, 4).minX"), 1.0)
	assertFloat(t, "bbox.maxY", testEval("bbox(1, 2, 3, 4).maxY"), 4.0)
	assertFloat(t, "bbox.width", testEval("bbox(1, 2, 5, 4).width"), 4.0)
	assertFloat(t, "bbox.height", testEval("bbox(1, 2, 5, 8).height"), 6.0)
}

// ── Projection ──

func TestProjection(t *testing.T) {
	// Project (0, 0) should stay (0, 0) in both systems
	result := testEval("project(point(0, 0), 4326, 3857)")
	pt, ok := result.(*object.Point)
	if !ok {
		t.Fatalf("project: expected Point, got %T (%s)", result, result.Inspect())
	}
	if math.Abs(pt.Coord.X) > 0.01 || math.Abs(pt.Coord.Y) > 0.01 {
		t.Errorf("project (0,0): expected (0,0), got (%g,%g)", pt.Coord.X, pt.Coord.Y)
	}

	// Round-trip: 4326 -> 3857 -> 4326
	result2 := testEval("project(project(point(-73.9857, 40.7484), 4326, 3857), 3857, 4326)")
	pt2 := result2.(*object.Point)
	if math.Abs(pt2.Coord.X-(-73.9857)) > 0.001 || math.Abs(pt2.Coord.Y-40.7484) > 0.001 {
		t.Errorf("project round-trip: expected (-73.9857, 40.7484), got (%g, %g)", pt2.Coord.X, pt2.Coord.Y)
	}
}

// ── Bearing / Destination ──

func TestBearing(t *testing.T) {
	// Due east
	result := testEval("bearing(point(0, 0), point(1, 0))")
	f := result.(*object.Float)
	if math.Abs(f.Value-90) > 0.5 {
		t.Errorf("bearing: expected ~90, got %g", f.Value)
	}
}

func TestMidpoint(t *testing.T) {
	result := testEval("midpoint(point(0, 0), point(10, 10))")
	pt := result.(*object.Point)
	if pt.Coord.X != 5 || pt.Coord.Y != 5 {
		t.Errorf("midpoint: expected (5, 5), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestAlong(t *testing.T) {
	result := testEval("along(linestring(point(0, 0), point(10, 0)), 0.5)")
	pt := result.(*object.Point)
	if math.Abs(pt.Coord.X-5) > 0.01 || math.Abs(pt.Coord.Y) > 0.01 {
		t.Errorf("along 50%%: expected (5, 0), got (%g, %g)", pt.Coord.X, pt.Coord.Y)
	}
}

// ── Feature ──

func TestFeature(t *testing.T) {
	result := testEval(`feature(point(1, 2), {"name": "test"})`)
	f, ok := result.(*object.Feature)
	if !ok {
		t.Fatalf("feature: expected Feature, got %T (%s)", result, result.Inspect())
	}
	pt := f.Geom.(*object.Point)
	if pt.Coord.X != 1 || pt.Coord.Y != 2 {
		t.Errorf("feature geometry: expected (1,2), got (%g,%g)", pt.Coord.X, pt.Coord.Y)
	}
}

func TestFeaturePropertyAccess(t *testing.T) {
	result := testEval(`feature(point(1, 2), {"name": "test"}).name`)
	s, ok := result.(*object.String)
	if !ok {
		t.Fatalf("feature.name: expected String, got %T (%s)", result, result.Inspect())
	}
	if s.Value != "test" {
		t.Errorf("feature.name: expected 'test', got %q", s.Value)
	}
}

// ── Accessors ──

func TestGeomType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`geomType(point(1, 2))`, "Point"},
		{`geomType(linestring(point(0,0), point(1,1)))`, "LineString"},
		{`geomType(parseWKT("POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"))`, "Polygon"},
	}
	for _, tt := range tests {
		result := testEval(tt.input)
		s, ok := result.(*object.String)
		if !ok {
			t.Errorf("%s: expected String, got %T (%s)", tt.input, result, result.Inspect())
			continue
		}
		if s.Value != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.input, tt.expected, s.Value)
		}
	}
}

func TestNumPoints(t *testing.T) {
	assertInt(t, "numPoints", testEval(`numPoints(linestring(point(0,0), point(1,1), point(2,2)))`), 3)
}

func TestIsValid(t *testing.T) {
	assertBool(t, "valid point", testEval("isValid(point(1, 2))"), true)
	assertBool(t, "valid polygon", testEval(`isValid(parseWKT("POLYGON ((0 0, 1 0, 1 1, 0 1, 0 0))"))`), true)
}
