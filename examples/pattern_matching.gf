-- pattern_matching.gf: Match expressions and Option/Result types

-- Basic pattern matching
fn describe(x) {
  match x {
    0 => "zero",
    1 => "one",
    n if n > 0 => "positive",
    _ => "negative"
  }
}

println(describe(0))
println(describe(1))
println(describe(42))
println(describe(-5))

-- Option type
let values = [10, 20, 30]
let first = head(values)
println("First element:", first)

let empty = []
let none_val = head(empty)
println("Head of empty:", none_val)

-- Result type
let ok_result = Ok(42)
let err_result = Err("something went wrong")

println("Ok:", ok_result)
println("Err:", err_result)
println("Is ok?", ok_result.isOk())
println("Is err?", err_result.isErr())

-- Unwrap with default
let opt = Some(100)
println("Unwrapped:", opt.unwrap())
