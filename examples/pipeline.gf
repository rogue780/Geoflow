-- pipeline.gf: Pipeline and composition examples

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
