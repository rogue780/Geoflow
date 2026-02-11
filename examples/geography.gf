-- geography.gf: Geography module (std.geo.geog) - geodetic calculations in meters

import std.geo.geog as geog
import std.math

-- ── Geographic Points (WGS84, SRID 4326) ──
let nyc = geog.Point(-74.006, 40.7128)
let london = geog.Point(-0.1278, 51.5074)
let tokyo = geog.Point(139.6917, 35.6895)
let sydney = geog.Point(151.2093, -33.8688)

println("NYC:", nyc)
println("London:", london)
println("Tokyo:", tokyo)
println("Sydney:", sydney)

-- ── Great-Circle Distance (meters) ──
let dNYCLon = geog.distanceGreatCircle(nyc, london)
println("NYC → London:", math.round(dNYCLon / 1000), "km")

let dNYCTok = geog.distanceGreatCircle(nyc, tokyo)
println("NYC → Tokyo:", math.round(dNYCTok / 1000), "km")

let dLonSyd = geog.distanceGreatCircle(london, sydney)
println("London → Sydney:", math.round(dLonSyd / 1000), "km")

-- ── Bearing (degrees, 0-360) ──
let bNYCLon = geog.bearing(nyc, london)
println("Bearing NYC → London:", math.round(bNYCLon * 10) / 10, "°")

let bLonTok = geog.bearing(london, tokyo)
println("Bearing London → Tokyo:", math.round(bLonTok * 10) / 10, "°")

-- ── Destination (point at bearing + distance in meters) ──
let dest = geog.destination(nyc, 45, 100000)
println("100km NE of NYC:", dest)

let dest2 = geog.destination(london, 90, 500000)
println("500km E of London:", dest2)

-- ── Midpoint ──
let mid = geog.midpoint(nyc, london)
println("Midpoint NYC-London:", mid)

let midTokSyd = geog.midpoint(tokyo, sydney)
println("Midpoint Tokyo-Sydney:", midTokSyd)

-- ── Geographic LineString ──
let flightPath = geog.LineString(nyc, london, tokyo)
println("Flight path:", flightPath)

-- ── Geodesic Area (square meters) ──
-- Create a geographic polygon using geom for the polygon type
import std.geo.geom as geom
let colorado = geom.parseWKT("POLYGON ((-109.05 37, -109.05 41, -102.05 41, -102.05 37, -109.05 37))")
let areaM2 = geog.areaGeodesic(colorado)
println("Colorado area:", math.round(areaM2 / 1000000), "km²")

println("Geography example complete")
