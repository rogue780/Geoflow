-- control_flow.gf: If/else, for loops, while loops, break/continue

-- If expressions return values
let x = 10
let parity = if x % 2 == 0 then "even" else "odd"
println(x, "is", parity)

-- For loops
println("Counting:")
for i in range(1, 6) {
  println("  ", i)
}

-- For with list
let fruits = ["apple", "banana", "cherry"]
for fruit in fruits {
  println("Fruit:", fruit)
}

-- While loop
let mut n = 1
while n <= 100 {
  n := n * 2
}
println("First power of 2 > 100:", n)

-- Break
let mut found = -1
for i in range(100) {
  if i * i > 50 then {
    found := i
    break
  }
}
println("First n where n^2 > 50:", found)

-- String iteration
for ch in "GeoFlow" {
  print(ch, " ")
}
println("")

-- Nested loops with enumerate
let items = ["a", "b", "c"]
for (i, item) in items.enumerate()  {
  println("Item", i, ":", item)
}
