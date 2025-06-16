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
