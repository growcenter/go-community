package herrx

import "go-community/internal/pkg/locale"

func init() {
	locale.Register(locale.Default, map[string]string{
		"COMMON_DEPRECATED_ENDPOINT": "Fitur ini sudah tidak tersedia. Mohon perbarui aplikasi Anda.",
		"COMMON_INVALID_INPUT":       "Ada isian yang belum benar. Mohon periksa kembali.",
		"COMMON_NOT_FOUND":           "Data yang Anda cari tidak ditemukan.",
		"COMMON_FORBIDDEN":           "Anda tidak memiliki akses untuk melakukan ini.",
		"COMMON_UNAUTHORIZED":        "Silakan masuk terlebih dahulu.",
		"COMMON_INTERNAL":            "Terjadi kesalahan di sisi kami. Silakan coba lagi.",
	})
}
