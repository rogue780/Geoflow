-- spatial_index.gf: Spatial indexing with R-Tree (std.geo.index)

import std.geo.index as index
import std.geo.geom as geom

-- ── Create an R-Tree ──
let tree = index.new()
println("Empty R-Tree size:", index.size(tree))

-- ── Insert geometries with associated data ──
-- insert(tree, geometry, item) returns a new tree
let tree = index.insert(tree, geom.point(0, 0), "Origin")
let tree = index.insert(tree, geom.point(5, 5), "Center")
let tree = index.insert(tree, geom.point(10, 10), "Far corner")
let tree = index.insert(tree, geom.point(-3, 2), "West")
let tree = index.insert(tree, geom.point(7, 1), "East")
println("R-Tree size after inserts:", index.size(tree))

-- ── Spatial Query (bounding box) ──
-- query(tree, bbox) returns items whose bounds intersect the query
let searchBox = geom.bbox(-1, -1, 6, 6)
let results = index.query(tree, searchBox)
println("Query [-1,-1 to 6,6]:", results)
println("  Found", len(results), "items")

-- Larger query
let bigBox = geom.bbox(-5, -5, 15, 15)
let allResults = index.query(tree, bigBox)
println("Query [-5,-5 to 15,15]:", allResults)
println("  Found", len(allResults), "items")

-- Query with no results
let emptyBox = geom.bbox(100, 100, 200, 200)
let noResults = index.query(tree, emptyBox)
println("Query [100,100 to 200,200]:", noResults)
println("  Found", len(noResults), "items")

-- ── Nearest Neighbor ──
-- nearest(tree, point, n) returns the n nearest items by bbox center
let searchPt = geom.point(4, 4)
let nearest1 = index.nearest(tree, searchPt, 1)
println("Nearest 1 to (4,4):", nearest1)

let nearest3 = index.nearest(tree, searchPt, 3)
println("Nearest 3 to (4,4):", nearest3)

-- ── Building a city index ──
let cities = index.new()
let cities = index.insert(cities, geom.point(-74.006, 40.713), "New York")
let cities = index.insert(cities, geom.point(-87.630, 41.878), "Chicago")
let cities = index.insert(cities, geom.point(-118.244, 34.052), "Los Angeles")
let cities = index.insert(cities, geom.point(-75.165, 39.953), "Philadelphia")
let cities = index.insert(cities, geom.point(-122.419, 37.775), "San Francisco")

println("\nCity index size:", index.size(cities))

-- Find cities in the northeast US
let neBox = geom.bbox(-80, 38, -70, 42)
let neCities = index.query(cities, neBox)
println("Northeast US cities:", neCities)

-- Find nearest cities to Chicago
let nearChicago = index.nearest(cities, geom.point(-87.630, 41.878), 3)
println("3 nearest to Chicago:", nearChicago)

println("Spatial index example complete")
