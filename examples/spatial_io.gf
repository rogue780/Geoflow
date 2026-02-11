-- spatial_io.gf: Spatial I/O formats (std.geo.io)

import std.geo.geom as geom
import std.geo.io as geoio

-- Create test geometries
let pt = geom.point(-74.006, 40.7128)
let line = geom.linestring(geom.point(0, 0), geom.point(1, 1), geom.point(2, 0))
let poly = geom.parseWKT("POLYGON ((0 0, 10 0, 10 10, 0 10, 0 0))")

-- ══════════════════════════════════════
-- ── WKT (Well-Known Text) ──
-- ══════════════════════════════════════
-- WKT is handled by geom (parseWKT/toWKT)
println("=== WKT ===")
let wkt = geom.toWKT(pt)
println("Point WKT:", wkt)
let ptBack = geom.parseWKT(wkt)
println("Parsed back:", ptBack)

let lineWkt = geom.toWKT(line)
println("LineString WKT:", lineWkt)

let polyWkt = geom.toWKT(poly)
println("Polygon WKT:", polyWkt)

-- ══════════════════════════════════════
-- ── GeoJSON ──
-- ══════════════════════════════════════
println("\n=== GeoJSON ===")
let geojson = geom.toGeoJSON(pt)
println("Point GeoJSON:", geojson)
let ptFromJson = geom.parseGeoJSON(geojson)
println("Parsed back:", ptFromJson)

let lineJson = geom.toGeoJSON(line)
println("LineString GeoJSON:", lineJson)

let polyJson = geom.toGeoJSON(poly)
println("Polygon GeoJSON:", polyJson)

-- ══════════════════════════════════════
-- ── WKB (Well-Known Binary, hex-encoded) ──
-- ══════════════════════════════════════
println("\n=== WKB ===")
let wkb = geoio.toWKB(pt)
println("Point WKB (hex):", wkb)
let ptFromWkb = geoio.fromWKB(wkb)
println("Parsed back:", ptFromWkb)

let lineWkb = geoio.toWKB(line)
println("LineString WKB length:", len(lineWkb), "chars")
let lineFromWkb = geoio.fromWKB(lineWkb)
println("Parsed back:", lineFromWkb)

let polyWkb = geoio.toWKB(poly)
println("Polygon WKB length:", len(polyWkb), "chars")
let polyFromWkb = geoio.fromWKB(polyWkb)
println("Parsed back:", polyFromWkb)

-- ══════════════════════════════════════
-- ── KML (Keyhole Markup Language) ──
-- ══════════════════════════════════════
println("\n=== KML ===")
let kml = geoio.toKML(pt)
println("Point KML:", kml)
let ptFromKml = geoio.fromKML(kml)
println("Parsed back:", ptFromKml)

let lineKml = geoio.toKML(line)
println("LineString KML:", lineKml)
let lineFromKml = geoio.fromKML(lineKml)
println("Parsed back:", lineFromKml)

let polyKml = geoio.toKML(poly)
println("Polygon KML:", polyKml)
let polyFromKml = geoio.fromKML(polyKml)
println("Parsed back:", polyFromKml)

println("\nSpatial I/O example complete")
