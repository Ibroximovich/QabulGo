package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1. Clinic Search - Status filtering, only approved clinics appear
func TestPublicSearch_Clinics_StatusFilter(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	lat := 41.311081
	lon := 69.240562

	// Approved Clinic
	cApproved := &models.Clinic{
		ID:          uuid.New(),
		Name:        "Shifo Nur Med Approved",
		Description: "A'lo darajadagi xususiy klinika",
		Phone:       "+998712001122",
		Email:       "info@shifonur.uz",
		Region:      "Toshkent shahri",
		District:    "Yunusobod tumani",
		Address:     "Amir Temur ko'chasi 1",
		Status:      models.ClinicStatusApproved,
		Latitude:    &lat,
		Longitude:   &lon,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cApproved, nil)

	// Pending Clinic
	cPending := &models.Clinic{
		ID:          uuid.New(),
		Name:        "Kutilayotgan Klinika",
		Phone:       "+998712003344",
		Region:      "Toshkent shahri",
		District:    "Chilonzor tumani",
		Address:     "Bunyodkor ko'chasi 5",
		Status:      models.ClinicStatusPending,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cPending, nil)

	// Rejected Clinic
	cRejected := &models.Clinic{
		ID:          uuid.New(),
		Name:        "Rad Etilgan Klinika",
		Phone:       "+998712005566",
		Region:      "Samarqand viloyati",
		District:    "Samarqand sh.",
		Address:     "Registon ko'chasi 10",
		Status:      models.ClinicStatusRejected,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cRejected, nil)

	// Suspended Clinic
	cSuspended := &models.Clinic{
		ID:          uuid.New(),
		Name:        "To'xtatilgan Klinika",
		Phone:       "+998712007788",
		Region:      "Farg'ona viloyati",
		District:    "Farg'ona sh.",
		Address:     "Mustaqillik ko'chasi 20",
		Status:      models.ClinicStatusSuspended,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cSuspended, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/clinics", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("kutilgan status 200, olindi %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                      `json:"success"`
		Data    []dto.PublicClinicItemDTO `json:"data"`
		Meta    struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Meta.Total != 1 {
		t.Fatalf("kutilgan jami approved klinika 1 ta, olindi %d", resp.Meta.Total)
	}

	if len(resp.Data) != 1 || resp.Data[0].ID != cApproved.ID {
		t.Fatalf("faqat approved klinika chiqishi kerak edi, olindi: %+v", resp.Data)
	}
}

// 2. Clinic Search - Filters (region, district, specialty, search, pagination)
func TestPublicSearch_Clinics_Filters(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	specCardio := &models.Specialty{ID: uuid.New(), Name: "Kardiologiya", Slug: "kardiologiya"}
	specDentist := &models.Specialty{ID: uuid.New(), Name: "Stomatologiya", Slug: "stomatologiya"}
	_ = app.SpecialtyRepo.Create(nil, specCardio)
	_ = app.SpecialtyRepo.Create(nil, specDentist)

	c1 := &models.Clinic{
		ID:          uuid.New(),
		Name:        "Toshkent Kardiocenter",
		Description: "Yurak qon tomir kasalliklari markazi",
		Phone:       "+998712001111",
		Region:      "Toshkent shahri",
		District:    "Yunusobod tumani",
		Address:     "Amir Temur ko'chasi 100",
		Status:      models.ClinicStatusApproved,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, c1, nil)

	c2 := &models.Clinic{
		ID:          uuid.New(),
		Name:        "Samarqand Dental Shifo",
		Description: "Tish davolash va protezlash",
		Phone:       "+998662002222",
		Region:      "Samarqand viloyati",
		District:    "Urgut tumani",
		Address:     "Navoiy ko'chasi 50",
		Status:      models.ClinicStatusApproved,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, c2, nil)

	// Shifokorlarni bog'laymiz
	doc1 := &models.Doctor{ID: uuid.New(), FirstName: "Ali", LastName: "Valiyev", ExperienceYears: 10}
	_, _ = app.DoctorRepo.CreateDoctorWithClinic(nil, c1.ID, doc1, specCardio.ID, 150000, 30, true)

	doc2 := &models.Doctor{ID: uuid.New(), FirstName: "Hasan", LastName: "Husanov", ExperienceYears: 8}
	_, _ = app.DoctorRepo.CreateDoctorWithClinic(nil, c2.ID, doc2, specDentist.ID, 80000, 20, true)

	// Viloyat bo'yicha filter
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/clinics?region=Toshkent+shahri", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	var respRegion struct {
		Data []dto.PublicClinicItemDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respRegion)
	if len(respRegion.Data) != 1 || respRegion.Data[0].ID != c1.ID {
		t.Errorf("viloyat filtri to'g'ri ishlamadi: %+v", respRegion.Data)
	}

	// Mutaxassislik bo'yicha filter
	reqSpec, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics?specialty_id=%s", specDentist.ID), nil)
	wSpec := httptest.NewRecorder()
	app.Router.ServeHTTP(wSpec, reqSpec)
	var respSpec struct {
		Data []dto.PublicClinicItemDTO `json:"data"`
	}
	_ = json.Unmarshal(wSpec.Body.Bytes(), &respSpec)
	if len(respSpec.Data) != 1 || respSpec.Data[0].ID != c2.ID {
		t.Errorf("mutaxassislik filtri to'g'ri ishlamadi: %+v", respSpec.Data)
	}

	// Qidiruv (nomi bo'yicha)
	reqSearch, _ := http.NewRequest(http.MethodGet, "/api/v1/public/clinics?search=dental", nil)
	wSearch := httptest.NewRecorder()
	app.Router.ServeHTTP(wSearch, reqSearch)
	var respSearch struct {
		Data []dto.PublicClinicItemDTO `json:"data"`
	}
	_ = json.Unmarshal(wSearch.Body.Bytes(), &respSearch)
	if len(respSearch.Data) != 1 || respSearch.Data[0].ID != c2.ID {
		t.Errorf("qidiruv filtri to'g'ri ishlamadi: %+v", respSearch.Data)
	}
}

// 3. Clinic Details & Clinic Doctors
func TestPublicSearch_ClinicDetails(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	lat := 41.311081
	lon := 69.240562

	cApproved := &models.Clinic{
		ID:          uuid.New(),
		Name:        "City Med Markazi",
		Phone:       "+998712008899",
		Region:      "Toshkent shahri",
		District:    "Yunusobod tumani",
		Address:     "Bodomzor ko'chasi 15",
		Status:      models.ClinicStatusApproved,
		Latitude:    &lat,
		Longitude:   &lon,
		OwnerID:     ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cApproved, []models.ClinicWorkingHour{
		{ClinicID: cApproved.ID, Weekday: 1, OpeningTime: "08:00", ClosingTime: "18:00", IsClosed: false},
	})

	cPending := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Pending Clinic",
		Phone:    "+998712001100",
		Region:   "Toshkent shahri",
		District: "Olmazor tumani",
		Address:  "Qorasaroy 1",
		Status:   models.ClinicStatusPending,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cPending, nil)

	// Approved klinika tafsilotlari - 200
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s", cApproved.ID), nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("kutilgan status 200, olindi %d", w.Code)
	}

	var detailResp struct {
		Data dto.ClinicDetailDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &detailResp)
	if detailResp.Data.Name != "City Med Markazi" || len(detailResp.Data.WorkingHours) != 1 {
		t.Errorf("klinika tafsilotlari noto'g'ri: %+v", detailResp.Data)
	}

	// Pending klinika tafsilotlari - 404
	reqPending, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s", cPending.ID), nil)
	wPending := httptest.NewRecorder()
	app.Router.ServeHTTP(wPending, reqPending)
	if wPending.Code != http.StatusNotFound {
		t.Fatalf("pending klinika uchun kutilgan status 404, olindi %d", wPending.Code)
	}

	// Shifokorlar ro'yxati
	reqDocs, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s/doctors", cApproved.ID), nil)
	wDocs := httptest.NewRecorder()
	app.Router.ServeHTTP(wDocs, reqDocs)
	if wDocs.Code != http.StatusOK {
		t.Fatalf("klinika shifokorlari uchun kutilgan 200, olindi %d", wDocs.Code)
	}
}

// 4. Nearby Clinics Search - Haversine, Radius, and Sorting
func TestPublicSearch_NearbyClinics(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	// Markaz: Toshkent Amir Temur xiyoboni (41.311081, 69.279737)
	centerLat := 41.311081
	centerLon := 69.279737

	// 1. Yaqin klinika (~2.3 km)
	latNear := 41.332000
	lonNear := 69.282000
	cNear := &models.Clinic{
		ID:        uuid.New(),
		Name:      "Yaqin Klinika Minor",
		Phone:     "+998712001111",
		Region:    "Toshkent shahri",
		District:  "Yunusobod",
		Address:   "Minor",
		Status:    models.ClinicStatusApproved,
		Latitude:  &latNear,
		Longitude: &lonNear,
		OwnerID:   ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cNear, nil)

	// 2. O'rta masofadagi klinika (~30 km)
	latMid := 41.468000
	lonMid := 69.582000
	cMid := &models.Clinic{
		ID:        uuid.New(),
		Name:      "Chirchiq Med Center",
		Phone:     "+998702002222",
		Region:    "Toshkent viloyati",
		District:  "Chirchiq",
		Address:   "Chirchiq markazi",
		Status:    models.ClinicStatusApproved,
		Latitude:  &latMid,
		Longitude: &lonMid,
		OwnerID:   ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cMid, nil)

	// 3. Koordinatasi yo'q klinika (chiqmasligi shart)
	cNoCoords := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Koordinatasiz Klinika",
		Phone:    "+998712003333",
		Region:   "Toshkent shahri",
		District: "Yunusobod",
		Address:  "Kiritilmagan",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cNoCoords, nil)

	// 4. Yaqin, lekin Pending holatdagi klinika (chiqmasligi shart)
	latPending := 41.312000
	lonPending := 69.280000
	cPending := &models.Clinic{
		ID:        uuid.New(),
		Name:      "Yaqin Pending Klinika",
		Phone:     "+998712004444",
		Region:    "Toshkent shahri",
		District:  "Yunusobod",
		Address:   "Markaz",
		Status:    models.ClinicStatusPending,
		Latitude:  &latPending,
		Longitude: &lonPending,
		OwnerID:   ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cPending, nil)

	// Radius 10 km bo'yicha qidiruv -> Faqat cNear chiqishi kerak
	url10km := fmt.Sprintf("/api/v1/public/clinics/nearby?latitude=%f&longitude=%f&radius_km=10", centerLat, centerLon)
	req10, _ := http.NewRequest(http.MethodGet, url10km, nil)
	w10 := httptest.NewRecorder()
	app.Router.ServeHTTP(w10, req10)

	if w10.Code != http.StatusOK {
		t.Fatalf("kutilgan status 200, olindi %d: %s", w10.Code, w10.Body.String())
	}

	var resp10 struct {
		Data []dto.NearbyClinicItemDTO `json:"data"`
	}
	_ = json.Unmarshal(w10.Body.Bytes(), &resp10)

	if len(resp10.Data) != 1 || resp10.Data[0].ID != cNear.ID {
		t.Fatalf("10 km radiusda faqat cNear chiqishi kerak edi, olindi: %d ta", len(resp10.Data))
	}
	if resp10.Data[0].DistanceKM <= 0 || resp10.Data[0].DistanceKM > 5.0 {
		t.Errorf("kutilgan masofa taxminan 2-3 km, olindi: %f", resp10.Data[0].DistanceKM)
	}

	// Radius 50 km bo'yicha qidiruv -> cNear (yaqinroq) va cMid (uzoqroq) chiqishi kerak
	url50km := fmt.Sprintf("/api/v1/public/clinics/nearby?latitude=%f&longitude=%f&radius_km=50", centerLat, centerLon)
	req50, _ := http.NewRequest(http.MethodGet, url50km, nil)
	w50 := httptest.NewRecorder()
	app.Router.ServeHTTP(w50, req50)

	var resp50 struct {
		Data []dto.NearbyClinicItemDTO `json:"data"`
	}
	_ = json.Unmarshal(w50.Body.Bytes(), &resp50)

	if len(resp50.Data) != 2 {
		t.Fatalf("50 km radiusda 2 ta klinika chiqishi kerak edi, olindi: %d ta", len(resp50.Data))
	}
	if resp50.Data[0].ID != cNear.ID || resp50.Data[1].ID != cMid.ID {
		t.Errorf("klinikalar masofa bo'yicha saralanmagan: %+v", resp50.Data)
	}

	// Noto'g'ri koordinatalar uchun 400 Bad Request
	reqInvalid, _ := http.NewRequest(http.MethodGet, "/api/v1/public/clinics/nearby?latitude=100&longitude=500", nil)
	wInvalid := httptest.NewRecorder()
	app.Router.ServeHTTP(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("noto'g'ri koordinata uchun kutilgan 400, olindi %d", wInvalid.Code)
	}
}

// 5. Doctor Search - Active vs Inactive, Multiple Clinics, Sorting and Price Filter
func TestPublicSearch_Doctors(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	specNevro := &models.Specialty{ID: uuid.New(), Name: "Nevrolog", Slug: "nevrolog"}
	_ = app.SpecialtyRepo.Create(nil, specNevro)

	cA := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Klinika Alfa",
		Region:   "Toshkent shahri",
		District: "Yunusobod",
		Address:  "Ko'cha A",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cA, nil)

	cB := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Klinika Beta",
		Region:   "Toshkent shahri",
		District: "Chilonzor",
		Address:  "Ko'cha B",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, cB, nil)

	// Shifokor 1: ikkala klinikada ishlaydi (narxlari har xil)
	docMulti := &models.Doctor{
		ID:              uuid.New(),
		FirstName:       "Jasur",
		LastName:        "Karimov",
		ExperienceYears: 12,
		Bio:             "Yuqori toifali nevropatolog",
	}
	cdA, _ := app.DoctorRepo.CreateDoctorWithClinic(nil, cA.ID, docMulti, specNevro.ID, 120000, 30, true)
	cdB, _ := app.DoctorRepo.CreateDoctorWithClinic(nil, cB.ID, docMulti, specNevro.ID, 180000, 45, true)

	// Shifokor 2: nofaol (is_active = false)
	docInactive := &models.Doctor{
		ID:              uuid.New(),
		FirstName:       "Nofaol",
		LastName:        "Shifokor",
		ExperienceYears: 5,
	}
	_, _ = app.DoctorRepo.CreateDoctorWithClinic(nil, cA.ID, docInactive, specNevro.ID, 90000, 20, false)

	// Shifokor 3: soft-deleted
	docDeleted := &models.Doctor{
		ID:              uuid.New(),
		FirstName:       "O'chirilgan",
		LastName:        "Shifokor",
		ExperienceYears: 3,
	}
	cdDel, _ := app.DoctorRepo.CreateDoctorWithClinic(nil, cA.ID, docDeleted, specNevro.ID, 70000, 15, true)
	now := time.Now()
	cdDel.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	_ = app.DoctorRepo.UpdateClinicDoctor(nil, cdDel, docDeleted)

	// Umumiy qidiruv -> faqat docMulti (2 ta ish joyi bilan) chiqishi kerak
	reqAll, _ := http.NewRequest(http.MethodGet, "/api/v1/public/doctors", nil)
	wAll := httptest.NewRecorder()
	app.Router.ServeHTTP(wAll, reqAll)

	var respAll struct {
		Data []dto.PublicDoctorItemDTO `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	_ = json.Unmarshal(wAll.Body.Bytes(), &respAll)

	if respAll.Meta.Total != 2 {
		t.Fatalf("kutilgan natijalar soni 2 ta (jasur ikkita klinikada), olindi %d", respAll.Meta.Total)
	}

	// Narx bo'yicha filter: min_price = 150000 -> faqat Klinika Beta (180000) chiqishi kerak
	reqPrice, _ := http.NewRequest(http.MethodGet, "/api/v1/public/doctors?min_price=150000", nil)
	wPrice := httptest.NewRecorder()
	app.Router.ServeHTTP(wPrice, reqPrice)

	var respPrice struct {
		Data []dto.PublicDoctorItemDTO `json:"data"`
	}
	_ = json.Unmarshal(wPrice.Body.Bytes(), &respPrice)

	if len(respPrice.Data) != 1 || respPrice.Data[0].ClinicDoctorID != cdB.ID {
		t.Errorf("narx filtri to'g'ri ishlamadi: %+v", respPrice.Data)
	}

	// Shifokor batafsil ma'lumotlari: GET /api/v1/public/doctors/:id
	reqDetail, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/doctors/%s", docMulti.ID), nil)
	wDetail := httptest.NewRecorder()
	app.Router.ServeHTTP(wDetail, reqDetail)

	if wDetail.Code != http.StatusOK {
		t.Fatalf("kutilgan status 200, olindi %d", wDetail.Code)
	}

	var respDetail struct {
		Data dto.PublicDoctorDetailDTO `json:"data"`
	}
	_ = json.Unmarshal(wDetail.Body.Bytes(), &respDetail)

	if len(respDetail.Data.Workplaces) != 2 {
		t.Errorf("shifokorning ikkala klinikadagi ish joyi ko'rinishi kerak, olindi %d ta", len(respDetail.Data.Workplaces))
	}

	// ClinicDoctorID orqali ham qidirish ishlaydi
	reqDetailByCdID, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/doctors/%s", cdA.ID), nil)
	wDetailByCdID := httptest.NewRecorder()
	app.Router.ServeHTTP(wDetailByCdID, reqDetailByCdID)

	if wDetailByCdID.Code != http.StatusOK {
		t.Fatalf("clinic_doctor_id orqali ham 200 bo'lishi kerak, olindi %d", wDetailByCdID.Code)
	}
}

// 6. Security & Privacy: No leak of passwords, hashes, tokens or admin data in public responses
func TestPublicSearch_SecurityNoLeak(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	c := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Test Xavfsizlik Klinikasi",
		Phone:    "+998712009999",
		Region:   "Toshkent shahri",
		District: "Yunusobod",
		Address:  "Ko'cha 1",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, c, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/clinics", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	bodyStr := w.Body.String()

	// Hech qanday maxfiy maydon bo'lmasligi shart
	forbiddenKeys := []string{"password", "password_hash", "token", "owner_id", "email_verified_at", "terms_accepted"}
	for _, key := range forbiddenKeys {
		if fmt.Sprintf("\"%s\"", key) == bodyStr {
			t.Errorf("public endpointda maxfiy maydon oshkor bo'ldi: %s", key)
		}
	}
}

// 7. Regions Catalog Endpoint
func TestPublicSearch_RegionsCatalog(t *testing.T) {
	app := setupClinicTestApp()

	ownerID := uuid.New()
	c1 := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Toshkent Med",
		Region:   "Toshkent shahri",
		District: "Yunusobod",
		Address:  "Manzil 1",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	c2 := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Surxon Med",
		Region:   "Surxondaryo viloyati",
		District: "Termiz sh.",
		Address:  "Manzil 2",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	_ = app.ClinicRepo.Create(nil, c1, nil)
	_ = app.ClinicRepo.Create(nil, c2, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/regions", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("kutilgan status 200, olindi %d", w.Code)
	}

	var resp struct {
		Data []dto.RegionItemDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) != 2 {
		t.Errorf("2 ta viloyat katalogi kutilgan edi, olindi %d", len(resp.Data))
	}
}
