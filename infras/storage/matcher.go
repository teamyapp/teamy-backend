package storage

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type DataType string

const (
	IDDataType       DataType = "ID"
	StringDataType   DataType = "STRING"
	IntDataType      DataType = "INT"
	FloatDataType    DataType = "FLOAT"
	BoolDataType     DataType = "BOOL"
	DateTimeDataType DataType = "DATE_TIME"
)

var dataTypes = map[DataType]bool{
	IDDataType:       true,
	StringDataType:   true,
	IntDataType:      true,
	FloatDataType:    true,
	BoolDataType:     true,
	DateTimeDataType: true,
}

type Comparison string

const (
	EqualToComparison              Comparison = "EQUAL_TO"
	MatchPatternComparison         Comparison = "MATCH_PATTERN"
	LessThanComparison             Comparison = "LESS_THAN"
	LessThanOrEqualToComparison    Comparison = "LESS_THAN_OR_EQUAL_TO"
	GreaterThanComparison          Comparison = "GREATER_THAN"
	GreaterThanOrEqualToComparison Comparison = "GREATER_THAN_OR_EQUAL_TO"
)

var comparisons = map[Comparison]bool{
	EqualToComparison:              true,
	MatchPatternComparison:         true,
	LessThanComparison:             true,
	LessThanOrEqualToComparison:    true,
	GreaterThanComparison:          true,
	GreaterThanOrEqualToComparison: true,
}

type Matcher struct {
	DataType   DataType
	Comparison Comparison
	Pattern    string
}

func (m Matcher) Validate() error {
	_, ok := dataTypes[m.DataType]
	if !ok {
		return fmt.Errorf("invalid dataType: %v", m.DataType)
	}

	_, ok = comparisons[m.Comparison]
	if !ok {
		return fmt.Errorf("invalid comparison: %v", m.Comparison)
	}

	return nil
}

func (m Matcher) match(value string) (bool, error) {
	switch m.Comparison {
	case EqualToComparison:
		return m.equalTo(value)
	case MatchPatternComparison:
		if m.DataType != StringDataType {
			return false, fmt.Errorf("match pattern only accept strings: dataType=%v", m.DataType)
		}

		return regexp.MatchString(m.Pattern, value)
	case LessThanComparison:
		return m.lessThan(value)
	case LessThanOrEqualToComparison:
		return m.lessThanOrEqualTo(value)
	case GreaterThanComparison:
		return m.greaterThan(value)
	case GreaterThanOrEqualToComparison:
		return m.greaterThanOrEqualTo(value)
	default:
		return false, fmt.Errorf("unknown comparision: %v", m.Comparison)
	}
}

func (m Matcher) equalTo(value string) (bool, error) {
	switch m.DataType {
	case IDDataType:
		return m.matchID(value, func(value uint64, pattern uint64) bool {
			return value == pattern
		})
	case IntDataType:
		return m.matchInt(value, func(value int64, pattern int64) bool {
			return value == pattern
		})
	case FloatDataType:
		return m.matchFloat(value, func(value float64, pattern float64) bool {
			return value == pattern
		})
	case BoolDataType:
		return m.matchBool(value, func(value bool, pattern bool) bool {
			return value == pattern
		})
	case DateTimeDataType:
		return m.matchTime(value, func(value time.Time, pattern time.Time) bool {
			return value.Equal(pattern)
		})
	default:
		return false, fmt.Errorf(
			"unsupported data type for comparison: comparison=%v dataType=%v",
			m.Comparison, m.DataType)
	}
}

func (m Matcher) lessThan(value string) (bool, error) {
	switch m.DataType {
	case IDDataType:
		return m.matchID(value, func(value uint64, pattern uint64) bool {
			return value < pattern
		})
	case IntDataType:
		return m.matchInt(value, func(value int64, pattern int64) bool {
			return value < pattern
		})
	case FloatDataType:
		return m.matchFloat(value, func(value float64, pattern float64) bool {
			return value < pattern
		})
	case DateTimeDataType:
		return m.matchTime(value, func(value time.Time, pattern time.Time) bool {
			return value.Before(pattern)
		})
	default:
		return false, fmt.Errorf(
			"unsupported data type for comparison: comparison=%v dataType=%v",
			m.Comparison, m.DataType)
	}
}

func (m Matcher) lessThanOrEqualTo(value string) (bool, error) {
	switch m.DataType {
	case IDDataType:
		return m.matchID(value, func(value uint64, pattern uint64) bool {
			return value <= pattern
		})
	case IntDataType:
		return m.matchInt(value, func(value int64, pattern int64) bool {
			return value <= pattern
		})
	case FloatDataType:
		return m.matchFloat(value, func(value float64, pattern float64) bool {
			return value <= pattern
		})
	case DateTimeDataType:
		return m.matchTime(value, func(value time.Time, pattern time.Time) bool {
			return !value.After(pattern)
		})
	default:
		return false, fmt.Errorf(
			"unsupported data type for comparison: comparison=%v dataType=%v",
			m.Comparison, m.DataType)
	}
}

func (m Matcher) greaterThan(value string) (bool, error) {
	switch m.DataType {
	case IDDataType:
		return m.matchID(value, func(value uint64, pattern uint64) bool {
			return value > pattern
		})
	case IntDataType:
		return m.matchInt(value, func(value int64, pattern int64) bool {
			return value > pattern
		})
	case FloatDataType:
		return m.matchFloat(value, func(value float64, pattern float64) bool {
			return value > pattern
		})
	case DateTimeDataType:
		return m.matchTime(value, func(value time.Time, pattern time.Time) bool {
			return value.After(pattern)
		})
	default:
		return false, fmt.Errorf(
			"unsupported data type for comparison: comparison=%v dataType=%v",
			m.Comparison, m.DataType)
	}
}

func (m Matcher) greaterThanOrEqualTo(value string) (bool, error) {
	switch m.DataType {
	case IDDataType:
		return m.matchID(value, func(value uint64, pattern uint64) bool {
			return value >= pattern
		})
	case IntDataType:
		return m.matchInt(value, func(value int64, pattern int64) bool {
			return value >= pattern
		})
	case FloatDataType:
		return m.matchFloat(value, func(value float64, pattern float64) bool {
			return value >= pattern
		})
	case DateTimeDataType:
		return m.matchTime(value, func(value time.Time, pattern time.Time) bool {
			return !value.Before(pattern)
		})
	default:
		return false, fmt.Errorf(
			"unsupported data type for comparison: comparison=%v dataType=%v",
			m.Comparison, m.DataType)
	}
}

func (m Matcher) matchID(value string, compare func(value uint64, pattern uint64) bool) (bool, error) {
	val, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return false, err
	}

	ptn, err := strconv.ParseUint(m.Pattern, 10, 64)
	if err != nil {
		return false, err
	}

	return compare(val, ptn), nil
}

func (m Matcher) matchInt(value string, compare func(value int64, pattern int64) bool) (bool, error) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false, err
	}

	ptn, err := strconv.ParseInt(m.Pattern, 10, 64)
	if err != nil {
		return false, err
	}

	return compare(val, ptn), nil
}

func (m Matcher) matchFloat(value string, compare func(value float64, pattern float64) bool) (bool, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false, err
	}

	ptn, err := strconv.ParseFloat(m.Pattern, 64)
	if err != nil {
		return false, err
	}

	return compare(val, ptn), nil
}

func (m Matcher) matchBool(value string, compare func(value bool, pattern bool) bool) (bool, error) {
	val, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	ptn, err := strconv.ParseBool(m.Pattern)
	if err != nil {
		return false, err
	}

	return compare(val, ptn), nil
}

func (m Matcher) matchTime(value string, compare func(value time.Time, pattern time.Time) bool) (bool, error) {
	val, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false, err
	}

	ptn, err := time.Parse(time.RFC3339, m.Pattern)
	if err != nil {
		return false, err
	}

	return compare(val, ptn), nil
}
