-- math.gf: Math library showcase with fully qualified imports

import std.math
import std.math.stats as stats
import std.math.linalg as linalg
import std.math.complex as cmplx
import std.math.numeric as numeric

-- ── Constants ──
println("Pi:", math.PI())
println("E:", math.E())
println("Tau:", math.TAU())
println("Phi:", math.PHI())

-- ── Trigonometry ──
let angle = math.toRadians(45)
println("sin(45°):", math.sin(angle))
println("cos(45°):", math.cos(angle))
println("tan(45°):", math.tan(angle))
println("atan2(1, 1) in degrees:", math.toDegrees(math.atan2(1, 1)))

-- Hyperbolic
println("sinh(1):", math.sinh(1))
println("cosh(1):", math.cosh(1))
println("tanh(1):", math.tanh(1))

-- Inverse trig
println("asin(0.5) deg:", math.toDegrees(math.asin(0.5)))
println("acos(0.5) deg:", math.toDegrees(math.acos(0.5)))

-- ── Rounding & Parts ──
println("floor(3.7):", math.floor(3.7))
println("ceil(3.2):", math.ceil(3.2))
println("round(3.5):", math.round(3.5))
println("trunc(3.9):", math.trunc(3.9))
println("frac(3.75):", math.frac(3.75))

-- ── Exponential / Logarithmic ──
println("exp(1):", math.exp(1))
println("log(E()):", math.log(math.E()))
println("log2(1024):", math.log2(1024))
println("log10(10000):", math.log10(10000))
println("pow(2, 10):", math.pow(2, 10))
println("cbrt(27):", math.cbrt(27))
println("hypot(3, 4):", math.hypot(3, 4))

-- ── Sign / Clamp / Lerp / Smoothstep ──
println("sign(-42):", math.sign(-42))
println("clamp(150, 0, 100):", math.clamp(150, 0, 100))
println("lerp(0, 100, 0.75):", math.lerp(0, 100, 0.75))
println("smoothstep(0, 10, 5):", math.smoothstep(0, 10, 5))

-- ── Number Theory ──
println("gcd(48, 18):", math.gcd(48, 18))
println("lcm(12, 8):", math.lcm(12, 8))
println("factorial(10):", math.factorial(10))
println("isPrime(17):", math.isPrime(17))
println("isPrime(15):", math.isPrime(15))

-- ── Checks (abs, min, max, sqrt are global builtins) ──
println("abs(-7):", abs(-7))
println("min(3, 8):", min(3, 8))
println("max(3, 8):", max(3, 8))
println("sqrt(144):", sqrt(144))
println("isNaN(NAN):", math.isNaN(math.NAN()))
println("isFinite(42):", math.isFinite(42))
println("isInfinite(INFINITY):", math.isInfinite(math.INFINITY()))

-- ══════════════════════════════════════
-- ── Statistics (std.math.stats) ──
-- ══════════════════════════════════════
let data = [2, 4, 4, 4, 5, 5, 7, 9]
println("\nData:", data)
println("Mean:", stats.mean(data))
println("Median:", stats.median(data))
println("Mode:", stats.mode(data))
println("Variance:", stats.variance(data))
println("Std Dev:", stats.stddev(data))
println("SEM:", stats.sem(data))
println("Skewness:", stats.skewness(data))
println("Kurtosis:", stats.kurtosis(data))
println("25th percentile:", stats.percentile(data, 25))
println("75th percentile:", stats.percentile(data, 75))
println("IQR:", stats.iqr(data))

-- Z-scores
println("Z-scores:", stats.zscore(data))

-- Moving mean
println("Moving mean (window=3):", stats.movingMean(data, 3))

-- Correlation and linear regression
let xs = [1, 2, 3, 4, 5]
let ys = [2.1, 3.9, 6.2, 7.8, 10.1]
println("Correlation:", stats.correlation(xs, ys))
println("Covariance:", stats.covariance(xs, ys))
let reg = stats.linreg(xs, ys)
println("Linear regression: slope =", reg[0], ", intercept =", reg[1], ", R² =", reg[2])

-- ══════════════════════════════════════
-- ── Linear Algebra (std.math.linalg) ──
-- ══════════════════════════════════════

-- Vectors
let v1 = linalg.vec(1, 2, 3)
let v2 = linalg.vec(4, 5, 6)
println("\nv1:", v1)
println("v2:", v2)
println("v1 + v2:", v1 + v2)
println("v1 - v2:", v1 - v2)
println("v1 * 3:", v1 * 3)
println("dot(v1, v2):", linalg.dot(v1, v2))
println("cross(i, j):", linalg.cross(linalg.vec(1, 0, 0), linalg.vec(0, 1, 0)))
println("magnitude(v1):", linalg.magnitude(v1))
println("normalize(vec(3, 4)):", linalg.normalize(linalg.vec(3, 4)))
println("v1.x:", v1.x, " v1.y:", v1.y, " v1.z:", v1.z)

-- Matrices
let m = linalg.mat([1, 2], [3, 4])
println("\nMatrix:", m)
println("Transpose:", linalg.transpose(m))
println("Determinant:", linalg.determinant(m))
println("Identity 3x3:", linalg.identity(3))
println("Zeros 2x2:", linalg.zeros(2, 2))

let a = linalg.mat([1, 2], [3, 4])
let b = linalg.mat([5, 6], [7, 8])
println("A * B:", linalg.matMul(a, b))

-- Matrix-vector multiplication
let transform = linalg.mat([2, 0], [0, 3])
let point = linalg.vec(5, 7)
println("Transform", point, "->", linalg.matVecMul(transform, point))

-- ══════════════════════════════════════
-- ── Complex Numbers (std.math.complex) ──
-- ══════════════════════════════════════
let z1 = cmplx.complex(3, 4)
let z2 = cmplx.complex(1, -2)
println("\nz1:", z1)
println("z2:", z2)
println("z1 + z2:", z1 + z2)
println("z1 * z2:", z1 * z2)
println("|z1|:", cmplx.complexMag(z1))
println("phase(z1):", cmplx.complexPhase(z1))
println("conjugate(z1):", cmplx.conjugate(z1))
println("z1.real:", z1.real, " z1.imag:", z1.imag)

-- ══════════════════════════════════════
-- ── Numerical Utilities (std.math.numeric) ──
-- ══════════════════════════════════════
println("\nlinspace(0, 1, 5):", numeric.linspace(0, 1, 5))
println("arange(0, 10, 2.5):", numeric.arange(0, 10, 2.5))

-- ══════════════════════════════════════
-- ── Putting it all together: 3D distance ──
-- ══════════════════════════════════════
fn dist3d(p1, p2) {
  linalg.magnitude(p1 - p2)
}
let pointA = linalg.vec(1, 2, 3)
let pointB = linalg.vec(4, 6, 3)
println("\nDistance from", pointA, "to", pointB, ":", dist3d(pointA, pointB))

-- ══════════════════════════════════════
-- ── Numeric Literal Formats ──
-- ══════════════════════════════════════
println("\n--- Numeric Literal Formats ---")

-- Hexadecimal (0x prefix)
println("0xFF:", 0xFF)
println("0xCAFE:", 0xCAFE)

-- Binary (0b prefix)
println("0b1010:", 0b1010)
println("0b11111111:", 0b11111111)

-- Octal (0o prefix)
println("0o777:", 0o777)
println("0o644:", 0o644)

-- Scientific notation (e/E)
println("2.5e10:", 2.5e10)
println("1.0E-5:", 1.0E-5)
println("6.022e23:", 6.022e23)
