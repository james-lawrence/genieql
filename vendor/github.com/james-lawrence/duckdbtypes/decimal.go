package duckdbtypes

import (
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
)

type Decimal struct {
	Width uint8
	Scale uint8
	Value *big.Int
}

func (d *Decimal) Float64() float64 {
	scale := big.NewInt(int64(d.Scale))
	factor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), scale, nil))
	value := new(big.Float).SetInt(d.Value)
	value.Quo(value, factor)
	f, _ := value.Float64()
	return f
}

func (src *Decimal) AssignTo(dst any) error {
	return float64AssignTo(src.Float64(), dst)
}

func (d *Decimal) String() string {
	// Get the sign, and return early if zero
	if d.Value.Sign() == 0 {
		return "0"
	}

	// Remove the sign from the string integer value
	var signStr string
	scaleless := d.Value.String()
	if d.Value.Sign() < 0 {
		signStr = "-"
		scaleless = scaleless[1:]
	}

	// Remove all zeros from the right side
	zeroTrimmed := strings.TrimRightFunc(scaleless, func(r rune) bool { return r == '0' })
	scale := int(d.Scale) - (len(scaleless) - len(zeroTrimmed))

	// If the string is still bigger than the scale factor, output it without a decimal point
	if scale <= 0 {
		return signStr + zeroTrimmed + strings.Repeat("0", -1*scale)
	}

	// Pad a number with 0.0's if needed
	if len(zeroTrimmed) <= scale {
		return fmt.Sprintf("%s0.%s%s", signStr, strings.Repeat("0", scale-len(zeroTrimmed)), zeroTrimmed)
	}
	return signStr + zeroTrimmed[:len(zeroTrimmed)-scale] + "." + zeroTrimmed[len(zeroTrimmed)-scale:]
}

func float64AssignTo(srcVal float64, dst any) error {
	switch v := dst.(type) {
	case *float32:
		*v = float32(srcVal)
	case *float64:
		*v = srcVal
	default:
		if v := reflect.ValueOf(dst); v.Kind() == reflect.Ptr {
			el := v.Elem()
			switch el.Kind() {
			// if dst is a type alias of a float32 or 64, set dst val
			case reflect.Float32, reflect.Float64:
				el.SetFloat(srcVal)
				return nil
			// if dst is a pointer to pointer, strip the pointer and try again
			case reflect.Ptr:
				if el.IsNil() {
					// allocate destination
					el.Set(reflect.New(el.Type().Elem()))
				}
				return float64AssignTo(srcVal, el.Interface())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				i64 := int64(srcVal)
				if float64(i64) == srcVal {
					return int64AssignTo(i64, dst)
				}
			}
		}
		return fmt.Errorf("cannot assign %v into %T", srcVal, dst)
	}

	return fmt.Errorf("cannot assign %v into %T", srcVal, dst)
}

func int64AssignTo(srcVal int64, dst interface{}) error {
	switch v := dst.(type) {
	case *int:
		// if srcVal < int64(minInt) {
		// 	return fmt.Errorf("%d is less than minimum value for int", srcVal)
		// } else if srcVal > int64(maxInt) {
		// 	return fmt.Errorf("%d is greater than maximum value for int", srcVal)
		// }
		*v = int(srcVal)
	case *int8:
		if srcVal < math.MinInt8 {
			return fmt.Errorf("%d is less than minimum value for int8", srcVal)
		} else if srcVal > math.MaxInt8 {
			return fmt.Errorf("%d is greater than maximum value for int8", srcVal)
		}
		*v = int8(srcVal)
	case *int16:
		if srcVal < math.MinInt16 {
			return fmt.Errorf("%d is less than minimum value for int16", srcVal)
		} else if srcVal > math.MaxInt16 {
			return fmt.Errorf("%d is greater than maximum value for int16", srcVal)
		}
		*v = int16(srcVal)
	case *int32:
		if srcVal < math.MinInt32 {
			return fmt.Errorf("%d is less than minimum value for int32", srcVal)
		} else if srcVal > math.MaxInt32 {
			return fmt.Errorf("%d is greater than maximum value for int32", srcVal)
		}
		*v = int32(srcVal)
	case *int64:
		*v = srcVal
	case *uint:
		if srcVal < 0 {
			return fmt.Errorf("%d is less than zero for uint", srcVal)
			// } else if uint64(srcVal) > uint64(maxUint) {
			// 	return fmt.Errorf("%d is greater than maximum value for uint", srcVal)
		}
		*v = uint(srcVal)
	case *uint8:
		if srcVal < 0 {
			return fmt.Errorf("%d is less than zero for uint8", srcVal)
		} else if srcVal > math.MaxUint8 {
			return fmt.Errorf("%d is greater than maximum value for uint8", srcVal)
		}
		*v = uint8(srcVal)
	case *uint16:
		if srcVal < 0 {
			return fmt.Errorf("%d is less than zero for uint16", srcVal)
		} else if srcVal > math.MaxUint16 {
			return fmt.Errorf("%d is greater than maximum value for uint16", srcVal)
		}
		*v = uint16(srcVal)
	case *uint32:
		if srcVal < 0 {
			return fmt.Errorf("%d is less than zero for uint32", srcVal)
		} else if srcVal > math.MaxUint32 {
			return fmt.Errorf("%d is greater than maximum value for uint32", srcVal)
		}
		*v = uint32(srcVal)
	case *uint64:
		if srcVal < 0 {
			return fmt.Errorf("%d is less than zero for uint64", srcVal)
		}
		*v = uint64(srcVal)
	case sql.Scanner:
		return v.Scan(srcVal)
	default:
		if v := reflect.ValueOf(dst); v.Kind() == reflect.Ptr {
			el := v.Elem()
			switch el.Kind() {
			// if dst is a pointer to pointer, strip the pointer and try again
			case reflect.Ptr:
				if el.IsNil() {
					// allocate destination
					el.Set(reflect.New(el.Type().Elem()))
				}
				return int64AssignTo(srcVal, el.Interface())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if el.OverflowInt(int64(srcVal)) {
					return fmt.Errorf("cannot put %d into %T", srcVal, dst)
				}
				el.SetInt(int64(srcVal))
				return nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if srcVal < 0 {
					return fmt.Errorf("%d is less than zero for %T", srcVal, dst)
				}
				if el.OverflowUint(uint64(srcVal)) {
					return fmt.Errorf("cannot put %d into %T", srcVal, dst)
				}
				el.SetUint(uint64(srcVal))
				return nil
			}
		}
		return fmt.Errorf("cannot assign %v into %T", srcVal, dst)
	}

	return fmt.Errorf("cannot assign %v into %T", srcVal, dst)
}
