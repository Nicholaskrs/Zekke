package helpers

import (
	"errors"
	"regexp"
	"strings"
)

var (
	imageExtRegex   = regexp.MustCompile(`\.(jpg|jpeg|png|webp)$`)
	extContentTypes = map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"webp": "image/webp",
		"pdf":  "application/pdf",
	}
)

func ExtractImageExtension(fileName string) (ext string, contentType string, err error) {

	sl := imageExtRegex.FindStringSubmatch(fileName)
	if len(sl) != 2 {
		err = errors.New("invalid image extension")
		return "", "", err
	}
	ext = strings.ToLower(sl[1])

	// Additional check, just in case there's a vulnerability in the regex.
	if ext != "jpg" && ext != "jpeg" && ext != "png" && ext != "webp" {
		err = errors.New("invalid image extension")
		return "", "", err
	}

	return ext, extContentTypes[ext], nil
}
