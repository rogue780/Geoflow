-- crs.gf: Coordinate Reference Systems (std.geo.crs)

import std.geo.crs as crs

-- ── Well-Known CRS Constants ──
let wgs = crs.WGS84
println("WGS 84:", wgs)
println("  EPSG:", wgs.epsg)
println("  Name:", wgs.name)
println("  Geographic?", wgs.isGeographic)
println("  Projected?", wgs.isProjected)
println("  Units:", wgs.units)

let merc = crs.WebMercator
println("\nWeb Mercator:", merc)
println("  EPSG:", merc.epsg)
println("  Name:", merc.name)
println("  Geographic?", merc.isGeographic)
println("  Projected?", merc.isProjected)
println("  Units:", merc.units)

let nad83 = crs.NAD83
println("\nNAD83:", nad83)
println("  EPSG:", nad83.epsg)

let nad27 = crs.NAD27
println("NAD27:", nad27)
println("  EPSG:", nad27.epsg)

let etrs = crs.ETRS89
println("ETRS89:", etrs)
println("  EPSG:", etrs.epsg)

-- ── fromEPSG Lookup ──
let found = crs.fromEPSG(4326)
println("\nfromEPSG(4326):", found)

let found2 = crs.fromEPSG(3857)
println("fromEPSG(3857):", found2)

let notFound = crs.fromEPSG(99999)
println("fromEPSG(99999):", notFound)

-- ── UTM Zone Creation ──
let utm18n = crs.UTM(18, true)
println("\nUTM Zone 18N:", utm18n)
println("  EPSG:", utm18n.epsg)
println("  Name:", utm18n.name)
println("  Projected?", utm18n.isProjected)
println("  Units:", utm18n.units)

let utm33s = crs.UTM(33, false)
println("\nUTM Zone 33S:", utm33s)
println("  EPSG:", utm33s.epsg)
println("  Name:", utm33s.name)

-- Common US UTM zones
let utm10n = crs.UTM(10, true)
let utm17n = crs.UTM(17, true)
println("\nUTM 10N (West Coast) EPSG:", utm10n.epsg)
println("UTM 17N (East Coast) EPSG:", utm17n.epsg)

println("CRS example complete")
