package cloud

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService interface {
	UploadImage(ctx context.Context, file multipart.File, folder string) (string, error)
	DeleteImage(ctx context.Context, secureURL string) error
}

type cloudinaryService struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudinaryURL string) (CloudinaryService, error) {
	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, err
	}
	return &cloudinaryService{client: cld}, nil
}

func (s *cloudinaryService) UploadImage(ctx context.Context, file multipart.File, folder string) (string, error) {
	uploadResult, err := s.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

func (s *cloudinaryService) DeleteImage(ctx context.Context, secureURL string) error {
	// Extract public_id from secureURL
	// Format: https://res.cloudinary.com/<cloud_name>/image/upload/v<version>/<folder>/<filename>.<ext>
	// We want <folder>/<filename>
	
	parts := strings.Split(secureURL, "/upload/")
	if len(parts) != 2 {
		return nil // Not a standard Cloudinary upload URL, skip deletion
	}

	pathParts := strings.Split(parts[1], "/")
	// pathParts could be: ["v12345678", "products", "myimage.jpg"]
	
	// Remove the version tag if it starts with 'v' and is followed by numbers
	var publicIDParts []string
	for _, part := range pathParts {
		if strings.HasPrefix(part, "v") && len(part) > 1 && part[1] >= '0' && part[1] <= '9' {
			continue // Skip version
		}
		publicIDParts = append(publicIDParts, part)
	}

	if len(publicIDParts) == 0 {
		return nil
	}

	publicIDWithExt := strings.Join(publicIDParts, "/")
	
	// Remove extension
	lastDotIndex := strings.LastIndex(publicIDWithExt, ".")
	publicID := publicIDWithExt
	if lastDotIndex > 0 {
		publicID = publicIDWithExt[:lastDotIndex]
	}

	_, err := s.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	
	return err
}
