-- file_io.gf: File I/O with std.io

import std.io

-- ══════════════════════════════════════
-- ── Path Operations ──
-- ══════════════════════════════════════
let path = io.join("/tmp", "geoflow", "data", "test.csv")
println("join:", path)
println("dirname:", io.dirname(path))
println("basename:", io.basename(path))
println("extension:", io.extension(path))
println("absolute(\".\"):", io.absolute("."))

-- ══════════════════════════════════════
-- ── Write & Read Files ──
-- ══════════════════════════════════════
let testDir = io.join("/tmp", "geoflow_io_test")
let testFile = io.join(testDir, "hello.txt")

-- Create directory
io.createDirAll(testDir)
println("\nDirectory created:", io.isDir(testDir))

-- Write a file
io.writeFile(testFile, "Hello, GeoFlow!\nLine 2\nLine 3\n")
println("File exists:", io.exists(testFile))
println("Is file:", io.isFile(testFile))
println("Is dir:", io.isDir(testFile))
println("File size:", io.fileSize(testFile))

-- Read entire file
let content = io.readFile(testFile)
println("readFile:", content)

-- Read as lines
let lines = io.readLines(testFile)
println("readLines:", lines)

-- ══════════════════════════════════════
-- ── Append ──
-- ══════════════════════════════════════
io.appendFile(testFile, "Line 4 (appended)\n")
let updated = io.readLines(testFile)
println("After append:", updated)

-- ══════════════════════════════════════
-- ── Copy & Move ──
-- ══════════════════════════════════════
let copyDest = io.join(testDir, "hello_copy.txt")
io.copy(testFile, copyDest)
println("\nCopy exists:", io.exists(copyDest))

let moveDest = io.join(testDir, "hello_moved.txt")
io.move(copyDest, moveDest)
println("Original after move:", io.exists(copyDest))
println("Moved file exists:", io.exists(moveDest))

-- ══════════════════════════════════════
-- ── Directory Listing ──
-- ══════════════════════════════════════
let entries = io.listDir(testDir)
println("\nDirectory contents:", entries)

-- ══════════════════════════════════════
-- ── Cleanup ──
-- ══════════════════════════════════════
io.removeFile(moveDest)
io.removeFile(testFile)
println("\nAfter cleanup:", io.listDir(testDir))

-- ══════════════════════════════════════
-- ── Error Handling ──
-- ══════════════════════════════════════
let badRead = io.readFile("/tmp/nonexistent_file_12345.txt")
println("Read nonexistent:", badRead)

println("\nFile I/O example complete")
