package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding/charmap"
)

var (
	winspool             = syscall.NewLazyDLL("winspool.drv")
	procOpenPrinter      = winspool.NewProc("OpenPrinterW")
	procClosePrinter     = winspool.NewProc("ClosePrinter")
	procStartDocPrinter  = winspool.NewProc("StartDocPrinterW")
	procEndDocPrinter    = winspool.NewProc("EndDocPrinter")
	procStartPagePrinter = winspool.NewProc("StartPagePrinter")
	procEndPagePrinter   = winspool.NewProc("EndPagePrinter")
	procWritePrinter     = winspool.NewProc("WritePrinter")
)

// Configuration
const (
	PRINTER_NAME = "XP-80C"
)

type DOC_INFO_1 struct {
	pDocName    *uint16
	pOutputFile *uint16
	pDatatype   *uint16
}

func main() {
	printerName := syscall.StringToUTF16Ptr(PRINTER_NAME) // เปลี่ยนเป็นชื่อเครื่องพิมพ์จริง
	var hPrinter syscall.Handle

	// OpenPrinter
	r1, _, err := procOpenPrinter.Call(
		uintptr(unsafe.Pointer(printerName)),
		uintptr(unsafe.Pointer(&hPrinter)),
		0,
	)
	if r1 == 0 {
		fmt.Println("OpenPrinter failed:", err)
		return
	}
	defer procClosePrinter.Call(uintptr(hPrinter))

	docName := syscall.StringToUTF16Ptr("Receipt")
	dataType := syscall.StringToUTF16Ptr("RAW") // ส่งข้อมูล ESC/POS ตรงๆ
	di := DOC_INFO_1{
		pDocName:    docName,
		pOutputFile: nil,
		pDatatype:   dataType,
	}

	// StartDocPrinter
	r1, _, err = procStartDocPrinter.Call(
		uintptr(hPrinter),
		1,
		uintptr(unsafe.Pointer(&di)),
	)
	if r1 == 0 {
		fmt.Println("StartDocPrinter failed:", err)
		return
	}
	defer procEndDocPrinter.Call(uintptr(hPrinter))

	// StartPagePrinter
	r1, _, err = procStartPagePrinter.Call(uintptr(hPrinter))
	if r1 == 0 {
		fmt.Println("StartPagePrinter failed:", err)
		return
	}
	defer procEndPagePrinter.Call(uintptr(hPrinter))

	// // เนื้อหาที่จะพิมพ์ (UTF-8 → CP874)
	text := "\n=== ร้าน ABC ราคา: 100.00 บาท ส่งฟรี ถึงบ้าน ===\n\n"
	encoder := charmap.Windows874.NewEncoder()
	thaiData, _ := encoder.Bytes([]byte(text))
	fmt.Printf("text: %x", thaiData)
	// bytesText := []byte(text)
	// // คำสั่ง ESC/POS ตัดกระดาษ (Full cut)
	cancelMultibyte := []byte{0x1C, 0x2E}
	// cut := []byte{0x1B, 0x69}
	// reset := []byte{0x1B, 0x40} // ESC @ (reset)
	// // setFontThai := []byte{0x1B, 0x4D, 0x01} // ESC M 1 (เลือกฟอนต์)
	// setCodePage := []byte{0x1B, 0x74, 26} // ESC t 70 (PC874 THAI)

	// finalData := append(reset, cancelMultibyte...)
	// // finalData = append(finalData, setFontThai...)
	// finalData = append(finalData, setCodePage...)
	// finalData = append(finalData, thaiData...)
	// finalData = append(finalData, cut...)

	// Header
	// header := []byte("Code page " + fmt.Sprintf("%d", n) + "\n  0123456789ABCDEF0123456789ABCDEF\n\n\n")

	// สร้าง string ของ chars 128-255 (extended chars)
	var chars []byte
	for i := 128; i <= 255; i++ {
		chars = append(chars, byte(i))
	}
	reset := []byte{0x1B, 0x40}
	// cut := []byte{0x1B, 0x69}

	finalData := append(reset, cancelMultibyte...)
	// Print rows (แบ่งเป็น 4 แถว, แถวละ 32 chars)
	for x := 0; x < 256; x++ {
		// n := 255 // ลองค่าเช่น 96, 16, 20, etc.
		setCodePage := []byte{0x1B, 0x74, byte(x)}
		finalData = append(finalData, setCodePage...)
		for y := 0; y < 4; y++ {
			rowHeader := []byte(fmt.Sprintf("%X ", y+8)) // เช่น "8 ", "9 ", etc.
			row := chars[y*32 : (y+1)*32]
			finalData = append(finalData, rowHeader...)
			finalData = append(finalData, row...)
			finalData = append(finalData, []byte("\n")...)
		}
	}

	// รวม data: reset + set code page + header + rows + cut

	// finalData = append(finalD
	// ata, header...)
	// แล้ว append rows อย่างข้างบน
	// finalData = append(finalData, cut...)
	var written uint32
	r1, _, err = procWritePrinter.Call(
		uintptr(hPrinter),
		uintptr(unsafe.Pointer(&finalData[0])),
		uintptr(len(finalData)),
		uintptr(unsafe.Pointer(&written)),
	)
	if r1 == 0 {
		fmt.Println("WritePrinter failed:", err)
		return
	}

	fmt.Println("พิมพ์สำเร็จ:", written, "bytes")
}
