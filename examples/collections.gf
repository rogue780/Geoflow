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

-- ── Advanced List Methods ──
println("\n-- Advanced List Methods --")

-- filterMap: apply function, keep non-None results
let fm = [1, 2, 3, 4, 5].filterMap(\x -> if x % 2 == 0 then Some(x * 10) else None)
println("filterMap (even*10):", fm)

-- distinctBy: deduplicate by key function
let db = [1, -1, 2, -2, 3].distinctBy(\x -> if x < 0 then -x else x)
println("distinctBy abs:", db)

-- sortWith: custom comparator
let sw = [3, 1, 4, 1, 5].sortWith(\a, b -> if a < b then -1 else if a > b then 1 else 0)
println("sortWith:", sw)

-- minBy / maxBy
let words = ["hello", "hi", "hey", "howdy"]
println("minBy length:", words.minBy(\w -> len(w)))
println("maxBy length:", words.maxBy(\w -> len(w)))

-- zipWith: zip + combine
let zw = [1, 2, 3].zipWith([10, 20, 30], \a, b -> a + b)
println("zipWith (+):", zw)

-- unzip: split list of tuples
let uz = [(1, "a"), (2, "b"), (3, "c")].unzip()
println("unzip:", uz)

-- intersperse: insert separator
let is = [1, 2, 3].intersperse(0)
println("intersperse 0:", is)

-- interleave: alternate elements
let il = [1, 3, 5].interleave([2, 4, 6])
println("interleave:", il)

-- windowed: sliding windows
let win = [1, 2, 3, 4, 5].windowed(3)
println("windowed(3):", win)
let win2 = [1, 2, 3, 4, 5].windowed(3, 2)
println("windowed(3, 2):", win2)

-- foldRight: right-to-left fold
let fr = [1, 2, 3].foldRight([], \elem, acc -> acc.append(elem))
println("foldRight:", fr)

-- scan: fold with intermediates
let sc = [1, 2, 3, 4].scan(0, \acc, x -> acc + x)
println("scan (+):", sc)

-- insert: insert at index
let ins = [1, 2, 4, 5].insert(2, 3)
println("insert(2, 3):", ins)

-- remove: remove at index
let rm = [1, 2, 99, 3, 4].remove(2)
println("remove(2):", rm)

-- ── Advanced Map Methods ──
println("\n-- Advanced Map Methods --")

let m1 = {"a": "1", "b": "2"}
let m2 = {"b": "20", "c": "3"}

-- putAll: merge all pairs
println("putAll:", m1.putAll(m2))

-- update: apply function to value at key
println("update a:", m1.update("a", \v -> v + "!"))
println("update missing:", m1.update("z", \v -> v + "!"))

-- updateOrInsert: update if exists, insert default if not
println("updateOrInsert a:", m1.updateOrInsert("a", "default", \v -> v + "!"))
println("updateOrInsert z:", m1.updateOrInsert("z", "default", \v -> v + "!"))

-- filterKeys: keep entries where predicate matches key
let m3 = {"apple": "1", "banana": "2", "avocado": "3"}
println("filterKeys (starts with a):", m3.filterKeys(\k -> k[0] == "a"))

-- filterValues: keep entries where predicate matches value
println("filterValues (> '1'):", m3.filterValues(\v -> v > "1"))

-- ── Advanced Set Methods ──
println("\n-- Advanced Set Methods --")

let s1 = collections.of(1, 2, 3)
let s2 = collections.of(4, 5, 6)
let s3 = collections.of(3, 4, 5)

-- isDisjoint: true if no shared elements
println("s1.isDisjoint(s2):", s1.isDisjoint(s2))
println("s1.isDisjoint(s3):", s1.isDisjoint(s3))

-- symmetricDifference: elements in either but not both
println("s1.symmetricDifference(s3):", s1.symmetricDifference(s3))

-- toggle: remove if present, add if absent
println("s1.toggle(2):", s1.toggle(2))
println("s1.toggle(10):", s1.toggle(10))

-- map: transform elements
println("s1.map(*2):", s1.map(\x -> x * 2))

-- filter: keep matching elements
println("s1.filter(>1):", s1.filter(\x -> x > 1))

-- flatMap: map + union
println("s1.flatMap:", s1.flatMap(\x -> collections.of(x, x * 10)))
