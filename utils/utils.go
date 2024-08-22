package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"path/filepath"

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

	resizedImg := resize.Resize(600, 0, img, resize.Lanczos2)

	return resizedImg, nil
}

func ImageToMultipartFileHeader(img image.Image, filename string, typeOf string, buf *bytes.Buffer) (*multipart.FileHeader, error) {

	if typeOf == "image/png" {
		err := png.Encode(buf, img)

		if err != nil {
			return nil, err
		}
	} else {
		err := jpeg.Encode(buf, img, nil)

		if err != nil {
			return nil, err
		}
	}

	fileHeader := &multipart.FileHeader{
		Filename: filepath.Base(filename),
		Header:   textproto.MIMEHeader{},
		Size:     int64(buf.Len()),
	}

	fileHeader.Header.Set("Content-Disposition", "form-data; name=\"file\"; filename=\""+filename+"\"")

	fileHeader.Header.Set("Content-Type", typeOf)

	return fileHeader, nil
}
