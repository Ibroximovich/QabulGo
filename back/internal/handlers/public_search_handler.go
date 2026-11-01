package handlers

import (
	"errors"
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PublicSearchHandler struct {
	searchService services.PublicSearchService
}

func NewPublicSearchHandler(searchService services.PublicSearchService) *PublicSearchHandler {
	return &PublicSearchHandler{searchService: searchService}
}

// SearchClinics godoc
// @Summary Bemorlar uchun tasdiqlangan klinikalarni qidirish va filtrlash
// @Description Nomi, manzili, viloyat, tuman va mutaxassislik bo‘yicha ommaviy qidiruv. Faqat tasdiqlangan (approved) klinikalar qaytariladi.
// @Tags Public Search
// @Produce json
// @Param search query string false "Klinika nomi, manzili yoki tavsifi bo‘yicha qidiruv"
// @Param region query string false "Viloyat (masalan: Toshkent shahri)"
// @Param district query string false "Tuman / Shahar (masalan: Yunusobod tumani)"
// @Param specialty_id query string false "Mutaxassislik ID (UUID)"
// @Param page query int false "Sahifa raqami (default 1)"
// @Param limit query int false "Sahifadagi elementlar soni (default 12, max 50)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.PublicClinicItemDTO} "Klinikalar ro‘yxati"
// @Failure 400 {object} dto.ErrorResponse "Parametrlar xatosi"
// @Failure 500 {object} dto.ErrorResponse "Server xatosi"
// @Router /api/v1/public/clinics [get]
func (h *PublicSearchHandler) SearchClinics(c *gin.Context) {
	var filter dto.PublicClinicFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qidiruv parametrlari", nil))
		return
	}

	clinics, total, err := h.searchService.SearchClinics(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Klinikalarni qidirishda xatolik yuz berdi", nil))
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 12
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tasdiqlangan klinikalar ro‘yxati",
		"data":    clinics,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// SearchNearbyClinics godoc
// @Summary Bemor lokatsiyasiga eng yaqin tasdiqlangan klinikalarni topish
// @Description Berilgan latitude va longitude koordinatalari atrofidagi klinikalarni Haversine formulasi bo‘yicha masofasi (km) bilan tartiblab qaytaradi.
// @Tags Public Search
// @Produce json
// @Param latitude query number true "Bemor kenglik koordinatasi (-90 .. 90)"
// @Param longitude query number true "Bemor uzunlik koordinatasi (-180 .. 180)"
// @Param radius_km query number false "Qidiruv radiusi kilometrda (default 10 km, max 100 km)"
// @Param specialty_id query string false "Mutaxassislik bo‘yicha filtr (UUID)"
// @Param page query int false "Sahifa raqami (default 1)"
// @Param limit query int false "Sahifadagi elementlar soni (default 12, max 50)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.NearbyClinicItemDTO} "Eng yaqin klinikalar ro‘yxati"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri koordinata yoki parametrlar"
// @Failure 500 {object} dto.ErrorResponse "Server xatosi"
// @Router /api/v1/public/clinics/nearby [get]
func (h *PublicSearchHandler) SearchNearbyClinics(c *gin.Context) {
	var filter dto.NearbyClinicFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Koordinatalar kiritilishi shart: latitude (-90..90) va longitude (-180..180)", nil))
		return
	}

	clinics, total, err := h.searchService.SearchNearbyClinics(c.Request.Context(), &filter)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCoordinates) || errors.Is(err, services.ErrInvalidRadius) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Yaqin klinikalarni qidirishda xatolik yuz berdi", nil))
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 12
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Yaqin oradagi klinikalar",
		"data":    clinics,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetClinicDetails godoc
// @Summary Tasdiqlangan klinikaning to‘liq ommaviy profili
// @Description Klinika ma’lumotlari, ish soatlari, mutaxassisliklari va faol shifokorlari
// @Tags Public Search
// @Produce json
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika tafsilotlari"
// @Failure 400 {object} dto.ErrorResponse "Klinika ID formati noto‘g‘ri"
// @Failure 404 {object} dto.ErrorResponse "Klinika topilmadi"
// @Router /api/v1/public/clinics/{id} [get]
func (h *PublicSearchHandler) GetClinicDetails(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	clinic, err := h.searchService.GetClinicDetails(c.Request.Context(), clinicID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi yoki hali tasdiqlanmagan", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika tafsilotlari", clinic))
}

// GetClinicDoctors godoc
// @Summary Tasdiqlangan klinikaning barcha faol shifokorlari
// @Description Klinika shifokorlari, ularning mutaxassisliklari va narxlari
// @Tags Public Search
// @Produce json
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.ClinicDoctorItemDTO} "Klinika shifokorlari"
// @Failure 400 {object} dto.ErrorResponse "Klinika ID formati noto‘g‘ri"
// @Failure 404 {object} dto.ErrorResponse "Klinika topilmadi"
// @Router /api/v1/public/clinics/{id}/doctors [get]
func (h *PublicSearchHandler) GetClinicDoctors(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctors, err := h.searchService.GetClinicDoctors(c.Request.Context(), clinicID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi yoki hali tasdiqlanmagan", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika shifokorlari ro‘yxati", doctors))
}

// SearchDoctors godoc
// @Summary Shifokorlarni ommaviy qidirish va filtrlash
// @Description Ism, familiya, mutaxassislik, hudud, narx oralig‘i bo‘yicha qidirish va saralash
// @Tags Public Search
// @Produce json
// @Param search query string false "Shifokor ismi, familiyasi yoki tavsifi"
// @Param region query string false "Viloyat"
// @Param district query string false "Tuman"
// @Param specialty_id query string false "Mutaxassislik ID"
// @Param clinic_id query string false "Muayyan klinika ID"
// @Param min_price query number false "Minimal narx (so‘m)"
// @Param max_price query number false "Maksimal narx (so‘m)"
// @Param sort query string false "Saralash (price_asc, price_desc, experience_desc, name_asc)"
// @Param page query int false "Sahifa (default 1)"
// @Param limit query int false "Miqdor (default 12, max 50)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.PublicDoctorItemDTO} "Shifokorlar ro‘yxati"
// @Failure 500 {object} dto.ErrorResponse "Server xatosi"
// @Router /api/v1/public/doctors [get]
func (h *PublicSearchHandler) SearchDoctors(c *gin.Context) {
	var filter dto.PublicDoctorFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri parametrlar", nil))
		return
	}

	doctors, total, err := h.searchService.SearchDoctors(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Shifokorlarni qidirishda xatolik yuz berdi", nil))
		return
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 12
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shifokorlar ro‘yxati",
		"data":    doctors,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetDoctorDetails godoc
// @Summary Shifokorning batafsil profili va ish joylari
// @Description Shifokor ma’lumotlari va u qabul qiladigan barcha tasdiqlangan klinikalar hamda narxlari
// @Tags Public Search
// @Produce json
// @Param id path string true "Shifokor UUID yoki ClinicDoctor UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.PublicDoctorDetailDTO} "Shifokor tafsilotlari"
// @Failure 400 {object} dto.ErrorResponse "ID formati noto‘g‘ri"
// @Failure 404 {object} dto.ErrorResponse "Shifokor topilmadi"
// @Router /api/v1/public/doctors/{id} [get]
func (h *PublicSearchHandler) GetDoctorDetails(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("ID formati noto‘g‘ri", nil))
		return
	}

	detail, err := h.searchService.GetDoctorDetails(c.Request.Context(), targetID)
	if err != nil {
		if errors.Is(err, repositories.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Shifokor topilmadi yoki tasdiqlangan klinikalarda faol emas", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokor tafsilotlari", detail))
}

// GetRegionsCatalog godoc
// @Summary Klinika mavjud hududlar (viloyatlar va tumanlar) katalogi
// @Description Tasdiqlangan klinikalar mavjud bo‘lgan hududlar ro‘yxati
// @Tags Public Search
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=[]dto.RegionItemDTO} "Hududlar katalogi"
// @Router /api/v1/public/regions [get]
func (h *PublicSearchHandler) GetRegionsCatalog(c *gin.Context) {
	catalog, err := h.searchService.GetRegionsCatalog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Katalogni olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Hududlar katalogi", catalog))
}

