package main

import (
	"log"

	"github.com/jazzy-crane/printer"
)

func main() {
	// เลือกใช้เครื่องพิมพ์เริ่มต้น หรือเปลี่ยนชื่อเครื่องพิมพ์ตามต้องการ
	pName, err := printer.Default()
	if err != nil {
		log.Fatal("ไม่สามารถดึงชื่อเครื่องพิมพ์เริ่มต้น:", err)
	}

	p, err := printer.Open(pName)
	if err != nil {
		log.Fatal("ไม่สามารถเปิดเครื่องพิมพ์:", err)
	}
	defer p.Close()

	docName := "Receipt"
	jobID, err := p.StartDocument(docName, "", "RAW")
	if err != nil {
		log.Fatal("StartDocument ผิดพลาด:", err)
	}

	if err = p.StartPage(); err != nil {
		log.Fatal("StartPage ผิดพลาด:", err)
	}

	receiptText := "ร้าน ABC\nเลขที่ใบเสร็จ: 12345\nราคา: ฿100.00\n"
	_, err = p.Write([]byte(receiptText))
	if err != nil {
		log.Fatal("Write ผิดพลาด:", err)
	}

	if err = p.EndPage(); err != nil {
		log.Fatal("EndPage ผิดพลาด:", err)
	}

	if err = p.EndDocument(); err != nil {
		log.Fatal("EndDocument ผิดพลาด:", err)
	}

	log.Println("พิมพ์สำเร็จ (Job ID:", jobID, ")")
}
