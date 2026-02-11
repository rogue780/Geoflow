-- geospatial.gf: Geospatial core showcase with fully qualified imports

import std.geo.geom as geom
import std.math

-- ── WKT Literal Syntax ──
let pt = #POINT(10 20)#
println("WKT literal point:", pt)
println("  x:", pt.x, " y:", pt.y)

-- ── Point Constructors ──
let nyc = geom.point(-74.006, 40.7128)
let london = geom.point(-0.1278, 51.5074)
let tokyo = geom.point(139.6917, 35.6895)
println("NYC:", nyc)
println("London:", london)
println("Tokyo:", tokyo)

-- ── Geodesic Distance ──
let d = geom.distanceGeodesic(nyc, london)
println("NYC to London:", math.round(d), "km")

let d2 = geom.distanceGeodesic(nyc, tokyo)
println("NYC to Tokyo:", math.round(d2), "km")

-- ── Bearing ──
let b = geom.bearing(nyc, london)
println("Bearing NYC -> London:", math.round(b * 10) / 10, "degrees")

-- ── Midpoint ──
let mid = geom.midpoint(nyc, london)
println("Midpoint NYC-London:", mid)

-- ── LineString & Perimeter ──
let route = geom.linestring(
  geom.point(0, 0),
  geom.point(3, 0),
  geom.point(3, 4),
  geom.point(0, 4)
)
println("Route:", route)
println("Route length:", geom.perimeter(route))
println("Route start:", route.startPoint)
println("Route end:", route.endPoint)
println("Route closed?", route.isClosed)

-- ── Point along a line ──
let halfway = geom.along(route, 0.5)
println("Halfway along route:", halfway)

-- ── Polygon & Area ──
let square = geom.parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
println("Square:", square)
println("Area:", geom.area(square))

-- Polygon with hole
let donut = geom.parseWKT("POLYGON ((0 0, 20 0, 20 20, 0 20, 0 0), (5 5, 15 5, 15 15, 5 15, 5 5))")
println("Donut area:", geom.area(donut), "(= 400 - 100)")

-- ── Centroid ──
let c = geom.centroid(square)
println("Square centroid:", c)

-- ── Bounding Box ──
let bb = geom.envelope(route)
println("Route bounding box:", bb)
println("  width:", bb.width, " height:", bb.height)
println("  center:", bb.center)

-- ── Spatial Predicates ──
let inside = geom.point(5, 5)
let outside = geom.point(15, 5)
println("Point (5,5) inside square?", geom.contains(square, inside))
println("Point (15,5) inside square?", geom.contains(square, outside))
println("Point (5,5) within square?", geom.within(inside, square))

-- ── Intersects & Disjoint ──
let poly1 = geom.parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")
let poly2 = geom.parseWKT("POLYGON ((5 5, 15 5, 15 15, 5 15, 5 5))")
let poly3 = geom.parseWKT("POLYGON ((20 20, 30 20, 30 30, 20 30, 20 20))")
println("poly1 intersects poly2?", geom.intersects(poly1, poly2))
println("poly1 intersects poly3?", geom.intersects(poly1, poly3))
println("poly1 disjoint poly3?", geom.disjoint(poly1, poly3))

-- ── Convex Hull ──
let pts = geom.multipoint(
  geom.point(0, 0), geom.point(10, 0), geom.point(5, 5),
  geom.point(10, 10), geom.point(0, 10), geom.point(5, 3)
)
let hull = geom.convexHull(pts)
println("Convex hull:", hull)

-- ── Simplify ──
let detailed = geom.linestring(
  geom.point(0, 0), geom.point(1, 0.1), geom.point(2, -0.1),
  geom.point(3, 0.05), geom.point(4, 0), geom.point(5, 0.1)
)
let simplified = geom.simplify(detailed, 0.2)
println("Original points:", geom.numPoints(detailed))
println("Simplified points:", geom.numPoints(simplified))

-- ── Buffer ──
let buffered = geom.buffer(geom.point(0, 0), 5, 8)
println("Buffer around origin:", geom.numPoints(buffered), "vertices")

-- ── GeoJSON Round-trip ──
let geojson = geom.toGeoJSON(nyc)
println("GeoJSON:", geojson)
let parsed = geom.parseGeoJSON(geojson)
println("Parsed back:", parsed)

-- ── Feature & Properties ──
let f = geom.feature(nyc, {"name": "New York City", "population": 8336817})
println("Feature:", f)
println("  name:", f.name)
println("  geometry type:", f.geomType)

-- ── Projection (WGS84 <-> Web Mercator) ──
let projected = geom.project(nyc, 4326, 3857)
println("NYC in Web Mercator:", projected)
let backToWGS = geom.project(projected, 3857, 4326)
println("Back to WGS84:", backToWGS)

-- ── Destination ──
let dest = geom.destination(nyc, 45, 100)
println("100km NE of NYC:", dest)

-- ── Geometry Type & Validation ──
println("geomType(point):", geom.geomType(nyc))
println("geomType(linestring):", geom.geomType(route))
println("geomType(polygon):", geom.geomType(square))
println("isValid(square):", geom.isValid(square))

-- ── Geometry Properties ──
let coords = geom.coordinates(route)
println("Route coordinates:", coords)
println("Route numPoints:", geom.numPoints(route))

-- ── Euclidean Distance ──
let dist = geom.distance(geom.point(0, 0), geom.point(3, 4))
println("Euclidean distance (0,0)→(3,4):", dist)

-- ── Putting it all together: find cities within a fence ──
let fence = geom.parseWKT("POLYGON ((-80 35, -70 35, -70 45, -80 45, -80 35))")
let cities = [
  geom.point(-74.006, 40.7128),
  geom.point(-87.6298, 41.8781),
  geom.point(-75.1652, 39.9526),
  geom.point(-118.2437, 34.0522)
]

println("Cities within the US East Coast fence:")
for city in cities {
  if geom.contains(fence, city) {
    println("  ", city)
  }
}
