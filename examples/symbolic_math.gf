-- symbolic_math.gf: Symbolic mathematics and numerical methods showcase
--
-- This example demonstrates GeoFlow's symbolic math engine for building,
-- differentiating, and simplifying mathematical expressions, as well as
-- numerical methods for root finding, integration, differentiation,
-- polynomial operations, and interpolation.

import std.math.symbolic as sym
import std.math.numeric as num
import std.math

-- ======================================================================
-- 1. Symbolic Variables and Constants
-- ======================================================================
println("--- Symbolic Variables and Constants ---")

-- Create symbolic variables
let x = sym.var("x")
let y = sym.var("y")
let t = sym.var("t")
println("Variable x:", x)
println("Variable y:", y)
println("Variable t:", t)

-- Create numeric literal expressions
let two = sym.num(2)
let three = sym.num(3)
println("Numeric 2:", two)
println("Numeric 3:", three)

-- Named numeric constants can be created with sym.num
-- (sym.const is also available but "const" is a reserved keyword in GeoFlow,
-- so it cannot be accessed via dot notation on a module)
let pi_approx = sym.num(3.14159)
let gravity = sym.num(9.81)
println("Constant pi:", pi_approx)
println("Constant g:", gravity)

-- ======================================================================
-- 2. Symbolic Arithmetic
-- ======================================================================
println("\n--- Symbolic Arithmetic ---")

-- Basic arithmetic using sym.add, sym.sub, sym.mul, sym.div, sym.pow
let sum_expr = sym.add(x, sym.num(1))
println("x + 1:", sum_expr)

let diff_expr = sym.sub(x, sym.num(5))
println("x - 5:", diff_expr)

let prod_expr = sym.mul(sym.num(2), x)
println("2 * x:", prod_expr)

let quot_expr = sym.div(x, sym.num(3))
println("x / 3:", quot_expr)

let pow_expr = sym.pow(x, sym.num(2))
println("x^2:", pow_expr)

-- Negation
let neg_expr = sym.neg(x)
println("-x:", neg_expr)

-- Compound expressions: 2*x^2 + 3*x + 1
let quadratic = sym.add(
    sym.add(
        sym.mul(sym.num(2), sym.pow(x, sym.num(2))),
        sym.mul(sym.num(3), x)
    ),
    sym.num(1)
)
println("2*x^2 + 3*x + 1:", quadratic)

-- Operator overloading: Expr objects support +, -, *, / infix operators
let overloaded = x + sym.num(1)
println("x + 1 (via operator):", overloaded)

let overloaded2 = x * sym.num(3)
println("x * 3 (via operator):", overloaded2)

-- ======================================================================
-- 3. Symbolic Functions
-- ======================================================================
println("\n--- Symbolic Functions ---")

-- Trigonometric functions
let sin_x = sym.sin(x)
let cos_x = sym.cos(x)
println("sin(x):", sin_x)
println("cos(x):", cos_x)

-- Exponential and logarithmic
let exp_x = sym.exp(x)
let log_x = sym.log(x)
println("exp(x):", exp_x)
println("log(x):", log_x)

-- Square root (internally represented as x^0.5)
let sqrt_x = sym.sqrt(x)
println("sqrt(x):", sqrt_x)

-- Composed functions: sin(x^2)
let sin_x2 = sym.sin(sym.pow(x, sym.num(2)))
println("sin(x^2):", sin_x2)

-- exp(2*x)
let exp_2x = sym.exp(sym.mul(sym.num(2), x))
println("exp(2*x):", exp_2x)

-- ======================================================================
-- 4. Differentiation
-- ======================================================================
println("\n--- Symbolic Differentiation ---")

-- d/dx(x^2) = 2*x
let dx2 = sym.diff(sym.pow(x, sym.num(2)), "x")
println("d/dx(x^2) =", dx2)

-- d/dx(x^3) = 3*x^2
let dx3 = sym.diff(sym.pow(x, sym.num(3)), "x")
println("d/dx(x^3) =", dx3)

-- d/dx(sin(x)) = cos(x)
let dsin = sym.diff(sym.sin(x), "x")
println("d/dx(sin(x)) =", dsin)

-- d/dx(cos(x)) = -sin(x)
let dcos = sym.diff(sym.cos(x), "x")
println("d/dx(cos(x)) =", dcos)

-- d/dx(exp(x)) = exp(x)
let dexp = sym.diff(sym.exp(x), "x")
println("d/dx(exp(x)) =", dexp)

-- d/dx(log(x)) = 1/x
let dlog = sym.diff(sym.log(x), "x")
println("d/dx(log(x)) =", dlog)

-- Chain rule: d/dx(sin(x^2)) = cos(x^2) * 2*x
let dsin_x2 = sym.diff(sym.sin(sym.pow(x, sym.num(2))), "x")
println("d/dx(sin(x^2)) =", dsin_x2)

-- Product rule: d/dx(x * sin(x)) = sin(x) + x * cos(x)
let dprod = sym.diff(sym.mul(x, sym.sin(x)), "x")
println("d/dx(x * sin(x)) =", dprod)

-- Differentiating with respect to a variable not in the expression gives 0
let dy_from_x = sym.diff(sym.pow(x, sym.num(2)), "y")
println("d/dy(x^2) =", dy_from_x)

-- ======================================================================
-- 5. Simplification
-- ======================================================================
println("\n--- Simplification ---")

-- 0 + x simplifies to x
let s1 = sym.simplify(sym.add(sym.num(0), x))
println("simplify(0 + x) =", s1)

-- x + 0 simplifies to x
let s2 = sym.simplify(sym.add(x, sym.num(0)))
println("simplify(x + 0) =", s2)

-- 1 * x simplifies to x
let s3 = sym.simplify(sym.mul(sym.num(1), x))
println("simplify(1 * x) =", s3)

-- 0 * x simplifies to 0
let s4 = sym.simplify(sym.mul(sym.num(0), x))
println("simplify(0 * x) =", s4)

-- x^0 simplifies to 1
let s5 = sym.simplify(sym.pow(x, sym.num(0)))
println("simplify(x^0) =", s5)

-- x^1 simplifies to x
let s6 = sym.simplify(sym.pow(x, sym.num(1)))
println("simplify(x^1) =", s6)

-- x - x simplifies to 0
let s7 = sym.simplify(sym.sub(x, x))
println("simplify(x - x) =", s7)

-- x / x simplifies to 1
let s8 = sym.simplify(sym.div(x, x))
println("simplify(x / x) =", s8)

-- Constant folding: 2 + 3 simplifies to 5
let s9 = sym.simplify(sym.add(sym.num(2), sym.num(3)))
println("simplify(2 + 3) =", s9)

-- Expand: (x + 1) * (x - 1) distributes to x*x - x + x - 1, then simplifies
let expanded = sym.expand(sym.mul(sym.add(x, sym.num(1)), sym.sub(x, sym.num(1))))
println("expand((x + 1)*(x - 1)) =", expanded)

-- ======================================================================
-- 6. Substitution and Realization
-- ======================================================================
println("\n--- Substitution and Realization ---")

-- Substitute a numeric value for a variable (dot method on Expr)
let expr_xplus3 = sym.add(x, sym.num(3))
println("Expression: x + 3")

-- expr.substitute("x", value) returns a new Expr with x replaced
let substituted = expr_xplus3.substitute("x", sym.num(10))
println("After substituting x = 10:", substituted)

-- expr.realize(bindings) evaluates the expression to a float
let result1 = expr_xplus3.realize({"x": 5.0})
println("Realize x + 3 with x=5:", result1)

-- Realize a more complex expression: 2*x^2 + 3*x + 1 at x=4
let quadratic2 = sym.add(
    sym.add(
        sym.mul(sym.num(2), sym.pow(x, sym.num(2))),
        sym.mul(sym.num(3), x)
    ),
    sym.num(1)
)
let quad_val = quadratic2.realize({"x": 4.0})
println("Realize 2*x^2 + 3*x + 1 at x=4:", quad_val)

-- Multi-variable expressions
let multi_expr = sym.add(sym.mul(x, y), sym.num(10))
let multi_result = multi_expr.realize({"x": 3.0, "y": 7.0})
println("Realize x*y + 10 at x=3, y=7:", multi_result)

-- Using sym.substitute as a standalone function (3 arguments)
let sub_result = sym.substitute(sym.add(x, y), "x", sym.num(5))
println("sym.substitute(x + y, x, 5) =", sub_result)

-- Using sym.realize as a standalone function
let real_result = sym.realize(sym.add(x, sym.num(7)), {"x": 3.0})
println("sym.realize(x + 7, {x: 3}) =", real_result)

-- toString method for converting expression to a string value
let expr_str = sym.mul(sym.num(2), x).toString()
println("toString(2*x):", expr_str)

-- isConstant and freeSymbols methods
let const_expr = sym.num(42)
println("Is 42 constant?", const_expr.isConstant())
println("Is x+1 constant?", sum_expr.isConstant())

let free_syms = quadratic2.freeSymbols()
println("Free symbols in 2*x^2 + 3*x + 1:", free_syms)

-- ======================================================================
-- 7. Numerical Root Finding
-- ======================================================================
println("\n--- Numerical Root Finding ---")

-- Bisection method: find where f(x) = x^2 - 2 = 0
-- (i.e., find sqrt(2) in the interval [1, 2])
let f1 = \x -> x * x - 2
let root_bisect = num.bisect(f1, 1.0, 2.0)
println("Bisection: sqrt(2) ~=", root_bisect)

-- Newton's method: find where f(x) = x^2 - 4 = 0
-- (i.e., find 2, starting from x0 = 3)
let f2 = \x -> x * x - 4
let root_newton = num.newton(f2, 3.0)
println("Newton: root of x^2 - 4 ~=", root_newton)

-- Bisection with custom tolerance
let f3 = \x -> x * x * x - 27
let root_cube = num.bisect(f3, 2.0, 4.0, 1e-10)
println("Bisection: cbrt(27) ~=", root_cube)

-- Newton's method for cos(x) = 0 near pi/2
let f4 = \x -> x * x - 9
let root_3 = num.newton(f4, 4.0)
println("Newton: root of x^2 - 9 ~=", root_3)

-- ======================================================================
-- 8. Numerical Integration
-- ======================================================================
println("\n--- Numerical Integration ---")

-- Integrate x^2 from 0 to 1 (exact answer: 1/3)
-- num.quad returns a Tuple (result, error_estimate)
let f_sq = \x -> x * x
let integral1 = num.quad(f_sq, 0.0, 1.0)
println("Integral of x^2 from 0 to 1:", integral1[0], " (error:", integral1[1], ")")

-- Integrate sin(x) from 0 to pi (exact answer: 2)
let f_sin = \x -> math.sin(x)
let integral2 = num.quad(f_sin, 0.0, math.PI())
println("Integral of sin(x) from 0 to pi:", integral2[0], " (error:", integral2[1], ")")

-- Integrate exp(x) from 0 to 1 (exact answer: e - 1 ~ 1.71828)
let f_exp = \x -> math.exp(x)
let integral3 = num.quad(f_exp, 0.0, 1.0)
println("Integral of exp(x) from 0 to 1:", integral3[0], " (error:", integral3[1], ")")

-- ======================================================================
-- 9. Numerical Differentiation
-- ======================================================================
println("\n--- Numerical Differentiation ---")

-- Derivative of x^2 at x=3 (exact answer: 6)
let f_x2 = \x -> x * x
let deriv1 = num.diff(f_x2, 3.0)
println("d/dx(x^2) at x=3:", deriv1)

-- Derivative of sin(x) at x=0 (exact answer: cos(0) = 1)
let f_sinx = \x -> math.sin(x)
let deriv2 = num.diff(f_sinx, 0.0)
println("d/dx(sin(x)) at x=0:", deriv2)

-- Derivative of exp(x) at x=1 (exact answer: e ~ 2.71828)
let f_expx = \x -> math.exp(x)
let deriv3 = num.diff(f_expx, 1.0)
println("d/dx(exp(x)) at x=1:", deriv3)

-- Derivative with custom step size
let deriv4 = num.diff(f_x2, 5.0, 1e-6)
println("d/dx(x^2) at x=5 (h=1e-6):", deriv4)

-- ======================================================================
-- 10. Polynomial Operations
-- ======================================================================
println("\n--- Polynomial Operations ---")

-- Polynomial evaluation using Horner's method
-- Coefficients are [highest degree, ..., lowest degree]
-- p(x) = x^2 + 0*x - 1 = x^2 - 1
-- Coefficients: [1.0, 0.0, -1.0]
let coeffs = [1.0, 0.0, -1.0]
let p_at_2 = num.polyval(coeffs, 2.0)
println("p(x) = x^2 - 1, p(2) =", p_at_2)

let p_at_0 = num.polyval(coeffs, 0.0)
println("p(x) = x^2 - 1, p(0) =", p_at_0)

-- Polynomial: 2*x^3 - x + 4
let cubic_coeffs = [2.0, 0.0, -1.0, 4.0]
let p_cubic = num.polyval(cubic_coeffs, 3.0)
println("p(x) = 2*x^3 - x + 4, p(3) =", p_cubic)

-- Polynomial fitting: fit a line (degree 1) to data points
let xs_fit = [1.0, 2.0, 3.0, 4.0, 5.0]
let ys_fit = [2.1, 3.9, 6.1, 7.9, 10.0]
let fit_coeffs = num.polyfit(xs_fit, ys_fit, 1)
println("Linear fit coefficients (slope, intercept):", fit_coeffs)

-- Evaluate the fitted polynomial at x = 6
let predicted = num.polyval(fit_coeffs, 6.0)
println("Predicted value at x=6:", predicted)

-- Quadratic fit to parabolic data
let xs_quad = [0.0, 1.0, 2.0, 3.0, 4.0]
let ys_quad = [1.0, 2.0, 5.0, 10.0, 17.0]
let quad_coeffs = num.polyfit(xs_quad, ys_quad, 2)
println("Quadratic fit coefficients:", quad_coeffs)

-- ======================================================================
-- 11. Interpolation
-- ======================================================================
println("\n--- Interpolation ---")

-- Create a linear interpolation function from data points
let xs_interp = [0.0, 1.0, 2.0, 3.0, 4.0]
let ys_interp = [0.0, 1.0, 4.0, 9.0, 16.0]

let interp_fn = num.interp1d(xs_interp, ys_interp)

-- Evaluate at known points
println("interp(0.0) =", interp_fn(0.0))
println("interp(2.0) =", interp_fn(2.0))
println("interp(4.0) =", interp_fn(4.0))

-- Evaluate between data points (linear interpolation)
println("interp(0.5) =", interp_fn(0.5))
println("interp(1.5) =", interp_fn(1.5))
println("interp(2.5) =", interp_fn(2.5))
println("interp(3.5) =", interp_fn(3.5))

-- Boundary behavior: clamps to endpoint values
println("interp(-1.0) =", interp_fn(-1.0))
println("interp(5.0) =", interp_fn(5.0))

-- ======================================================================
-- 12. Utility Functions: linspace and arange
-- ======================================================================
println("\n--- Utility Functions ---")

-- linspace: create n evenly spaced points between start and end (inclusive)
let ls1 = num.linspace(0.0, 1.0, 5)
println("linspace(0, 1, 5):", ls1)

let ls2 = num.linspace(-1.0, 1.0, 3)
println("linspace(-1, 1, 3):", ls2)

let ls3 = num.linspace(0.0, 10.0, 11)
println("linspace(0, 10, 11):", ls3)

-- arange: create values from start to end (exclusive) with a step
let ar1 = num.arange(0.0, 5.0, 1.0)
println("arange(0, 5, 1):", ar1)
println("  length:", len(ar1))

let ar2 = num.arange(0.0, 1.0, 0.25)
println("arange(0, 1, 0.25):", ar2)

let ar3 = num.arange(0.0, 10.0, 2.5)
println("arange(0, 10, 2.5):", ar3)

-- Single-argument form: arange(end) starts from 0 with step 1
let ar4 = num.arange(5.0)
println("arange(5):", ar4)

-- Two-argument form: arange(start, end) with step 1
let ar5 = num.arange(3.0, 8.0)
println("arange(3, 8):", ar5)

-- ======================================================================
-- Bonus: Combining symbolic and numeric approaches
-- ======================================================================
println("\n--- Combining Symbolic and Numeric ---")

-- Build a symbolic expression, differentiate it, then realize at a point
let expr_x3 = sym.pow(x, sym.num(3))
let d_x3 = sym.diff(expr_x3, "x")
println("d/dx(x^3) =", d_x3)
let d_x3_at_2 = d_x3.realize({"x": 2.0})
println("  evaluated at x=2:", d_x3_at_2)

-- Verify numerically
let num_deriv_x3 = num.diff(\x -> x * x * x, 2.0)
println("  numerical derivative at x=2:", num_deriv_x3)

-- Build sin(x), differentiate to cos(x), realize at x=0
let sym_sinx = sym.sin(x)
let sym_dsinx = sym.diff(sym_sinx, "x")
println("d/dx(sin(x)) =", sym_dsinx)
let sym_dsinx_at_0 = sym_dsinx.realize({"x": 0.0})
println("  evaluated at x=0:", sym_dsinx_at_0)

println("\nDone.")
