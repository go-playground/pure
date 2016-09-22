package pure

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Benchmark2Param(b *testing.B) {

	path := "/users/13/contact/2"
	data := []byte("data")

	p := New()
	p.Get("/users/:uid/contact/:cid", func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})

	h := p.Serve()

	b.ResetTimer()

	b.Run("", func(b *testing.B) {
		for i := 0; i < b.N; i++ {

			r, _ := http.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				b.Error("BAD Request")
			}
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {

			for pb.Next() {
				r, _ := http.NewRequest(http.MethodGet, path, nil)
				w := httptest.NewRecorder()

				h.ServeHTTP(w, r)

				if w.Code != http.StatusOK {
					b.Error("BAD Request")
				}
			}
		})
	})
}
