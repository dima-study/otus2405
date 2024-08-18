package hw10programoptimization

import (
	"archive/zip"
	"testing"
)

// go test -v -count=1 -timeout=30s -tags bench .
func BenchmarkGetDomainStat(b *testing.B) {
	for range b.N {
		r, _ := zip.OpenReader("testdata/users.dat.zip")
		defer r.Close()

		data, _ := r.File[0].Open()

		b.StartTimer()
		stat, _ := GetDomainStat(data, "biz")
		b.StopTimer()

		_ = stat
	}
}
