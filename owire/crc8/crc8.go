// DALLAS 1-Wire crc8 table based implementation
// 
// 2013 Copyleft Alex Karev
//
// Implementation of table driven crc8 algorithm for (ROM) DALLAS 1-Wire protocol
// 
// Reflected algorithm used, as in [AppNote27], but basis for direct 
// implementation also provided
//
//
// References:
//
// [AppNote27] - DALLAS Application note 27
//
package crc8

import "fmt"

const (
	OWPOLY        = 0x31 // Dallas 1-wire crc8 poly:  x^8 + x^5 + x^4 + x^0
	OWPOLY_REFL   = 0x8c // Reflected poly:           x^0 + x^4 + x^5 + x^8

	BITS_PER_BYTE = 8
	TB_SIZE       = 1 << BITS_PER_BYTE
)

//
// Context interface section
//
type CRCContext interface {
	Calc(m []byte) byte
	Reset()
	CRC() byte
}

type CRC8TableContext struct {
	tb []byte
	crc byte
}

func (tcnx *CRC8TableContext) Calc(msg []byte) byte {
	tcnx.crc = calc( tcnx.crc, msg, tcnx.tb)
	return tcnx.crc
}

func (tcnx *CRC8TableContext) Reset() {
	tcnx.crc = 0
}

func (tcnx *CRC8TableContext) CRC() byte {
	return tcnx.crc
}

func (tcnx *CRC8TableContext) String() string {
	return fmt.Sprintf( "%02x", tcnx.crc)
}

func New() *CRC8TableContext {
	return &CRC8TableContext{ tb : genTbReflected(0, OWPOLY_REFL) }
}



//
// Private section
//
func calc(init byte, in []byte, tb []byte) (crc byte) {
	crc = init
	for _,v := range in { crc = tb[v ^ crc] }
	return
}

func genTb(init, poly byte) (tb []byte) {

	tb = make([]byte, TB_SIZE )

	var crc byte
	var zfl bool
	for i := 0; i < TB_SIZE; i++ {
		
		crc = byte(i)
		for j := 0; j < 8; j++ {

			// save value of hiest bit
			zfl = crc & 0x80 == 0
			crc <<= 1

			switch zfl { // check: should we select polynom or just shift crc register value
			case true: // zero flag = 1
			case false: // zero flag = 0
				crc ^= poly
			}

		}

		tb[i] = crc
	}

	return tb
}

func genTbReflected(init, poly byte) (tb []byte) {

	tb = make([]byte, TB_SIZE )

	var crc byte
	var zfl bool
	for i := 0; i < TB_SIZE; i++ {
		
		crc = byte(i)
		for j := 0; j < 8; j++ {

			// save value of hiest bit
			zfl = crc & 1 == 0
			crc >>= 1

			switch zfl { // check: should we select polynom or just shift crc register value
			case true: // zero flag = 1
			case false: // zero flag = 0
				crc ^= poly
			}

		}

		tb[i] = crc
	}

	return tb
}
