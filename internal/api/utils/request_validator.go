package utils

import (
	"fmt"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var ContentDataTypeToTag = map[dto.DataFormat]string{
	dto.DataFormatJSON: "json",
	dto.DataFormatXML:  "xml",
}

// ValidateDTO validates data transfer objects struct fields based on a specified tag (e.g., "json", "xml").
func ValidateDTO(v any, tag string) error {

	validate := validator.New()

	// Register a custom tag name function based on the provided tag
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		// Check if the specified tag exists and is not "-"
		tagValue := strings.Split(field.Tag.Get(tag), ",")[0]
		if tagValue != "-" && tagValue != "" {
			return tagValue
		}
		return field.Name // Fallback to field name if the tag is not set
	})

	// Perform validation
	err := validate.Struct(v)
	if err != nil {
		// Return validation errors in a readable format
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation: %s", e.Namespace(), e.ActualTag()))
			}
			return fmt.Errorf(strings.Join(errorMessages, "; "))
		}
		return err
	}

	return nil
}
