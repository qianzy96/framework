package diy_type

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/totoval/framework/helpers"
	"math/big"
	"strings"
)

// These constants define supported rounding modes.
const (
	ToNearestEven big.RoundingMode = iota // == IEEE 754-2008 roundTiesToEven
	ToNearestAway                         // == IEEE 754-2008 roundTiesToAway
	ToZero                                // == IEEE 754-2008 roundTowardZero
	AwayFromZero                          // no IEEE 754-2008 equivalent
	ToNegativeInf                         // == IEEE 754-2008 roundTowardNegative
	ToPositiveInf                         // == IEEE 754-2008 roundTowardPositive
)

type BF = big.Float
type BigFloat struct {
	BF
	normalCount  uint
	decimalCount uint
}

func (bf BigFloat) Value() (driver.Value, error) {
	//helpers.Dump(bf.BF.Prec(), bf.Text('f', 100), bf.String())
	return []byte(bf.String()), nil
}
func (bf *BigFloat) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return bf.scanBytes(src)
	case string:
		return bf.scanBytes([]byte(src))
	case nil:
		bf = nil
		return nil
	default:
		helpers.DD(src)
	}

	return fmt.Errorf("pq: cannot convert %T to BoolArray", src)
}

func (bf *BigFloat) scanBytes(src []byte) error {
	if err := bf.CreateFromString(string(src), ToNearestEven); err != nil {
		return err
	}
	return nil
}
func (bf *BigFloat) String() string {
	//helpers.Dump(bf.BF.Prec(), bf.BF.MinPrec())
	//if bf.decimalCount == 0 {
	//	return bf.Text('f', 62)
	//}
	return bf.Text('f', int(bf.decimalCount)/2)
}

func (bf *BigFloat) SetDecimal(d uint) {
	bf.decimalCount = d * 2
}

func (bf *BigFloat) CreateFromString(s string, mode big.RoundingMode) error {
	//parse number string
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		// There is no decimal point, we can just parse the original string as
		// an int
		bf.normalCount = uint(len(parts[0])) * 2
		bf.SetDecimal(0)
	} else if len(parts) == 2 {
		// strip the insignificant digits for more accurate comparisons.
		decimalPart := strings.TrimRight(parts[1], "0")
		bf.normalCount = uint(len(parts[0])) * 2
		bf.SetDecimal(uint(len(decimalPart)))
	} else {
		return errors.New("can't convert " + s + " to decimal")
	}

	helpers.Dump(bf.normalCount, bf.decimalCount)

	// string to BigFloat
	_bf, _, err := big.ParseFloat(s, 10, bf.normalCount*2+bf.decimalCount*2+8, mode)
	bf.BF = *_bf
	//bf.SetPrec(prec).SetMode(mode)
	//_, err := fmt.Sscan(s, &bf.BF)
	return err
}

func (bf *BigFloat) useBiggerDecimal(a BigFloat, b BigFloat){
	if a.decimalCount > b.decimalCount {
		bf.decimalCount = a.decimalCount
	}else{
		bf.decimalCount = b.decimalCount
	}
}

func (bf *BigFloat) Add(a BigFloat, b BigFloat) {
	bf.useBiggerDecimal(a, b)
	bf.BF.Add(&a.BF, &b.BF)
}
func (bf *BigFloat) Sub(a BigFloat, b BigFloat) {
	bf.useBiggerDecimal(a, b)
	bf.BF.Sub(&a.BF, &b.BF)
}
func (bf *BigFloat) Mul(a BigFloat, b BigFloat) {
	bf.useBiggerDecimal(a, b)
	bf.BF.Mul(&a.BF, &b.BF)
}
func (bf *BigFloat) Div(a BigFloat, b BigFloat) {
	bf.useBiggerDecimal(a, b)
	bf.BF.Quo(&a.BF, &b.BF)
}
func (bf *BigFloat) Abs(a BigFloat) {
	bf.BF.Abs(&a.BF)
}
func (bf *BigFloat) Cmp (a BigFloat) int {
	return bf.BF.Cmp(&a.BF)
}

//
//func main(){
//	a := BigFloat{}
//	a.SetString("10", 10)
//	b := BigFloat{}
//	b.SetString("11", 10)
//	c := BigFloat{}
//	c.Add(&a.BF, &b.BF)
//}