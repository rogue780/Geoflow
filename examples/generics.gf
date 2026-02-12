-- generics.gf: Generic type parameter syntax
--
-- GeoFlow is dynamically typed at runtime. Generic type parameters are
-- parsed and validated but erased before evaluation. They serve as
-- documentation and intent, not as runtime constraints.

-- 1. Generic identity function
fn identity<T>(x: T) -> T { x }

println("identity(42):", identity(42))
println("identity(hello):", identity("hello"))

-- 2. Multiple type parameters
fn first<A, B>(a: A, b: B) -> A { a }
fn second<A, B>(a: A, b: B) -> B { b }

println("first(1, two):", first(1, "two"))
println("second(1, two):", second(1, "two"))

-- 3. Generic higher-order function
fn apply<T, U>(f: Fn<(T) -> U>, x: T) -> U {
  f(x)
}

println("apply(double, 21):", apply(\x -> x * 2, 21))
println("apply(negate, 5):", apply(\x -> -x, 5))

-- 4. Type annotations with parameterized types
fn length<T>(list: List<T>) -> int {
  len(list)
}

println("length([1,2,3]):", length([1, 2, 3]))
println("length([a,b]):", length(["a", "b"]))

-- 5. Swap using generic pairs
fn swap<A, B>(a: A, b: B) -> List<B> {
  [b, a]
}

println("swap(1, hello):", swap(1, "hello"))

-- 6. Generic compose
fn compose<A, B, C>(f: Fn<(B) -> C>, g: Fn<(A) -> B>) -> Fn<(A) -> C> {
  \x -> f(g(x))
}

let double_then_add1 = compose(\x -> x + 1, \x -> x * 2)
println("compose(+1, *2)(10):", double_then_add1(10))

-- 7. Works the same without annotations (dynamic typing)
fn plain_identity(x) { x }
println("plain_identity(99):", plain_identity(99))

-- The generic versions above are functionally identical to their
-- untyped counterparts. Use generics when you want to communicate
-- the expected types to readers of your code.
