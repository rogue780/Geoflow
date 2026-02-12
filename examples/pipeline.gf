-- pipeline.gf: Pipeline and composition examples

import std.core

-- ── Forward Pipeline |> ──
-- The pipeline operator |> threads values through function calls
let result = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
  |> filter(\x -> x % 2 == 0)
  |> map(\x -> x ^ 2)
  |> sum()

println("Sum of squares of evens:", result)

-- Method chaining on lists
let items = [3, 1, 4, 1, 5, 9, 2, 6]
let sorted = sort(items)
println("Sorted:", sorted)

-- String methods and list operations
let words = "hello world geoflow"
let parts = words.split(" ")
let upper = parts.map(\w -> w.uppercase())
println("Uppercased words:", upper)

-- Chained list operations
let data = [1, 2, 3, 4, 5]
let doubled = data.map(\x -> x * 2)
println("Doubled:", doubled)

let evens = data.filter(\x -> x % 2 == 0)
println("Evens:", evens)

let total = data.reduce(0, \acc, x -> acc + x)
println("Sum:", total)

-- ── Reverse Pipeline <| ──
-- The reverse pipeline operator <| applies a function to a value: f <| x means f(x)
let neg = abs <| -42
println("abs <| -42:", neg)

fn double(x) { x * 2 }
fn addOne(x) { x + 1 }
println("double <| 5:", double <| 5)
println("addOne <| (double <| 3):", addOne <| (double <| 3))

-- ── Function Composition via |> ──
-- When both sides of |> are functions, it composes them instead of applying
let transform = (\x -> x + 5) |> (\x -> x * x)
println("transform(3):", transform(3))   -- 64: (3+5)^2

-- Chaining multiple compositions
let process = (\x -> x + 1) |> (\x -> x * 2) |> (\x -> x * x)
println("process(4):", process(4))   -- 100: ((4+1)*2)^2

-- Works with named functions too
let doubleAndAdd = double |> addOne
println("doubleAndAdd(5):", doubleAndAdd(5))  -- 11: 5*2+1

-- Reverse pipe composition: f <| g composes as f(g(x))
let revCompose = (\x -> x * x) <| (\x -> x + 5)
println("revCompose(3):", revCompose(3))  -- 64: (3+5)^2

-- ── Functional Composition with std.core ──
-- core.compose(f, g) returns a function h where h(x) = f(g(x))
let doubleAndAdd = core.compose(addOne, double)
println("compose(addOne, double)(5):", doubleAndAdd(5))

-- core.pipe(f, g) returns a function h where h(x) = g(f(x))
let doubleThenAdd = core.pipe(double, addOne)
println("pipe(double, addOne)(5):", doubleThenAdd(5))

-- core.identity returns its argument unchanged
println("identity(42):", core.identity(42))

-- ── Dot Composition ──
-- f . g creates a composed function where f(g(x)) (right-to-left, mathematical)
let doubleAndInc = addOne . double
println("(addOne . double)(5):", doubleAndInc(5))  -- 11

let chain = (\x -> x * 3) . (\x -> x + 1) . (\x -> x * 2)
println("chain(2):", chain(2))  -- 3 * ((2*2) + 1) = 15

-- ── Juxtaposition Composition ──
-- Space-separated functions compose left-to-right
let pipeline = double addOne
println("(double addOne)(5):", pipeline(5))  -- addOne(double(5)) = 11

-- Inline application: value followed by functions
println("5 double addOne:", 5 double addOne)  -- 11

-- With partial application
fn add(a, b) { a + b }
let inc = add(1)
let process = inc double
println("(inc double)(3):", process(3))  -- double(inc(3)) = 8
