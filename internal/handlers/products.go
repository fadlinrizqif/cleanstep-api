package handlers

import (
	"net/http"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductsHandler struct {
	App *app.App
}

func NewProductsHandler(app *app.App) *ProductsHandler {
	return &ProductsHandler{App: app}
}

func (h *ProductsHandler) CreateProducts(c *gin.Context) {
	type params struct {
		Name     string `json:"name"`
		Price    int32  `json:"price"`
		Category string `json:"category"`
		Stock    int32  `json:"stock"`
	}

	var productsDetail params
	if err := c.ShouldBindJSON(&productsDetail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	newProduct, err := h.App.DBqueries.CreateProduct(c.Request.Context(), database.CreateProductParams{
		Name:     productsDetail.Name,
		Price:    productsDetail.Price,
		Category: productsDetail.Category,
		Stock:    productsDetail.Stock,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, dto.ProductResponse{
		ID:       newProduct.ID,
		Name:     newProduct.Name,
		Category: newProduct.Category,
		Price:    newProduct.Price,
		Stock:    newProduct.Stock,
	})
}

func (h *ProductsHandler) GetAllProducts(c *gin.Context) {

	getProducts, err := h.App.DBqueries.GetAllProduct(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	var allProduct []dto.ProductResponse

	for _, product := range getProducts {
		allProduct = append(allProduct, dto.ProductResponse{
			ID:       product.ID,
			Name:     product.Name,
			Category: product.Category,
			Price:    product.Price,
			Stock:    product.Stock,
		})
	}

	totalProduct := len(allProduct)

	c.JSON(http.StatusOK, dto.GetProductResponse{
		Data:  allProduct,
		Total: int32(totalProduct),
	})

}

func (h *ProductsHandler) GetProducts(c *gin.Context) {
	productID, err := uuid.Parse(c.Request.PathValue("productID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	getProduct, err := h.App.DBqueries.GetProduct(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, dto.ProductResponse{
		ID:       getProduct.ID,
		Name:     getProduct.Name,
		Price:    getProduct.Price,
		Category: getProduct.Category,
		Stock:    getProduct.Stock,
	})

}
