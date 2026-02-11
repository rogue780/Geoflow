# GeoFlow Language Specification & Implementation Plan

**Version:** 0.1.0-draft  
**Target Runtime:** Go 1.22+  
**Purpose:** Composition-oriented functional language for geospatial data transformation and ETL pipelines

---

## Table of Contents

1. [Philosophy & Design Principles](#1-philosophy--design-principles)
2. [Lexical Structure](#2-lexical-structure)
3. [Type System](#3-type-system)
4. [Syntax Reference](#4-syntax-reference)
5. [Composition Model](#5-composition-model)
6. [Standard Library](#6-standard-library)
7. [Math Library](#7-math-library)
8. [Geospatial Library](#8-geospatial-library)
9. [Implementation Plan](#9-implementation-plan)
10. [File Structure](#10-file-structure)

---

## 1. Philosophy & Design Principles

### 1.1 Core Belief

Every function can be decomposed into a chain of more primitive functions. GeoFlow embraces this by making composition the fundamental operation of the language.

### 1.2 Design Principles

1. **Composition Over Application**: `f g h` means "apply h, then g, then f" (left-to-right data flow, inside-out evaluation)
2. **Static Types, Dynamic Escape Hatch**: Strong static typing with `auto` for inference when needed
3. **Symbolic-Numeric Bridge**: Symbolic math expressions that can be "realized" to concrete values
4. **Geospatial Native**: First-class support for OGC geometries, geography vs geometry distinction, coordinate reference systems
5. **Pipeline Oriented**: Optimized for ETL and streaming data transformation workflows

### 1.3 Influences

- **Haskell**: Type system, composition operator semantics
- **F#**: Pipeline operator, pragmatic functional approach  
- **PostGIS**: Geography/Geometry distinction, spatial operations
- **APL/J**: Tacit programming, array-oriented computation
- **Forth/Joy**: Stack-based composition inspiration

### 1.4 Memory Model

GeoFlow uses garbage collection with immutable-by-default semantics. This section documents the design rationale and tradeoffs.

#### 1.4.1 Design Philosophy

GeoFlow prioritizes developer productivity and correctness over maximum performance. The memory model reflects this:

1. **No explicit pointers or references** in user code
2. **Garbage collected** runtime (leveraging Go's GC)
3. **Immutable by default** data structures
4. **Value semantics** at the language level
5. **Structural sharing** for efficiency under the hood

#### 1.4.2 Value Semantics

All values in GeoFlow behave as if they are passed by value. Assigning a variable or passing an argument conceptually creates a copy:

```geoflow
-- Primitives: true copies
let a = 42
let b = a           -- b is independent of a
-- modifying b (if it were mutable) would not affect a

-- Collections: immutable, so "copy" is safe and cheap
let points = [p1, p2, p3]
let alias = points  -- alias and points refer to same immutable data
-- this is safe because neither can be mutated

-- Function arguments: value semantics
fn process(data: List<Point>) -> List<Point> {
    -- data behaves as a copy
    -- cannot modify caller's data
    data.map(\p -> p.buffer(10))
}

let original = [p1, p2, p3]
let result = process(original)
-- original is unchanged, guaranteed
```

#### 1.4.3 Immutability by Default

All values are immutable unless explicitly using mutable constructs:

```geoflow
-- Immutable binding (default)
let x = [1, 2, 3]
-- x := [4, 5, 6]     -- ERROR: cannot reassign immutable binding
-- x.push(4)          -- ERROR: no such method, List is immutable

-- "Modifying" returns a new value
let y = x.append(4)   -- y is [1,2,3,4], x is still [1,2,3]

-- Mutable binding (allows reassignment)
let mut counter = 0
counter := counter + 1  -- OK: reassigns the binding

-- But the VALUES are still immutable
let mut items = [1, 2, 3]
let snapshot = items
items := items.append(4)  
-- items is now [1,2,3,4]
-- snapshot is still [1,2,3] (no spooky action at a distance)
```

#### 1.4.4 Internal Representation

While the language presents value semantics, the runtime uses efficient representations:

**Primitives (int, float, bool, byte):**
- Stored directly (unboxed) where possible
- Truly copied on assignment

**Small Strings:**
- Interned or copy-on-write

**Collections (List, Map, Set):**
- Implemented as persistent data structures
- Structural sharing: "modified" versions share most memory with original
- Example: appending to a 1-million element list doesn't copy 1 million elements

**Large Objects (DataFrame, Raster, GeometryCollection):**
- Reference counted or GC-managed
- Copy-on-write semantics
- Immutable facade over efficient internal representation

**Geometry Objects:**
- Immutable value objects
- Coordinate arrays shared when derived geometries are created

#### 1.4.5 Mutable Escape Hatches

For performance-critical code, GeoFlow provides explicit mutable containers:

```geoflow
import std.collections.mutable as mut

-- Mutable list for accumulation
let buffer = mut.List<Point>()
for point in stream {
    buffer.push(point)      -- actual in-place mutation
    if buffer.length() >= 1000 {
        flush(buffer.freeze())  -- convert to immutable
        buffer.clear()
    }
}

-- Mutable map for aggregation
let counts = mut.Map<string, int>()
for event in events {
    let key = event.category
    counts.update(key, 0, \n -> n + 1)
}
let result = counts.freeze()
```

**Rules for mutable containers:**
- Cannot be stored in immutable data structures
- Cannot escape the function that created them (enforced at runtime)
- Must be explicitly converted to immutable via `.freeze()` to return/store

#### 1.4.6 Comparison with Other Approaches

| Aspect | Rust | GeoFlow | Java/Go |
|--------|------|---------|---------|
| Memory management | Ownership + RAII | GC | GC |
| Default mutability | Mutable | Immutable | Mutable |
| Aliasing control | Borrow checker | Immutability | Programmer discipline |
| Passing semantics | Move + borrow | Value (structural sharing) | Reference |
| Null safety | Option type | Option type | Nullable refs |
| Learning curve | Steep | Gentle | Moderate |
| Runtime overhead | Minimal | GC + sharing | GC |
| Compile-time cost | Borrow checker | Type checking | Type checking |

**Why not Rust-style ownership?**

Rust's ownership model provides zero-cost memory safety, but introduces complexity:
- Lifetime annotations on function signatures
- Fighting the borrow checker for valid patterns
- Learning curve that distracts from the domain (geospatial ETL)

For GeoFlow's target use case (data transformation pipelines), the tradeoffs favor simplicity:
- ETL workloads are batch or streaming, not real-time latency-critical
- Developer productivity matters more than microsecond latencies
- Correctness through immutability is easier to reason about

**Why not pure reference semantics like Java?**

Reference semantics with mutability leads to:
- Defensive copying ("should I clone this?")
- Aliasing bugs (unexpected mutation through shared reference)
- Difficulty reasoning about data flow in pipelines

Immutable value semantics eliminates these classes of bugs.

#### 1.4.7 Performance Considerations

**GC Pauses:**
- Go's GC is low-latency (sub-millisecond typical)
- For most ETL workloads, imperceptible
- Streaming pipelines may need tuning for sustained throughput

**Structural Sharing Overhead:**
- Small overhead per operation (tree traversal)
- Pays off when "copying" large structures
- Net win for typical functional transformation patterns

**When to Use Mutable Containers:**
- Building large collections element-by-element
- Aggregating into maps with many updates
- Hot loops in streaming pipelines

**Memory Arenas (Future):**
- For batch processing of large datasets
- Allocate all working memory from arena
- Free entire arena at end of batch
- Not in v1, but architected to allow addition

#### 1.4.8 Guarantees

GeoFlow provides these memory safety guarantees:

1. **No null pointer dereferences**: Option types for nullable values
2. **No use-after-free**: GC prevents this
3. **No data races from aliasing**: Immutability prevents this
4. **No iterator invalidation**: Immutable collections
5. **Deterministic behavior**: Same inputs always produce same outputs

What GeoFlow does NOT guarantee:

1. **Bounded memory usage**: GC may retain memory longer than ideal
2. **Deterministic latency**: GC pauses are unpredictable
3. **Minimal allocations**: Functional style allocates freely

---

## 2. Lexical Structure

### 2.1 Character Set

GeoFlow source files are UTF-8 encoded.

### 2.2 Comments

```geoflow
-- Single line comment

{- 
   Multi-line
   comment 
-}
```

### 2.3 Identifiers

```
identifier     = letter { letter | digit | '_' }
letter         = 'a'..'z' | 'A'..'Z' | '_'
digit          = '0'..'9'
```

### 2.4 Keywords

```
let       const     fn        type      struct    enum
if        then      else      match     with      
for       in        while     do        
return    yield     break     continue
true      false     nil       
auto      import    export    module
geom      geog      
```

### 2.5 Literals

```geoflow
-- Integers
42
-17
0xFF        -- hexadecimal
0b1010      -- binary
0o777       -- octal

-- Floats
3.14159
-2.5e10
1.0E-5

-- Strings
"hello world"
"line1\nline2"      -- escape sequences
`raw string`        -- no escape processing

-- WKT Literals (special syntax)
#POINT(0 0)#
#POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))#

-- Symbolic expressions
$x^2 + 3*x + 2$
$sin(x) * cos(y)$
```

### 2.6 Operators

```
-- Arithmetic
+   -   *   /   %   ^   

-- Comparison
==  !=  <   >   <=  >=

-- Logical
&&  ||  !

-- Composition (implicit via juxtaposition, or explicit)
|>      -- pipeline: x |> f  means f(x)
<|      -- reverse pipeline: f <| x means f(x)
.       -- function composition: f . g means fn(x) => f(g(x))

-- Assignment
=       -- binding
:=      -- mutable assignment

-- Type annotation
:       -- type annotation

-- Access
.       -- field/method access
[]      -- index access
```

---

## 3. Type System

### 3.1 Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `int` | 64-bit signed integer | `42` |
| `float` | 64-bit IEEE 754 | `3.14` |
| `bool` | Boolean | `true`, `false` |
| `string` | UTF-8 string | `"hello"` |
| `byte` | 8-bit unsigned | `0xFF` |
| `nil` | Absence of value | `nil` |
| `auto` | Inferred at assignment | `let x: auto = 42` |

### 3.2 Collection Types

```geoflow
-- List (ordered, homogeneous)
List<T>
[1, 2, 3, 4, 5]

-- Array (fixed-size, homogeneous)
Array<T, N>
Array<int, 5>

-- Map (key-value)
Map<K, V>
{"name": "Alice", "age": 30}

-- Set (unique values)
Set<T>
{1, 2, 3}

-- Tuple (fixed-size, heterogeneous)
Tuple<T1, T2, ...>
(1, "hello", true)

-- Option (nullable)
Option<T>
Some(42) | None

-- Result (error handling)
Result<T, E>
Ok(value) | Err(error)

-- Stream (lazy, potentially infinite)
Stream<T>
```

### 3.3 Mathematical Types

```geoflow
-- Symbolic expression (unevaluated math)
Expr

-- Vector (1D array with math operations)
Vector<T, N>
Vector<float, 3>

-- Matrix (2D array with linear algebra)
Matrix<T, M, N>
Matrix<float, 4, 4>

-- Complex numbers
Complex

-- Polynomial
Poly<T>

-- Series (statistical data)
Series<T>

-- DataFrame (tabular data)
DataFrame
```

### 3.4 Geospatial Types

GeoFlow distinguishes between **geometry** (planar, Cartesian) and **geography** (spheroidal, geodetic), mirroring PostGIS.

#### 3.4.1 Geometry Types (Planar)

```geoflow
-- 2D Geometry (default)
geom.Point
geom.LineString
geom.Polygon
geom.MultiPoint
geom.MultiLineString
geom.MultiPolygon
geom.GeometryCollection

-- 3D Geometry (Z coordinate)
geom.PointZ
geom.LineStringZ
geom.PolygonZ
-- ... etc

-- 2D + Measure
geom.PointM
geom.LineStringM

-- 3D + Measure  
geom.PointZM
geom.LineStringZM
```

#### 3.4.2 Geography Types (Spheroidal)

```geoflow
-- Geography uses geodetic calculations (great circles, etc.)
-- Default datum: WGS84 (EPSG:4326)

geog.Point
geog.LineString
geog.Polygon
geog.MultiPoint
geog.MultiLineString
geog.MultiPolygon
geog.GeometryCollection

-- 3D variants
geog.PointZ        -- includes altitude
geog.LineStringZ
-- ... etc
```

#### 3.4.3 Coordinate Reference Systems

```geoflow
-- CRS type
CRS

-- Built-in CRS constants
CRS.WGS84          -- EPSG:4326 (default for geog)
CRS.WebMercator    -- EPSG:3857
CRS.NAD83          -- EPSG:4269
CRS.NAD27          -- EPSG:4267
CRS.ETRS89         -- EPSG:4258

-- Custom CRS from EPSG code
CRS.fromEPSG(32610)    -- UTM Zone 10N

-- Custom CRS from Proj4 string
CRS.fromProj4("+proj=utm +zone=10 +datum=WGS84")

-- Custom CRS from WKT
CRS.fromWKT("GEOGCS[...]")
```

#### 3.4.4 Raster Types

```geoflow
-- Single-band raster
Raster<T>

-- Multi-band raster
RasterStack<T>

-- Raster cell/pixel
Cell<T>
```

### 3.5 Function Types

```geoflow
-- Function type syntax
Fn<(Args) -> Return>

-- Examples
Fn<(int, int) -> int>
Fn<(geom.Point) -> float>
Fn<() -> string>

-- Generic functions
Fn<T>((T) -> T)
Fn<T, U>((T) -> U)
```

### 3.6 Custom Types

```geoflow
-- Struct (product type)
type Person = struct {
    name: string,
    age: int,
    location: geog.Point
}

-- Enum (sum type)
type Status = enum {
    Active,
    Inactive,
    Pending(string)
}

-- Type alias
type Coordinate = Tuple<float, float>
type PointList = List<geom.Point>
```

### 3.7 Type Inference with `auto`

```geoflow
-- Compiler infers type from right-hand side
let x: auto = 42           -- inferred as int
let y: auto = "hello"      -- inferred as string
let p: auto = #POINT(0 0)# -- inferred as geom.Point

-- auto in function parameters (dynamic dispatch)
fn process(data: auto) -> auto {
    -- type determined at runtime
    match typeof(data) {
        int => data * 2,
        string => data.uppercase(),
        _ => data
    }
}
```

---

## 4. Syntax Reference

### 4.1 Variable Binding

```geoflow
-- Immutable binding (default)
let x = 42
let name: string = "GeoFlow"

-- Mutable binding
let mut counter = 0
counter := counter + 1

-- Constant (compile-time)
const PI = 3.14159265359
const MAX_POINTS = 10000
```

### 4.2 Function Definition

```geoflow
-- Named function
fn add(a: int, b: int) -> int {
    a + b
}

-- With type parameters (generics)
fn identity<T>(x: T) -> T {
    x
}

-- Anonymous function (lambda)
let double = fn(x: int) -> int { x * 2 }

-- Shorthand lambda
let double = \x -> x * 2
let add = \x, y -> x + y

-- Function with default parameters
fn greet(name: string, greeting: string = "Hello") -> string {
    "{greeting}, {name}!"
}

-- Variadic function
fn sum(numbers: ...int) -> int {
    numbers.reduce(0, \acc, n -> acc + n)
}
```

### 4.3 Control Flow

```geoflow
-- If expression (always returns value)
let result = if x > 0 then "positive" else "non-positive"

-- If with multiple branches
let category = if x < 0 then
    "negative"
else if x == 0 then
    "zero"
else
    "positive"

-- Match expression (pattern matching)
let description = match shape {
    geom.Point(x, y) => "Point at ({x}, {y})",
    geom.LineString(points) => "Line with {points.length()} points",
    geom.Polygon(rings) => "Polygon with {rings.length()} rings",
    _ => "Unknown shape"
}

-- Match with guards
let grade = match score {
    s if s >= 90 => "A",
    s if s >= 80 => "B",
    s if s >= 70 => "C",
    s if s >= 60 => "D",
    _ => "F"
}

-- For loop
for point in points {
    log.info(point.toString())
}

-- For with index
for (i, point) in points.enumerate() {
    log.info("{i}: {point}")
}

-- While loop
while condition {
    -- body
}

-- Loop with break/continue
for x in items {
    if x < 0 then continue
    if x > 100 then break
    process(x)
}
```

### 4.4 Error Handling

```geoflow
-- Result type for recoverable errors
fn divide(a: float, b: float) -> Result<float, string> {
    if b == 0.0 then
        Err("Division by zero")
    else
        Ok(a / b)
}

-- Using Result
let result = divide(10.0, 2.0)
match result {
    Ok(value) => log.info("Result: {value}"),
    Err(msg) => log.error("Error: {msg}")
}

-- Propagate errors with ?
fn calculate(x: float) -> Result<float, string> {
    let a = divide(x, 2.0)?    -- returns early if Err
    let b = divide(a, 3.0)?
    Ok(b)
}

-- Try-catch for exceptions (rare, for panics)
try {
    riskyOperation()
} catch e {
    log.error("Caught: {e}")
}

-- Assert (panics on failure)
assert(x > 0, "x must be positive")
assert.equal(a, b)
assert.notNil(value)
```

### 4.5 Modules and Imports

```geoflow
-- Module declaration (top of file)
module myproject.processing

-- Import entire module
import std.math
import std.geo

-- Import specific items
import std.math.{sin, cos, tan}
import std.geo.{Point, Polygon}

-- Import with alias
import std.math as m
import std.geo.transform as t

-- Export (items are private by default)
export fn publicFunction() { ... }
export type PublicType = ...
```

---

## 5. Composition Model

### 5.1 Core Concept

In GeoFlow, `f g h` means: apply h first, then g, then f. Data flows left-to-right through the chain.

```geoflow
-- These are equivalent:
let result = f(g(h(x)))
let result = x |> h |> g |> f
let result = x h g f
```

### 5.2 Pipeline Operator

```geoflow
-- Pipeline: passes left side as first argument to right side
x |> f           -- equivalent to f(x)
x |> f(y)        -- equivalent to f(x, y)
x |> f |> g |> h -- equivalent to h(g(f(x)))
```

#### Polymorphic Composition

When **both sides** of `|>` are functions (or callable objects), the pipeline operator
**composes** them into a new function instead of applying:

```geoflow
-- Both sides are functions → returns a ComposedFunction
let transform = (\x -> x + 5) |> (\x -> x * x)
transform(3)   -- 64: first adds 5 (→8), then squares (→64)

-- Chaining composes naturally (ComposedFunction is itself callable)
let process = (\x -> x + 1) |> (\x -> x * 2) |> (\x -> x * x)
process(4)     -- 100: 4→5→10→100

-- Works with named functions
fn double(x) { x * 2 }
fn addOne(x) { x + 1 }
let doubleAndAdd = double |> addOne
doubleAndAdd(5)  -- 11

-- Reverse pipe composes with swapped order: f <| g → f(g(x))
let r = (\x -> x * x) <| (\x -> x + 5)
r(3)   -- 64: same as transform above
```

A value on the left side is still applied as before — composition only triggers
when the left side is also a function, builtin, or composed function.

### 5.3 Composition Operator

```geoflow
-- Compose functions without applying
let fg = f . g           -- fg(x) = f(g(x))
let pipeline = h . g . f -- pipeline(x) = h(g(f(x)))

-- Use composed function
let result = pipeline(input)
```

### 5.4 Juxtaposition Composition

```geoflow
-- Space-separated identifiers form a composition chain
-- Read left-to-right for data flow

-- Example: read file, parse points, buffer, intersect, count
let pipeline = read parsePoints buffer(100) intersect(regions) count

-- Apply to input
let result = "data.wkt" pipeline

-- Or inline
let result = "data.wkt" read parsePoints buffer(100) intersect(regions) count
```

### 5.5 Partial Application

```geoflow
-- Functions are automatically curried
fn add(a: int, b: int) -> int { a + b }

let add5 = add(5)      -- partially applied, waiting for second arg
let result = add5(3)   -- 8

-- Useful in pipelines
let results = points |> map(buffer(100)) |> filter(intersects(region))
```

### 5.6 Function Chaining Methods

```geoflow
-- Methods can be chained with dot notation
let result = points
    .filter(\p -> p.x > 0)
    .map(\p -> p.buffer(10))
    .reduce(geom.union)

-- Equivalent pipeline form
let result = points 
    |> filter(\p -> p.x > 0)
    |> map(\p -> p.buffer(10))
    |> reduce(geom.union)
```

---

## 6. Standard Library

### 6.1 Core (`std.core`)

```geoflow
-- Type inspection
typeof(value) -> string
instanceof<T>(value) -> bool

-- Comparison
min<T: Ord>(a: T, b: T) -> T
max<T: Ord>(a: T, b: T) -> T
clamp<T: Ord>(value: T, low: T, high: T) -> T

-- Identity and constant
identity<T>(x: T) -> T
const<T>(x: T) -> Fn<(auto) -> T>

-- Function manipulation
compose<A, B, C>(f: Fn<(B) -> C>, g: Fn<(A) -> B>) -> Fn<(A) -> C>
pipe<A, B, C>(f: Fn<(A) -> B>, g: Fn<(B) -> C>) -> Fn<(A) -> C>
flip<A, B, C>(f: Fn<(A, B) -> C>) -> Fn<(B, A) -> C>
curry<A, B, C>(f: Fn<(A, B) -> C>) -> Fn<(A) -> Fn<(B) -> C>>
```

### 6.2 Strings (`std.string`)

```geoflow
-- Construction
string.from<T>(value: T) -> string
string.repeat(s: string, n: int) -> string
string.join(parts: List<string>, sep: string) -> string

-- Properties
s.length() -> int
s.isEmpty() -> bool
s.isBlank() -> bool     -- empty or only whitespace

-- Case conversion
s.uppercase() -> string
s.lowercase() -> string
s.capitalize() -> string
s.titleCase() -> string

-- Trimming
s.trim() -> string
s.trimStart() -> string
s.trimEnd() -> string
s.strip(chars: string) -> string

-- Search
s.contains(sub: string) -> bool
s.startsWith(prefix: string) -> bool
s.endsWith(suffix: string) -> bool
s.indexOf(sub: string) -> Option<int>
s.lastIndexOf(sub: string) -> Option<int>
s.count(sub: string) -> int

-- Extraction
s.charAt(i: int) -> Option<string>
s.substring(start: int, end: int) -> string
s.slice(start: int, end: int = -1) -> string
s.split(sep: string) -> List<string>
s.splitAt(i: int) -> Tuple<string, string>
s.lines() -> List<string>
s.words() -> List<string>

-- Modification
s.replace(old: string, new: string) -> string
s.replaceAll(old: string, new: string) -> string
s.replaceFirst(old: string, new: string) -> string
s.insert(i: int, sub: string) -> string
s.remove(start: int, end: int) -> string
s.reverse() -> string
s.padStart(len: int, char: string = " ") -> string
s.padEnd(len: int, char: string = " ") -> string

-- Regex
s.matches(pattern: string) -> bool
s.findAll(pattern: string) -> List<string>
s.replaceRegex(pattern: string, replacement: string) -> string
s.capture(pattern: string) -> Option<List<string>>

-- Parsing
s.parseInt() -> Result<int, string>
s.parseFloat() -> Result<float, string>
s.parseBool() -> Result<bool, string>

-- Formatting
string.format(template: string, args: ...auto) -> string
-- Example: string.format("Point({}, {})", x, y)

-- Interpolation is built-in
let msg = "Coordinates: ({x}, {y})"
```

### 6.3 Collections (`std.collections`)

#### 6.3.1 List Operations

```geoflow
-- Construction
List.empty<T>() -> List<T>
List.of<T>(items: ...T) -> List<T>
List.repeat<T>(item: T, n: int) -> List<T>
List.range(start: int, end: int, step: int = 1) -> List<int>
List.generate<T>(n: int, f: Fn<(int) -> T>) -> List<T>

-- Properties
list.length() -> int
list.isEmpty() -> bool
list.nonEmpty() -> bool

-- Access
list.first() -> Option<T>
list.last() -> Option<T>
list.get(i: int) -> Option<T>
list.head() -> Option<T>
list.tail() -> List<T>
list.init() -> List<T>

-- Search
list.contains(item: T) -> bool
list.indexOf(item: T) -> Option<int>
list.find(pred: Fn<(T) -> bool>) -> Option<T>
list.findIndex(pred: Fn<(T) -> bool>) -> Option<int>
list.any(pred: Fn<(T) -> bool>) -> bool
list.all(pred: Fn<(T) -> bool>) -> bool
list.none(pred: Fn<(T) -> bool>) -> bool

-- Transformation
list.map<U>(f: Fn<(T) -> U>) -> List<U>
list.flatMap<U>(f: Fn<(T) -> List<U>>) -> List<U>
list.filter(pred: Fn<(T) -> bool>) -> List<T>
list.filterMap<U>(f: Fn<(T) -> Option<U>>) -> List<U>
list.take(n: int) -> List<T>
list.drop(n: int) -> List<T>
list.takeWhile(pred: Fn<(T) -> bool>) -> List<T>
list.dropWhile(pred: Fn<(T) -> bool>) -> List<T>
list.slice(start: int, end: int) -> List<T>
list.reverse() -> List<T>
list.sort() -> List<T>
list.sortBy<K: Ord>(f: Fn<(T) -> K>) -> List<T>
list.sortWith(cmp: Fn<(T, T) -> int>) -> List<T>
list.distinct() -> List<T>
list.distinctBy<K>(f: Fn<(T) -> K>) -> List<T>

-- Aggregation
list.reduce<U>(init: U, f: Fn<(U, T) -> U>) -> U
list.fold<U>(init: U, f: Fn<(U, T) -> U>) -> U  -- alias for reduce
list.foldRight<U>(init: U, f: Fn<(T, U) -> U>) -> U
list.scan<U>(init: U, f: Fn<(U, T) -> U>) -> List<U>
list.sum() -> T                -- requires T: Numeric
list.product() -> T            -- requires T: Numeric
list.min() -> Option<T>        -- requires T: Ord
list.max() -> Option<T>        -- requires T: Ord
list.minBy<K: Ord>(f: Fn<(T) -> K>) -> Option<T>
list.maxBy<K: Ord>(f: Fn<(T) -> K>) -> Option<T>

-- Combination
list.concat(other: List<T>) -> List<T>
list.append(item: T) -> List<T>
list.prepend(item: T) -> List<T>
list.insert(i: int, item: T) -> List<T>
list.remove(i: int) -> List<T>
list.zip<U>(other: List<U>) -> List<Tuple<T, U>>
list.zipWith<U, V>(other: List<U>, f: Fn<(T, U) -> V>) -> List<V>
list.unzip<A, B>() -> Tuple<List<A>, List<B>>  -- requires T = Tuple<A, B>
list.flatten<U>() -> List<U>  -- requires T = List<U>
list.intersperse(sep: T) -> List<T>
list.interleave(other: List<T>) -> List<T>

-- Grouping
list.partition(pred: Fn<(T) -> bool>) -> Tuple<List<T>, List<T>>
list.groupBy<K>(f: Fn<(T) -> K>) -> Map<K, List<T>>
list.chunked(size: int) -> List<List<T>>
list.windowed(size: int, step: int = 1) -> List<List<T>>

-- Iteration
list.forEach(f: Fn<(T) -> unit>) -> unit
list.enumerate() -> List<Tuple<int, T>>
```

#### 6.3.2 Map Operations

```geoflow
-- Construction
Map.empty<K, V>() -> Map<K, V>
Map.of<K, V>(pairs: ...Tuple<K, V>) -> Map<K, V>
Map.fromList<K, V>(pairs: List<Tuple<K, V>>) -> Map<K, V>

-- Properties
map.size() -> int
map.isEmpty() -> bool
map.keys() -> List<K>
map.values() -> List<V>
map.entries() -> List<Tuple<K, V>>

-- Access
map.get(key: K) -> Option<V>
map.getOrDefault(key: K, default: V) -> V
map.contains(key: K) -> bool

-- Modification
map.put(key: K, value: V) -> Map<K, V>
map.putAll(other: Map<K, V>) -> Map<K, V>
map.remove(key: K) -> Map<K, V>
map.update(key: K, f: Fn<(V) -> V>) -> Map<K, V>
map.updateOrInsert(key: K, default: V, f: Fn<(V) -> V>) -> Map<K, V>

-- Transformation
map.mapValues<U>(f: Fn<(V) -> U>) -> Map<K, U>
map.mapKeys<J>(f: Fn<(K) -> J>) -> Map<J, V>
map.filter(pred: Fn<(K, V) -> bool>) -> Map<K, V>
map.filterKeys(pred: Fn<(K) -> bool>) -> Map<K, V>
map.filterValues(pred: Fn<(V) -> bool>) -> Map<K, V>

-- Combination
map.merge(other: Map<K, V>, resolve: Fn<(V, V) -> V>) -> Map<K, V>
```

#### 6.3.3 Set Operations

```geoflow
-- Construction
Set.empty<T>() -> Set<T>
Set.of<T>(items: ...T) -> Set<T>
Set.fromList<T>(items: List<T>) -> Set<T>

-- Properties
set.size() -> int
set.isEmpty() -> bool
set.toList() -> List<T>

-- Membership
set.contains(item: T) -> bool
set.isSubset(other: Set<T>) -> bool
set.isSuperset(other: Set<T>) -> bool
set.isDisjoint(other: Set<T>) -> bool

-- Modification
set.add(item: T) -> Set<T>
set.remove(item: T) -> Set<T>
set.toggle(item: T) -> Set<T>

-- Set operations
set.union(other: Set<T>) -> Set<T>
set.intersection(other: Set<T>) -> Set<T>
set.difference(other: Set<T>) -> Set<T>
set.symmetricDifference(other: Set<T>) -> Set<T>

-- Transformation
set.map<U>(f: Fn<(T) -> U>) -> Set<U>
set.filter(pred: Fn<(T) -> bool>) -> Set<T>
set.flatMap<U>(f: Fn<(T) -> Set<U>>) -> Set<U>
```

### 6.4 Arrays and DataFrames (`std.data`)

#### 6.4.1 Array (Fixed-size, Numeric-optimized)

```geoflow
-- Construction
Array.zeros<T: Numeric>(shape: ...int) -> Array<T>
Array.ones<T: Numeric>(shape: ...int) -> Array<T>
Array.fill<T>(shape: ...int, value: T) -> Array<T>
Array.fromList<T>(data: List<T>) -> Array<T>
Array.range(start: float, end: float, step: float) -> Array<float>
Array.linspace(start: float, end: float, n: int) -> Array<float>
Array.logspace(start: float, end: float, n: int) -> Array<float>

-- Properties
arr.shape() -> List<int>
arr.ndim() -> int
arr.size() -> int
arr.dtype() -> string

-- Indexing
arr[i]                  -- single element
arr[i:j]                -- slice
arr[i, j]               -- 2D access
arr[i:j, k:l]           -- 2D slice

-- Reshaping
arr.reshape(shape: ...int) -> Array<T>
arr.flatten() -> Array<T>
arr.transpose() -> Array<T>
arr.T                   -- shorthand for transpose

-- Element-wise operations (all return new Array)
arr + other             -- addition
arr - other             -- subtraction
arr * other             -- multiplication
arr / other             -- division
arr ** n                -- power
arr.abs() -> Array<T>
arr.sqrt() -> Array<float>
arr.exp() -> Array<float>
arr.log() -> Array<float>
arr.sin() -> Array<float>
arr.cos() -> Array<float>
-- ... all math functions work element-wise

-- Aggregation
arr.sum() -> T
arr.mean() -> float
arr.std() -> float
arr.var() -> float
arr.min() -> T
arr.max() -> T
arr.argmin() -> int
arr.argmax() -> int
arr.cumsum() -> Array<T>
arr.cumprod() -> Array<T>

-- Axis operations
arr.sum(axis: int) -> Array<T>
arr.mean(axis: int) -> Array<float>
-- ... etc with axis parameter
```

#### 6.4.2 Vector (1D with Linear Algebra)

```geoflow
-- Construction
Vector.zeros<T: Numeric>(n: int) -> Vector<T>
Vector.ones<T: Numeric>(n: int) -> Vector<T>
Vector.fromList<T>(data: List<T>) -> Vector<T>
Vector.unit(dim: int, n: int) -> Vector<float>  -- unit vector in dimension dim

-- Properties
vec.length() -> int
vec.norm() -> float             -- Euclidean norm
vec.norm(p: int) -> float       -- p-norm
vec.normalized() -> Vector<float>

-- Operations
vec + other -> Vector<T>
vec - other -> Vector<T>
vec * scalar -> Vector<T>
vec.dot(other: Vector<T>) -> T
vec.cross(other: Vector<T>) -> Vector<T>  -- 3D only
vec.outer(other: Vector<T>) -> Matrix<T>
vec.angle(other: Vector<T>) -> float

-- Projection
vec.projectOnto(other: Vector<T>) -> Vector<T>
vec.rejectFrom(other: Vector<T>) -> Vector<T>
```

#### 6.4.3 Matrix (2D with Linear Algebra)

```geoflow
-- Construction
Matrix.zeros<T: Numeric>(m: int, n: int) -> Matrix<T>
Matrix.ones<T: Numeric>(m: int, n: int) -> Matrix<T>
Matrix.identity(n: int) -> Matrix<float>
Matrix.diagonal(values: Vector<T>) -> Matrix<T>
Matrix.fromRows(rows: List<Vector<T>>) -> Matrix<T>
Matrix.fromCols(cols: List<Vector<T>>) -> Matrix<T>
Matrix.fromList2D(data: List<List<T>>) -> Matrix<T>

-- Properties
mat.rows() -> int
mat.cols() -> int
mat.shape() -> Tuple<int, int>
mat.isSquare() -> bool
mat.isSymmetric() -> bool
mat.trace() -> T
mat.rank() -> int

-- Access
mat[i, j] -> T
mat.row(i: int) -> Vector<T>
mat.col(j: int) -> Vector<T>
mat.diagonal() -> Vector<T>
mat.submatrix(rowStart: int, rowEnd: int, colStart: int, colEnd: int) -> Matrix<T>

-- Arithmetic
mat + other -> Matrix<T>
mat - other -> Matrix<T>
mat * scalar -> Matrix<T>
mat * vec -> Vector<T>          -- matrix-vector multiplication
mat * other -> Matrix<T>        -- matrix multiplication
mat ** n -> Matrix<T>           -- matrix power

-- Transformations
mat.transpose() -> Matrix<T>
mat.T -> Matrix<T>              -- shorthand
mat.inverse() -> Result<Matrix<float>, string>
mat.pseudoInverse() -> Matrix<float>

-- Decompositions
mat.lu() -> Tuple<Matrix<float>, Matrix<float>, Matrix<int>>  -- L, U, P
mat.qr() -> Tuple<Matrix<float>, Matrix<float>>               -- Q, R
mat.svd() -> Tuple<Matrix<float>, Vector<float>, Matrix<float>>  -- U, S, V
mat.cholesky() -> Result<Matrix<float>, string>               -- L where A = LL^T
mat.eigen() -> Tuple<Vector<Complex>, Matrix<Complex>>        -- eigenvalues, eigenvectors

-- Solving
mat.solve(b: Vector<T>) -> Result<Vector<float>, string>      -- solves Ax = b
mat.leastSquares(b: Vector<T>) -> Vector<float>               -- least squares solution

-- Norms
mat.norm() -> float             -- Frobenius norm
mat.norm(p: string) -> float    -- "fro", "1", "inf", "2"
mat.condition() -> float        -- condition number
mat.determinant() -> T
```

#### 6.4.4 DataFrame (Tabular Data)

```geoflow
-- Construction
DataFrame.empty() -> DataFrame
DataFrame.fromRows(rows: List<Map<string, auto>>) -> DataFrame
DataFrame.fromCols(cols: Map<string, List<auto>>) -> DataFrame
DataFrame.fromCSV(path: string, options: CSVOptions = default) -> Result<DataFrame, string>
DataFrame.fromJSON(path: string) -> Result<DataFrame, string>

-- Properties
df.columns() -> List<string>
df.dtypes() -> Map<string, string>
df.shape() -> Tuple<int, int>
df.rowCount() -> int
df.colCount() -> int

-- Access
df[colName] -> Series<auto>
df[colNames: List<string>] -> DataFrame
df.row(i: int) -> Map<string, auto>
df.rows(indices: List<int>) -> DataFrame
df.head(n: int = 5) -> DataFrame
df.tail(n: int = 5) -> DataFrame
df.sample(n: int) -> DataFrame

-- Column operations
df.select(cols: ...string) -> DataFrame
df.drop(cols: ...string) -> DataFrame
df.rename(mapping: Map<string, string>) -> DataFrame
df.addColumn(name: string, values: Series<auto>) -> DataFrame
df.addColumn(name: string, f: Fn<(Map<string, auto>) -> auto>) -> DataFrame
df.withColumn(name: string, f: Fn<(Series<auto>) -> Series<auto>>) -> DataFrame

-- Row operations
df.filter(pred: Fn<(Map<string, auto>) -> bool>) -> DataFrame
df.filterByColumn(col: string, pred: Fn<(auto) -> bool>) -> DataFrame
df.where(condition: string) -> DataFrame  -- SQL-like syntax
df.sortBy(col: string, ascending: bool = true) -> DataFrame
df.sortBy(cols: List<Tuple<string, bool>>) -> DataFrame
df.distinct() -> DataFrame
df.distinctBy(cols: ...string) -> DataFrame
df.dropNulls() -> DataFrame
df.dropNulls(cols: ...string) -> DataFrame
df.fillNulls(value: auto) -> DataFrame
df.fillNulls(mapping: Map<string, auto>) -> DataFrame

-- Transformation
df.map(f: Fn<(Map<string, auto>) -> Map<string, auto>>) -> DataFrame
df.apply(col: string, f: Fn<(auto) -> auto>) -> DataFrame

-- Aggregation
df.groupBy(cols: ...string) -> GroupedDataFrame
df.agg(aggregations: Map<string, string>) -> DataFrame
-- aggregation functions: "sum", "mean", "min", "max", "count", "first", "last", "std", "var"

grouped.agg(aggregations: Map<string, string>) -> DataFrame
grouped.sum() -> DataFrame
grouped.mean() -> DataFrame
grouped.count() -> DataFrame

-- Joins
df.join(other: DataFrame, on: string, how: string = "inner") -> DataFrame
df.join(other: DataFrame, left: string, right: string, how: string = "inner") -> DataFrame
-- how: "inner", "left", "right", "outer", "cross"

df.concat(other: DataFrame) -> DataFrame
df.union(other: DataFrame) -> DataFrame

-- Pivoting
df.pivot(index: string, columns: string, values: string) -> DataFrame
df.melt(idVars: List<string>, valueVars: List<string>) -> DataFrame

-- I/O
df.toCSV(path: string, options: CSVOptions = default) -> Result<unit, string>
df.toJSON(path: string) -> Result<unit, string>
df.toList() -> List<Map<string, auto>>
df.toMap() -> Map<string, List<auto>>

-- Statistics
df.describe() -> DataFrame      -- summary statistics for numeric columns
df.corr() -> Matrix<float>      -- correlation matrix
df.cov() -> Matrix<float>       -- covariance matrix
```

#### 6.4.5 Series (1D DataFrame Column)

```geoflow
-- Properties
series.name() -> string
series.dtype() -> string
series.length() -> int

-- Access
series[i] -> auto
series[i:j] -> Series<T>

-- Statistics
series.sum() -> T
series.mean() -> float
series.median() -> float
series.mode() -> T
series.std() -> float
series.var() -> float
series.min() -> T
series.max() -> T
series.quantile(q: float) -> float
series.describe() -> Map<string, float>

-- Transformation
series.map<U>(f: Fn<(T) -> U>) -> Series<U>
series.filter(pred: Fn<(T) -> bool>) -> Series<T>
series.sort() -> Series<T>
series.unique() -> Series<T>
series.valueCounts() -> Map<T, int>

-- Missing values
series.isNull() -> Series<bool>
series.notNull() -> Series<bool>
series.dropNulls() -> Series<T>
series.fillNulls(value: T) -> Series<T>

-- String operations (when T = string)
series.str.uppercase() -> Series<string>
series.str.lowercase() -> Series<string>
series.str.contains(pattern: string) -> Series<bool>
series.str.replace(old: string, new: string) -> Series<string>
series.str.split(sep: string) -> Series<List<string>>

-- Datetime operations (when T = DateTime)
series.dt.year() -> Series<int>
series.dt.month() -> Series<int>
series.dt.day() -> Series<int>
series.dt.hour() -> Series<int>
series.dt.dayOfWeek() -> Series<int>
```

### 6.5 Logging (`std.log`)

```geoflow
import std.log

-- Log levels
log.Level = enum { Trace, Debug, Info, Warn, Error, Fatal }

-- Basic logging
log.trace(message: string)
log.debug(message: string)
log.info(message: string)
log.warn(message: string)
log.error(message: string)
log.fatal(message: string)  -- also terminates program

-- Structured logging
log.info("Processing point", { "x": p.x, "y": p.y, "id": id })

-- With formatting
log.info("Processed {count} points in {time}ms", { "count": n, "time": elapsed })

-- Configuration
log.setLevel(level: log.Level)
log.setFormat(format: string)  -- e.g., "[{level}] {time} - {message}"
log.setOutput(writer: Writer)
log.addOutput(writer: Writer)

-- Context/scoped logging
let logger = log.withContext({ "module": "geo", "operation": "buffer" })
logger.info("Starting buffer operation")

-- Timed operations
log.timed("Buffer operation", fn() {
    -- operation here
})
-- Output: "Buffer operation completed in 123ms"

-- Conditional logging
log.debugIf(condition, "Debug message")

-- Outputs
log.ConsoleOutput(colored: bool = true) -> Writer
log.FileOutput(path: string, rotate: bool = false) -> Writer
log.JSONOutput(writer: Writer) -> Writer
```

### 6.6 Assertions (`std.assert`)

```geoflow
import std.assert

-- Basic assertions (panic on failure)
assert(condition: bool)
assert(condition: bool, message: string)

-- Value assertions
assert.equal<T>(actual: T, expected: T)
assert.notEqual<T>(actual: T, expected: T)
assert.same<T>(actual: T, expected: T)     -- reference equality
assert.nil<T>(value: Option<T>)
assert.notNil<T>(value: Option<T>)
assert.ok<T, E>(result: Result<T, E>)
assert.err<T, E>(result: Result<T, E>)

-- Numeric assertions
assert.approximately(actual: float, expected: float, epsilon: float = 1e-10)
assert.greaterThan<T: Ord>(actual: T, expected: T)
assert.lessThan<T: Ord>(actual: T, expected: T)
assert.greaterOrEqual<T: Ord>(actual: T, expected: T)
assert.lessOrEqual<T: Ord>(actual: T, expected: T)
assert.between<T: Ord>(value: T, low: T, high: T)
assert.positive(value: float)
assert.negative(value: float)
assert.finite(value: float)
assert.NaN(value: float)

-- Collection assertions
assert.empty<T>(collection: List<T>)
assert.notEmpty<T>(collection: List<T>)
assert.length<T>(collection: List<T>, expected: int)
assert.contains<T>(collection: List<T>, item: T)
assert.containsAll<T>(collection: List<T>, items: List<T>)
assert.containsNone<T>(collection: List<T>, items: List<T>)

-- String assertions
assert.startsWith(actual: string, prefix: string)
assert.endsWith(actual: string, suffix: string)
assert.contains(actual: string, substring: string)
assert.matches(actual: string, pattern: string)
assert.blank(actual: string)

-- Type assertions
assert.type<T>(value: auto)
assert.instanceof<T>(value: auto)

-- Exception assertions
assert.throws(f: Fn<() -> auto>)
assert.throws<E>(f: Fn<() -> auto>)
assert.doesNotThrow(f: Fn<() -> auto>)

-- Geospatial assertions
assert.geo.valid(geometry: geom.Geometry)
assert.geo.equal(a: geom.Geometry, b: geom.Geometry, tolerance: float = 1e-10)
assert.geo.intersects(a: geom.Geometry, b: geom.Geometry)
assert.geo.contains(container: geom.Geometry, contained: geom.Geometry)
assert.geo.within(inner: geom.Geometry, outer: geom.Geometry)
assert.geo.disjoint(a: geom.Geometry, b: geom.Geometry)
assert.geo.nearEqual(a: geog.Point, b: geog.Point, toleranceMeters: float)

-- Soft assertions (collect failures, report at end)
let soft = assert.soft()
soft.equal(a, b)
soft.greaterThan(x, y)
soft.assertAll()  -- throws if any failed
```

### 6.7 I/O (`std.io`)

```geoflow
import std.io

-- File reading
io.readFile(path: string) -> Result<string, IOError>
io.readBytes(path: string) -> Result<List<byte>, IOError>
io.readLines(path: string) -> Result<List<string>, IOError>

-- File writing
io.writeFile(path: string, content: string) -> Result<unit, IOError>
io.writeBytes(path: string, content: List<byte>) -> Result<unit, IOError>
io.appendFile(path: string, content: string) -> Result<unit, IOError>

-- File info
io.exists(path: string) -> bool
io.isFile(path: string) -> bool
io.isDir(path: string) -> bool
io.fileSize(path: string) -> Result<int, IOError>
io.modifiedTime(path: string) -> Result<DateTime, IOError>

-- Directory operations
io.listDir(path: string) -> Result<List<string>, IOError>
io.createDir(path: string) -> Result<unit, IOError>
io.createDirAll(path: string) -> Result<unit, IOError>
io.removeFile(path: string) -> Result<unit, IOError>
io.removeDir(path: string) -> Result<unit, IOError>
io.removeDirAll(path: string) -> Result<unit, IOError>
io.copy(src: string, dst: string) -> Result<unit, IOError>
io.move(src: string, dst: string) -> Result<unit, IOError>

-- Path utilities
io.path.join(parts: ...string) -> string
io.path.dirname(path: string) -> string
io.path.basename(path: string) -> string
io.path.extension(path: string) -> string
io.path.absolute(path: string) -> string
io.path.relative(path: string, base: string) -> string

-- Streaming
io.Reader trait
io.Writer trait
io.openFile(path: string, mode: string) -> Result<File, IOError>
-- mode: "r", "w", "a", "rw"

-- Stdin/Stdout/Stderr
io.stdin: Reader
io.stdout: Writer
io.stderr: Writer
io.print(value: auto)
io.println(value: auto)
io.readLine() -> Result<string, IOError>
```

### 6.8 Time (`std.time`)

```geoflow
import std.time

-- Types
DateTime
Duration
Instant

-- Current time
time.now() -> DateTime
time.instant() -> Instant      -- monotonic clock for measuring

-- Construction
DateTime.parse(s: string, format: string = "ISO8601") -> Result<DateTime, string>
DateTime.fromUnix(seconds: int) -> DateTime
DateTime.fromUnixMillis(millis: int) -> DateTime
DateTime.new(year: int, month: int, day: int, hour: int = 0, minute: int = 0, second: int = 0) -> DateTime

Duration.seconds(n: int) -> Duration
Duration.minutes(n: int) -> Duration
Duration.hours(n: int) -> Duration
Duration.days(n: int) -> Duration
Duration.millis(n: int) -> Duration

-- DateTime components
dt.year() -> int
dt.month() -> int
dt.day() -> int
dt.hour() -> int
dt.minute() -> int
dt.second() -> int
dt.millisecond() -> int
dt.dayOfWeek() -> int       -- 1 = Monday, 7 = Sunday
dt.dayOfYear() -> int
dt.weekOfYear() -> int
dt.quarter() -> int

-- DateTime arithmetic
dt + duration -> DateTime
dt - duration -> DateTime
dt1 - dt2 -> Duration

-- DateTime comparison
dt1 < dt2
dt1 == dt2
dt.isBefore(other: DateTime) -> bool
dt.isAfter(other: DateTime) -> bool
dt.isBetween(start: DateTime, end: DateTime) -> bool

-- Formatting
dt.format(pattern: string) -> string
dt.toISO8601() -> string
dt.toUnix() -> int
dt.toUnixMillis() -> int

-- Manipulation
dt.withYear(year: int) -> DateTime
dt.withMonth(month: int) -> DateTime
dt.withDay(day: int) -> DateTime
dt.startOfDay() -> DateTime
dt.endOfDay() -> DateTime
dt.startOfMonth() -> DateTime
dt.startOfYear() -> DateTime

-- Timezone
dt.inTimezone(tz: string) -> DateTime
dt.toUTC() -> DateTime
dt.timezone() -> string
dt.offset() -> Duration

-- Duration operations
dur.toSeconds() -> float
dur.toMinutes() -> float
dur.toHours() -> float
dur.toDays() -> float
dur.toMillis() -> int

-- Measuring
let start = time.instant()
-- ... work ...
let elapsed = start.elapsed() -> Duration
```

---

## 7. Math Library

### 7.1 Basic Math (`std.math`)

```geoflow
import std.math

-- Constants
math.PI: float
math.E: float
math.TAU: float         -- 2*PI
math.PHI: float         -- golden ratio
math.INF: float         -- infinity
math.NEG_INF: float
math.NaN: float

-- Basic functions
math.abs(x: float) -> float
math.sign(x: float) -> int      -- -1, 0, or 1
math.floor(x: float) -> int
math.ceil(x: float) -> int
math.round(x: float) -> int
math.trunc(x: float) -> int
math.frac(x: float) -> float    -- fractional part

-- Powers and roots
math.sqrt(x: float) -> float
math.cbrt(x: float) -> float
math.pow(base: float, exp: float) -> float
math.exp(x: float) -> float
math.exp2(x: float) -> float
math.expm1(x: float) -> float   -- e^x - 1, accurate for small x

-- Logarithms
math.log(x: float) -> float     -- natural log
math.log2(x: float) -> float
math.log10(x: float) -> float
math.log1p(x: float) -> float   -- log(1+x), accurate for small x
math.logBase(base: float, x: float) -> float

-- Trigonometry
math.sin(x: float) -> float
math.cos(x: float) -> float
math.tan(x: float) -> float
math.asin(x: float) -> float
math.acos(x: float) -> float
math.atan(x: float) -> float
math.atan2(y: float, x: float) -> float

-- Hyperbolic
math.sinh(x: float) -> float
math.cosh(x: float) -> float
math.tanh(x: float) -> float
math.asinh(x: float) -> float
math.acosh(x: float) -> float
math.atanh(x: float) -> float

-- Angle conversion
math.toRadians(degrees: float) -> float
math.toDegrees(radians: float) -> float

-- Comparison and bounds
math.min(a: float, b: float) -> float
math.max(a: float, b: float) -> float
math.clamp(x: float, low: float, high: float) -> float

-- Special values
math.isNaN(x: float) -> bool
math.isInf(x: float) -> bool
math.isFinite(x: float) -> bool

-- Modular arithmetic
math.mod(a: int, b: int) -> int
math.gcd(a: int, b: int) -> int
math.lcm(a: int, b: int) -> int
math.divMod(a: int, b: int) -> Tuple<int, int>

-- Combinatorics
math.factorial(n: int) -> int
math.permutations(n: int, r: int) -> int
math.combinations(n: int, r: int) -> int
math.binomial(n: int, k: int) -> int

-- Number theory
math.isPrime(n: int) -> bool
math.primeFactors(n: int) -> List<int>
math.nextPrime(n: int) -> int
math.nthPrime(n: int) -> int

-- Interpolation
math.lerp(a: float, b: float, t: float) -> float
math.inverseLerp(a: float, b: float, value: float) -> float
math.smoothstep(edge0: float, edge1: float, x: float) -> float
```

### 7.2 Statistics (`std.math.stats`)

```geoflow
import std.math.stats

-- Descriptive statistics
stats.mean(data: List<float>) -> float
stats.median(data: List<float>) -> float
stats.mode(data: List<float>) -> List<float>
stats.variance(data: List<float>, ddof: int = 0) -> float
stats.std(data: List<float>, ddof: int = 0) -> float
stats.sem(data: List<float>) -> float           -- standard error of mean
stats.skewness(data: List<float>) -> float
stats.kurtosis(data: List<float>) -> float

-- Quantiles
stats.quantile(data: List<float>, q: float) -> float
stats.percentile(data: List<float>, p: float) -> float
stats.quartiles(data: List<float>) -> Tuple<float, float, float>
stats.iqr(data: List<float>) -> float           -- interquartile range

-- Range statistics
stats.min(data: List<float>) -> float
stats.max(data: List<float>) -> float
stats.range(data: List<float>) -> float
stats.sum(data: List<float>) -> float
stats.product(data: List<float>) -> float

-- Weighted statistics
stats.weightedMean(data: List<float>, weights: List<float>) -> float
stats.weightedVariance(data: List<float>, weights: List<float>) -> float

-- Correlation and covariance
stats.cov(x: List<float>, y: List<float>) -> float
stats.corr(x: List<float>, y: List<float>) -> float
stats.spearman(x: List<float>, y: List<float>) -> float
stats.kendall(x: List<float>, y: List<float>) -> float
stats.autocorr(data: List<float>, lag: int) -> float

-- Regression
stats.linearRegression(x: List<float>, y: List<float>) -> LinearModel
type LinearModel = struct {
    slope: float,
    intercept: float,
    rSquared: float,
    stdErr: float
}
model.predict(x: float) -> float
model.predict(xs: List<float>) -> List<float>

stats.polynomialRegression(x: List<float>, y: List<float>, degree: int) -> PolyModel

-- Distributions
stats.normal(mean: float, std: float) -> Distribution
stats.uniform(low: float, high: float) -> Distribution
stats.exponential(rate: float) -> Distribution
stats.poisson(lambda: float) -> Distribution
stats.binomial(n: int, p: float) -> Distribution
stats.gamma(shape: float, scale: float) -> Distribution
stats.beta(alpha: float, beta: float) -> Distribution
stats.chisquare(df: int) -> Distribution
stats.t(df: int) -> Distribution
stats.f(df1: int, df2: int) -> Distribution

-- Distribution methods
dist.pdf(x: float) -> float         -- probability density
dist.cdf(x: float) -> float         -- cumulative distribution
dist.ppf(p: float) -> float         -- percent point (inverse CDF)
dist.sample() -> float
dist.sample(n: int) -> List<float>
dist.mean() -> float
dist.variance() -> float

-- Hypothesis testing
stats.tTest(sample1: List<float>, sample2: List<float>) -> TestResult
stats.pairedTTest(sample1: List<float>, sample2: List<float>) -> TestResult
stats.oneSampleTTest(sample: List<float>, mu: float) -> TestResult
stats.anova(groups: ...List<float>) -> TestResult
stats.chiSquareTest(observed: List<int>, expected: List<float>) -> TestResult
stats.ksTest(data: List<float>, dist: Distribution) -> TestResult

type TestResult = struct {
    statistic: float,
    pValue: float,
    degreesOfFreedom: int
}

-- Confidence intervals
stats.confidenceInterval(data: List<float>, level: float = 0.95) -> Tuple<float, float>
stats.bootstrapCI(data: List<float>, statFn: Fn<(List<float>) -> float>, nResamples: int = 1000, level: float = 0.95) -> Tuple<float, float>

-- Normalization
stats.zscore(data: List<float>) -> List<float>
stats.minMaxScale(data: List<float>, newMin: float = 0.0, newMax: float = 1.0) -> List<float>
stats.standardize(data: List<float>) -> List<float>

-- Outlier detection
stats.outliers(data: List<float>, method: string = "iqr") -> List<int>  -- indices
stats.trimmedMean(data: List<float>, proportionToCut: float) -> float
stats.winsorize(data: List<float>, limits: Tuple<float, float>) -> List<float>

-- Moving statistics
stats.movingMean(data: List<float>, window: int) -> List<float>
stats.movingStd(data: List<float>, window: int) -> List<float>
stats.ewma(data: List<float>, alpha: float) -> List<float>  -- exponentially weighted moving average
```

### 7.3 Symbolic Math (`std.math.symbolic`)

```geoflow
import std.math.symbolic as sym

-- Expression type
Expr

-- Variables
let x = sym.var("x")
let y = sym.var("y")
let z = sym.var("z")

-- Constants
sym.const(value: float) -> Expr
sym.pi -> Expr
sym.e -> Expr

-- Building expressions (overloaded operators)
let f = x^2 + 3*x + 2
let g = sym.sin(x) * sym.cos(y)
let h = sym.exp(-x^2)

-- Arithmetic (all return Expr)
e1 + e2
e1 - e2
e1 * e2
e1 / e2
e1 ^ e2
-e1

-- Functions (all return Expr)
sym.sin(e: Expr) -> Expr
sym.cos(e: Expr) -> Expr
sym.tan(e: Expr) -> Expr
sym.asin(e: Expr) -> Expr
sym.acos(e: Expr) -> Expr
sym.atan(e: Expr) -> Expr
sym.sinh(e: Expr) -> Expr
sym.cosh(e: Expr) -> Expr
sym.tanh(e: Expr) -> Expr
sym.exp(e: Expr) -> Expr
sym.log(e: Expr) -> Expr
sym.sqrt(e: Expr) -> Expr
sym.abs(e: Expr) -> Expr

-- Calculus: Differentiation
sym.diff(expr: Expr, var: Expr) -> Expr
sym.diff(expr: Expr, var: Expr, n: int) -> Expr  -- nth derivative
sym.gradient(expr: Expr, vars: List<Expr>) -> List<Expr>
sym.jacobian(exprs: List<Expr>, vars: List<Expr>) -> Matrix<Expr>
sym.hessian(expr: Expr, vars: List<Expr>) -> Matrix<Expr>

-- Examples:
let f = x^3 + 2*x^2 - x + 1
let df = sym.diff(f, x)          -- 3*x^2 + 4*x - 1
let d2f = sym.diff(f, x, 2)      -- 6*x + 4

-- Calculus: Integration
sym.integrate(expr: Expr, var: Expr) -> Expr                           -- indefinite
sym.integrate(expr: Expr, var: Expr, lower: Expr, upper: Expr) -> Expr -- definite
sym.doubleIntegrate(expr: Expr, var1: Expr, var2: Expr, bounds: ...) -> Expr
sym.tripleIntegrate(expr: Expr, var1: Expr, var2: Expr, var3: Expr, bounds: ...) -> Expr

-- Example:
let f = x^2
let F = sym.integrate(f, x)              -- x^3/3
let area = sym.integrate(f, x, 0, 1)     -- 1/3

-- Limits
sym.limit(expr: Expr, var: Expr, point: Expr) -> Expr
sym.limit(expr: Expr, var: Expr, point: Expr, dir: string) -> Expr  -- "left" or "right"

-- Series expansion
sym.taylor(expr: Expr, var: Expr, point: Expr, order: int) -> Expr
sym.maclaurin(expr: Expr, var: Expr, order: int) -> Expr

-- Simplification
sym.simplify(expr: Expr) -> Expr
sym.expand(expr: Expr) -> Expr
sym.factor(expr: Expr) -> Expr
sym.collect(expr: Expr, var: Expr) -> Expr
sym.cancel(expr: Expr) -> Expr
sym.trigSimplify(expr: Expr) -> Expr
sym.rationalize(expr: Expr) -> Expr

-- Solving
sym.solve(equation: Expr, var: Expr) -> List<Expr>
sym.solve(equations: List<Expr>, vars: List<Expr>) -> List<Map<Expr, Expr>>
sym.roots(polynomial: Expr, var: Expr) -> List<Complex>

-- Example:
let solutions = sym.solve(x^2 - 5*x + 6, x)  -- [2, 3]

-- Differential equations
sym.dsolve(ode: Expr, func: Expr, var: Expr) -> Expr
sym.dsolve(ode: Expr, func: Expr, var: Expr, ics: Map<Expr, float>) -> Expr  -- with initial conditions

-- Example:
let y = sym.func("y", x)         -- y(x)
let dy = sym.diff(y, x)          -- y'(x)
let ode = dy - 2*y               -- y' - 2y = 0
let solution = sym.dsolve(ode, y, x)  -- C*exp(2*x)

-- Substitution
expr.substitute(var: Expr, value: Expr) -> Expr
expr.substitute(mapping: Map<Expr, Expr>) -> Expr

-- Realization (symbolic -> numeric)
expr.realize(bindings: Map<string, float>) -> float
expr.realize(bindings: Map<Expr, float>) -> float

-- Example:
let f = x^2 + 3*x + 2
let value = f.realize({"x": 5})   -- 42.0

-- Conversion
expr.toString() -> string
expr.toLatex() -> string
expr.toMathML() -> string
expr.toFunction(vars: List<Expr>) -> Fn<(...float) -> float>

-- Example:
let f = x^2 + y
let numericFn = f.toFunction([x, y])
let result = numericFn(3.0, 2.0)  -- 11.0

-- Expression properties
expr.isPolynomial(var: Expr) -> bool
expr.degree(var: Expr) -> Option<int>
expr.coefficients(var: Expr) -> List<Expr>
expr.freeSymbols() -> Set<Expr>
expr.isConstant() -> bool
```

### 7.4 Linear Algebra (`std.math.linalg`)

```geoflow
import std.math.linalg as la

-- See Matrix and Vector types in std.data section
-- Additional linear algebra operations:

-- Solving linear systems
la.solve(A: Matrix<float>, b: Vector<float>) -> Result<Vector<float>, string>
la.solveLU(A: Matrix<float>, b: Vector<float>) -> Vector<float>
la.solveQR(A: Matrix<float>, b: Vector<float>) -> Vector<float>
la.solveSVD(A: Matrix<float>, b: Vector<float>) -> Vector<float>

-- Least squares
la.lstsq(A: Matrix<float>, b: Vector<float>) -> LeastSquaresResult
type LeastSquaresResult = struct {
    solution: Vector<float>,
    residuals: float,
    rank: int
}

-- Eigenvalue problems
la.eig(A: Matrix<float>) -> Tuple<Vector<Complex>, Matrix<Complex>>
la.eigvals(A: Matrix<float>) -> Vector<Complex>
la.eigsh(A: Matrix<float>, k: int) -> Tuple<Vector<float>, Matrix<float>>  -- symmetric, top k

-- Singular value decomposition
la.svd(A: Matrix<float>) -> Tuple<Matrix<float>, Vector<float>, Matrix<float>>
la.svdvals(A: Matrix<float>) -> Vector<float>
la.lowRankApprox(A: Matrix<float>, k: int) -> Matrix<float>

-- Matrix factorizations
la.lu(A: Matrix<float>) -> Tuple<Matrix<float>, Matrix<float>, Vector<int>>
la.qr(A: Matrix<float>) -> Tuple<Matrix<float>, Matrix<float>>
la.cholesky(A: Matrix<float>) -> Result<Matrix<float>, string>
la.schur(A: Matrix<float>) -> Tuple<Matrix<float>, Matrix<float>>

-- Matrix properties
la.rank(A: Matrix<float>) -> int
la.nullspace(A: Matrix<float>) -> Matrix<float>
la.columnspace(A: Matrix<float>) -> Matrix<float>
la.det(A: Matrix<float>) -> float
la.trace(A: Matrix<float>) -> float
la.cond(A: Matrix<float>, p: string = "2") -> float

-- Norms
la.norm(v: Vector<float>, p: int = 2) -> float
la.norm(A: Matrix<float>, p: string = "fro") -> float

-- Special matrices
la.hilbert(n: int) -> Matrix<float>
la.vandermonde(x: Vector<float>, n: int) -> Matrix<float>
la.toeplitz(c: Vector<float>, r: Vector<float>) -> Matrix<float>
la.companion(coeffs: Vector<float>) -> Matrix<float>

-- Kronecker and other products
la.kron(A: Matrix<float>, B: Matrix<float>) -> Matrix<float>
la.hadamard(A: Matrix<float>, B: Matrix<float>) -> Matrix<float>

-- Matrix exponential and functions
la.expm(A: Matrix<float>) -> Matrix<float>
la.logm(A: Matrix<float>) -> Matrix<float>
la.sqrtm(A: Matrix<float>) -> Matrix<float>
la.funm(A: Matrix<float>, f: Fn<(float) -> float>) -> Matrix<float>
```

### 7.5 Numerical Methods (`std.math.numeric`)

```geoflow
import std.math.numeric as num

-- Root finding
num.bisect(f: Fn<(float) -> float>, a: float, b: float, tol: float = 1e-10) -> float
num.newton(f: Fn<(float) -> float>, df: Fn<(float) -> float>, x0: float, tol: float = 1e-10) -> float
num.secant(f: Fn<(float) -> float>, x0: float, x1: float, tol: float = 1e-10) -> float
num.brent(f: Fn<(float) -> float>, a: float, b: float, tol: float = 1e-10) -> float
num.findRoots(f: Fn<(float) -> float>, interval: Tuple<float, float>, n: int = 100) -> List<float>

-- Optimization
num.minimize(f: Fn<(float) -> float>, bounds: Tuple<float, float>) -> OptResult
num.minimize(f: Fn<(Vector<float>) -> float>, x0: Vector<float>, method: string = "bfgs") -> OptResult
num.maximize(f: Fn<(float) -> float>, bounds: Tuple<float, float>) -> OptResult
num.gradientDescent(f: Fn<(Vector<float>) -> float>, grad: Fn<(Vector<float>) -> Vector<float>>, x0: Vector<float>, lr: float = 0.01, maxIter: int = 1000) -> Vector<float>

type OptResult = struct {
    x: auto,
    fval: float,
    iterations: int,
    converged: bool
}

-- Integration (quadrature)
num.quad(f: Fn<(float) -> float>, a: float, b: float) -> float
num.quadGauss(f: Fn<(float) -> float>, a: float, b: float, n: int = 5) -> float
num.romberg(f: Fn<(float) -> float>, a: float, b: float, tol: float = 1e-10) -> float
num.dblquad(f: Fn<(float, float) -> float>, ax: float, bx: float, ay: float, by: float) -> float
num.tplquad(f: Fn<(float, float, float) -> float>, bounds: ...) -> float

-- Differentiation
num.diff(f: Fn<(float) -> float>, x: float, h: float = 1e-8) -> float
num.diff2(f: Fn<(float) -> float>, x: float, h: float = 1e-5) -> float
num.gradient(f: Fn<(Vector<float>) -> float>, x: Vector<float>, h: float = 1e-8) -> Vector<float>
num.jacobian(f: Fn<(Vector<float>) -> Vector<float>>, x: Vector<float>) -> Matrix<float>
num.hessian(f: Fn<(Vector<float>) -> float>, x: Vector<float>) -> Matrix<float>

-- Interpolation
num.interp1d(xs: List<float>, ys: List<float>, kind: string = "linear") -> Fn<(float) -> float>
-- kinds: "linear", "cubic", "quadratic", "nearest"
num.interp2d(xs: List<float>, ys: List<float>, zs: Matrix<float>) -> Fn<(float, float) -> float>
num.spline(xs: List<float>, ys: List<float>) -> Spline
spline.evaluate(x: float) -> float
spline.derivative(x: float) -> float
spline.integral(a: float, b: float) -> float

-- Curve fitting
num.polyfit(xs: List<float>, ys: List<float>, degree: int) -> Vector<float>
num.polyval(coeffs: Vector<float>, x: float) -> float
num.fit(f: Fn<(float, ...float) -> float>, xs: List<float>, ys: List<float>, p0: List<float>) -> FitResult
type FitResult = struct {
    params: List<float>,
    covariance: Matrix<float>,
    rSquared: float
}

-- ODE solvers
num.odeint(f: Fn<(float, Vector<float>) -> Vector<float>>, y0: Vector<float>, t: List<float>) -> Matrix<float>
num.rk45(f: Fn<(float, Vector<float>) -> Vector<float>>, y0: Vector<float>, tSpan: Tuple<float, float>, maxStep: float = 0.1) -> ODESolution
type ODESolution = struct {
    t: List<float>,
    y: Matrix<float>
}
solution.evaluate(t: float) -> Vector<float>

-- FFT
num.fft(data: List<Complex>) -> List<Complex>
num.ifft(data: List<Complex>) -> List<Complex>
num.rfft(data: List<float>) -> List<Complex>
num.irfft(data: List<Complex>) -> List<float>
num.fftfreq(n: int, d: float = 1.0) -> List<float>
```

### 7.6 Complex Numbers (`std.math.complex`)

```geoflow
import std.math.complex

-- Type
Complex

-- Construction
Complex.new(real: float, imag: float) -> Complex
Complex.fromPolar(r: float, theta: float) -> Complex
complex.i -> Complex   -- imaginary unit

-- Properties
c.real() -> float
c.imag() -> float
c.abs() -> float      -- magnitude
c.arg() -> float      -- phase angle
c.conjugate() -> Complex

-- Arithmetic (overloaded operators)
c1 + c2
c1 - c2
c1 * c2
c1 / c2
c ^ n

-- Functions
complex.exp(c: Complex) -> Complex
complex.log(c: Complex) -> Complex
complex.sqrt(c: Complex) -> Complex
complex.sin(c: Complex) -> Complex
complex.cos(c: Complex) -> Complex
complex.tan(c: Complex) -> Complex
complex.sinh(c: Complex) -> Complex
complex.cosh(c: Complex) -> Complex
complex.tanh(c: Complex) -> Complex
```

---

## 8. Geospatial Library

### 8.1 Core Concepts

GeoFlow distinguishes between:
- **Geometry (`geom`)**: Planar/Cartesian operations. Distances in coordinate units.
- **Geography (`geog`)**: Spheroidal/geodetic operations. Distances in meters. Default datum: WGS84.

### 8.2 Geometry (`std.geo.geom`)

#### 8.2.1 Construction

```geoflow
import std.geo.geom

-- Point
geom.Point(x: float, y: float) -> geom.Point
geom.PointZ(x: float, y: float, z: float) -> geom.PointZ
geom.PointM(x: float, y: float, m: float) -> geom.PointM
geom.PointZM(x: float, y: float, z: float, m: float) -> geom.PointZM

-- LineString
geom.LineString(points: List<geom.Point>) -> geom.LineString
geom.LineStringZ(points: List<geom.PointZ>) -> geom.LineStringZ

-- Polygon (first ring is exterior, rest are holes)
geom.Polygon(rings: List<List<geom.Point>>) -> geom.Polygon
geom.Polygon(exterior: List<geom.Point>) -> geom.Polygon  -- no holes

-- Multi-geometries
geom.MultiPoint(points: List<geom.Point>) -> geom.MultiPoint
geom.MultiLineString(lines: List<geom.LineString>) -> geom.MultiLineString
geom.MultiPolygon(polys: List<geom.Polygon>) -> geom.MultiPolygon
geom.GeometryCollection(geoms: List<geom.Geometry>) -> geom.GeometryCollection

-- From WKT
geom.fromWKT(wkt: string) -> Result<geom.Geometry, string>
geom.fromWKB(wkb: bytes) -> Result<geom.Geometry, string>
geom.fromGeoJSON(json: string) -> Result<geom.Geometry, string>

-- WKT Literal syntax
#POINT(1 2)#
#LINESTRING(0 0, 1 1, 2 0)#
#POLYGON((0 0, 10 0, 10 10, 0 10, 0 0), (2 2, 2 8, 8 8, 8 2, 2 2))#
#POINT Z(1 2 3)#
```

#### 8.2.2 Properties

```geoflow
-- All geometries
g.type() -> string              -- "Point", "Polygon", etc.
g.isEmpty() -> bool
g.isValid() -> bool
g.isSimple() -> bool
g.dimension() -> int            -- 0=point, 1=line, 2=polygon
g.coordinateDimension() -> int  -- 2, 3, or 4 (with M)
g.hasZ() -> bool
g.hasM() -> bool
g.srid() -> Option<int>
g.setSRID(srid: int) -> Geometry
g.envelope() -> geom.Polygon    -- bounding box as polygon
g.bounds() -> Tuple<float, float, float, float>  -- minX, minY, maxX, maxY

-- Point specific
p.x -> float
p.y -> float
p.z -> Option<float>
p.m -> Option<float>

-- LineString specific
ls.numPoints() -> int
ls.pointN(n: int) -> geom.Point
ls.startPoint() -> geom.Point
ls.endPoint() -> geom.Point
ls.isClosed() -> bool
ls.isRing() -> bool
ls.points() -> List<geom.Point>

-- Polygon specific
poly.exteriorRing() -> geom.LineString
poly.numInteriorRings() -> int
poly.interiorRingN(n: int) -> geom.LineString
poly.interiorRings() -> List<geom.LineString>

-- Multi-geometry specific
mg.numGeometries() -> int
mg.geometryN(n: int) -> geom.Geometry
mg.geometries() -> List<geom.Geometry>
```

#### 8.2.3 Measurements

```geoflow
-- Length (for lines)
ls.length() -> float

-- Area (for polygons)
poly.area() -> float
poly.signedArea() -> float    -- positive=CCW, negative=CW

-- Perimeter
poly.perimeter() -> float

-- Distance between geometries
g1.distance(g2: geom.Geometry) -> float
g1.hausdorffDistance(g2: geom.Geometry) -> float
g1.frechetDistance(g2: geom.Geometry) -> float

-- Point calculations
p1.distanceTo(p2: geom.Point) -> float
p1.azimuthTo(p2: geom.Point) -> float  -- angle in radians
```

#### 8.2.4 Spatial Relationships

```geoflow
-- DE-9IM relationships
g1.equals(g2: geom.Geometry) -> bool
g1.disjoint(g2: geom.Geometry) -> bool
g1.intersects(g2: geom.Geometry) -> bool
g1.touches(g2: geom.Geometry) -> bool
g1.crosses(g2: geom.Geometry) -> bool
g1.within(g2: geom.Geometry) -> bool
g1.contains(g2: geom.Geometry) -> bool
g1.overlaps(g2: geom.Geometry) -> bool
g1.covers(g2: geom.Geometry) -> bool
g1.coveredBy(g2: geom.Geometry) -> bool
g1.relate(g2: geom.Geometry) -> string  -- DE-9IM matrix
g1.relate(g2: geom.Geometry, pattern: string) -> bool
```

#### 8.2.5 Spatial Operations

```geoflow
-- Set operations
g1.intersection(g2: geom.Geometry) -> geom.Geometry
g1.union(g2: geom.Geometry) -> geom.Geometry
g1.difference(g2: geom.Geometry) -> geom.Geometry
g1.symDifference(g2: geom.Geometry) -> geom.Geometry

-- Aggregate union
geom.unaryUnion(geoms: List<geom.Geometry>) -> geom.Geometry

-- Buffer
g.buffer(distance: float) -> geom.Polygon
g.buffer(distance: float, segments: int) -> geom.Polygon
g.buffer(distance: float, options: BufferOptions) -> geom.Polygon

type BufferOptions = struct {
    quadrantSegments: int,
    endCapStyle: string,    -- "round", "flat", "square"
    joinStyle: string,      -- "round", "mitre", "bevel"
    mitreLimit: float,
    singleSided: bool
}

-- Convex hull
g.convexHull() -> geom.Polygon

-- Centroid and center
g.centroid() -> geom.Point
g.pointOnSurface() -> geom.Point  -- guaranteed to be inside

-- Simplification
g.simplify(tolerance: float) -> geom.Geometry
g.simplifyPreserveTopology(tolerance: float) -> geom.Geometry
g.douglasPeucker(tolerance: float) -> geom.Geometry
g.visvalingamWhyatt(threshold: float) -> geom.Geometry

-- Densification
g.densify(maxSegmentLength: float) -> geom.Geometry

-- Validity repair
g.makeValid() -> geom.Geometry

-- Snapping
g.snap(reference: geom.Geometry, tolerance: float) -> geom.Geometry

-- Line operations
ls.reverse() -> geom.LineString
ls.merge(other: geom.LineString) -> geom.LineString
ls.substring(startFraction: float, endFraction: float) -> geom.LineString
ls.interpolate(fraction: float) -> geom.Point
ls.interpolatePoint(distance: float) -> geom.Point
ls.locatePoint(point: geom.Point) -> float  -- returns fraction along line

-- Polygon operations
poly.orientExteriorCCW() -> geom.Polygon
poly.removeHoles() -> geom.Polygon
```

#### 8.2.6 Affine Transformations

```geoflow
-- Transform operations (all return new geometry)
g.translate(dx: float, dy: float) -> geom.Geometry
g.translate(dx: float, dy: float, dz: float) -> geom.Geometry
g.scale(factor: float) -> geom.Geometry
g.scale(fx: float, fy: float) -> geom.Geometry
g.scale(fx: float, fy: float, origin: geom.Point) -> geom.Geometry
g.rotate(angle: float) -> geom.Geometry  -- radians, about origin
g.rotate(angle: float, origin: geom.Point) -> geom.Geometry
g.rotateZ(angle: float) -> geom.Geometry  -- for 3D
g.reflect(axis: string) -> geom.Geometry  -- "x", "y", or "xy"
g.shear(shx: float, shy: float) -> geom.Geometry
g.affine(matrix: Matrix<float, 3, 3>) -> geom.Geometry
g.affine3D(matrix: Matrix<float, 4, 4>) -> geom.Geometry
```

### 8.3 Geography (`std.geo.geog`)

Geography types mirror geometry but use geodetic calculations on a spheroid (default WGS84).

#### 8.3.1 Construction

```geoflow
import std.geo.geog

-- From lat/lon (note: lon, lat order to match GeoJSON/WKT convention)
geog.Point(lon: float, lat: float) -> geog.Point
geog.PointZ(lon: float, lat: float, altitude: float) -> geog.PointZ

-- Or explicit named parameters
geog.Point(latitude: float, longitude: float) -> geog.Point

-- From geometry with CRS
geog.fromGeometry(g: geom.Geometry, crs: CRS) -> geog.Geography

-- From WKT (assumes EPSG:4326)
geog.fromWKT(wkt: string) -> Result<geog.Geography, string>

-- LineString, Polygon, etc. similar to geom
geog.LineString(points: List<geog.Point>) -> geog.LineString
geog.Polygon(rings: List<List<geog.Point>>) -> geog.Polygon
```

#### 8.3.2 Properties

```geoflow
-- Point properties
p.latitude -> float
p.longitude -> float
p.lon -> float    -- alias
p.lat -> float    -- alias
p.altitude -> Option<float>
p.elevation -> Option<float>  -- alias
```

#### 8.3.3 Measurements (Geodetic)

```geoflow
-- Distance in meters (geodesic on WGS84 spheroid)
p1.distanceTo(p2: geog.Point) -> float
g1.distance(g2: geog.Geography) -> float

-- Length in meters
ls.length() -> float

-- Area in square meters
poly.area() -> float

-- Azimuth (bearing) in radians
p1.azimuthTo(p2: geog.Point) -> float

-- Perimeter in meters
poly.perimeter() -> float
```

#### 8.3.4 Geodetic Operations

```geoflow
-- Project a point given distance and bearing
p.project(distance: float, azimuth: float) -> geog.Point

-- Interpolate along great circle
p1.interpolate(p2: geog.Point, fraction: float) -> geog.Point

-- Densify along geodesic (add points along great circles)
g.densify(maxSegmentLength: float) -> geog.Geography  -- length in meters

-- Buffer (geodetic, returns geography)
g.buffer(distanceMeters: float) -> geog.Polygon

-- Spatial relationships (use spheroid)
g1.intersects(g2: geog.Geography) -> bool
g1.contains(g2: geog.Geography) -> bool
g1.within(g2: geog.Geography) -> bool
-- etc.

-- Convert back to geometry
g.toGeometry(crs: CRS) -> geom.Geometry
```

### 8.4 Coordinate Reference Systems (`std.geo.crs`)

```geoflow
import std.geo.crs

-- Built-in CRS constants
CRS.WGS84           -- EPSG:4326, geographic
CRS.WebMercator     -- EPSG:3857, projected (web maps)
CRS.NAD83           -- EPSG:4269, geographic
CRS.NAD27           -- EPSG:4267, geographic
CRS.ETRS89          -- EPSG:4258, European
CRS.PseudoMercator  -- alias for WebMercator

-- UTM zones
CRS.UTM(zone: int, north: bool = true) -> CRS
-- Examples:
CRS.UTM(10, true)   -- UTM Zone 10N (EPSG:32610)
CRS.UTM(18, false)  -- UTM Zone 18S (EPSG:32718)

-- State Plane (US)
CRS.StatePlane(zone: string) -> CRS
CRS.StatePlane("CA_I")  -- California Zone 1

-- From EPSG code
CRS.fromEPSG(code: int) -> Result<CRS, string>

-- From Proj4 string
CRS.fromProj4(proj4: string) -> Result<CRS, string>

-- From WKT
CRS.fromWKT(wkt: string) -> Result<CRS, string>

-- CRS properties
crs.epsg() -> Option<int>
crs.name() -> string
crs.isGeographic() -> bool
crs.isProjected() -> bool
crs.units() -> string         -- "degrees", "meters", "feet", etc.
crs.toProj4() -> string
crs.toWKT() -> string
crs.bounds() -> Option<Tuple<float, float, float, float>>

-- Datum
crs.datum() -> Datum
crs.ellipsoid() -> Ellipsoid

type Datum = struct {
    name: string,
    ellipsoid: Ellipsoid,
    toWGS84: List<float>      -- 7 Bursa-Wolf parameters
}

type Ellipsoid = struct {
    name: string,
    semiMajorAxis: float,
    inverseFlattening: float
}

-- Built-in ellipsoids
Ellipsoid.WGS84
Ellipsoid.GRS80
Ellipsoid.Clarke1866
Ellipsoid.Bessel1841
```

### 8.5 Transformations (`std.geo.transform`)

```geoflow
import std.geo.transform as t

-- Transform geometry between CRS
t.transform(g: geom.Geometry, from: CRS, to: CRS) -> geom.Geometry

-- Create reusable transformer
let transformer = t.Transformer(from: CRS, to: CRS)
transformer.transform(g: geom.Geometry) -> geom.Geometry
transformer.transformPoint(x: float, y: float) -> Tuple<float, float>
transformer.transformPoints(coords: List<Tuple<float, float>>) -> List<Tuple<float, float>>
transformer.inverse() -> Transformer

-- Convenience functions
t.toWebMercator(g: geom.Geometry, fromCRS: CRS = CRS.WGS84) -> geom.Geometry
t.fromWebMercator(g: geom.Geometry, toCRS: CRS = CRS.WGS84) -> geom.Geometry
t.toUTM(g: geom.Geometry, fromCRS: CRS = CRS.WGS84) -> Tuple<geom.Geometry, CRS>  -- auto-detects zone

-- Geography <-> Geometry
t.geographyToGeometry(g: geog.Geography, targetCRS: CRS) -> geom.Geometry
t.geometryToGeography(g: geom.Geometry, sourceCRS: CRS) -> geog.Geography

-- Pipeline transformations (complex multi-step)
let pipeline = t.Pipeline([
    t.Step.transform(CRS.NAD27, CRS.NAD83),
    t.Step.transform(CRS.NAD83, CRS.WGS84),
    t.Step.project(CRS.WebMercator)
])
pipeline.transform(g: geom.Geometry) -> geom.Geometry
```

### 8.6 Spatial Indexing (`std.geo.index`)

```geoflow
import std.geo.index

-- R-Tree
let rtree = index.RTree<T>()
rtree.insert(geometry: geom.Geometry, item: T)
rtree.insertAll(items: List<Tuple<geom.Geometry, T>>)
rtree.remove(geometry: geom.Geometry, item: T) -> bool
rtree.query(bounds: geom.Polygon) -> List<T>
rtree.queryPoint(point: geom.Point) -> List<T>
rtree.nearest(point: geom.Point, n: int = 1) -> List<T>
rtree.nearestWithDistance(point: geom.Point, n: int = 1) -> List<Tuple<T, float>>
rtree.size() -> int
rtree.clear()
rtree.bounds() -> Option<geom.Polygon>

-- Quadtree
let qtree = index.QuadTree<T>(bounds: geom.Polygon, maxDepth: int = 10)
qtree.insert(point: geom.Point, item: T)
qtree.query(bounds: geom.Polygon) -> List<T>
qtree.queryRadius(center: geom.Point, radius: float) -> List<T>

-- STRTree (static, for query-heavy workloads)
let strtree = index.STRTree<T>(items: List<Tuple<geom.Geometry, T>>)
strtree.query(bounds: geom.Polygon) -> List<T>
strtree.nearest(point: geom.Point, n: int = 1) -> List<T>

-- Grid index (for raster-like point data)
let grid = index.GridIndex<T>(bounds: geom.Polygon, cellSize: float)
grid.insert(point: geom.Point, item: T)
grid.query(bounds: geom.Polygon) -> List<T>
grid.getCell(point: geom.Point) -> List<T>
```

### 8.7 I/O Formats (`std.geo.io`)

```geoflow
import std.geo.io

-- WKT
io.toWKT(g: geom.Geometry) -> string
io.toWKT(g: geom.Geometry, precision: int) -> string
io.fromWKT(wkt: string) -> Result<geom.Geometry, string>
io.fromWKT(wkt: string, srid: int) -> Result<geom.Geometry, string>

-- WKB
io.toWKB(g: geom.Geometry) -> bytes
io.toWKB(g: geom.Geometry, endian: string) -> bytes  -- "big" or "little"
io.toEWKB(g: geom.Geometry) -> bytes  -- extended WKB with SRID
io.fromWKB(wkb: bytes) -> Result<geom.Geometry, string>
io.fromEWKB(ewkb: bytes) -> Result<geom.Geometry, string>

-- GeoJSON
io.toGeoJSON(g: geom.Geometry) -> string
io.toGeoJSON(g: geom.Geometry, properties: Map<string, auto>) -> string  -- Feature
io.toGeoJSON(features: List<Tuple<geom.Geometry, Map<string, auto>>>) -> string  -- FeatureCollection
io.fromGeoJSON(json: string) -> Result<geom.Geometry, string>
io.featuresFromGeoJSON(json: string) -> Result<List<Feature>, string>

type Feature = struct {
    geometry: geom.Geometry,
    properties: Map<string, auto>,
    id: Option<string>
}

-- KML
io.toKML(g: geom.Geometry) -> string
io.fromKML(kml: string) -> Result<List<geom.Geometry>, string>

-- GPX
io.fromGPX(gpx: string) -> Result<GPXData, string>
type GPXData = struct {
    tracks: List<geom.LineString>,
    routes: List<geom.LineString>,
    waypoints: List<geom.Point>,
    metadata: Map<string, string>
}

-- Shapefile (requires file I/O)
io.readShapefile(path: string) -> Result<ShapefileData, string>
io.writeShapefile(path: string, data: ShapefileData) -> Result<unit, string>
type ShapefileData = struct {
    geometries: List<geom.Geometry>,
    attributes: DataFrame,
    crs: Option<CRS>
}

-- GeoPackage
io.readGeoPackage(path: string) -> Result<GeoPackage, string>
io.writeGeoPackage(path: string, layers: Map<string, GeoPackageLayer>) -> Result<unit, string>

-- CSV with geometry
io.readGeoCSV(path: string, geomColumn: string, geomFormat: string = "wkt") -> Result<DataFrame, string>
io.writeGeoCSV(path: string, df: DataFrame, geomColumn: string) -> Result<unit, string>
```

### 8.8 Raster Operations (`std.geo.raster`)

```geoflow
import std.geo.raster

-- Raster type
Raster<T>

-- Construction
raster.fromFile(path: string) -> Result<Raster<float>, string>
raster.fromArray(data: Array<T>, bounds: geom.Polygon, crs: CRS) -> Raster<T>
raster.empty<T>(width: int, height: int, bounds: geom.Polygon, crs: CRS, nodata: T) -> Raster<T>

-- Properties
r.width() -> int
r.height() -> int
r.shape() -> Tuple<int, int>
r.bounds() -> geom.Polygon
r.crs() -> CRS
r.cellSize() -> Tuple<float, float>
r.nodata() -> Option<T>
r.dtype() -> string

-- Access
r[row, col] -> T
r.getValue(point: geom.Point) -> Option<T>  -- sample at point
r.toArray() -> Array<T>

-- Sampling
r.sample(point: geom.Point, method: string = "nearest") -> Option<T>
-- methods: "nearest", "bilinear", "cubic"
r.sampleLine(line: geom.LineString, interval: float) -> List<T>

-- Raster algebra (returns new Raster)
r1 + r2
r1 - r2
r1 * r2
r1 / r2
r * scalar
r.apply(f: Fn<(T) -> T>) -> Raster<T>
r.combine(other: Raster<U>, f: Fn<(T, U) -> V>) -> Raster<V>

-- Resampling
r.resample(targetCellSize: float, method: string = "bilinear") -> Raster<T>
r.resample(targetWidth: int, targetHeight: int, method: string = "bilinear") -> Raster<T>
r.reproject(targetCRS: CRS, method: string = "bilinear") -> Raster<T>

-- Analysis
r.slope() -> Raster<float>       -- terrain slope in degrees
r.aspect() -> Raster<float>      -- terrain aspect in degrees
r.hillshade(azimuth: float = 315, altitude: float = 45) -> Raster<float>
r.contours(interval: float) -> List<geom.LineString>
r.focal(kernel: Matrix<float>, f: Fn<(List<T>) -> T>) -> Raster<T>
r.focalMean(size: int) -> Raster<float>
r.focalMax(size: int) -> Raster<T>
r.focalMin(size: int) -> Raster<T>

-- Zonal statistics
r.zonalStats(zones: Raster<int>) -> Map<int, ZonalStats>
r.zonalStats(zones: List<geom.Polygon>) -> List<ZonalStats>
type ZonalStats = struct {
    count: int,
    sum: float,
    mean: float,
    min: float,
    max: float,
    std: float
}

-- Vectorization
r.toPolygons() -> List<Tuple<geom.Polygon, T>>
r.toPoints() -> List<Tuple<geom.Point, T>>

-- I/O
r.toFile(path: string, format: string = "GeoTIFF") -> Result<unit, string>
```

### 8.9 Spatial Analysis (`std.geo.analysis`)

```geoflow
import std.geo.analysis as analysis

-- Clustering
analysis.dbscan(points: List<geom.Point>, eps: float, minPts: int) -> List<List<geom.Point>>
analysis.kmeans(points: List<geom.Point>, k: int) -> List<List<geom.Point>>
analysis.hdbscan(points: List<geom.Point>, minClusterSize: int) -> List<List<geom.Point>>

-- Nearest neighbor
analysis.nearestNeighbor(points: List<geom.Point>) -> NearestNeighborResult
type NearestNeighborResult = struct {
    observedMeanDistance: float,
    expectedMeanDistance: float,
    ratio: float,
    zScore: float,
    pValue: float
}

-- Spatial autocorrelation
analysis.moransI(values: List<float>, locations: List<geom.Point>) -> MoransResult
analysis.getisOrdG(values: List<float>, locations: List<geom.Point>) -> List<float>

-- Point pattern analysis
analysis.kernelDensity(points: List<geom.Point>, bounds: geom.Polygon, cellSize: float, bandwidth: float) -> Raster<float>
analysis.quadratCount(points: List<geom.Point>, bounds: geom.Polygon, nx: int, ny: int) -> Matrix<int>

-- Network analysis
analysis.shortestPath(network: Network, start: geom.Point, end: geom.Point) -> Path
analysis.serviceArea(network: Network, origin: geom.Point, maxCost: float) -> geom.Polygon
analysis.closestFacility(network: Network, origins: List<geom.Point>, facilities: List<geom.Point>) -> List<Path>

-- Voronoi and Delaunay
analysis.voronoi(points: List<geom.Point>) -> List<geom.Polygon>
analysis.voronoi(points: List<geom.Point>, bounds: geom.Polygon) -> List<geom.Polygon>
analysis.delaunay(points: List<geom.Point>) -> List<geom.Polygon>  -- triangles

-- Alpha shapes
analysis.alphaShape(points: List<geom.Point>, alpha: float) -> geom.Polygon
analysis.concaveHull(points: List<geom.Point>, concavity: float = 0.5) -> geom.Polygon

-- Interpolation
analysis.idw(points: List<Tuple<geom.Point, float>>, query: geom.Point, power: float = 2) -> float
analysis.idwGrid(points: List<Tuple<geom.Point, float>>, bounds: geom.Polygon, cellSize: float, power: float = 2) -> Raster<float>
analysis.kriging(points: List<Tuple<geom.Point, float>>, bounds: geom.Polygon, cellSize: float) -> Raster<float>
analysis.tin(points: List<Tuple<geom.Point, float>>) -> TIN

-- Viewshed
analysis.viewshed(dem: Raster<float>, observerPoint: geom.PointZ, maxDistance: float) -> Raster<bool>

-- Least cost path
analysis.leastCostPath(costSurface: Raster<float>, start: geom.Point, end: geom.Point) -> geom.LineString
```

### 8.10 H3 and S2 Indexing (`std.geo.h3`, `std.geo.s2`)

```geoflow
import std.geo.h3

-- H3 hexagonal indexing
h3.latLonToCell(lat: float, lon: float, resolution: int) -> H3Index
h3.pointToCell(p: geog.Point, resolution: int) -> H3Index
h3.cellToPoint(cell: H3Index) -> geog.Point
h3.cellToBoundary(cell: H3Index) -> geog.Polygon
h3.cellToChildren(cell: H3Index, childResolution: int) -> List<H3Index>
h3.cellToParent(cell: H3Index, parentResolution: int) -> H3Index
h3.gridDisk(cell: H3Index, k: int) -> List<H3Index>
h3.gridRing(cell: H3Index, k: int) -> List<H3Index>
h3.gridPathCells(start: H3Index, end: H3Index) -> List<H3Index>
h3.cellArea(cell: H3Index) -> float  -- in square meters
h3.gridDistance(a: H3Index, b: H3Index) -> int
h3.polyfill(polygon: geog.Polygon, resolution: int) -> List<H3Index>
h3.cellsToMultiPolygon(cells: List<H3Index>) -> geog.MultiPolygon
h3.compactCells(cells: List<H3Index>) -> List<H3Index>
h3.uncompactCells(cells: List<H3Index>, resolution: int) -> List<H3Index>

import std.geo.s2

-- S2 spherical geometry
s2.latLonToCell(lat: float, lon: float, level: int) -> S2CellId
s2.pointToCell(p: geog.Point, level: int) -> S2CellId
s2.cellToPoint(cell: S2CellId) -> geog.Point
s2.cellToBoundary(cell: S2CellId) -> geog.Polygon
s2.cellToChildren(cell: S2CellId) -> List<S2CellId>
s2.cellToParent(cell: S2CellId) -> S2CellId
s2.cellNeighbors(cell: S2CellId) -> List<S2CellId>
s2.covering(polygon: geog.Polygon, maxCells: int = 8, maxLevel: int = 30) -> List<S2CellId>
s2.interiorCovering(polygon: geog.Polygon, maxCells: int = 8, maxLevel: int = 30) -> List<S2CellId>
```

---

## 9. Implementation Plan

### Phase 1: Core Language (Weeks 1-4)

**Week 1: Lexer and Tokens**
- Implement token types for all keywords, operators, literals
- Build lexer with position tracking (line, column)
- Support WKT literal syntax (`#POINT(...)#`)
- Support symbolic expression syntax (`$x^2 + 1$`)
- Handle comments (single-line `--`, multi-line `{- -}`)
- Unit tests for lexer

**Week 2: Parser and AST**
- Define AST node types for all language constructs
- Implement recursive descent parser
- Handle operator precedence (composition, arithmetic, comparison, logical)
- Parse function definitions, let bindings, control flow
- Parse type annotations
- Unit tests for parser

**Week 3: Type System**
- Implement type representation (primitives, generics, functions)
- Build type checker
- Implement type inference for `auto`
- Handle generic type instantiation
- Type checking for all expressions and statements
- Unit tests for type checker

**Week 4: Basic Evaluation**
- Implement evaluation environment (scopes, bindings)
- Build tree-walking interpreter
- Implement primitive operations (arithmetic, comparison, logical)
- Implement function calls and closures
- Implement control flow (if, match, for, while)
- REPL loop
- Integration tests

### Phase 2: Composition and Core Libraries (Weeks 5-8)

**Week 5: Composition System**
- Implement pipeline operator (`|>`)
- Implement composition operator (`.`)
- Implement juxtaposition composition (space-separated)
- Implement automatic currying and partial application
- Tests for composition semantics

**Week 6: Collections and Data Structures**
- Implement List, Map, Set, Tuple
- Implement Array, Vector, Matrix types
- Implement all collection operations
- Bridge to Go slices/maps for efficiency

**Week 7: DataFrame**
- Implement Series type
- Implement DataFrame type
- CSV/JSON parsing
- GroupBy operations
- Joins
- Performance optimization for large datasets

**Week 8: String and I/O**
- Implement string library
- Implement file I/O
- Implement logging library
- Implement assertion library
- Implement time library

### Phase 3: Math Libraries (Weeks 9-12)

**Week 9: Basic Math and Statistics**
- Implement `std.math` (basic functions, trig, etc.)
- Implement `std.math.stats` (descriptive statistics)
- Implement distributions
- Hypothesis testing

**Week 10: Linear Algebra**
- Implement `std.math.linalg`
- Matrix operations and decompositions
- Bind to BLAS/LAPACK or use pure Go implementations
- Optimize for performance

**Week 11: Symbolic Math**
- Implement `Expr` type for symbolic expressions
- Expression building and simplification
- Differentiation (automatic differentiation or symbolic rules)
- Basic integration
- Equation solving

**Week 12: Numerical Methods**
- Root finding algorithms
- Optimization
- Numerical integration
- ODE solvers
- Interpolation and curve fitting

### Phase 4: Geospatial Core (Weeks 13-18)

**Week 13: Geometry Types**
- Implement all OGC geometry types (Point, LineString, Polygon, Multi*, Collection)
- Implement 2D, Z, M, ZM variants
- WKT/WKB parsing and serialization
- Geometry validation

**Week 14: Geometry Operations**
- Bind to GEOS library for spatial operations
- Implement spatial relationships (intersects, contains, etc.)
- Implement set operations (union, intersection, difference)
- Buffer, convex hull, simplification

**Week 15: Geography Types**
- Implement geography types mirroring geometry
- Geodetic distance calculations
- Great circle operations
- Area and length on spheroid (Karney's algorithm or GeographicLib)

**Week 16: CRS and Transformations**
- Implement CRS type
- Bind to PROJ library for transformations
- Implement common CRS definitions
- Transform between geometry and geography

**Week 17: Spatial I/O**
- GeoJSON
- KML/GPX
- Shapefile (bind to GDAL or pure Go)
- GeoPackage
- GeoCSV

**Week 18: Spatial Indexing**
- R-Tree implementation
- STRTree for static data
- QuadTree for points
- Grid index

### Phase 5: Advanced Geo and Polish (Weeks 19-22)

**Week 19: Raster Support**
- Raster type
- GeoTIFF reading/writing
- Raster algebra
- Resampling and reprojection

**Week 20: Spatial Analysis**
- Clustering algorithms
- Interpolation (IDW, Kriging)
- Voronoi/Delaunay
- Network analysis basics

**Week 21: H3 and S2**
- Bind to H3 library
- Bind to S2 library
- Integration with geography types

**Week 22: Optimization and Documentation**
- Performance profiling and optimization
- Memory optimization for large datasets
- Complete API documentation
- Tutorial and examples
- Integration test suite

### Phase 6: ETL Integration (Weeks 23-24)

**Week 23: Streaming and Pipelines**
- Implement lazy Stream type
- Streaming file readers
- Parallel processing with goroutines
- Backpressure handling

**Week 24: Event Processing**
- Event source connectors (Kafka, etc.)
- Event sink connectors
- Windowing operations
- Aggregation over time

---

## 10. File Structure

```
geoflow/
├── go.mod
├── go.sum
├── main.go                     # Entry point, CLI
├── cmd/
│   └── geoflow/
│       └── main.go             # CLI application
├── internal/
│   ├── token/
│   │   └── token.go            # Token types
│   ├── lexer/
│   │   ├── lexer.go            # Lexer implementation
│   │   └── lexer_test.go
│   ├── ast/
│   │   ├── ast.go              # AST node definitions
│   │   └── ast_string.go       # Pretty printing
│   ├── parser/
│   │   ├── parser.go           # Parser implementation
│   │   ├── parser_test.go
│   │   └── precedence.go       # Operator precedence
│   ├── types/
│   │   ├── types.go            # Type representations
│   │   ├── checker.go          # Type checker
│   │   ├── inference.go        # Type inference
│   │   └── checker_test.go
│   ├── object/
│   │   ├── object.go           # Runtime value types
│   │   ├── environment.go      # Scope/binding environment
│   │   └── builtins.go         # Built-in functions
│   ├── eval/
│   │   ├── eval.go             # Evaluator
│   │   ├── eval_test.go
│   │   └── composition.go      # Composition handling
│   └── repl/
│       └── repl.go             # REPL implementation
├── pkg/
│   ├── stdlib/
│   │   ├── core/
│   │   │   └── core.go
│   │   ├── string/
│   │   │   └── string.go
│   │   ├── collections/
│   │   │   ├── list.go
│   │   │   ├── map.go
│   │   │   ├── set.go
│   │   │   └── stream.go
│   │   ├── data/
│   │   │   ├── array.go
│   │   │   ├── vector.go
│   │   │   ├── matrix.go
│   │   │   ├── series.go
│   │   │   └── dataframe.go
│   │   ├── math/
│   │   │   ├── basic.go
│   │   │   ├── stats.go
│   │   │   ├── symbolic/
│   │   │   │   ├── expr.go
│   │   │   │   ├── diff.go
│   │   │   │   ├── integrate.go
│   │   │   │   └── solve.go
│   │   │   ├── linalg/
│   │   │   │   ├── linalg.go
│   │   │   │   └── decomp.go
│   │   │   ├── numeric/
│   │   │   │   ├── roots.go
│   │   │   │   ├── optimize.go
│   │   │   │   ├── integrate.go
│   │   │   │   └── ode.go
│   │   │   └── complex.go
│   │   ├── io/
│   │   │   └── io.go
│   │   ├── time/
│   │   │   └── time.go
│   │   ├── log/
│   │   │   └── log.go
│   │   └── assert/
│   │       └── assert.go
│   └── geo/
│       ├── geom/
│       │   ├── point.go
│       │   ├── linestring.go
│       │   ├── polygon.go
│       │   ├── multi.go
│       │   ├── collection.go
│       │   ├── operations.go   # Spatial operations
│       │   └── relations.go    # Spatial relationships
│       ├── geog/
│       │   ├── point.go
│       │   ├── linestring.go
│       │   ├── polygon.go
│       │   ├── geodetic.go     # Geodetic calculations
│       │   └── operations.go
│       ├── crs/
│       │   ├── crs.go
│       │   ├── builtin.go      # Built-in CRS definitions
│       │   └── transform.go
│       ├── io/
│       │   ├── wkt.go
│       │   ├── wkb.go
│       │   ├── geojson.go
│       │   ├── shapefile.go
│       │   └── geopackage.go
│       ├── index/
│       │   ├── rtree.go
│       │   ├── strtree.go
│       │   └── quadtree.go
│       ├── raster/
│       │   ├── raster.go
│       │   ├── algebra.go
│       │   └── io.go
│       ├── analysis/
│       │   ├── cluster.go
│       │   ├── interpolate.go
│       │   ├── voronoi.go
│       │   └── network.go
│       ├── h3/
│       │   └── h3.go           # H3 bindings
│       └── s2/
│           └── s2.go           # S2 bindings
├── examples/
│   ├── hello.gf
│   ├── pipeline.gf
│   ├── geospatial.gf
│   ├── math.gf
│   └── etl.gf
├── docs/
│   ├── language-reference.md
│   ├── stdlib-reference.md
│   ├── geo-reference.md
│   └── tutorials/
│       ├── getting-started.md
│       ├── pipelines.md
│       └── geospatial.md
└── test/
    ├── integration/
    └── fixtures/
```

---

## Appendix A: Example Programs

### A.1 Hello World

```geoflow
-- hello.gf
"Hello, GeoFlow!" |> println
```

### A.2 Basic Pipeline

```geoflow
-- pipeline.gf
let numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

let result = numbers
    |> filter(\x -> x % 2 == 0)
    |> map(\x -> x * x)
    |> sum

println("Sum of squares of evens: {result}")  -- 220
```

### A.3 Symbolic Math

```geoflow
-- math.gf
import std.math.symbolic as sym

let x = sym.var("x")
let f = x^3 - 6*x^2 + 11*x - 6

let df = sym.diff(f, x)
println("f(x) = {f}")
println("f'(x) = {df}")

let roots = sym.solve(f, x)
println("Roots: {roots}")  -- [1, 2, 3]

let integral = sym.integrate(f, x, 0, 1)
println("Integral from 0 to 1: {integral.realize({})}")
```

### A.4 Geospatial Processing

```geoflow
-- geo.gf
import std.geo.geom
import std.geo.geog
import std.geo.io
import std.geo.crs

-- Create points (lon, lat order)
let sf = geog.Point(-122.4194, 37.7749)
let la = geog.Point(-118.2437, 34.0522)

-- Geodetic distance
let dist = sf.distanceTo(la)
println("SF to LA: {dist / 1000} km")  -- ~559 km

-- Load polygons
let regions = io.fromGeoJSON(io.readFile("regions.geojson")?)

-- Pipeline: read points, filter valid, buffer, intersect
let result = "points.csv"
    |> io.readGeoCSV("geometry", "wkt")?
    |> df.filterByColumn("geometry", \g -> g.isValid())
    |> df.apply("geometry", \g -> g.buffer(100))
    |> df.filter(\row -> row["geometry"].intersects(regions[0]))

println("Found {result.rowCount()} points in region")

-- Transform to UTM for accurate area calculation
let transformer = crs.Transformer(CRS.WGS84, CRS.UTM(10, true))
let localRegion = transformer.transform(regions[0])
println("Region area: {localRegion.area()} sq meters")
```

### A.5 ETL Pipeline

```geoflow
-- etl.gf
import std.geo.io
import std.geo.geom
import std.geo.geog
import std.geo.analysis

-- Define processing pipeline
let processEvents = fn(events: Stream<Event>) -> Stream<ProcessedEvent> {
    events
        |> filter(\e -> e.location.isValid())
        |> map(\e -> {
            ...e,
            localTime: e.timestamp.inTimezone(e.timezone),
            region: findRegion(e.location)
        })
        |> window(Duration.minutes(5))
        |> aggregate(\w -> {
            count: w.count(),
            centroid: analysis.centroid(w.map(\e -> e.location)),
            avgValue: w.map(\e -> e.value).mean()
        })
}

-- Run pipeline
let source = EventSource.kafka("events-topic")
let sink = EventSink.postgres("processed_events")

source
    |> processEvents
    |> sink.write
```

---

## Appendix B: Grammar (EBNF)

```ebnf
program        = { statement } ;
statement      = letStmt | fnStmt | typeStmt | exprStmt ;

letStmt        = "let" [ "mut" ] IDENT [ ":" type ] "=" expr ;
fnStmt         = "fn" IDENT [ typeParams ] "(" [ params ] ")" [ "->" type ] block ;
typeStmt       = "type" IDENT [ typeParams ] "=" typeExpr ;

typeParams     = "<" IDENT { "," IDENT } ">" ;
params         = param { "," param } ;
param          = IDENT ":" type [ "=" expr ] ;

type           = simpleType | genericType | fnType | "auto" ;
simpleType     = IDENT ;
genericType    = IDENT "<" type { "," type } ">" ;
fnType         = "Fn" "<" "(" [ type { "," type } ] ")" "->" type ">" ;

expr           = pipeline ;
pipeline       = composition { "|>" composition } ;
composition    = logicalOr { logicalOr } ;  -- juxtaposition
logicalOr      = logicalAnd { "||" logicalAnd } ;
logicalAnd     = equality { "&&" equality } ;
equality       = comparison { ( "==" | "!=" ) comparison } ;
comparison     = additive { ( "<" | ">" | "<=" | ">=" ) additive } ;
additive       = multiplicative { ( "+" | "-" ) multiplicative } ;
multiplicative = power { ( "*" | "/" | "%" ) power } ;
power          = unary { "^" unary } ;
unary          = ( "!" | "-" ) unary | call ;
call           = primary { "(" [ args ] ")" | "." IDENT | "[" expr "]" } ;
args           = expr { "," expr } ;

primary        = IDENT | INT | FLOAT | STRING | "true" | "false" | "nil"
               | wktLiteral | symbolicLiteral
               | listLiteral | mapLiteral | tupleLiteral
               | lambda | ifExpr | matchExpr | forExpr | block
               | "(" expr ")" ;

wktLiteral     = "#" WKT_CONTENT "#" ;
symbolicLiteral = "$" SYMBOLIC_EXPR "$" ;
listLiteral    = "[" [ expr { "," expr } ] "]" ;
mapLiteral     = "{" [ mapEntry { "," mapEntry } ] "}" ;
mapEntry       = ( STRING | IDENT ) ":" expr ;
tupleLiteral   = "(" expr "," expr { "," expr } ")" ;

lambda         = "\" params "->" expr
               | "fn" "(" [ params ] ")" [ "->" type ] block ;

ifExpr         = "if" expr "then" expr [ "else" expr ] ;
matchExpr      = "match" expr "{" { matchArm } "}" ;
matchArm       = pattern [ "if" expr ] "=>" expr "," ;
pattern        = IDENT | literal | constructor | "_" ;

forExpr        = "for" pattern "in" expr block ;
block          = "{" { statement } [ expr ] "}" ;
```

---

**End of Specification**

This document should provide Claude Code with everything needed to implement GeoFlow. Start with Phase 1 (lexer, parser, type system, basic evaluation) and iterate from there.
