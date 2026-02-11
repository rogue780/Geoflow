-- geospatial.gf: Geospatial core showcase

-- ── WKT Literal Syntax ──
let pt = #POINT(10 20)#
println("WKT literal point:", pt)
println("  x:", pt.x, " y:", pt.y)

-- ── Point Constructors ──
let nyc = point(-74.006, 40.7128)
let london = point(-0.1278, 51.5074)
let tokyo = point(139.6917, 35.6895)
println("NYC:", nyc)
println("London:", london)
println("Tokyo:", tokyo)

-- ── Geodesic Distance ──
let d = distanceGeodesic(nyc, london)
println("NYC to London:", round(d), "km")

let d2 = distanceGeodesic(nyc, tokyo)
println("NYC to Tokyo:", round(d2), "km")

-- ── Bearing ──
let b = bearing(nyc, london)
println("Bearing NYC -> London:", round(b * 10) / 10, "degrees")

-- ── Midpoint ──
let mid = midpoint(nyc, london)
println("Midpoint NYC-London:", mid)

-- ── LineString & Perimeter ──
let route = linestring(
  point(0, 0),
  point(3, 0),
  point(3, 4),
  point(0, 4)
)
println("Route:", route)
println("Route length:", perimeter(route))
println("Route start:", route.startPoint)
println("Route end:", route.endPoint)
println("Route closed?", route.isClosed)

-- ── Point along a line ──
let halfway = along(route, 0.5)
println("Halfway along route:", halfway)

-- ── Polygon & Area ──
let square = parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
println("Square:", square)
println("Area:", area(square))

-- Polygon with hole
let donut = parseWKT("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")
println("Donut area:", area(donut), "(= 400 - 100)")

-- ── Centroid ──
let c = centroid(square)
println("Square centroid:", c)

-- ── Bounding Box ──
let bb = envelope(route)
println("Route bounding box:", bb)
println("  width:", bb.width, " height:", bb.height)
println("  center:", bb.center)

-- ── Spatial Predicates ──
let inside = point(5, 5)
let outside = point(15, 5)
println("Point (5,5) inside square?", contains(square, inside))
println("Point (15,5) inside square?", contains(square, outside))
println("Point (5,5) within square?", within(inside, square))

-- ── Intersects & Disjoint ──
let poly1 = parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
let poly2 = parseWKT("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
let poly3 = parseWKT("POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20))")
println("poly1 intersects poly2?", intersects(poly1, poly2))
println("poly1 intersects poly3?", intersects(poly1, poly3))
println("poly1 disjoint poly3?", disjoint(poly1, poly3))

-- ── Convex Hull ──
let pts = multipoint(
  point(0, 0), point(10, 0), point(5, 5),
  point(10, 10), point(0, 10), point(5, 3)
)
let hull = convexHull(pts)
println("Convex hull:", hull)

-- ── Simplify ──
let detailed = linestring(
  point(0, 0), point(1, 0.1), point(2, -0.1),
  point(3, 0.05), point(4, 0), point(5, 0.1)
)
let simplified = simplify(detailed, 0.2)
println("Original points:", numPoints(detailed))
println("Simplified points:", numPoints(simplified))

-- ── Buffer ──
let buffered = buffer(point(0, 0), 5, 8)
println("Buffer around origin:", numPoints(buffered), "vertices")

-- ── GeoJSON Round-trip ──
let geojson = toGeoJSON(nyc)
println("GeoJSON:", geojson)
let parsed = parseGeoJSON(geojson)
println("Parsed back:", parsed)

-- ── Feature & Properties ──
let f = feature(nyc, {"name": "New York City", "population": 8336817})
println("Feature:", f)
println("  name:", f.name)
println("  geometry type:", f.geomType)

-- ── Projection (WGS84 <-> Web Mercator) ──
let projected = project(nyc, 4326, 3857)
println("NYC in Web Mercator:", projected)
let backToWGS = project(projected, 3857, 4326)
println("Back to WGS84:", backToWGS)

-- ── Destination ──
let dest = destination(nyc, 45, 100)
println("100km NE of NYC:", dest)

-- ── Geometry Type Checking ──
println("geomType(point):", geomType(nyc))
println("geomType(linestring):", geomType(route))
println("geomType(polygon):", geomType(square))

-- ── Putting it all together: find points within a fence ──
let fence = parseWKT("POLYGON ((-80 35, -70 35, -70 45, -80 45, -80 35))")
let cities = [
  point(-74.006, 40.7128),
  point(-87.6298, 41.8781),
  point(-75.1652, 39.9526),
  point(-118.2437, 34.0522)
]

println("Cities within the US East Coast fence:")
for city in cities {
  if contains(fence, city) {
    println("  ", city)
  }
}
