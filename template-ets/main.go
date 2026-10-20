package main 

import (
	"net/http"
	"encoding/json"
	"strings"
	"sync"
	"strconv"
	"time"
)

type SumberDaya struct {
	ID int `json:"id"`
	Nama string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok int `json:"stok"`
}

type Peminjaman struct {
	ID int `json:"id"`
	SumberDayaID int `json:"sumber_daya_id"`
	Peminjam string `json:"peminjam"`
	Tanggal string `json:"tanggal"`
	SlotMulai int `json:"slot_mulai"`
	SlotSelesai int `json:"slot_selesai"`
	Jumlah int `json:"jumlah"`
	Status string `json:"status"`
}

type memoryStore struct {
	mu sync.Mutex
	sumberDaya []SumberDaya
	peminjaman []Peminjaman
	idSumberDaya int
	idPeminjaman int
}

func newMemoryStore() *memoryStore {
	return &memoryStore {
		sumberDaya: []SumberDaya{},
		peminjaman: []Peminjaman{},
		idSumberDaya: 1,
		idPeminjaman: 1,
	}
}

type inputPeminjaman struct {
	SumberDayaID *int `json:"sumber_daya_id"`
	Peminjam *string `json:"peminjam"`
	Tanggal *string `json:"tanggal"`
	SlotMulai *int `json:"slot_mulai"`
	SlotSelesai *int `json:"slot_selesai"`
	Jumlah *int `json:"jumlah"`
}

func NewHandler() http.Handler {
	store := newMemoryStore()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /sumber-daya", store.membuatSumberDaya)
	mux.HandleFunc("GET /sumber-daya", store.daftarSemuaSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", store.detailSumberDaya)

	mux.HandleFunc("POST /peminjaman", store.membuatPeminjaman)
	mux.HandleFunc("GET /peminjaman", store.daftarSemuaPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", store.detailPeminjaman)

	return mux 
}

func (s *memoryStore) membuatSumberDaya(w http.ResponseWriter, r *http.Request) {
	var input SumberDaya
	if err := json.NewDecoder(r.Body).Decode(&input); 
	err != nil {
		writeError(w, http.StatusBadRequest, "request JSON tidak valid")
		return
	}

	input.Nama = strings.TrimSpace(input.Nama)
	input.Jenis = strings.TrimSpace(input.Jenis)

	if input.Nama == "" {
		writeError(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}

	if len([]rune(input.Nama)) > 50 {
		writeError(w, http.StatusBadRequest, "nama tidak boleh lebih dari 50 karakter")
		return
	}

	if input.Jenis != "alat" && input.Jenis != "ruang" {
		writeError(w, http.StatusBadRequest, "jenis harus 'alat' atau 'ruang'")
		return
	}

	if input.Stok < 1 {
		writeError(w, http.StatusBadRequest, "stok harus minimal 1")
		return
	}

	s.mu.Lock()
	input.ID = s.idSumberDaya
	s.idSumberDaya++
	s.sumberDaya = append(s.sumberDaya, input)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}


func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *memoryStore) daftarSemuaSumberDaya(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	daftar := append([]SumberDaya{}, s.sumberDaya...)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(daftar)

}

func (s *memoryStore) detailSumberDaya(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id harus angka")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.sumberDaya {
		if item.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(item)
			return
		}
	}
	writeError(w, http.StatusNotFound, "tidak dapat menemukan sumber daya dengan id tersebut")
}

func tanggalValid(tanggal string) bool {
	parsed, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return false
	}

	return parsed.Format("2006-01-02") == tanggal
}

func (s *memoryStore) membuatPeminjaman(w http.ResponseWriter, r *http.Request) {
	var input inputPeminjaman
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "request JSON tidak valid")
		return
	}

	if input.SumberDayaID == nil || input.Peminjam == nil || input.Tanggal == nil || input.SlotMulai == nil || input.SlotSelesai == nil || input.Jumlah == nil {
		writeError(w, http.StatusBadRequest, "semua field harus diisi")
		return
	}

	peminjam := strings.TrimSpace(*input.Peminjam)
	tanggal := strings.TrimSpace(*input.Tanggal)

	if peminjam == "" {
		writeError(w, http.StatusBadRequest, "peminjam tidak boleh kosong")
		return
	}

	if (len([]rune(peminjam)) > 50) {
		writeError(w, http.StatusBadRequest, "peminjam tidak boleh lebih dari 50 karakter")
		return
	}

	if !tanggalValid(tanggal) {
		writeError(w, http.StatusBadRequest, "tanggal harus dalam format YYYY-MM-DD")
		return
	}

	if *input.SlotMulai < 1 || *input.SlotMulai > 8 {
		*input.SlotSelesai < 1 || *input.SlotSelesai > 8 || *input.SlotMulai >= *input.SlotSelesai {
			writeError(w, http.StatusBadRequest, "slot_mulai dan slot_selesai harus antara 1-8 dan slot_mulai < slot_selesai")
			return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var sumber *SumberDaya
	for i : = range s.sumberDaya {
		if s.sumberDaya[i].ID == *input.SumberDayaID {
			sumber = &s.sumberDaya[i]
			break
		}
	}

	if sumber == nil {
		writeError(w, http.StatusBadRequest, "sumber daya tidak ditemukan")
		return
	}

	
}

func main() {
	http.ListenAndServe(":8080", NewHandler())
}