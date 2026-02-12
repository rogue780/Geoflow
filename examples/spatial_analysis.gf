-- spatial_analysis.gf: Spatial analysis functions from std.geo.analysis
--
-- Demonstrates nearest neighbor search, clustering (DBSCAN, K-means),
-- Voronoi diagrams, Delaunay triangulation, IDW interpolation,
-- concave hull, and alpha shapes.

import std.geo.analysis as analysis
import std.geo.geom as geom

-- ============================================================================
-- 1. Nearest Neighbor
-- ============================================================================
-- Find the closest point to a target from a list of candidates.

println("-- Nearest Neighbor --")

let target = #POINT(0 0)#
let candidates = [
  #POINT(1 0)#,
  #POINT(0 2)#,
  #POINT(3 3)#,
  #POINT(-1 5)#,
  #POINT(4 1)#
]

let nearest = analysis.nearestNeighbor(target, candidates)
println("Target:", target)
println("Nearest neighbor:", nearest)

-- ============================================================================
-- 2. K-Nearest Neighbors
-- ============================================================================
-- Find the k closest points to a target.

println("\n-- K-Nearest Neighbors --")

let k = 3
let kNearest = analysis.kNearestNeighbors(target, candidates, k)
println("Target:", target)
println(k, "nearest neighbors:", kNearest)

-- ============================================================================
-- 3. DBSCAN Clustering
-- ============================================================================
-- Density-based spatial clustering. Groups nearby points into clusters
-- based on a distance threshold (eps) and minimum number of neighbors (minPts).
-- Points that do not belong to any cluster are treated as noise and excluded.

println("\n-- DBSCAN Clustering --")

let dbscanPoints = [
  -- Cluster A: tightly grouped near the origin
  #POINT(0 0)#,
  #POINT(0.1 0.1)#,
  #POINT(0.2 0)#,
  #POINT(0 0.2)#,
  -- Cluster B: tightly grouped near (5, 5)
  #POINT(5 5)#,
  #POINT(5.1 5.1)#,
  #POINT(5.2 5)#,
  #POINT(5 5.2)#,
  -- Noise: far from both clusters
  #POINT(20 20)#
]

let eps = 1.0
let minPts = 2
let clusters = analysis.dbscan(dbscanPoints, eps, minPts)

println("Input points:", len(dbscanPoints))
println("Epsilon:", eps, " MinPts:", minPts)
println("Number of clusters found:", len(clusters))

let mut i = 0
for cluster in clusters {
  let i = i + 1
  println("  Cluster", i, "has", len(cluster), "points")
}

-- ============================================================================
-- 4. K-Means Clustering
-- ============================================================================
-- Partition points into k groups using Lloyd's algorithm.

println("\n-- K-Means Clustering --")

let kmeansPoints = [
  -- Group near origin
  #POINT(0 0)#,
  #POINT(1 0)#,
  #POINT(0 1)#,
  #POINT(1 1)#,
  -- Group near (10, 10)
  #POINT(10 10)#,
  #POINT(11 10)#,
  #POINT(10 11)#,
  #POINT(11 11)#
]

let k = 2
let kmClusters = analysis.kmeans(kmeansPoints, k)

println("Input points:", len(kmeansPoints))
println("Requested k:", k)
println("Clusters returned:", len(kmClusters))

let mut j = 0
for cluster in kmClusters {
  let j = j + 1
  println("  Cluster", j, ":", cluster)
}

-- ============================================================================
-- 5. Voronoi Diagrams
-- ============================================================================
-- Generate approximate Voronoi cells for a set of sites within a bounding box.
-- Each cell is a polygon containing all space closest to its site.

println("\n-- Voronoi Diagrams --")

let sites = [
  #POINT(2 2)#,
  #POINT(8 2)#,
  #POINT(5 8)#
]

-- The bounds define the region in which cells are computed.
let bounds = geom.bbox(0, 0, 10, 10)
let cells = analysis.voronoi(sites, bounds)

println("Sites:", len(sites))
println("Voronoi cells generated:", len(cells))

let mut ci = 0
for cell in cells {
  let ci = ci + 1
  println("  Cell", ci, ":", cell)
}

-- ============================================================================
-- 6. Delaunay Triangulation
-- ============================================================================
-- Compute a Delaunay triangulation from a set of points.
-- Returns a list of triangle polygons.

println("\n-- Delaunay Triangulation --")

let triPoints = [
  #POINT(0 0)#,
  #POINT(10 0)#,
  #POINT(5 8)#,
  #POINT(3 4)#,
  #POINT(7 3)#
]

let triangles = analysis.delaunay(triPoints)

println("Input points:", len(triPoints))
println("Triangles generated:", len(triangles))

let mut ti = 0
for tri in triangles {
  let ti = ti + 1
  println("  Triangle", ti, ":", tri)
}

-- ============================================================================
-- 7. IDW Interpolation (Inverse Distance Weighting)
-- ============================================================================
-- Estimate an unknown value at a target location based on known values
-- at surrounding points. Closer points have more influence.

println("\n-- IDW Interpolation --")

let samplePoints = [
  #POINT(0 0)#,
  #POINT(10 0)#,
  #POINT(0 10)#,
  #POINT(10 10)#
]

-- Known values at each sample point (e.g., temperature readings)
let values = [15.0, 25.0, 20.0, 30.0]

-- Interpolate at the center of the sample area
let interpTarget = #POINT(5 5)#
let estimated = analysis.idw(samplePoints, values, interpTarget)
println("Sample points:", samplePoints)
println("Known values:", values)
println("Target:", interpTarget)
println("Interpolated value:", estimated)

-- IDW with a custom power parameter (default is 2.0).
-- Higher power gives more weight to the nearest points.
let estimatedP3 = analysis.idw(samplePoints, values, interpTarget, 3.0)
println("Interpolated value (power=3.0):", estimatedP3)

-- Interpolating at a known point returns the exact value
let atKnown = analysis.idw(samplePoints, values, #POINT(0 0)#)
println("Value at known point (0,0):", atKnown)

-- ============================================================================
-- 8. Concave Hull and Alpha Shape
-- ============================================================================
-- Compute the concave hull (tight boundary) of a point set.
-- Alpha shape is a generalization parameterized by an alpha value.

println("\n-- Concave Hull --")

let hullPoints = [
  #POINT(0 0)#,
  #POINT(10 0)#,
  #POINT(10 10)#,
  #POINT(0 10)#,
  #POINT(5 5)#,
  #POINT(3 7)#,
  #POINT(8 2)#
]

let cHull = analysis.concaveHull(hullPoints)
println("Points:", len(hullPoints))
println("Concave hull:", cHull)

println("\n-- Alpha Shape --")

-- Without an alpha parameter (uses default)
let shape1 = analysis.alphaShape(hullPoints)
println("Alpha shape (default):", shape1)

-- With a specific alpha parameter
let shape2 = analysis.alphaShape(hullPoints, 0.5)
println("Alpha shape (alpha=0.5):", shape2)

println("\nSpatial analysis examples complete.")
