package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MaxSKS menghitung batas SKS berdasarkan IPK mahasiswa.
// Rules:
// - IPK >= 3.00: 24 SKS
// - 2.50 <= IPK < 3.00: 21 SKS
// - IPK < 2.50: 18 SKS
func MaxSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

var tahunAkademikRegex = regexp.MustCompile(`^(\d{4})/(\d{4})-(Ganjil|Genap)$`)

// ValidateTahunAkademik memastikan format tahun akademik valid (mis. "2026/2027-Ganjil")
// dan tahun kedua selisih 1 tahun dari tahun pertama.
func ValidateTahunAkademik(ta string) error {
	matches := tahunAkademikRegex.FindStringSubmatch(strings.TrimSpace(ta))
	if len(matches) != 4 {
		return fmt.Errorf("format tahun_akademik tidak valid (contoh yang benar: '2026/2027-Ganjil')")
	}

	y1, _ := strconv.Atoi(matches[1])
	y2, _ := strconv.Atoi(matches[2])

	if y2 != y1+1 {
		return fmt.Errorf("tahun kedua (%d) harus tepat satu tahun setelah tahun pertama (%d)", y2, y1)
	}

	return nil
}

// ValidateAngkatan memastikan angkatan berupa 4 digit angka dan <= tahun berjalan.
func ValidateAngkatan(angkatan int) error {
	tahunBerjalan := time.Now().Year()
	if angkatan < 1900 || angkatan > tahunBerjalan {
		return fmt.Errorf("angkatan harus 4 digit angka dan tidak boleh melebihi tahun berjalan (%d)", tahunBerjalan)
	}
	return nil
}
