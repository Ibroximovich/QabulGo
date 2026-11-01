package handlers

import (
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storageService services.StorageService
}

func NewUploadHandler(storageService services.StorageService) *UploadHandler {
	return &UploadHandler{storageService: storageService}
}

// UploadImage godoc
// @Summary Rasm yuklash (klinika logotipi, cover yoki shifokor rasmi)
// @Description Faqat JPEG, PNG, WebP formatlari, maksimal hajm 5MB
// @Tags Uploads
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Rasm fayli"
// @Param folder query string false "Papka nomi (masalan: clinics yoki doctors)"
// @Success 201 {object} dto.SuccessResponse{data=dto.UploadResponse} "Rasm muvaffaqiyatli yuklandi"
// @Failure 400 {object} dto.ErrorResponse "Fayl formati noto‘g‘ri yoki hajm katta"
// @Router /api/v1/upload [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Fayl topilmadi. 'file' maydonida rasm yuboring.", nil))
		return
	}

	folder := c.DefaultQuery("folder", "general")
	if folder != "clinics" && folder != "doctors" && folder != "general" {
		folder = "general"
	}

	fileURL, err := h.storageService.SaveUploadedFile(fileHeader, folder)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	resp := dto.UploadResponse{
		URL:      fileURL,
		Filename: fileHeader.Filename,
		Size:     fileHeader.Size,
		MimeType: fileHeader.Header.Get("Content-Type"),
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Rasm muvaffaqiyatli yuklandi", resp))
}

