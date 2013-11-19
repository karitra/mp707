package mp707
import ( 
	"testing"
//	"fmt"
)

func TestLibUSB(t *testing.T) {

	if r := InitLib(); !r {
		t.Error("Failed to init libusb")
	}

	Lookup()
	DesposeLib()
}
