/*

Low level MP707 usb termometer functions

Based on protocol implemetation in:
 http://code.google.com/p/bmcontrol/downloads/detail?name=bmcontrol-1.1.tar.gz

*/

package mp707


/*
 #cgo LDFLAGS: -lusb-1.0
 #include <stdint.h>
 #include <libusb-1.0/libusb.h>

// struct foo {
//   int a;
// };

 libusb_device *get_dev(libusb_device **list, int offset) {
    return list[offset];
 }
*/
import "C"

import (
	"fmt"
	"container/list" 
//	"os"
)

const (
	VENDOR_ID   = 0x16c0
	PRODUCT_ID  = 0x05df
	TIMEOUT_CTL = 100     // ms
	WR_RETRIES  = 5       // from bmcontrol
	ROM_PORTS   = 64      // from bmcontrol
	CMD_SIZE    = 8       // from bmcontrol
)

type UsbErrorCode int
type UsbErrorDesc string

type UsbError struct {
	code UsbErrorCode
	desc UsbErrorDesc
}

func MakeUsbError(code C.int) *UsbError {
	return &UsbError{
		UsbErrorCode(code) , 
		UsbErrorDesc(C.GoString(C.libusb_error_name(code))) }
}

func (ue UsbError) Error() string {
	return fmt.Sprintf("Error(%d): %s", ue.code, ue.desc)
}


type MP707Dev struct {
	id uint8 
	dev *C.struct_libusb_device_handle
	outb [CMD_SIZE]byte
	inb  [CMD_SIZE]byte
}

func (t *MP707Dev) clOutBuff() {
	for _, v = range t.outb { v = 0	}
}

func (t *MP707Dev) clInBuff() {
	for _, v = range t.nb {	v = 0 }
}

func (t *MP707Dev) clearBuffers() {
	t.clOutBuff()
	t.clInBuff ()
}

func InitLib() bool {
	v := C.libusb_init(nil)
	if v != 0 {
		return false
	} 

	return true
}

func DesposeLib() {
	C.libusb_exit(nil)
}

func extractDesc(dev *C.libusb_device) (id uint8, vendor int, product int, err *UsbError) {
	desc := &C.struct_libusb_device_descriptor{}
	//var a C.struct_foo

	if r := C.libusb_get_device_descriptor(dev, desc); r != C.int(0) {
	 	return 0, 0, 0, MakeUsbError( C.int(r) )
	}

	return uint8(desc.iSerialNumber), int(desc.idVendor), int(desc.idProduct), nil
}

func Lookup() (*list.List, *UsbError) {
	var li **C.struct_libusb_device

	n := C.libusb_get_device_list( nil, &li)
	if n <= 0 {
		return nil, MakeUsbError(C.int(n))
	}

	defer C.libusb_free_device_list( li, 1)

	// fmt.Println("Devices => ", n)

	mp707List := list.New()

	var ue *UsbError
	for i := 0; i < int(n); i++ {
		//fmt.Printf("Value: %v\n", C.get_dev(list, C.int(i)))
		if id, vendor, product, err := extractDesc(C.get_dev(li, C.int(i))); err != nil {
			// fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			ue = err
			break
		} else if vendor == VENDOR_ID && product == PRODUCT_ID {
			//fmt.Printf("Vendor => %04x, Product => %04x\n", vendor, product ) 
			handle  := &C.struct_libusb_device_handle{}
			if err  := C.libusb_open( C.get_dev(li, C.int(i)), &handle ); err != 0 {
				// fmt.Fprintf(os.Stderr, "Open error: %s\n", MakeUsbError(err))
				ue = MakeUsbError(err)
				continue
			}

			mp707List.PushBack( &MP707Dev{ id: id, dev: handle } )
		}
	}

	// fmt.Println("List size: ", mp707List.Len() )

	return mp707List, ue
}


// bmRequest bits config
// 
// 12345678
// -        Transfer direction: 0 host->device, 1 device -> host
//  --      Type: 0 std, 1 class, 2 vendor, 3 reserved
//    ----- Recipient: 0 device, 1 if, 2 endpoint, 3 other
//
// USB_SET_FEATURE
func (t *MP707Dev) usbWriteCtl() *UsbError {
	if r := C.libusb_control_transfer( t.dev, 0x21, 0x09, 0x300, 0, t.outb, len(t.outb), TIMEOUT ); r != 0 {
		return MakeUsbError(r)
	}

	return nil
}

// USB_GET_FEATURE
func (t *MP707Dev) usbReadCtl() *UsbError {
	if r := C.libusb_control_transfer( t.dev, 0xA1, 0x01, 0x300, 0, t.inb, len(t.inb), TIMEOUT ); r != 0 {
		return MakeUsbError(r)
	}

	return nil
}



// OW_RESET
func (t *MP707Dev) Reset() bool {
	success := fals

	t.clOutBuff()

	t.outb[0] = 0x18 // 0001 1000
	t.outb[1] = 0x48 // 0100 1000

	for i := 0; i < WR_RETRIES; i++ {

		if err := t.usbWriteCtl(); err != nil {
			continue
		} 
		
		time.Sleep( 1 * time.Millisecond )

		if err := t.usbReadCtl(); err != nil {
			continue
		} 
		
		if t.inb[0] == 0x18 && t.inb[1] == 0x48 && t.inb[2] == 0x00 {
			success = true
			break
		}
	}

	return success
}

func (t *MP707Dev) WriteByte(d byte) bool {
	result := false
	t.clearBuffers()

	t.outb[0] = 0x18
	t.outb[1] = 0x88
	t.outb[2] = d

	if !t.usbWriteCtl() {
		return result
	}

	time.Sleep( 1 * time.Millisecond )
	
	if !t.usbReadCtl() {
		return result
	}

	return result = t.inb[0] == 0x18 && t.inb[1] == 0x88 && t.inb[2] == d
}


func (t *MP707Dev) Read2Bit (bit byte, err bool) {

	t.clearBuffers()

	t.outb[0] = 0x18
	t.outb[1] = 0x82
	t.outb[2] = 0x01
	t.outb[3] = 0x01

	if !t.usbWriteCtl() {
		return false
	}

	time.Sleep( 1 * time.Millisecond )
	
	if !t.usbReadCtl() {
		return false
	}

	result := t.inb[0] == 0x18 && t.inb[1] == 0x82
	return t.inb[2] + t.inb[3] << 1, result
}

func (t *MP707Dev) WriteBit(bit byte) bool {
//						if !t.WriteBit(bit) {
	t.clearBuffers()

	t.outb[0] = 0x18
	t.outb[1] = 0x81
	t.outb[2] = bit & 0x01

	if !t.usbWriteCtl() {
		return false
	}

	time.Sleep( 1 * time.Millisecond )
	
	if !.usbReadCtl() {
		return false 
	}

	return t.inb[0] == 0x18 && t.inb[1] == 0x81 && t.inb[2] == bit & 0x01
}

func crc8(crc byte, d byte) byte {
	r := crc
	
	for i := 0; i < 8; i++ {
		if r^(d>>i) == 1 {
			r = r ^ 0x18 >> 1 | 0x80
		} else {
			r = r >> 1 & 0x7f
		}
	}

	return r
}
	

// SEARCH_ROM
func (t *MP707Dev) SearchRom(uint64 rom_next, uint32 pl) bool {

	result := false
	cl     := make( []bool,   ROM_PORTS )
	rl     := make( []uint64, ROM_PORTS )

	var bit uint8
	var err bool
	var rom,crc uint64

	for n := 0; n < WR_RETRIES && !result; n++ {
		if t.Reset() {
			result = t.WriteByte(0xF0)
		}

		if result {
			for i := 0; i < ROM_PORTS; i++ {
				if result {

					if bit, err = t.Read2Bit(); !err {
						switch bit & 0x03 {
						case 0: // collision?
							if pl < i {
								cl[i] = true
								rl[i] = rom
								bit   = 0
							} else { // pl >= i
								bit = rom_next>>i & 1
							}

							if !t.WriteBit(bit) {
								result = false
								// i      = ROM_PORTS
								break
							}

							if bit == 1 {
								rom += 1<<i
							}

						case 1:
							if !t.WriteBit(1) {
								result = false
								//i      = ROM_PORTS
								break
							} else {
								rom += (1<<i)
							}
						case 2:
							if !t.WriteBit(0) {
								result = false
								//i      = ROM_PORTS
								break
							}
						case 3:
							result = false
							//i      = ROM_PORTS
							break
						}
					} else { // if Read2Bit
						result = false 
						break
					} 
				} // if result
			} // for i..ROM_PORTS
		} // if result
		
		if rom == 0 {
			result = false
		}

		if result {
			crc = 0
			for j := 0; j < 8; j++ {
				crc = crc8(crc, (rom >> (j<<3)) & 0xff )
				result = crc == 0
			}
		}
	} // for retries
	
	if !result {
		return result
	} else {
		//t.
	}

	for i := 0; i < ROM_PORTS; i++ {
		if cl[i] {
			t.SearchRom( rl[i] | 1<<i, i )
		}
	}

	return result
}
