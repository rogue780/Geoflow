-- error_handling.gf: Option, Result, try-catch, and error propagation

-- ── Option Type ──
let some = Some(42)
let none = head([])

println("Some(42):", some)
println("None:", none)
println("Some(42).isSome:", some.isSome())
println("Some(42).isNone:", some.isNone())
println("None.isSome:", none.isSome())
println("None.isNone:", none.isNone())
println("Some(42).unwrap:", some.unwrap())
println("None.unwrapOr(0):", none.unwrapOr(0))

-- Pattern matching on Option
fn describeOpt(opt) {
  match opt {
    Some(x) => "Got value: " + toString(x),
    None => "Nothing here"
  }
}
println(describeOpt(some))
println(describeOpt(none))

-- ── Result Type ──
let ok = Ok(100)
let err = Err("file not found")

println("Ok(100):", ok)
println("Err(...):", err)
println("Ok.isOk:", ok.isOk())
println("Ok.isErr:", ok.isErr())
println("Err.isOk:", err.isOk())
println("Err.isErr:", err.isErr())
println("Ok.unwrap:", ok.unwrap())

-- ── Try-Catch ──
-- try-catch captures runtime errors and allows recovery
let safe = try {
  let x = 1 / 0
  x
} catch e {
  "Caught error: " + e
}
println("Try-catch result:", safe)

-- Successful try returns the value
let good = try {
  42
} catch e {
  0
}
println("Successful try:", good)

-- ── Error Propagation with ? ──
-- The ? operator unwraps Ok/Some or short-circuits on Err/None
fn safeDivide(a, b) {
  if b == 0 {
    Err("division by zero")
  } else {
    Ok(a / b)
  }
}

fn calculate() {
  let x = safeDivide(100, 4)?
  let y = safeDivide(x, 5)?
  Ok(y)
}

println("calculate():", calculate())

-- ── Combining Patterns ──
fn safeHead(list) {
  let h = head(list)
  match h {
    Some(x) => Ok(x),
    None => Err("empty list")
  }
}

println("safeHead([1,2,3]):", safeHead([1, 2, 3]))
println("safeHead([]):", safeHead([]))
