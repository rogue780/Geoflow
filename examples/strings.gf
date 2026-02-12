-- strings.gf: String methods and std.string module

import std.string

-- ── Module Functions ──
println("string.from(42):", string.from(42))
println("string.from(3.14):", string.from(3.14))
println("string.repeat(\"ab\", 3):", string.repeat("ab", 3))
println("string.join([\"a\",\"b\",\"c\"], \"-\"):", string.join(["a", "b", "c"], "-"))

-- ── Instance Methods ──
let s = "  Hello, GeoFlow World!  "
println("Original:", s)

-- Case conversion
println("uppercase:", s.uppercase())
println("lowercase:", s.lowercase())
println("capitalize:", "hello world".capitalize())

-- Trimming
println("trim:", s.trim())
println("trimStart:", s.trimStart())
println("trimEnd:", s.trimEnd())

-- Search
let text = "hello world hello"
println("contains(\"world\"):", text.contains("world"))
println("startsWith(\"hello\"):", text.startsWith("hello"))
println("endsWith(\"hello\"):", text.endsWith("hello"))
println("indexOf(\"world\"):", text.indexOf("world"))
println("lastIndexOf(\"hello\"):", text.lastIndexOf("hello"))
println("count(\"hello\"):", text.count("hello"))

-- Character access
println("charAt(0):", text.charAt(0))
println("length:", text.length())

-- Splitting
println("split(\" \"):", text.split(" "))
println("lines:", "one\ntwo\nthree".lines())
println("words:", "one two three".words())

-- Extraction
println("substring(6, 11):", text.substring(6, 11))

-- Replacement
println("replace(\"hello\", \"hi\"):", text.replace("hello", "hi"))
println("replaceAll(\"hello\", \"hi\"):", text.replaceAll("hello", "hi"))

-- Transform
println("reverse:", "abcde".reverse())
println("repeat(3):", "ha".repeat(3))

-- Padding
println("padStart(10, \".\"):", "hi".padStart(10, "."))
println("padEnd(10, \".\"):", "hi".padEnd(10, "."))

-- Predicates
println("isEmpty(\"\"):", "".isEmpty())
println("isEmpty(\"a\"):", "a".isEmpty())
println("isBlank(\"  \"):", "  ".isBlank())
println("isBlank(\"a\"):", "a".isBlank())

-- Parsing
println("\"42\".toInt:", "42".toInt())
println("\"3.14\".toFloat:", "3.14".toFloat())

-- ── Escape Sequences ──
println("\n--- Escape Sequences ---")
println("Tab:\tEnd")
println("Newline:\nEnd")
println("Carriage return:\rEnd")
println("Backslash: \\")
println("Quote: \"hello\"")

-- ── Raw Strings (backtick) ──
println("\n--- Raw Strings ---")
let raw = `This is a raw string: \n \t \\ are literal`
println("Raw:", raw)
let path = `C:\Users\data\new_file.txt`
println("Path:", path)
