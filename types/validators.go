package types

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ValidateRangeNumbers[K comparable, T int | int8 | int16 | int32 | int64 | float32 | float64 | uint8 | uint16 | uint32 | uint64](v T, rangeRule string) error {
	if rangeRule == "" {
		return nil
	}

	if !strings.Contains(rangeRule, "..") {
		return fmt.Errorf("incorrect range value %s", rangeRule)
	}

	nums := strings.Split(rangeRule, "..")
	if len(nums) != 2 {
		return fmt.Errorf("range value (%s) is incorrect, it should be separated by .. (two dots)", rangeRule)
	}

	num1, err := strconv.ParseFloat(nums[0], 64)
	if err != nil {
		return fmt.Errorf("range start value (%s) is not a number", nums[0])
	}
	num2, err := strconv.ParseFloat(nums[1], 64)
	if err != nil {
		return fmt.Errorf("range end value (%s) is not a number", nums[1])
	}

	if v < T(num1) || v > T(num2) {
		return fmt.Errorf("number %v is outside of the range %s", v, rangeRule)
	}
	return nil
}

func ValidateRangeTimeDuration(v time.Duration, rule string) error {
	if rule == "" {
		return nil
	}

	if !strings.Contains(rule, "..") {
		return fmt.Errorf("incorrect range value %s", rule)
	}

	nums := strings.Split(rule, "..")
	if len(nums) != 2 {
		return fmt.Errorf("range value (%s) is incorrect, it should be separated by .. (two dots)", rule)
	}

	num1, err := time.ParseDuration(nums[0])
	if err != nil {
		return fmt.Errorf("range start value (%s) is not a time.Duration parseable value", nums[0])
	}
	num2, err := time.ParseDuration(nums[1])
	if err != nil {
		return fmt.Errorf("range end value (%s) is not a time.Duration parseable value", nums[1])
	}

	if v < num1 || v > num2 {
		return fmt.Errorf("time.Duration %v is outside of the range %s", v, rule)
	}
	return nil
}

func ValidateGreaterThan[K comparable, T string | int | int8 | int16 | int32 | int64 | float32 | float64 | uint8 | uint16 | uint32 | uint64](val T, rule T) error {
	if val > rule {
		return nil
	}
	return fmt.Errorf("value %v must be greater than %v", val, rule)
}

func ValidateGreaterThanTimeDuration(val time.Duration, rule string) error {
	var ruleTd, err = time.ParseDuration(rule)
	if err != nil {
		return fmt.Errorf("GreaterThan rule for time.Duration %s is incorrect", rule)
	}
	if val > ruleTd {
		return nil
	}
	return fmt.Errorf("value %v must be greater than %v", val, rule)
}
func ValidateLessThanTimeDuration(val time.Duration, rule string) error {
	var ruleTd, err = time.ParseDuration(rule)
	if err != nil {
		return fmt.Errorf("LessThan rule for time.Duration %s is incorrect", rule)
	}
	if val < ruleTd {
		return nil
	}
	return fmt.Errorf("value %v must be less than %v", val, rule)
}

func ValidateLessThan[K comparable, T string | int | int8 | int16 | int32 | int64 | float32 | float64 | uint8 | uint16 | uint32 | uint64](val T, rule T) error {
	if val < rule {
		return nil
	}
	return fmt.Errorf("value %v must be greater than %v", val, rule)
}

// ValidateNumbersSet checks to see if a given numeric value exists among a passed set of similarly typed values
// Note: if the passed set is empty, the validation skips and returns no error
func ValidateNumbersSet[K comparable, T int | int8 | int16 | int32 | int64 | float32 | float64 | uint8 | uint16 | uint32 | uint64](val T, setRule string) error {
	if setRule == "" {
		return nil
	}
	var setStr = strings.Split(setRule, ",")
	var set = make([]T, 0)
	for _, v := range setStr {
		var i, err = strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}
		set = append(set, T(i))
	}
	if set == nil || len(set) == 0 {
		return nil
	}
	for _, v := range set {
		if val == v {
			return nil
		}
	}
	return errors.New("the given value is not among the allowed set")
}

func ValidateStringSet(val string, setRule string) error {
	if setRule == "" {
		return nil
	}
	var set = strings.Split(setRule, ",")
	if set == nil || len(set) == 0 {
		return nil
	}
	for _, v := range set {
		if val == v {
			return nil
		}
	}
	return fmt.Errorf("the given value %s is not among the allowed set", val)
}

func ValidateNumbers[k comparable, T int | int8 | int16 | int32 | int64 | float32 | float64 | uint8 | uint16 | uint32 | uint64](val T,
	valRules map[string]string) error {
	if v, ok := valRules[VdSet]; ok {
		return ValidateNumbersSet[T](val, v)
	} else if v, ok := valRules[VdRange]; ok {
		return ValidateRangeNumbers[T](val, v)
	} else if v, ok := valRules[VdGt]; ok {
		return ValidateRangeNumbers[T](val, v)
	} else if v, ok := valRules[VdLt]; ok {
		return ValidateRangeNumbers[T](val, v)
	}
	return nil
}

func ValidateStrings(val string,
	valRules map[string]string) error {
	if v, ok := valRules[VdSet]; ok {
		return ValidateStringSet(val, v)
	} else if v, ok := valRules[VdGt]; ok {
		return ValidateGreaterThan[string](val, v)
	} else if v, ok := valRules[VdLt]; ok {
		return ValidateLessThan[string](val, v)
	}
	return nil
}

func ValidateTimeDurations(t time.Duration,
	valRules map[string]string) error {
	if v, ok := valRules[VdRange]; ok {
		return ValidateRangeTimeDuration(t, v)
	} else if v, ok := valRules[VdGt]; ok {
		return ValidateGreaterThanTimeDuration(t, v)
	} else if v, ok := valRules[VdLt]; ok {
		return ValidateLessThanTimeDuration(t, v)
	}
	return nil
}
