package service

import (
	"os"
	"os/exec"
	"testing"
)

func TestParsePDFSample(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	f, err := os.Open("../../../sample_data/SPEI092025053100172422 (1).pdf")
	if err != nil {
		t.Fatalf("open sample pdf: %v", err)
	}
	defer f.Close()
	svc := NewAdInvoiceService(nil, nil, nil)
	inv, err := svc.parsePDF(f)
	if err != nil {
		t.Logf("inv: %+v", inv)
		t.Fatalf("parsePDF error: %v", err)
	}
	if inv.InvoiceNo != "SPEI092025053100172422" {
		t.Errorf("invoice number = %s", inv.InvoiceNo)
	}
	if inv.InvoiceDate.Format("02/01/2006") != "31/05/2025" {
		t.Errorf("invoice date = %s", inv.InvoiceDate.Format("02/01/2006"))
	}
	if inv.Total != 4162500.00 {
		t.Errorf("total = %f", inv.Total)
	}
}

func TestParsePDFSampleMultiLine(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	f, err := os.Open("../../../sample_data/SPEI092024093000132653.pdf")
	if err != nil {
		t.Fatalf("open sample pdf: %v", err)
	}
	defer f.Close()
	svc := NewAdInvoiceService(nil, nil, nil)
	inv, err := svc.parsePDF(f)
	if err != nil {
		t.Logf("inv: %+v", inv)
		t.Fatalf("parsePDF error: %v", err)
	}
	if inv.InvoiceNo != "SPEI092024093000132653" {
		t.Errorf("invoice number = %s", inv.InvoiceNo)
	}
	if inv.InvoiceDate.Format("02/01/2006") != "30/09/2024" {
		t.Errorf("invoice date = %s", inv.InvoiceDate.Format("02/01/2006"))
	}
	if inv.Total != 2073150.00 {
		t.Errorf("total = %f", inv.Total)
	}
}

func TestParseSplitInvoiceNumber(t *testing.T) {
	lines := []string{
		"Faktur",
		"No. Faktur",
		"SPEI09202408",
		"3100129594",
		"Username",
		"Pelanggan",
		"mrest0re",
		"Tanggal Invoice",
		"31/08/2024",
		"Total",
		"(Termasuk PPN jika ada)",
		"1,691,420.00",
	}
	inv := parseInvoiceText(lines)
	if inv.InvoiceNo != "SPEI092024083100129594" {
		t.Errorf("invoice number = %s", inv.InvoiceNo)
	}
	if inv.InvoiceDate.Format("02/01/2006") != "31/08/2024" {
		t.Errorf("invoice date = %s", inv.InvoiceDate.Format("02/01/2006"))
	}
	if inv.Total != 1691420.00 {
		t.Errorf("total = %f", inv.Total)
	}
}

func TestParseSecondSampleSplit(t *testing.T) {
	lines := []string{
		"Faktur",
		"No. Faktur",
		"SPEI09202407",
		"3100117166",
		"Username",
		"Pelanggan",
		"mrest0re",
		"Tanggal Invoice",
		"31/07/2024",
		"Total",
		"(Termasuk PPN jika ada)",
		"220,900.00",
	}
	inv := parseInvoiceText(lines)
	if inv.InvoiceNo != "SPEI092024073100117166" {
		t.Errorf("invoice number = %s", inv.InvoiceNo)
	}
	if inv.InvoiceDate.Format("02/01/2006") != "31/07/2024" {
		t.Errorf("invoice date = %s", inv.InvoiceDate.Format("02/01/2006"))
	}
	if inv.Total != 220900.00 {
		t.Errorf("total = %f", inv.Total)
	}
}

func TestParseTotalSameLine(t *testing.T) {
	lines := []string{
		"Faktur",
		"No. Faktur",
		"SPEI092025053100289327",
		"Username",
		"Pelanggan",
		"mrest0re",
		"Tanggal Invoice",
		"31/05/2025",
		"Total (Termasuk PPN jika ada) 1,310,750.00",
	}
	inv := parseInvoiceText(lines)
	if inv.InvoiceNo != "SPEI092025053100289327" {
		t.Errorf("invoice number = %s", inv.InvoiceNo)
	}
	if inv.InvoiceDate.Format("02/01/2006") != "31/05/2025" {
		t.Errorf("invoice date = %s", inv.InvoiceDate.Format("02/01/2006"))
	}
	if inv.Total != 1310750.00 {
		t.Errorf("total = %f", inv.Total)
	}
}

func TestAdInvoiceAmountPositive(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	files := []string{
		"../../../sample_data/SPEI092025053100172422 (1).pdf",
	}
	svc := NewAdInvoiceService(nil, nil, nil)
	for _, fp := range files {
		f, err := os.Open(fp)
		if err != nil {
			t.Fatalf("open %s: %v", fp, err)
		}
		inv, err := svc.parsePDF(f)
		f.Close()
		if err != nil {
			t.Fatalf("parse %s: %v", fp, err)
		}
		if inv.Total <= 0 {
			t.Errorf("invoice %s total <= 0", fp)
		}
	}
}
