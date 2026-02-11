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
