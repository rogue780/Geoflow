-- assertions.gf: Testing assertions with std.assert

import std.assert as check

-- ── Basic Assertions ──
check.assert(true)
check.assert(1 + 1 == 2, "basic math")

-- ── Equality ──
check.equal(42, 42)
check.notEqual(1, 2)

-- ── Numeric Comparisons ──
check.approximately(3.14159, 3.14, 0.01)
check.greaterThan(10, 5)
check.lessThan(5, 10)
check.greaterOrEqual(10, 10)
check.lessOrEqual(5, 10)
check.between(7, 1, 10)

-- ── Nil Checks ──
check.isNil(nil)
check.notNil(42)

-- ── Result Assertions ──
check.ok(Ok(42))
check.err(Err("oops"))

-- ── Collection Assertions ──
check.empty([])
check.notEmpty([1, 2, 3])
check.length([1, 2, 3], 3)
check.contains([1, 2, 3], 2)

-- ── String Assertions ──
check.contains("hello world", "world")
check.startsWith("hello world", "hello")
check.endsWith("hello world", "world")
check.matches("hello123", "^hello\\d+$")

-- ── Throws ──
-- assert.throws checks if its argument is an Error
fn failingOp() {
  let x = 1 / 0
  x
}
let result = try { failingOp() } catch e { Err(e) }
check.err(result)

println("All assertions passed!")
