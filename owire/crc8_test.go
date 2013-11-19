package owire

import "testing"
//import "fmt"

func TestCRC8(t *testing.T) {
	//	dumpTb()

	// crc: a2
	testSeq1 := []byte{ 0x02, 	0x1c, 	0xb8,	0x01, 	0x00, 	0x00,   0x00  }
	// crc: 9c
	testSeq2 := []byte{	0x28,	0xc2,	0x84,	0xbd,	0x01,	0x00,	0x00  }

	crc := New()
	crc.Calc(testSeq1)

	// fmt.Printf("test crc(0xA2): %v\n", crc)
	if crc.CRC() != 0xa2 {
		t.Fail()
	}

	crc.Reset()
	crc.Calc(testSeq2)

	// fmt.Printf("test crc(0x9C): %v\n", crc)
	if (crc.CRC() != 0x9c) {
		t.Fail()
	}
}

func BenchmarkTableGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		genTbReflected(0, OWPOLY_REFL)
	}
}

func BenchmarkROMCRCCalc(b *testing.B) {
	testSeq := []byte{	0x28,	0xc2,	0x84,	0xbd,	0x01,	0x00,	0x00  }
	crc     := New()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crc.Calc(testSeq)
	}
}
