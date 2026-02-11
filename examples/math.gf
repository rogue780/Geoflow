-- math.gf: Math library showcase

-- ── Constants ──
println("Pi:", PI())
println("E:", E())
println("Tau:", TAU())

-- ── Trigonometry ──
let angle = toRadians(45)
println("sin(45°):", sin(angle))
println("cos(45°):", cos(angle))
println("tan(45°):", tan(angle))
println("atan2(1, 1) in degrees:", toDegrees(atan2(1, 1)))

-- ── Rounding ──
println("floor(3.7):", floor(3.7))
println("ceil(3.2):", ceil(3.2))
println("round(3.5):", round(3.5))

-- ── Exponential / Logarithmic ──
println("exp(1):", exp(1))
println("log(E()):", log(E()))
println("log2(1024):", log2(1024))
println("log10(10000):", log10(10000))
println("cbrt(27):", cbrt(27))
println("hypot(3, 4):", hypot(3, 4))

-- ── Sign / Clamp / Lerp ──
println("sign(-42):", sign(-42))
println("clamp(150, 0, 100):", clamp(150, 0, 100))
println("lerp(0, 100, 0.75):", lerp(0, 100, 0.75))

-- ── GCD / LCM / Factorial ──
println("gcd(48, 18):", gcd(48, 18))
println("lcm(12, 8):", lcm(12, 8))
println("factorial(10):", factorial(10))

-- ── Statistics ──
let data = [2, 4, 4, 4, 5, 5, 7, 9]
println("Data:", data)
println("Mean:", mean(data))
println("Median:", median(data))
println("Variance:", variance(data))
println("Std Dev:", stddev(data))
println("25th percentile:", percentile(data, 25))
println("75th percentile:", percentile(data, 75))

-- Correlation and linear regression
let xs = [1, 2, 3, 4, 5]
let ys = [2.1, 3.9, 6.2, 7.8, 10.1]
println("Correlation:", correlation(xs, ys))
let reg = linreg(xs, ys)
println("Linear regression: slope =", reg[0], ", intercept =", reg[1], ", R² =", reg[2])

-- ── Vectors ──
let v1 = vec(1, 2, 3)
let v2 = vec(4, 5, 6)
println("v1:", v1)
println("v2:", v2)
println("v1 + v2:", v1 + v2)
println("v1 - v2:", v1 - v2)
println("v1 * 3:", v1 * 3)
println("dot(v1, v2):", dot(v1, v2))
println("cross(i, j):", cross(vec(1, 0, 0), vec(0, 1, 0)))
println("magnitude(v1):", magnitude(v1))
println("normalize(vec(3, 4)):", normalize(vec(3, 4)))
println("v1.x:", v1.x, " v1.y:", v1.y, " v1.z:", v1.z)

-- ── Matrices ──
let m = mat([1, 2], [3, 4])
println("Matrix:", m)
println("Transpose:", transpose(m))
println("Determinant:", determinant(m))
println("Identity 3x3:", identity(3))

let a = mat([1, 2], [3, 4])
let b = mat([5, 6], [7, 8])
println("A * B:", matMul(a, b))

-- Matrix-vector multiplication
let transform = mat([2, 0], [0, 3])
let point = vec(5, 7)
println("Transform", point, "->", matVecMul(transform, point))

-- ── Complex Numbers ──
let z1 = complex(3, 4)
let z2 = complex(1, -2)
println("z1:", z1)
println("z2:", z2)
println("z1 + z2:", z1 + z2)
println("z1 * z2:", z1 * z2)
println("|z1|:", complexMag(z1))
println("conjugate(z1):", conjugate(z1))
println("z1.real:", z1.real, " z1.imag:", z1.imag)

-- ── Numerical Utilities ──
println("linspace(0, 1, 5):", linspace(0, 1, 5))
println("arange(0, 10, 2.5):", arange(0, 10, 2.5))

-- ── Putting it all together: distance between 3D points ──
fn distance(p1, p2) {
  magnitude(p1 - p2)
}
let pointA = vec(1, 2, 3)
let pointB = vec(4, 6, 3)
println("Distance from", pointA, "to", pointB, ":", distance(pointA, pointB))
