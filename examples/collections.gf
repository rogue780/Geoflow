-- collections.gf: Lists, maps, tuples

-- Lists
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

-- Maps
let person = {"name": "Alice", "age": "30", "city": "Portland"}
println("Person:", person)
println("Name:", person["name"])

-- Tuples
let point = (3.14, 2.71)
println("Point:", point)
println("X:", point[0])
println("Y:", point[1])

-- Nested collections
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
