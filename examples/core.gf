-- core.gf: Functional utilities from std.core

import std.core

-- ── core.typeof ──
-- Returns the type of a value as a string
println("typeof(42):", core.typeof(42))
println("typeof(3.14):", core.typeof(3.14))
println("typeof(true):", core.typeof(true))
println("typeof(\"hi\"):", core.typeof("hello"))
println("typeof([1,2]):", core.typeof([1, 2]))
println("typeof(nil):", core.typeof(nil))

-- ── core.instanceof ──
-- Checks if a value is of a given type
println("instanceof(42, \"int\"):", core.instanceof(42, "int"))
println("instanceof(42, \"float\"):", core.instanceof(42, "float"))
println("instanceof(\"hi\", \"string\"):", core.instanceof("hi", "string"))

-- ── core.identity ──
-- Returns the argument unchanged
println("identity(99):", core.identity(99))
println("identity(\"hello\"):", core.identity("hello"))

-- identity is useful as a default transformer
let items = [1, 2, 3]
let same = items |> map(core.identity)
println("map(identity):", same)

-- ── core.compose ──
-- compose(f, g) returns h where h(x) = f(g(x))
fn double(x) { x * 2 }
fn negate(x) { -x }

let negDouble = core.compose(negate, double)
println("compose(negate, double)(5):", negDouble(5))

let doubleNeg = core.compose(double, negate)
println("compose(double, negate)(5):", doubleNeg(5))

-- ── core.pipe ──
-- pipe(f, g) returns h where h(x) = g(f(x))
let doubleThenNeg = core.pipe(double, negate)
println("pipe(double, negate)(5):", doubleThenNeg(5))

-- ── core.curry ──
-- curry(f) returns a curried version of a builtin function
println("curry example with compose:")
let pipeline = core.pipe(double, negate)
println("pipe(double, negate)(7):", pipeline(7))
