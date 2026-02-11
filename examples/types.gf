-- types.gf: Struct and enum types

-- ── Struct Definition ──
type Point2D = struct {
  x: auto,
  y: auto
}

let p = Point2D(3.0, 4.0)
println("Point2D:", p)
println("  x:", p.x, " y:", p.y)

type Person = struct {
  name: auto,
  age: auto,
  city: auto
}

let alice = Person("Alice", 30, "Portland")
println("Person:", alice)
println("  Name:", alice.name)
println("  Age:", alice.age)
println("  City:", alice.city)

-- ── Enum Definition ──
type Color = enum {
  Red,
  Green,
  Blue
}

let c = Red
println("Color:", c)

-- Enum with payloads
type Shape = enum {
  Circle(auto),
  Rectangle(auto, auto),
  Triangle(auto, auto)
}

let s1 = Circle(5.0)
let s2 = Rectangle(3.0, 4.0)
let s3 = Triangle(6.0, 2.0)
println("Shape 1:", s1)
println("Shape 2:", s2)
println("Shape 3:", s3)

-- ── Pattern Matching on Enums ──
fn describeShape(shape) {
  match shape {
    Circle(r) => "Circle with radius " + toString(r),
    Rectangle(w, h) => "Rectangle " + toString(w) + "x" + toString(h),
    Triangle(b, h) => "Triangle base=" + toString(b) + " height=" + toString(h),
    _ => "Unknown shape"
  }
}

println(describeShape(s1))
println(describeShape(s2))
println(describeShape(s3))

-- ── Structs in Collections ──
let people = [
  Person("Alice", 30, "Portland"),
  Person("Bob", 25, "Seattle"),
  Person("Charlie", 35, "Denver")
]

println("People:")
for p in people {
  println("  ", p.name, "age", p.age, "from", p.city)
}

-- Filter structs
let over30 = people |> filter(\p -> p.age >= 30)
println("Age >= 30:", len(over30), "people")
