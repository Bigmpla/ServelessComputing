package ReedSolomon

import (
	"errors"
	"fmt"
)

type Matrix struct {
	rows    int
	columns int
	data    [][]byte
}

func NewMatrix(initRows, initColumns int) *Matrix {
	data := make([][]byte, initRows)
	for r := 0; r < initRows; r++ {
		data[r] = make([]byte, initColumns)
	}
	return &Matrix{rows: initRows, columns: initColumns, data: data}
}

func NewMatrixFromData(initData [][]byte) (*Matrix, error) {
	rows := len(initData)
	if rows == 0 {
		return nil, errors.New("data cannot be empty")
	}
	columns := len(initData[0])
	for _, row := range initData {
		if len(row) != columns {
			return nil, errors.New("not all rows have the same number of columns")
		}
	}
	data := make([][]byte, rows)
	for r := 0; r < rows; r++ {
		data[r] = append([]byte{}, initData[r]...)
	}
	return &Matrix{rows: rows, columns: columns, data: data}, nil
}

func IdentityMatrix(size int) *Matrix {
	result := NewMatrix(size, size)
	for i := 0; i < size; i++ {
		result.Set(i, i, 1)
	}
	return result
}

func (m *Matrix) ToString() string {
	result := "["
	for r := 0; r < m.rows; r++ {
		if r > 0 {
			result += ", "
		}
		result += "["
		for c := 0; c < m.columns; c++ {
			if c > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%d", m.data[r][c])
		}
		result += "]"
	}
	result += "]"
	return result
}

func (m *Matrix) ToBigString() string {
	result := ""
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			value := int(m.Get(r, c))
			if value < 0 {
				value += 256
			}
			result += fmt.Sprintf("%02x ", value)
		}
		result += "\n"
	}
	return result
}

func (m *Matrix) GetColumns() int {
	return m.columns
}

func (m *Matrix) GetRows() int {
	return m.rows
}

func (m *Matrix) Get(r, c int) byte {
	if r < 0 || r >= m.rows || c < 0 || c >= m.columns {
		panic("index out of range")
	}
	return m.data[r][c]
}

func (m *Matrix) Set(r, c int, value byte) {
	if r < 0 || r >= m.rows || c < 0 || c >= m.columns {
		panic("index out of range")
	}
	m.data[r][c] = value
}

func (m *Matrix) Equals(other *Matrix) bool {
	if m.rows != other.rows || m.columns != other.columns {
		return false
	}
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			if m.data[r][c] != other.data[r][c] {
				return false
			}
		}
	}
	return true
}

func (m *Matrix) Times(right *Matrix) *Matrix {
	if m.columns != right.rows {
		panic("columns of left matrix must match rows of right matrix")
	}
	result := NewMatrix(m.rows, right.columns)
	for r := 0; r < m.rows; r++ {
		for c := 0; c < right.columns; c++ {
			var value byte = 0
			for i := 0; i < m.columns; i++ {
				value ^= Multiply(m.Get(r, i), right.Get(i, c))
			}
			result.Set(r, c, value)
		}
	}
	return result
}

// Augment returns the concatenation of this matrix and another matrix.
func (m *Matrix) Augment(right *Matrix) *Matrix {
	if m.rows != right.rows {
		panic("Matrices do not have the same number of rows")
	}

	result := NewMatrix(m.rows, m.columns+right.columns)
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.columns; c++ {
			result.data[r][c] = m.data[r][c]
		}
		for c := 0; c < right.columns; c++ {
			result.data[r][m.columns+c] = right.data[r][c]
		}
	}
	return result
}

func (m *Matrix) Submatrix(rmin, cmin, rmax, cmax int) *Matrix {

	result := NewMatrix(rmax-rmin, cmax-cmin)
	for r := rmin; r < rmax; r++ {
		for c := cmin; c < cmax; c++ {
			result.data[r-rmin][c-cmin] = m.data[r][c]
		}
	}
	return result
}

func (m *Matrix) GetRow(row int) []byte {
	result := make([]byte, m.columns)
	for c := 0; c < m.columns; c++ {
		result[c] = m.Get(row, c)
	}
	return result
}

func (m *Matrix) SwapRows(r1, r2 int) {
	if r1 < 0 || r1 >= m.rows || r2 < 0 || r2 >= m.rows {
		panic("Row index out of range")
	}
	m.data[r1], m.data[r2] = m.data[r2], m.data[r1]
}

func (m *Matrix) Invert() *Matrix {
	if m.rows != m.columns {
		panic("Only square matrices can be inverted")
	}
	work := m.Augment(IdentityMatrix(m.rows))
	work.GaussianElimination()
	return work.Submatrix(0, m.rows, m.columns, 2*m.columns)
}

func (m *Matrix) GaussianElimination() {
	for r := 0; r < m.rows; r++ {
		if m.data[r][r] == 0 {
			for rowBelow := r + 1; rowBelow < m.rows; rowBelow++ {
				if m.data[rowBelow][r] != 0 {
					m.SwapRows(r, rowBelow)
					break
				}
			}
		}
		if m.data[r][r] == 0 {
			panic("Matrix is singular")
		}

		if m.data[r][r] != 1 {
			scale := Divide(1, m.data[r][r])
			for c := 0; c < m.columns; c++ {
				m.data[r][c] = Multiply(m.data[r][c], scale)
			}
		}

		for rowBelow := r + 1; rowBelow < m.rows; rowBelow++ {
			if m.data[rowBelow][r] != 0 {
				scale := m.data[rowBelow][r]
				for c := 0; c < m.columns; c++ {
					m.data[rowBelow][c] ^= Multiply(scale, m.data[r][c])
				}
			}
		}
	}
	for d := 0; d < m.rows; d++ {
		for rowAbove := 0; rowAbove < d; rowAbove++ {
			if m.data[rowAbove][d] != 0 {
				scale := m.data[rowAbove][d]
				for c := 0; c < m.columns; c++ {
					m.data[rowAbove][c] ^= Multiply(scale, m.data[d][c])
				}
			}
		}
	}
}
