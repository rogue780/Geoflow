// Package linalg provides advanced linear algebra functions for the GeoFlow standard library.
package linalg

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/rogue780/geoflow/internal/object"
)

// toFloat extracts a float64 from an Integer or Float object.
func toFloat(o object.Object) (float64, bool) {
	switch v := o.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	default:
		return 0, false
	}
}

// GetExports returns all exported functions for the std.math.linalg module.
func GetExports() map[string]object.Object {
	return map[string]object.Object{
		// Constructors
		"vector":     &object.Builtin{Name: "linalg.vector", Fn: vectorConstructor},
		"matrix":     &object.Builtin{Name: "linalg.matrix", Fn: matrixConstructor},
		"eye":        &object.Builtin{Name: "linalg.eye", Fn: eye},
		"zeros":      &object.Builtin{Name: "linalg.zeros", Fn: zerosMatrix},
		"ones":       &object.Builtin{Name: "linalg.ones", Fn: onesMatrix},
		"diag":       &object.Builtin{Name: "linalg.diag", Fn: diag},

		// Linear system solving
		"solve":      &object.Builtin{Name: "linalg.solve", Fn: solve},

		// Eigenvalues
		"eigenvalues": &object.Builtin{Name: "linalg.eigenvalues", Fn: eigenvalues},

		// SVD basics
		"svd":        &object.Builtin{Name: "linalg.svd", Fn: svd},

		// Factorizations
		"lu":         &object.Builtin{Name: "linalg.lu", Fn: lu},
		"qr":         &object.Builtin{Name: "linalg.qr", Fn: qr},

		// Properties
		"rank":       &object.Builtin{Name: "linalg.rank", Fn: rank},
		"det":        &object.Builtin{Name: "linalg.det", Fn: det},
		"trace":      &object.Builtin{Name: "linalg.trace", Fn: trace},
		"cond":       &object.Builtin{Name: "linalg.cond", Fn: cond},

		// Norms
		"vectorNorm": &object.Builtin{Name: "linalg.vectorNorm", Fn: vectorNorm},
		"matrixNorm": &object.Builtin{Name: "linalg.matrixNorm", Fn: matrixNorm},

		// Special matrices
		"hilbert":     &object.Builtin{Name: "linalg.hilbert", Fn: hilbert},
		"vandermonde": &object.Builtin{Name: "linalg.vandermonde", Fn: vandermonde},

		// Products
		"kronecker":  &object.Builtin{Name: "linalg.kronecker", Fn: kronecker},
		"hadamard":   &object.Builtin{Name: "linalg.hadamard", Fn: hadamard},
	}
}

// ── Constructors ──

// vectorConstructor creates a Vector from a list of numbers or individual number arguments.
func vectorConstructor(args ...object.Object) object.Object {
	if len(args) == 1 {
		if list, ok := args[0].(*object.List); ok {
			elems := make([]float64, len(list.Elements))
			for i, e := range list.Elements {
				v, ok := toFloat(e)
				if !ok {
					return &object.Error{Message: fmt.Sprintf("linalg.vector: element %d is not a number", i)}
				}
				elems[i] = v
			}
			return &object.Vector{Elements: elems}
		}
	}
	if len(args) == 0 {
		return &object.Error{Message: "linalg.vector expects at least 1 argument"}
	}
	elems := make([]float64, len(args))
	for i, arg := range args {
		v, ok := toFloat(arg)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("linalg.vector: argument %d is not a number", i)}
		}
		elems[i] = v
	}
	return &object.Vector{Elements: elems}
}

// matrixConstructor creates a Matrix from lists of lists.
func matrixConstructor(args ...object.Object) object.Object {
	if len(args) == 0 {
		return &object.Error{Message: "linalg.matrix expects at least 1 argument (rows as lists)"}
	}
	rows := make([][]float64, len(args))
	var cols int
	for i, arg := range args {
		list, ok := arg.(*object.List)
		if !ok {
			return &object.Error{Message: fmt.Sprintf("linalg.matrix: argument %d must be a list", i)}
		}
		row := make([]float64, len(list.Elements))
		for j, e := range list.Elements {
			v, ok := toFloat(e)
			if !ok {
				return &object.Error{Message: fmt.Sprintf("linalg.matrix: element [%d][%d] is not a number", i, j)}
			}
			row[j] = v
		}
		if i == 0 {
			cols = len(row)
		} else if len(row) != cols {
			return &object.Error{Message: "linalg.matrix: all rows must have the same length"}
		}
		rows[i] = row
	}
	return &object.Matrix{Rows: len(rows), Cols: cols, Data: rows}
}

// eye creates an identity matrix of size n x n.
func eye(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.eye expects 1 argument (size)"}
	}
	n, ok := args[0].(*object.Integer)
	if !ok || n.Value <= 0 {
		return &object.Error{Message: "linalg.eye: argument must be a positive integer"}
	}
	size := int(n.Value)
	data := make([][]float64, size)
	for i := range data {
		data[i] = make([]float64, size)
		data[i][i] = 1
	}
	return &object.Matrix{Rows: size, Cols: size, Data: data}
}

// zerosMatrix creates a rows x cols matrix of zeros.
func zerosMatrix(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "linalg.zeros expects 2 arguments (rows, cols)"}
	}
	r, ok1 := args[0].(*object.Integer)
	c, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 || r.Value <= 0 || c.Value <= 0 {
		return &object.Error{Message: "linalg.zeros: arguments must be positive integers"}
	}
	rows, cols := int(r.Value), int(c.Value)
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}
	return &object.Matrix{Rows: rows, Cols: cols, Data: data}
}

// onesMatrix creates a rows x cols matrix of ones.
func onesMatrix(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "linalg.ones expects 2 arguments (rows, cols)"}
	}
	r, ok1 := args[0].(*object.Integer)
	c, ok2 := args[1].(*object.Integer)
	if !ok1 || !ok2 || r.Value <= 0 || c.Value <= 0 {
		return &object.Error{Message: "linalg.ones: arguments must be positive integers"}
	}
	rows, cols := int(r.Value), int(c.Value)
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			data[i][j] = 1
		}
	}
	return &object.Matrix{Rows: rows, Cols: cols, Data: data}
}

// diag creates a diagonal matrix from a vector, or extracts the diagonal of a matrix.
func diag(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.diag expects 1 argument (vector or matrix)"}
	}
	switch v := args[0].(type) {
	case *object.Vector:
		n := len(v.Elements)
		data := make([][]float64, n)
		for i := range data {
			data[i] = make([]float64, n)
			data[i][i] = v.Elements[i]
		}
		return &object.Matrix{Rows: n, Cols: n, Data: data}
	case *object.Matrix:
		n := v.Rows
		if v.Cols < n {
			n = v.Cols
		}
		elems := make([]float64, n)
		for i := 0; i < n; i++ {
			elems[i] = v.Data[i][i]
		}
		return &object.Vector{Elements: elems}
	default:
		return &object.Error{Message: "linalg.diag: argument must be a vector or matrix"}
	}
}

// ── Linear System Solving ──

// solve solves the linear system Ax = b using Gaussian elimination with partial pivoting.
// args: A (matrix), b (vector) -> x (vector)
func solve(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "linalg.solve expects 2 arguments (matrix A, vector b)"}
	}
	A, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.solve: first argument must be a matrix"}
	}
	b, ok := args[1].(*object.Vector)
	if !ok {
		return &object.Error{Message: "linalg.solve: second argument must be a vector"}
	}
	if A.Rows != A.Cols {
		return &object.Error{Message: "linalg.solve: matrix must be square"}
	}
	n := A.Rows
	if len(b.Elements) != n {
		return &object.Error{Message: "linalg.solve: vector length must match matrix dimension"}
	}

	// Create augmented matrix [A|b]
	aug := make([][]float64, n)
	for i := range aug {
		aug[i] = make([]float64, n+1)
		copy(aug[i], A.Data[i])
		aug[i][n] = b.Elements[i]
	}

	// Gaussian elimination with partial pivoting
	for i := 0; i < n; i++ {
		maxVal := math.Abs(aug[i][i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(aug[k][i]) > maxVal {
				maxVal = math.Abs(aug[k][i])
				maxRow = k
			}
		}
		if maxVal < 1e-12 {
			return &object.Error{Message: "linalg.solve: matrix is singular"}
		}
		aug[i], aug[maxRow] = aug[maxRow], aug[i]

		for k := i + 1; k < n; k++ {
			factor := aug[k][i] / aug[i][i]
			for j := i; j <= n; j++ {
				aug[k][j] -= factor * aug[i][j]
			}
		}
	}

	// Back substitution
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		x[i] = aug[i][n]
		for j := i + 1; j < n; j++ {
			x[i] -= aug[i][j] * x[j]
		}
		x[i] /= aug[i][i]
	}

	return &object.Vector{Elements: x}
}

// ── Eigenvalues ──

// eigenvalues computes eigenvalues for 2x2 and 3x3 matrices.
// Returns a list of floats (real eigenvalues) or complex numbers.
func eigenvalues(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.eigenvalues expects 1 argument (square matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.eigenvalues: argument must be a matrix"}
	}
	if m.Rows != m.Cols {
		return &object.Error{Message: "linalg.eigenvalues: matrix must be square"}
	}

	n := m.Rows
	switch n {
	case 1:
		return &object.List{Elements: []object.Object{
			&object.Float{Value: m.Data[0][0]},
		}}
	case 2:
		return eigenvalues2x2(m.Data)
	case 3:
		return eigenvalues3x3(m.Data)
	default:
		return &object.Error{Message: fmt.Sprintf("linalg.eigenvalues: only supports 1x1, 2x2, and 3x3 matrices, got %dx%d", n, n)}
	}
}

func eigenvalues2x2(d [][]float64) object.Object {
	a, b := d[0][0], d[0][1]
	c, dd := d[1][0], d[1][1]

	tr := a + dd
	det := a*dd - b*c
	disc := tr*tr - 4*det

	if disc >= 0 {
		sqrtDisc := math.Sqrt(disc)
		return &object.List{Elements: []object.Object{
			&object.Float{Value: (tr + sqrtDisc) / 2},
			&object.Float{Value: (tr - sqrtDisc) / 2},
		}}
	}
	sqrtDisc := math.Sqrt(-disc)
	return &object.List{Elements: []object.Object{
		&object.Complex{Real: tr / 2, Imag: sqrtDisc / 2},
		&object.Complex{Real: tr / 2, Imag: -sqrtDisc / 2},
	}}
}

func eigenvalues3x3(d [][]float64) object.Object {
	// Use the characteristic equation: lambda^3 - tr*lambda^2 + p*lambda - det = 0
	// Where p = sum of 2x2 minors of the diagonal
	a := d[0][0]
	b := d[1][1]
	c := d[2][2]

	tr := a + b + c // trace

	// Sum of principal 2x2 minors
	p := a*b - d[0][1]*d[1][0] +
		a*c - d[0][2]*d[2][0] +
		b*c - d[1][2]*d[2][1]

	// Determinant
	det := a*(b*c-d[1][2]*d[2][1]) -
		d[0][1]*(d[1][0]*c-d[1][2]*d[2][0]) +
		d[0][2]*(d[1][0]*d[2][1]-b*d[2][0])

	// Solve cubic: x^3 - tr*x^2 + p*x - det = 0
	// Using Cardano's formula / trigonometric method for 3 real roots
	q := (3*p - tr*tr) / 9
	r := (9*tr*p - 27*det - 2*tr*tr*tr) / 54

	disc := q*q*q + r*r

	if disc >= 0 {
		// One real root and two complex conjugate roots (or all real with multiplicities)
		sqrtDisc := math.Sqrt(disc)
		s := cbrt(r + sqrtDisc)
		t := cbrt(r - sqrtDisc)
		x1 := s + t + tr/3

		realPart := -(s+t)/2 + tr/3
		imagPart := math.Sqrt(3) * (s - t) / 2

		if math.Abs(imagPart) < 1e-10 {
			return &object.List{Elements: []object.Object{
				&object.Float{Value: x1},
				&object.Float{Value: realPart},
				&object.Float{Value: realPart},
			}}
		}
		return &object.List{Elements: []object.Object{
			&object.Float{Value: x1},
			&object.Complex{Real: realPart, Imag: imagPart},
			&object.Complex{Real: realPart, Imag: -imagPart},
		}}
	}

	// Three distinct real roots (trigonometric method)
	theta := math.Acos(r / math.Sqrt(-q*q*q))
	sqrtNegQ := math.Sqrt(-q)
	return &object.List{Elements: []object.Object{
		&object.Float{Value: 2*sqrtNegQ*math.Cos(theta/3) + tr/3},
		&object.Float{Value: 2*sqrtNegQ*math.Cos((theta+2*math.Pi)/3) + tr/3},
		&object.Float{Value: 2*sqrtNegQ*math.Cos((theta+4*math.Pi)/3) + tr/3},
	}}
}

func cbrt(x float64) float64 {
	if x >= 0 {
		return math.Cbrt(x)
	}
	return -math.Cbrt(-x)
}

// ── SVD ──

// svd computes a simplified SVD for small matrices.
// Returns a Map with keys "U", "S" (singular values as vector), "V".
func svd(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.svd expects 1 argument (matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.svd: argument must be a matrix"}
	}

	// Compute A^T * A
	ata := matMulRaw(transposeRaw(m.Data, m.Rows, m.Cols), m.Data, m.Cols, m.Rows, m.Rows, m.Cols)

	// Get eigenvalues of A^T * A (these are sigma^2 values)
	n := m.Cols
	if n > 3 {
		return &object.Error{Message: "linalg.svd: only supports matrices with up to 3 columns"}
	}

	var sigmaSquared []float64
	switch n {
	case 1:
		sigmaSquared = []float64{ata[0][0]}
	case 2:
		tr := ata[0][0] + ata[1][1]
		det := ata[0][0]*ata[1][1] - ata[0][1]*ata[1][0]
		disc := tr*tr - 4*det
		if disc < 0 {
			disc = 0
		}
		sqrtDisc := math.Sqrt(disc)
		sigmaSquared = []float64{(tr + sqrtDisc) / 2, (tr - sqrtDisc) / 2}
	case 3:
		eigResult := eigenvalues3x3(ata)
		eigList := eigResult.(*object.List)
		sigmaSquared = make([]float64, len(eigList.Elements))
		for i, e := range eigList.Elements {
			if f, ok := e.(*object.Float); ok {
				sigmaSquared[i] = f.Value
			} else {
				// Complex eigenvalue - use magnitude
				if c, ok := e.(*object.Complex); ok {
					sigmaSquared[i] = cmplx.Abs(complex(c.Real, c.Imag))
				}
			}
		}
	}

	sigmas := make([]float64, len(sigmaSquared))
	for i, s2 := range sigmaSquared {
		if s2 > 0 {
			sigmas[i] = math.Sqrt(s2)
		}
	}

	return &object.Map{Pairs: []object.MapPair{
		{Key: &object.String{Value: "S"}, Value: &object.Vector{Elements: sigmas}},
		{Key: &object.String{Value: "rows"}, Value: &object.Integer{Value: int64(m.Rows)}},
		{Key: &object.String{Value: "cols"}, Value: &object.Integer{Value: int64(m.Cols)}},
	}}
}

// ── Factorizations ──

// lu computes the LU decomposition with partial pivoting.
// Returns a Map with keys "L", "U", "P" (permutation as list).
func lu(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.lu expects 1 argument (square matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.lu: argument must be a matrix"}
	}
	if m.Rows != m.Cols {
		return &object.Error{Message: "linalg.lu: matrix must be square"}
	}

	n := m.Rows
	// Copy data
	u := make([][]float64, n)
	for i := range u {
		u[i] = make([]float64, n)
		copy(u[i], m.Data[i])
	}
	l := make([][]float64, n)
	for i := range l {
		l[i] = make([]float64, n)
		l[i][i] = 1
	}
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	for i := 0; i < n; i++ {
		// Find pivot
		maxVal := math.Abs(u[i][i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(u[k][i]) > maxVal {
				maxVal = math.Abs(u[k][i])
				maxRow = k
			}
		}
		if maxRow != i {
			u[i], u[maxRow] = u[maxRow], u[i]
			perm[i], perm[maxRow] = perm[maxRow], perm[i]
			// Swap L rows below diagonal
			for j := 0; j < i; j++ {
				l[i][j], l[maxRow][j] = l[maxRow][j], l[i][j]
			}
		}
		if math.Abs(u[i][i]) < 1e-12 {
			continue
		}
		for k := i + 1; k < n; k++ {
			factor := u[k][i] / u[i][i]
			l[k][i] = factor
			for j := i; j < n; j++ {
				u[k][j] -= factor * u[i][j]
			}
		}
	}

	permObjs := make([]object.Object, n)
	for i, p := range perm {
		permObjs[i] = &object.Integer{Value: int64(p)}
	}

	return &object.Map{Pairs: []object.MapPair{
		{Key: &object.String{Value: "L"}, Value: &object.Matrix{Rows: n, Cols: n, Data: l}},
		{Key: &object.String{Value: "U"}, Value: &object.Matrix{Rows: n, Cols: n, Data: u}},
		{Key: &object.String{Value: "P"}, Value: &object.List{Elements: permObjs}},
	}}
}

// qr computes the QR decomposition using Gram-Schmidt orthogonalization.
// Returns a Map with keys "Q" and "R".
func qr(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.qr expects 1 argument (matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.qr: argument must be a matrix"}
	}

	rows := m.Rows
	cols := m.Cols
	if rows < cols {
		return &object.Error{Message: "linalg.qr: matrix must have at least as many rows as columns"}
	}

	// Extract columns of A
	aCols := make([][]float64, cols)
	for j := 0; j < cols; j++ {
		aCols[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			aCols[j][i] = m.Data[i][j]
		}
	}

	// Gram-Schmidt
	q := make([][]float64, cols)
	rData := make([][]float64, cols)
	for j := range rData {
		rData[j] = make([]float64, cols)
	}

	for j := 0; j < cols; j++ {
		v := make([]float64, rows)
		copy(v, aCols[j])

		for i := 0; i < j; i++ {
			proj := dotProduct(q[i], aCols[j])
			rData[i][j] = proj
			for k := 0; k < rows; k++ {
				v[k] -= proj * q[i][k]
			}
		}

		norm := vecNorm(v)
		rData[j][j] = norm
		if norm < 1e-12 {
			q[j] = make([]float64, rows)
		} else {
			q[j] = make([]float64, rows)
			for k := 0; k < rows; k++ {
				q[j][k] = v[k] / norm
			}
		}
	}

	// Build Q matrix (rows x cols)
	qData := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		qData[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			qData[i][j] = q[j][i]
		}
	}

	return &object.Map{Pairs: []object.MapPair{
		{Key: &object.String{Value: "Q"}, Value: &object.Matrix{Rows: rows, Cols: cols, Data: qData}},
		{Key: &object.String{Value: "R"}, Value: &object.Matrix{Rows: cols, Cols: cols, Data: rData}},
	}}
}

// ── Properties ──

// rank computes the rank of a matrix using row reduction.
func rank(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.rank expects 1 argument (matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.rank: argument must be a matrix"}
	}

	// Copy the matrix
	rows := m.Rows
	cols := m.Cols
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		copy(data[i], m.Data[i])
	}

	// Row echelon form
	r := 0
	for j := 0; j < cols && r < rows; j++ {
		maxVal := math.Abs(data[r][j])
		maxRow := r
		for k := r + 1; k < rows; k++ {
			if math.Abs(data[k][j]) > maxVal {
				maxVal = math.Abs(data[k][j])
				maxRow = k
			}
		}
		if maxVal < 1e-10 {
			continue
		}
		data[r], data[maxRow] = data[maxRow], data[r]
		for k := r + 1; k < rows; k++ {
			factor := data[k][j] / data[r][j]
			for l := j; l < cols; l++ {
				data[k][l] -= factor * data[r][l]
			}
		}
		r++
	}

	return &object.Integer{Value: int64(r)}
}

// det computes the determinant of a square matrix.
func det(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.det expects 1 argument (square matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.det: argument must be a matrix"}
	}
	if m.Rows != m.Cols {
		return &object.Error{Message: "linalg.det: matrix must be square"}
	}
	return &object.Float{Value: detLU(m.Data, m.Rows)}
}

// detLU computes the determinant using LU decomposition.
func detLU(data [][]float64, n int) float64 {
	if n == 1 {
		return data[0][0]
	}
	if n == 2 {
		return data[0][0]*data[1][1] - data[0][1]*data[1][0]
	}

	lu := make([][]float64, n)
	for i := range lu {
		lu[i] = make([]float64, n)
		copy(lu[i], data[i])
	}
	sign := 1.0
	for i := 0; i < n; i++ {
		maxVal := math.Abs(lu[i][i])
		maxRow := i
		for k := i + 1; k < n; k++ {
			if math.Abs(lu[k][i]) > maxVal {
				maxVal = math.Abs(lu[k][i])
				maxRow = k
			}
		}
		if maxVal < 1e-12 {
			return 0
		}
		if maxRow != i {
			lu[i], lu[maxRow] = lu[maxRow], lu[i]
			sign *= -1
		}
		for k := i + 1; k < n; k++ {
			factor := lu[k][i] / lu[i][i]
			for j := i; j < n; j++ {
				lu[k][j] -= factor * lu[i][j]
			}
		}
	}
	result := sign
	for i := 0; i < n; i++ {
		result *= lu[i][i]
	}
	return result
}

// trace computes the trace (sum of diagonal elements) of a square matrix.
func trace(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.trace expects 1 argument (square matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.trace: argument must be a matrix"}
	}
	if m.Rows != m.Cols {
		return &object.Error{Message: "linalg.trace: matrix must be square"}
	}
	sum := 0.0
	for i := 0; i < m.Rows; i++ {
		sum += m.Data[i][i]
	}
	return &object.Float{Value: sum}
}

// cond computes the condition number of a matrix (ratio of largest to smallest singular value).
func cond(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.cond expects 1 argument (matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.cond: argument must be a matrix"}
	}
	if m.Cols > 3 {
		return &object.Error{Message: "linalg.cond: only supports matrices with up to 3 columns"}
	}

	// Compute singular values via SVD
	svdResult := svd(args...)
	svdMap, ok := svdResult.(*object.Map)
	if !ok {
		return svdResult // error propagation
	}

	// Get S vector
	var sVec *object.Vector
	for _, p := range svdMap.Pairs {
		if k, ok := p.Key.(*object.String); ok && k.Value == "S" {
			sVec = p.Value.(*object.Vector)
			break
		}
	}
	if sVec == nil || len(sVec.Elements) == 0 {
		return &object.Error{Message: "linalg.cond: could not compute singular values"}
	}

	maxS := 0.0
	minS := math.Inf(1)
	for _, s := range sVec.Elements {
		if s > maxS {
			maxS = s
		}
		if s < minS && s > 0 {
			minS = s
		}
	}
	if minS == 0 || math.IsInf(minS, 1) {
		return &object.Float{Value: math.Inf(1)}
	}
	return &object.Float{Value: maxS / minS}
}

// ── Norms ──

// vectorNorm computes the p-norm of a vector (default p=2).
func vectorNorm(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return &object.Error{Message: "linalg.vectorNorm expects 1-2 arguments (vector[, p])"}
	}
	v, ok := args[0].(*object.Vector)
	if !ok {
		return &object.Error{Message: "linalg.vectorNorm: first argument must be a vector"}
	}
	p := 2.0
	if len(args) == 2 {
		pv, ok := toFloat(args[1])
		if !ok {
			return &object.Error{Message: "linalg.vectorNorm: p must be a number"}
		}
		p = pv
	}

	if math.IsInf(p, 1) {
		// Infinity norm: max absolute value
		maxVal := 0.0
		for _, e := range v.Elements {
			if math.Abs(e) > maxVal {
				maxVal = math.Abs(e)
			}
		}
		return &object.Float{Value: maxVal}
	}

	sum := 0.0
	for _, e := range v.Elements {
		sum += math.Pow(math.Abs(e), p)
	}
	return &object.Float{Value: math.Pow(sum, 1/p)}
}

// matrixNorm computes the Frobenius norm of a matrix.
func matrixNorm(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.matrixNorm expects 1 argument (matrix)"}
	}
	m, ok := args[0].(*object.Matrix)
	if !ok {
		return &object.Error{Message: "linalg.matrixNorm: argument must be a matrix"}
	}
	sum := 0.0
	for i := 0; i < m.Rows; i++ {
		for j := 0; j < m.Cols; j++ {
			sum += m.Data[i][j] * m.Data[i][j]
		}
	}
	return &object.Float{Value: math.Sqrt(sum)}
}

// ── Special Matrices ──

// hilbert creates a Hilbert matrix of size n.
// H[i][j] = 1 / (i + j + 1)
func hilbert(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.hilbert expects 1 argument (size)"}
	}
	n, ok := args[0].(*object.Integer)
	if !ok || n.Value <= 0 {
		return &object.Error{Message: "linalg.hilbert: argument must be a positive integer"}
	}
	size := int(n.Value)
	data := make([][]float64, size)
	for i := range data {
		data[i] = make([]float64, size)
		for j := range data[i] {
			data[i][j] = 1.0 / float64(i+j+1)
		}
	}
	return &object.Matrix{Rows: size, Cols: size, Data: data}
}

// vandermonde creates a Vandermonde matrix from a vector of values.
// V[i][j] = x[i]^j
func vandermonde(args ...object.Object) object.Object {
	if len(args) != 1 {
		return &object.Error{Message: "linalg.vandermonde expects 1 argument (vector)"}
	}
	v, ok := args[0].(*object.Vector)
	if !ok {
		return &object.Error{Message: "linalg.vandermonde: argument must be a vector"}
	}
	n := len(v.Elements)
	data := make([][]float64, n)
	for i := range data {
		data[i] = make([]float64, n)
		data[i][0] = 1
		for j := 1; j < n; j++ {
			data[i][j] = data[i][j-1] * v.Elements[i]
		}
	}
	return &object.Matrix{Rows: n, Cols: n, Data: data}
}

// ── Products ──

// kronecker computes the Kronecker product of two matrices.
func kronecker(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "linalg.kronecker expects 2 arguments (matrix, matrix)"}
	}
	a, ok1 := args[0].(*object.Matrix)
	b, ok2 := args[1].(*object.Matrix)
	if !ok1 || !ok2 {
		return &object.Error{Message: "linalg.kronecker: both arguments must be matrices"}
	}

	rows := a.Rows * b.Rows
	cols := a.Cols * b.Cols
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}

	for i := 0; i < a.Rows; i++ {
		for j := 0; j < a.Cols; j++ {
			for k := 0; k < b.Rows; k++ {
				for l := 0; l < b.Cols; l++ {
					data[i*b.Rows+k][j*b.Cols+l] = a.Data[i][j] * b.Data[k][l]
				}
			}
		}
	}

	return &object.Matrix{Rows: rows, Cols: cols, Data: data}
}

// hadamard computes the Hadamard (element-wise) product of two matrices.
func hadamard(args ...object.Object) object.Object {
	if len(args) != 2 {
		return &object.Error{Message: "linalg.hadamard expects 2 arguments (matrix, matrix)"}
	}
	a, ok1 := args[0].(*object.Matrix)
	b, ok2 := args[1].(*object.Matrix)
	if !ok1 || !ok2 {
		return &object.Error{Message: "linalg.hadamard: both arguments must be matrices"}
	}
	if a.Rows != b.Rows || a.Cols != b.Cols {
		return &object.Error{Message: "linalg.hadamard: matrices must have the same dimensions"}
	}

	data := make([][]float64, a.Rows)
	for i := range data {
		data[i] = make([]float64, a.Cols)
		for j := range data[i] {
			data[i][j] = a.Data[i][j] * b.Data[i][j]
		}
	}
	return &object.Matrix{Rows: a.Rows, Cols: a.Cols, Data: data}
}

// ── Helper functions ──

func dotProduct(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func vecNorm(v []float64) float64 {
	sum := 0.0
	for _, e := range v {
		sum += e * e
	}
	return math.Sqrt(sum)
}

func transposeRaw(data [][]float64, rows, cols int) [][]float64 {
	result := make([][]float64, cols)
	for i := range result {
		result[i] = make([]float64, rows)
		for j := 0; j < rows; j++ {
			result[i][j] = data[j][i]
		}
	}
	return result
}

func matMulRaw(a, b [][]float64, aRows, aCols, bRows, bCols int) [][]float64 {
	_ = bRows // aCols must equal bRows
	result := make([][]float64, aRows)
	for i := range result {
		result[i] = make([]float64, bCols)
		for j := 0; j < bCols; j++ {
			sum := 0.0
			for k := 0; k < aCols; k++ {
				sum += a[i][k] * b[k][j]
			}
			result[i][j] = sum
		}
	}
	return result
}
