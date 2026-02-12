{- h3s2.gf: H3 and S2 spatial indexing examples

   H3 provides hexagonal hierarchical spatial indexing.
   S2 provides spherical geometry cell-based spatial indexing.

   Both systems encode geographic coordinates into discrete cell identifiers
   at varying resolutions/levels, enabling efficient spatial queries. -}

import std.geo.h3 as h3
import std.geo.s2 as s2

-- ============================================================
-- PART 1: H3 Hexagonal Spatial Indexing
-- ============================================================

println("=== H3 Hexagonal Spatial Indexing ===")
println("")

-- ── 1. H3 Index Creation ──
-- h3.fromLatLon(lat, lon, resolution) creates an H3 cell index.
-- Resolution ranges from 0 (coarsest) to 15 (finest).

let sf = h3.fromLatLon(37.7749, -122.4194, 5)
println("San Francisco (res 5):", sf)

let sf_fine = h3.fromLatLon(37.7749, -122.4194, 9)
println("San Francisco (res 9):", sf_fine)

let nyc = h3.fromLatLon(40.7128, -74.0060, 5)
println("New York City (res 5):", nyc)

-- ── 2. H3 Properties ──
-- h3.resolution returns the resolution level of an H3 index.
-- h3.isValid checks whether an H3 index has a valid resolution.

println("")
println("--- H3 Properties ---")
println("Resolution of sf:", h3.resolution(sf))
println("Resolution of sf_fine:", h3.resolution(sf_fine))
println("Is sf valid?", h3.isValid(sf))
println("Is sf_fine valid?", h3.isValid(sf_fine))

-- ── 3. H3 Neighbors (k-Ring) ──
-- h3.kRing(index, k) returns the disk of cells within grid distance k.
-- k=0 returns just the cell itself, k=1 returns the cell and its immediate
-- neighbors, and so on.

println("")
println("--- H3 k-Ring (Neighbors) ---")

let ring0 = h3.kRing(sf, 0)
println("k-Ring(sf, 0) count:", len(ring0))

let ring1 = h3.kRing(sf, 1)
println("k-Ring(sf, 1) count:", len(ring1))

let ring2 = h3.kRing(sf, 2)
println("k-Ring(sf, 2) count:", len(ring2))

-- ── 4. H3 Distance ──
-- h3.distance computes the approximate grid distance between two H3 cells
-- at the same resolution (Chebyshev distance on the grid).

println("")
println("--- H3 Grid Distance ---")

let a = h3.fromLatLon(37.7749, -122.4194, 5)
let b = h3.fromLatLon(37.8, -122.4, 5)
let dist_ab = h3.distance(a, b)
println("Grid distance (SF center vs nearby, res 5):", dist_ab)

let c = h3.fromLatLon(40.7128, -74.0060, 5)
let dist_ac = h3.distance(a, c)
println("Grid distance (SF vs NYC, res 5):", dist_ac)

-- Higher resolution means more cells, so larger grid distances
let a_hi = h3.fromLatLon(37.7749, -122.4194, 9)
let b_hi = h3.fromLatLon(37.8, -122.4, 9)
let dist_hi = h3.distance(a_hi, b_hi)
println("Grid distance (SF center vs nearby, res 9):", dist_hi)

-- ── 5. H3 String Conversion ──
-- h3.toString converts an H3 index to a 16-character hex string.
-- h3.fromString parses a hex string back into an H3 index (returns a Result).

println("")
println("--- H3 String Conversion ---")

let hex_str = h3.toString(sf)
println("H3 hex string:", hex_str)

-- Round-trip: convert to string and back
let restored = h3.fromString(hex_str)?
println("Restored from string:", restored)
println("Restored is valid?", h3.isValid(restored))
println("Restored resolution:", h3.resolution(restored))

-- ── 6. H3 Cell Centroid ──
-- h3.toPoint returns the approximate center point of a cell.

println("")
println("--- H3 Cell Centroid ---")
let center = h3.toPoint(sf)
println("Centroid of SF cell (res 5):", center)

let center_fine = h3.toPoint(sf_fine)
println("Centroid of SF cell (res 9):", center_fine)


-- ============================================================
-- PART 2: S2 Spherical Geometry Cells
-- ============================================================

println("")
println("=== S2 Spherical Geometry Cells ===")
println("")

-- ── 7. S2 Cell Creation ──
-- s2.fromLatLon(lat, lon, level) creates an S2 cell ID.
-- Level ranges from 0 (coarsest) to 30 (finest).

let sf_s2 = s2.fromLatLon(37.7749, -122.4194, 10)
println("SF S2 cell (level 10):", sf_s2)

let sf_s2_fine = s2.fromLatLon(37.7749, -122.4194, 18)
println("SF S2 cell (level 18):", sf_s2_fine)

let nyc_s2 = s2.fromLatLon(40.7128, -74.0060, 10)
println("NYC S2 cell (level 10):", nyc_s2)

-- ── 8. S2 Properties ──
-- s2.level returns the level of an S2 cell.
-- s2.isValid checks whether an S2 cell has a valid level and face.

println("")
println("--- S2 Properties ---")
println("Level of sf_s2:", s2.level(sf_s2))
println("Level of sf_s2_fine:", s2.level(sf_s2_fine))
println("Is sf_s2 valid?", s2.isValid(sf_s2))
println("Is sf_s2_fine valid?", s2.isValid(sf_s2_fine))

-- ── 9. S2 Token Conversion ──
-- s2.toToken converts an S2 cell ID to a 16-character hex token.
-- s2.fromToken parses a hex token back into an S2 cell ID (returns a Result).

println("")
println("--- S2 Token Conversion ---")

let token = s2.toToken(sf_s2)
println("S2 token:", token)

-- Round-trip: convert to token and back
let restored_s2 = s2.fromToken(token)?
println("Restored from token:", restored_s2)
println("Restored is valid?", s2.isValid(restored_s2))
println("Restored level:", s2.level(restored_s2))

-- ── 10. S2 Cell Centroid ──
-- s2.toPoint returns the approximate center point of a cell.

println("")
println("--- S2 Cell Centroid ---")

let s2_center = s2.toPoint(sf_s2)
println("Centroid of SF cell (level 10):", s2_center)

let s2_center_fine = s2.toPoint(sf_s2_fine)
println("Centroid of SF cell (level 18):", s2_center_fine)

-- ── 11. S2 Containment ──
-- s2.contains checks whether an S2 cell contains a given point.
-- It encodes the point at the same level as the cell and checks equality.

println("")
println("--- S2 Containment ---")

let cell = s2.fromLatLon(37.7749, -122.4194, 12)
let pt_inside = s2.toPoint(cell)
println("Cell contains its own centroid?", s2.contains(cell, pt_inside))


-- ============================================================
-- PART 3: Comparing H3 and S2 for the Same Location
-- ============================================================

println("")
println("=== H3 vs S2 Comparison ===")
println("")

-- Encode the same location with both systems
let lat = 48.8566
let lon = 2.3522
println("Location: Paris (", lat, ",", lon, ")")

let paris_h3 = h3.fromLatLon(lat, lon, 7)
let paris_s2 = s2.fromLatLon(lat, lon, 14)

println("H3 (res 7):", paris_h3)
println("  Valid?", h3.isValid(paris_h3))
println("  Resolution:", h3.resolution(paris_h3))
println("  Hex:", h3.toString(paris_h3))
println("  Center:", h3.toPoint(paris_h3))

println("S2 (level 14):", paris_s2)
println("  Valid?", s2.isValid(paris_s2))
println("  Level:", s2.level(paris_s2))
println("  Token:", s2.toToken(paris_s2))
println("  Center:", s2.toPoint(paris_s2))

-- Neighbors in H3 vs containment check in S2
let paris_ring = h3.kRing(paris_h3, 1)
println("H3 neighbors (k=1):", len(paris_ring), "cells")

println("")
println("H3 and S2 spatial indexing example complete")
