package service

import (
	"testing"
)

func TestMaxSKS(t *testing.T) {
	tests := []struct {
		name string
		ipk  float64
		want int
	}{
		{"Batas IPK 4.00", 4.00, 24},
		{"Batas IPK 3.00", 3.00, 24},
		{"Batas IPK 2.99", 2.99, 21},
		{"Batas IPK 2.50", 2.50, 21},
		{"Batas IPK 2.49", 2.49, 18},
		{"Batas IPK 0.00", 0.00, 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxSKS(tt.ipk)
			if got != tt.want {
				t.Errorf("MaxSKS(%.2f) = %d; want %d", tt.ipk, got, tt.want)
			}
		})
	}
}

func TestValidateTahunAkademik(t *testing.T) {
	tests := []struct {
		name    string
		ta      string
		wantErr bool
	}{
		{"Valid Ganjil", "2026/2027-Ganjil", false},
		{"Valid Genap", "2025/2026-Genap", false},
		{"Format Salah Tanpa Semester", "2026/2027", true},
		{"Format Salah Pemisah Strip", "2026-2027-Ganjil", true},
		{"Selisih Tahun Salah (2 Tahun)", "2026/2028-Ganjil", true},
		{"Tahun Kedua Lebih Kecil", "2026/2025-Genap", true},
		{"Semester Tidak Valid", "2026/2027-Pendek", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTahunAkademik(tt.ta)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTahunAkademik(%q) error = %v, wantErr %v", tt.ta, err, tt.wantErr)
			}
		})
	}
}

func TestValidateAngkatan(t *testing.T) {
	tests := []struct {
		name     string
		angkatan int
		wantErr  bool
	}{
		{"Angkatan Valid 2024", 2024, false},
		{"Angkatan Valid 2020", 2020, false},
		{"Angkatan Masa Depan", 2099, true},
		{"Angkatan Kuno", 1850, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAngkatan(tt.angkatan)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAngkatan(%d) error = %v, wantErr %v", tt.angkatan, err, tt.wantErr)
			}
		})
	}
}
