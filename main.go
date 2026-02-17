package main

import (
	"encoding/json"
	"fmt"
	"go-backend/database"
	_ "go-backend/docs" // swag will generate this
	"go-backend/mcp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"

	// "log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rwcarlsen/goexif/exif"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Go Backend API
// @version 1.0
// @description Example API with GET, POST, and PATCH endpoints.

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // React dev server
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", pingHandler)

	r.POST("/upload", uploadImageHandler)

	r.POST("/chat", chatHandler)

	r.GET("/image/:filename", getImageHandler)

	r.GET("/images", getImagesHandler)

	r.GET("/tableData", getTableDataHandler)

	r.POST("/user", createUser)

	r.PATCH("/user/:id", updateUser)

	r.POST("/add", addItemHandler)

	r.GET("/item/:id", getItemHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}

// pingHandler godoc
// @Summary Ping endpoint
// @Description Returns pong
// @ID ping
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}

type Item = database.Item

// getItemHandler godoc
// @Summary Get an item by ID
// @Description Retrieve a JSON object stored in the database by its ID
// @Tags items
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} Item
// @Failure 404 {object} map[string]string "Item not found"
// @Failure 500 {object} map[string]string "Failed to read DB"
// @Router /item/{id} [get]
func getItemHandler(c *gin.Context) {
	id := c.Param("id")

	item, found, err := database.GetByID(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read DB"})
		return
	}
	if !found {
		c.JSON(404, gin.H{"error": "Item not found"})
		return
	}

	c.JSON(200, item)
}

// addItemHandler godoc
// @Summary Add a new item
// @Description Add a JSON object and store it in the database, returns the generated ID
// @Tags items
// @Accept json
// @Produce json
// @Param data body map[string]interface{} true "Item data"
// @Success 200 {object} map[string]string "id of the created item"
// @Failure 400 {object} map[string]string "Invalid JSON"
// @Failure 500 {object} map[string]string "Failed to read/write DB"
// @Router /add [post]
func addItemHandler(c *gin.Context) {
	var data map[string]interface{}

	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	id := database.NewID()
	item := database.Item{
		ID:   id,
		Data: data,
	}

	items, err := database.ReadDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read DB"})
		return
	}

	items = append(items, item)

	if err := database.WriteDB(items); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// getTableDataHandler godoc
// @Summary Get table data
// @Description Returns table data from JSON file
// @ID get-table-data
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /tableData [get]
func getTableDataHandler(c *gin.Context) {
	filepath := "./data/tableData.json"

	data, err := os.ReadFile(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}

// getImagesHandler godoc
// @Summary List all images
// @Description Returns a JSON array of filenames in the uploads folder
// @ID list-images
// @Produce json
// @Success 200 {array} string "List of image filenames"
// @Failure 500 {object} map[string]string
// @Router /images [get]
func getImagesHandler(c *gin.Context) {
	uploadsDir := "./images"

	files := []string{}

	err := filepath.WalkDir(uploadsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, d.Name())
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

// ChatMessage represents incoming chat JSON
type ChatMessage struct {
	Text string `json:"request" binding:"required" example:"Analyze scenario, List all machinery, Categorize machinery"`
}

type LLMMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type LLMParams struct {
    Messages []LLMMessage `json:"messages"`
}

type JsonRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params"`
    ID      int         `json:"id"`
}

type JsonRPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    Result  json.RawMessage `json:"result"`
    Error   interface{}     `json:"error,omitempty"`
    ID      int             `json:"id"`
}

type UserRequest struct {
    Request string `json:"request" binding:"required" example:"Analyze image + categorize machinery"`
}


// chatHandler godoc
// @Summary Send a chat message
// @Description Receives a message from UI and returns a JSON response
// @ID chat-message
// @Accept json
// @Produce json
// @Param message body ChatMessage true "Message JSON"
// @Success 200 {object} map[string]string
// @Router /chat [post]
// --- Gin handler ---
func chatHandler(c *gin.Context) {
    var req UserRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Detect tools
    toolsToRun := mcp.DetermineTools(req.Request)

    // Build JSON-RPC request for LLM
    messages := []mcp.LLMMessage{
        {Role: "system", Content: "You are a risk assessment assistant."},
        {Role: "user", Content: req.Request},
    }

    // Optionally prepend system message with tools
    if len(toolsToRun) > 0 {
        toolMsg := mcp.LLMMessage{
            Role:    "system",
            Content: fmt.Sprintf("Use the following tools for this request: %v", toolsToRun),
        }
        messages = append([]mcp.LLMMessage{toolMsg}, messages...)
    }

    rpcReq := &mcp.JsonRPCRequest{
        JSONRPC: "2.0",
        Method:  "llm/message",
        Params:  mcp.LLMParams{Messages: messages},
        ID:      1,
    }

    // Send to LLM (mock or real)
    rpcResp, err := mcp.SendToLLM("", rpcReq) // "" = mock
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Include detected tools in the response
    respMap := map[string]interface{}{}
    json.Unmarshal(rpcResp.Result, &respMap)
    respMap["tools_detected"] = toolsToRun
    respBytes, _ := json.Marshal(respMap)
    rpcResp.Result = respBytes

    c.JSON(http.StatusOK, rpcResp)
}

// uploadImageHandler godoc
// @Summary Upload an image
// @Description Uploads an image and returns filename, size, and EXIF metadata
// @ID upload-image
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file"
// @Success 200 {object} map[string]interface{}
// @Router /upload [post]
func uploadImageHandler(c *gin.Context) {
	// Retrieve uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	// Ensure uploads folder exists
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	savePath := "uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Open file to read EXIF and image size
	f, err := os.Open(savePath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	// Decode image for width & height
	img, _, err := image.Decode(f)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Image decode error: %v", err)})
		return
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Reset file pointer for EXIF
	f.Seek(0, 0)
	exifData := make(map[string]string)
	x, err := exif.Decode(f)
	if err == nil {
		// Extract a few common EXIF tags
		tags := []exif.FieldName{
			exif.Model, exif.Make, exif.DateTime, exif.FocalLength, exif.ExposureTime,
		}
		for _, tag := range tags {
			if val, err := x.Get(tag); err == nil {
				exifData[string(tag)] = val.String()
			}
		}
	}

	// Return JSON response
	c.JSON(http.StatusOK, gin.H{
		"filename": file.Filename,
		"path":     savePath,
		"width":    width,
		"height":   height,
		"exif":     exifData,
	})
}

// getImageHandler godoc
// @Summary Serve an image
// @Description Returns an image file as response
// @ID get-image
// @Accept  json
// @Produce image/png, image/jpeg
// @Param filename path string true "Image filename"
// @Success 200 {file} file
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /image/{filename} [get]
func getImageHandler(c *gin.Context) {
	filename := c.Param("filename") // from /image/:filename

	if filename == "" {
		c.JSON(400, gin.H{"error": "Filename required"})
		return
	}

	filepath := "./images/" + filename
	c.File(filepath) // Gin will set Content-Type automatically based on file extension
}

// createUser godoc
// @Summary Create a new user
// @Description Creates a user with given data
// @ID create-user
// @Accept json
// @Produce json
// @Param user body map[string]string true "User data"
// @Success 201 {object} map[string]interface{}
// @Router /user [post]
func createUser(c *gin.Context) {
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"data": body})
}

// updateUser godoc
// @Summary Update user
// @Description Update a user by ID
// @ID update-user
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body map[string]string true "User data"
// @Success 200 {object} map[string]interface{}
// @Router /user/{id} [patch]
func updateUser(c *gin.Context) {
	id := c.Param("id")
	var body map[string]string
	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"id":   id,
		"data": body,
	})
}
