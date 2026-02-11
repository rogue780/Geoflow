-- hello.gf: Hello World in GeoFlow

println("Hello, GeoFlow!")

-- Variable bindings
let name = "World"
println("Hello, {name}!")

-- Immutable by default
let x = 42
println("The answer is", x)

-- Mutable with 'let mut'
let mut counter = 0
counter := counter + 1
println("Counter:", counter)
