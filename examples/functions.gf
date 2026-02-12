-- functions.gf: Function definitions and closures

-- Named function
fn greet(name) {
  "Hello, " + name + "!"
}
println(greet("GeoFlow"))

-- Recursive factorial
fn factorial(n) {
  if n <= 1 then 1 else n * factorial(n - 1)
}
println("10! =", factorial(10))

-- Higher-order functions
fn apply(f, x) {
  f(x)
}
let doubled = apply(\x -> x * 2, 21)
println("Doubled:", doubled)

-- Closures
fn makeCounter() {
  let mut count = 0
  fn increment() {
    count := count + 1
    count
  }
  increment
}

-- Fibonacci
fn fib(n) {
  if n <= 1 then n
  else fib(n - 1) + fib(n - 2)
}
println("fib(10) =", fib(10))

-- Function with default parameters
fn power(base, exp = 2) {
  base ^ exp
}
println("3^2 =", power(3))
println("2^10 =", power(2, 10))

-- Multiple defaults
fn range_sum(start, end, step = 1) {
  let mut total = 0
  let mut i = start
  while i <= end {
    total := total + i
    i := i + step
  }
  total
}
println("sum 1..10 by 1:", range_sum(1, 10))
println("sum 1..10 by 2:", range_sum(1, 10, 2))

-- Partial application with defaults
fn make_greeter(greeting = "Hello") {
  \name -> greeting + ", " + name + "!"
}
let hello = make_greeter()
let hi = make_greeter("Hi")
println(hello("World"))
println(hi("World"))

-- Variadic functions
fn sum(...numbers) {
  numbers.reduce(0, \acc, n -> acc + n)
}
println("sum(1,2,3):", sum(1, 2, 3))
println("sum():", sum())

-- Variadic with required params
fn addTo(base, ...numbers) {
  base + numbers.reduce(0, \acc, n -> acc + n)
}
println("addTo(100, 1, 2, 3):", addTo(100, 1, 2, 3))
println("addTo(100):", addTo(100))
