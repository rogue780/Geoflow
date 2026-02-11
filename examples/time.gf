-- time.gf: DateTime and Duration with std.time

import std.time

-- ── Constructors ──
let now = time.now()
println("Now:", now)

let d = time.date(2024, 7, 4)
println("Date:", d)

let dt = time.datetime(2024, 12, 25, 10, 30, 0)
println("DateTime:", dt)

let epoch = time.fromUnix(0)
println("Unix epoch:", epoch)

let epoch2000 = time.fromUnix(946684800)
println("Y2K:", epoch2000)

-- ── Parsing ──
let parsed = time.parse("2024-06-15", "YYYY-MM-DD")
println("Parsed:", parsed)

let parsedRFC = time.parse("2024-06-15T14:30:00Z", "RFC3339")
println("Parsed RFC3339:", parsedRFC)

-- ── DateTime Properties ──
println("Year:", dt.year)
println("Month:", dt.month)
println("Day:", dt.day)
println("Hour:", dt.hour)
println("Minute:", dt.minute)
println("Second:", dt.second)
println("Day of week:", dt.dayOfWeek)
println("Day of year:", dt.dayOfYear)
println("Unix timestamp:", dt.toUnix)

-- ── Formatting ──
println("ISO 8601:", dt.toISO8601())
println("Custom format:", dt.format("2006/01/02 15:04"))

-- ── Duration Constructors ──
let s = time.seconds(90)
println("90 seconds:", s)

let m = time.minutes(5)
println("5 minutes:", m)

let h = time.hours(2)
println("2 hours:", h)

let day = time.days(1)
println("1 day:", day)

let ms = time.millis(500)
println("500 millis:", ms)

-- ── Duration Conversions ──
println("90s in minutes:", s.toMinutes)
println("5min in seconds:", m.toSeconds)
println("2h in minutes:", h.toMinutes)
println("1 day in hours:", day.toHours)
println("500ms in millis:", ms.toMillis)

-- ── Time Calculations ──
let past = time.date(2024, 1, 1)
let elapsed = time.since(past)
println("Since 2024-01-01:", elapsed)

let future = time.datetime(2030, 1, 1, 0, 0, 0)
let remaining = time.until(future)
println("Until 2030-01-01:", remaining)

println("Time example complete")
