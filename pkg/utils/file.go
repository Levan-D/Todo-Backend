package utils

import (
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"net/http"
	"os"
)

func DownloadFileByURL(url string) (filePath string, err error) {
	//Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", errors.New("received non 200 response code")
	}

	filePath = fmt.Sprintf("%s/%s.jpeg", "/tmp", GenerateRandomString(12))

	//Create a empty file
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	//Write the bytes to the field
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func ResizeAndCropImageByPath(fullPath string, width int, height int) error {
	src, err := imaging.Open(fullPath)
	if err != nil {
		return err
	}

	// first crop
	if src.Bounds().Max.X < src.Bounds().Max.Y {
		src = imaging.CropAnchor(src, src.Bounds().Max.X, src.Bounds().Max.X, imaging.Center)
	} else if src.Bounds().Max.Y < src.Bounds().Max.X {
		src = imaging.CropAnchor(src, src.Bounds().Max.Y, src.Bounds().Max.Y, imaging.Center)
	}

	// next resize
	src = imaging.Resize(src, width, height, imaging.Lanczos)

	// save image
	err = imaging.Save(src, fullPath)
	if err != nil {
		return err
	}

	return nil
}

func CheckFileIsImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg":
		return true
	case "image/jpg":
		return true
	case "image/png":
		return true
	default:
		return false
	}
}
