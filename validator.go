package borm

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// errors errors type
type Errors map[string][]string

type validator struct {
	errors Errors
}

// ResetErrors clean all model errors
func (m *validator) ResetErrors() {
	m.errors = make(Errors)
}

// AddError adding error to record
func (m *validator) AddError(f string, t string) {
	if m.Valid() {
		m.errors = make(Errors)
	}
	m.errors[f] = append(m.errors[f], t)
}

// Valid returns true if no errors found in model
func (m *validator) Valid() bool {
	return len(m.errors) == 0
}

// GetErrors returns record errors
func (m *validator) GetErrors() Errors {
	return m.errors
}

// SetErrors set record errors
func (m *validator) SetErrors(e Errors) {
	m.errors = e
}

// ValidatePresence validates string for presence
//  m.ValidatePresence("Name", m.Name)
func (m *validator) ValidatePresence(f, v string) {
	if utf8.RuneCountInString(v) == 0 {
		m.AddError(f, "can't be blank")
	}
}

// ValidateLength validates string min, max length. -1 for any
//  m.ValidateLength("password", m.Password, 6, 18) // min 6, max 18
func (m *validator) ValidateLength(f, v string, min, max int) {
	if min > 0 {
		if utf8.RuneCountInString(v) < min {
			m.AddError(f, fmt.Sprint("minimum length is", min))
		}
	}
	if max > 0 {
		if utf8.RuneCountInString(v) > max {
			m.AddError(f, fmt.Sprint("maximum length is", max))
		}
	}
}

// ValidateInt validates int min, max. -1 for any
//  m.ValidateInt("number", 10, -1, 11)  // max 18
func (m *validator) ValidateInt(f string, v, min, max int) {
	if min > 0 {
		if v < min {
			m.AddError(f, fmt.Sprint("minimum length is", min))
		}
	}
	if max > 0 {
		if v > max {
			m.AddError(f, fmt.Sprint("maximum length is", max))
		}
	}
}

// ValidateInt64 validates int64 min, max. -1 for any
//  m.ValidateInt64("number", 10, 6, -1) // min 6
func (m *validator) ValidateInt64(f string, v, min, max int64) {
	if min > 0 {
		if v < min {
			m.AddError(f, fmt.Sprint("minimum length is", min))
		}
	}
	if max > 0 {
		if v > max {
			m.AddError(f, fmt.Sprint("maximum length is", max))
		}
	}
}

// ValidateFloat32 validates float32 min, max. -1 for any
//  m.ValidateFloat32("number", 10.2, -1, 11)
func (m *validator) ValidateFloat32(f string, v, min, max float32) {
	if min > 0 {
		if v < min {
			m.AddError(f, fmt.Sprint("minimum length is", min))
		}
	}
	if max > 0 {
		if v > max {
			m.AddError(f, fmt.Sprint("maximum length is", max))
		}
	}
}

// ValidateFloat64 validates float64 min, max. -1 for any
//  m.ValidateFloat64("number", 10.2, -1, 11)
func (m *validator) ValidateFloat64(f string, v, min, max float64) {
	if min > 0 {
		if v < min {
			m.AddError(f, fmt.Sprint("minimum length is", min))
		}
	}
	if max > 0 {
		if v > max {
			m.AddError(f, fmt.Sprint("maximum length is", max))
		}
	}
}

// ValidateFormat validates string format with regex string
//  m.ValidateFormat("ip address", u.IP, `\A(\d{1,3}\.){3}\d{1,3}\z`)
func (m *validator) ValidateFormat(f, v, reg string) {
	if r, _ := regexp.MatchString(reg, v); !r {
		m.AddError(f, "invalid format")
	}
}
