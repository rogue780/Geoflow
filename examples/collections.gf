-- collections.gf: Lists, maps, tuples, and sets

-- ── Lists ──
let numbers = [1, 2, 3, 4, 5]
println("List:", numbers)
println("Length:", len(numbers))
println("First:", numbers[0])
println("Last:", numbers[-1])

-- List methods
println("Reversed:", reverse(numbers))
println("Sum:", sum(numbers))
println("Product:", product(numbers))

-- Range
let r = range(1, 11)
println("Range 1-10:", r)

-- ── Maps ──
let person = {"name": "Alice", "age": "30", "city": "Portland"}
println("Person:", person)
println("Name:", person["name"])

-- ── Tuples ──
let coord = (3.14, 2.71)
println("Tuple:", coord)
println("X:", coord[0])
println("Y:", coord[1])

-- ── Nested Collections ──
let matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
println("Matrix:", matrix)
println("Center:", matrix[1][1])

-- List comprehension via map/filter
let squares = range(10) |> map(\x -> x ^ 2)
println("Squares:", squares)

let even_squares = range(10)
  |> filter(\x -> x % 2 == 0)
  |> map(\x -> x ^ 2)
println("Even squares:", even_squares)

-- ── Sets (std.collections) ──
import std.collections

-- Create sets
let empty = collections.empty()
println("Empty set:", empty)
println("Empty set size:", empty.size())

let fruits = collections.of("apple", "banana", "cherry")
println("Fruits:", fruits)
println("Fruits size:", fruits.size())

-- Create from list
let nums = collections.fromList([1, 2, 3, 2, 1])
println("Set from [1,2,3,2,1]:", nums)
println("Set size:", nums.size())

-- Set operations
let a = collections.of(1, 2, 3, 4)
let b = collections.of(3, 4, 5, 6)

println("a:", a)
println("b:", b)
println("a.contains(3):", a.contains(3))
println("a.contains(5):", a.contains(5))
println("Union:", a.union(b))
println("Intersection:", a.intersection(b))
println("Difference (a - b):", a.difference(b))

-- Add element
let c = a.add(10)
println("a.add(10):", c)

-- Convert back to list
println("a.toList:", a.toList())
