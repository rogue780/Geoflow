-- data_structures.gf: Array, Series, and DataFrame examples
-- Demonstrates the std.data module for numerical and tabular data.

import std.data

{- ============================================================
   Part 1: Array Construction
   Arrays are dense, float64-backed N-dimensional structures.
   ============================================================ -}

println("== Array Construction ==")

-- zeros: create an array filled with zeros
let z = data.zeros([4])
println("zeros([4]):", z)

-- ones: create an array filled with ones
let o = data.ones([2, 3])
println("ones([2,3]):", o)

-- full: create an array filled with a specific value
let f = data.full([3], 7.0)
println("full([3], 7.0):", f)

-- eye: create an identity matrix
let id = data.eye(3)
println("eye(3):", id)

-- linspace: evenly spaced values from start to stop (inclusive)
let ls = data.linspace(0, 1, 5)
println("linspace(0, 1, 5):", ls)

-- arange: values from start to stop (exclusive) with a step
let ar = data.arange(0, 10, 2)
println("arange(0, 10, 2):", ar)

-- fromList: create a 1D array from a flat list
let a1 = data.fromList([10, 20, 30, 40])
println("fromList([10,20,30,40]):", a1)

-- fromNested: create a 2D array from nested lists
let a2 = data.fromNested([[1, 2, 3], [4, 5, 6]])
println("fromNested([[1,2,3],[4,5,6]]):", a2)

{- ============================================================
   Part 2: Array Methods
   ============================================================ -}

println("\n== Array Methods ==")

let arr = data.fromList([1, 2, 3, 4, 5, 6])

-- shape: returns the dimensions as a list
println("shape:", arr.shape())

-- ndim: number of dimensions
println("ndim:", arr.ndim())

-- size: total number of elements
println("size:", arr.size())

-- dtype: data type (always "float" for Array)
println("dtype:", arr.dtype())

-- Aggregate functions
println("sum:", arr.sum())
println("mean:", arr.mean())
println("min:", arr.min())
println("max:", arr.max())

-- flatten: collapse to 1D
let mat = data.fromNested([[1, 2], [3, 4]])
println("2D array:", mat)
println("flattened:", mat.flatten())

-- reshape: change the shape without changing data
let reshaped = arr.reshape([2, 3])
println("reshaped [2,3]:", reshaped)

-- transpose: swap rows and columns of a 2D array
let m = data.fromNested([[1, 2, 3], [4, 5, 6]])
println("original:", m)
println("transposed:", m.transpose())
println("T (shorthand):", m.T)

-- toList: convert array data back to a GeoFlow list
let lst = data.fromList([10, 20, 30]).toList()
println("toList:", lst)

-- dot product (1D)
let v1 = data.fromList([1, 2, 3])
let v2 = data.fromList([4, 5, 6])
println("dot product:", v1.dot(v2))

-- matrix multiply (2D)
let mA = data.fromNested([[1, 2], [3, 4]])
let mB = data.fromNested([[5, 6], [7, 8]])
println("matrix multiply:", mA.dot(mB))

{- ============================================================
   Part 3: Series Construction and Methods
   A Series is a labeled, 1D collection of heterogeneous data.
   ============================================================ -}

println("\n== Series Construction ==")

-- Basic series from a list
let s1 = data.Series([10, 20, 30, 40, 50])
println("Series:", s1)

-- Series with a name
let s2 = data.Series([100, 200, 300], "revenue")
println("Named series:", s2)

-- Series with an index and a name
let s3 = data.Series([72, 85, 91], ["alice", "bob", "charlie"], "score")
println("Indexed series:", s3)

-- SeriesFromMap: keys become index labels
let s4 = data.SeriesFromMap({"x": 1, "y": 2, "z": 3})
println("Series from map:", s4)

println("\n== Series Methods ==")

let scores = data.Series([85, 92, 78, 95, 88, 73, 99], "test_scores")

-- name: access the series name (property, not a method call)
println("name:", scores.name)

-- length: number of elements
println("length:", scores.length())

-- values: get the underlying data as a list
println("values:", scores.values())

-- Aggregate methods
println("sum:", scores.sum())
println("mean:", scores.mean())
println("min:", scores.min())
println("max:", scores.max())

-- head and tail: view the first/last N elements (default 5)
println("head(3):", scores.head(3))
println("tail(3):", scores.tail(3))

-- map: transform each element
let doubled = scores.map(\x -> x * 2)
println("mapped (*2):", doubled)

-- filter: keep elements matching a predicate
let high = scores.filter(\x -> x >= 90)
println("filtered (>=90):", high)

-- sort: sort elements in ascending order
let sorted = scores.sort()
println("sorted:", sorted)

-- unique: deduplicate values
let dupes = data.Series([1, 2, 2, 3, 3, 3])
println("unique:", dupes.unique())

-- count: number of non-nil values
println("count:", scores.count())

{- ============================================================
   Part 4: DataFrame Construction and Methods
   A DataFrame is a tabular structure of named Series columns.
   ============================================================ -}

println("\n== DataFrame Construction ==")

-- From a map of column name -> list
let df = data.DataFrame({
  "name": ["Alice", "Bob", "Charlie", "Diana"],
  "age": [30, 25, 35, 28],
  "score": [92, 87, 95, 91]
})
println("DataFrame:", df)

-- From rows (list of maps)
let df2 = data.DataFrameFromRows([
  {"city": "Portland", "pop": 650000},
  {"city": "Seattle", "pop": 750000},
  {"city": "Boise", "pop": 230000}
])
println("From rows:", df2)

println("\n== DataFrame Methods ==")

-- shape: returns (rows, columns) as a tuple
println("shape:", df.shape())

-- columns: list of column names
println("columns:", df.columns())

-- length: number of rows
println("length:", df.length())

-- head and tail
println("head(2):", df.head(2))
println("tail(2):", df.tail(2))

-- row: access a single row as a map
println("row(0):", df.row(0))
println("row(2):", df.row(2))

-- select: pick specific columns
let subset = df.select("name", "score")
println("select name, score:", subset)

-- drop: remove columns
let dropped = df.drop("age")
println("drop age:", dropped)

-- Access a column directly by name (returns a Series)
println("df.score column:", df.score)

-- filter: keep rows where predicate is true
-- The predicate receives each row as a map
let seniors = df.filter(\row -> row["age"] >= 30)
println("filter (age >= 30):", seniors)

-- sort: order rows by a column (ascending by default)
let byAge = df.sort("age")
println("sort by age:", byAge)

-- sort descending (pass false as second argument)
let byScoreDesc = df.sort("score", false)
println("sort by score desc:", byScoreDesc)

-- withColumn: add or replace a column using a row-level function
let withBonus = df.withColumn("bonus", \row -> row["score"] * 10)
println("withColumn bonus:", withBonus)

-- describe: get a summary string
println("describe:", df.describe())

{- ============================================================
   Part 5: Putting It All Together
   A small ETL-style pipeline using data structures.
   ============================================================ -}

println("\n== Pipeline Example ==")

-- Build a dataset of sensor readings
let sensors = data.DataFrame({
  "sensor_id": ["S1", "S2", "S3", "S4", "S5"],
  "temp_c": [22, 35, 18, 41, 29],
  "humidity": [45, 80, 55, 90, 60]
})
println("Raw sensor data:", sensors)

-- Filter to hot sensors (temp >= 30)
let hot = sensors.filter(\row -> row["temp_c"] >= 30)
println("Hot sensors:", hot)

-- Add a Fahrenheit column
let withF = sensors.withColumn("temp_f", \row -> row["temp_c"] * 9 / 5 + 32)
println("With Fahrenheit:", withF)

-- Sort by temperature descending
let ranked = sensors.sort("temp_c", false)
println("Ranked by temp:", ranked)

-- Extract just the temperature series and compute stats
let temps = data.Series([22, 35, 18, 41, 29], "temp_c")
println("Temp mean:", temps.mean())
println("Temp min:", temps.min())
println("Temp max:", temps.max())

println("\nDone.")
