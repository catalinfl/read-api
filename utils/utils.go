package utils

import (
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/nfnt/resize"
)

func CompressPhoto(c *fiber.Ctx, file *multipart.FileHeader) (image.Image, error) {
	src, err := file.Open()

	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	defer src.Close()

	var img image.Image

	fileType := file.Header.Get("Content-Type")

	if fileType == "image/png" {
		img, err = png.Decode(src)
	} else {
		img, err = jpeg.Decode(src)
	}

	if err != nil {
		return nil, c.Status(400).JSON(fiber.Map{
			"data": "Invalid request",
		})
	}

	resizedImg := resize.Resize(300, 0, img, resize.Lanczos2)

	return resizedImg, nil
}
