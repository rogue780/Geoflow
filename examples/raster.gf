-- raster.gf: Raster grid construction and inspection

import std.geo.raster as raster

{-
  The raster module provides constructors for 2D grids of numeric data.
  Two constructors are available:
    - raster.empty(width, height, bounds, [fillValue])
    - raster.fromArray(data, width, height)
-}

-- ---------------------------------------------------------------
-- 1. Creating an empty raster with default fill (zeros)
-- ---------------------------------------------------------------

let bounds = {"minX": -180.0, "minY": -90.0, "maxX": 180.0, "maxY": 90.0}
let grid = raster.empty(10, 10, bounds)
println("Empty 10x10 raster:", grid)

-- ---------------------------------------------------------------
-- 2. Creating an empty raster with a custom fill value
-- ---------------------------------------------------------------

-- Fill every cell with -9999.0, a common "no data" sentinel
let nodata_grid = raster.empty(4, 4, bounds, -9999.0)
println("4x4 no-data raster:", nodata_grid)

-- Fill with a constant elevation (e.g., sea level = 0.0)
let sea_level = raster.empty(8, 6, bounds, 0.0)
println("8x6 sea-level raster:", sea_level)

-- ---------------------------------------------------------------
-- 3. Creating a raster from a flat data array
-- ---------------------------------------------------------------

-- A simple 3x3 grid with known values
let data_3x3 = [
  1.0, 2.0, 3.0,
  4.0, 5.0, 6.0,
  7.0, 8.0, 9.0
]
let small = raster.fromArray(data_3x3, 3, 3)
println("3x3 raster from array:", small)

-- ---------------------------------------------------------------
-- 4. Simulated elevation grid
-- ---------------------------------------------------------------

{-
  Imagine a 4x3 elevation grid (4 columns, 3 rows) covering
  a small area. Values represent meters above sea level.
-}

let elev_bounds = {"minX": -105.5, "minY": 39.5, "maxX": -105.0, "maxY": 40.0}

let elevations = [
  2100.0, 2250.0, 2400.0, 2350.0,
  2050.0, 2180.0, 2320.0, 2280.0,
  1980.0, 2050.0, 2200.0, 2150.0
]

let dem = raster.fromArray(elevations, 4, 3)
println("Elevation DEM (4x3):", dem)

-- ---------------------------------------------------------------
-- 5. Temperature grid
-- ---------------------------------------------------------------

{-
  A 5x4 grid of temperature readings (degrees C).
  Each value represents the average temperature in a grid cell.
-}

let temp_bounds = {"minX": -10.0, "minY": 50.0, "maxX": 0.0, "maxY": 55.0}

let temps = [
  12.5, 13.0, 13.2, 12.8, 12.0,
  14.0, 14.5, 14.8, 14.2, 13.5,
  15.5, 16.0, 16.3, 15.8, 15.0,
  16.0, 16.5, 17.0, 16.2, 15.5
]

let temp_grid = raster.fromArray(temps, 5, 4)
println("Temperature grid (5x4):", temp_grid)

-- ---------------------------------------------------------------
-- 6. Initializing a uniform raster for accumulation
-- ---------------------------------------------------------------

-- Start with a blank 20x20 canvas (all zeros), ready for data
let canvas_bounds = {"minX": 0.0, "minY": 0.0, "maxX": 100.0, "maxY": 100.0}
let canvas = raster.empty(20, 20, canvas_bounds, 0.0)
println("Blank 20x20 canvas:", canvas)

-- A larger raster initialized to 1.0 (e.g., mask where all cells are valid)
let mask = raster.empty(50, 50, canvas_bounds, 1.0)
println("50x50 mask raster:", mask)

-- ---------------------------------------------------------------
-- 7. Single-row and single-column rasters
-- ---------------------------------------------------------------

-- A transect: single row of sample values
let transect_data = [100.0, 150.0, 200.0, 180.0, 120.0, 90.0]
let transect = raster.fromArray(transect_data, 6, 1)
println("Transect (6x1):", transect)

-- A vertical profile: single column
let profile_data = [0.0, 25.0, 50.0, 75.0, 100.0]
let profile = raster.fromArray(profile_data, 1, 5)
println("Vertical profile (1x5):", profile)

println("Raster examples complete")
