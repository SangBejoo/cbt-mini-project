package interceptor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ValidationInterceptor provides global input validation and sanitization
type ValidationInterceptor struct {
	next         Interceptor
	sanitizer    *bluemonday.Policy
	strictPolicy *bluemonday.Policy
}

// ValidationError represents validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(err error) error {
	return ValidationError{
		Message: err.Error(),
	}
}

// NewValidationInterceptor creates a new validation interceptor
func NewValidationInterceptor() *ValidationInterceptor {
	return &ValidationInterceptor{
		sanitizer:    bluemonday.UGCPolicy(), // User Generated Content policy
		strictPolicy: bluemonday.StrictPolicy(), // Strict policy for sensitive fields
	}
}

// Do performs validation and sanitization on the input message
func (v *ValidationInterceptor) Do(ctx context.Context, input protoreflect.ProtoMessage) error {
	// Sanitize and validate the message
	if err := v.validateAndSanitizeMessage(input, ""); err != nil {
		return err
	}

	// Continue to next interceptor
	if v.next != nil {
		return v.next.Do(ctx, input)
	}
	return nil
}

// SetNext sets the next interceptor in the chain
func (v *ValidationInterceptor) SetNext(next Interceptor) {
	v.next = next
}

// validateAndSanitizeMessage validates and sanitizes all fields in a protobuf message
func (v *ValidationInterceptor) validateAndSanitizeMessage(msg protoreflect.ProtoMessage, path string) error {
	var errors []string

	m := msg.ProtoReflect()

	m.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		fieldName := string(fd.Name())
		currentPath := path
		if currentPath != "" {
			currentPath += "."
		}
		currentPath += fieldName

		// Skip validation for certain fields
		if v.shouldSkipField(fieldName) {
			return true
		}

		// Handle repeated fields
		if fd.IsList() {
			// List of strings
			if fd.Kind() == protoreflect.StringKind {
				list := val.List()
				for i := 0; i < list.Len(); i++ {
					s := list.Get(i).String()
					if err := v.validateStringField(s, fieldName); err != nil {
						errors = append(errors, fmt.Sprintf("%s[%d]: %s", currentPath, i, err.Error()))
						continue
					}
					sanitized := v.sanitizeText(s, fieldName)
					if sanitized != s {
						list.Set(i, protoreflect.ValueOfString(sanitized))
					}
				}
				return true
			}

			// List of messages
			if fd.Kind() == protoreflect.MessageKind {
				list := val.List()
				for i := 0; i < list.Len(); i++ {
					pm := list.Get(i).Message().Interface()
					if err := v.validateAndSanitizeMessage(pm, fmt.Sprintf("%s[%d]", currentPath, i)); err != nil {
						errors = append(errors, err.Error())
					}
				}
				return true
			}

			return true
		}

		// Handle maps
		if fd.IsMap() {
			mapVal := val.Map()
			
			mapVal.Range(func(k protoreflect.MapKey, val protoreflect.Value) bool {
				// Validate map key if string
				mapKeyFd := fd.MapKey()
				if mapKeyFd.Kind() == protoreflect.StringKind {
					keyStr := k.String()
					if err := v.validateStringField(keyStr, fieldName+"_key"); err != nil {
						errors = append(errors, fmt.Sprintf("%s[%s]: %s", currentPath, keyStr, err.Error()))
					}
					// Sanitize key if needed
					sanitizedKey := v.sanitizeText(keyStr, fieldName+"_key")
					if sanitizedKey != keyStr {
						// Note: Map keys are immutable in protobuf, can't modify directly
						// This is a limitation; for now, just validate
					}
				}

				// Validate map value based on value descriptor
				valueFd := fd.MapValue()
				if valueFd.Kind() == protoreflect.StringKind {
					valStr := val.String()
					if err := v.validateStringField(valStr, fieldName+"_value"); err != nil {
						errors = append(errors, fmt.Sprintf("%s[%s]: %s", currentPath, k.String(), err.Error()))
					}
					sanitizedVal := v.sanitizeText(valStr, fieldName+"_value")
					if sanitizedVal != valStr {
						mapVal.Set(k, protoreflect.ValueOfString(sanitizedVal))
					}
				} else if valueFd.Kind() == protoreflect.MessageKind {
					if val.Message().IsValid() {
						if err := v.validateAndSanitizeMessage(val.Message().Interface(), fmt.Sprintf("%s[%s]", currentPath, k.String())); err != nil {
							errors = append(errors, err.Error())
						}
					}
				}
				// Add other value types as needed
				return true
			})
			return true
		}

		switch fd.Kind() {
		case protoreflect.StringKind:
			s := val.String()
			if err := v.validateStringField(s, fieldName); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", currentPath, err.Error()))
				break
			}
			sanitized := v.sanitizeText(s, fieldName)
			if sanitized != s {
				m.Set(fd, protoreflect.ValueOfString(sanitized))
			}

		case protoreflect.MessageKind:
			// Nested message
			nested := val.Message()
			if nested.IsValid() {
				if err := v.validateAndSanitizeMessage(nested.Interface(), currentPath); err != nil {
					errors = append(errors, err.Error())
				}
			}

		case protoreflect.BoolKind:
			if err := v.validateBoolField(val.Bool(), fieldName); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", currentPath, err.Error()))
			}

		case protoreflect.FloatKind, protoreflect.DoubleKind:
			if err := v.validateFloatField(val.Float(), fieldName); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %s", currentPath, err.Error()))
			}

		default:
			// Validate integer fields
			if fd.Kind() == protoreflect.Int64Kind || fd.Kind() == protoreflect.Int32Kind {
				if err := v.validateIntField(val.Int(), fieldName); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %s", currentPath, err.Error()))
				}
			}
		}

		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// shouldSkipField determines if a field should be skipped during validation
func (v *ValidationInterceptor) shouldSkipField(fieldName string) bool {
	skipFields := []string{
		"id", "created_at", "updated_at", "deleted_at",
		"is_active", "status", "page", "limit", "search",
	}

	for _, skip := range skipFields {
		if fieldName == skip {
			return true
		}
	}
	return false
}

// validateStringField validates string fields
func (v *ValidationInterceptor) validateStringField(value, fieldName string) error {
	if value == "" {
		// Allow empty strings for optional fields
		return nil
	}

	// Check length limits
	if err := v.validateLength(value, fieldName); err != nil {
		return err
	}

	// Validate format based on field name
	switch {
	case strings.Contains(fieldName, "email"):
		return v.validateEmail(value)
	case strings.Contains(fieldName, "password"):
		return v.validatePassword(value)
	case strings.Contains(fieldName, "phone"):
		return v.validatePhone(value)
	case strings.Contains(fieldName, "date"):
		return v.validateDate(value)
	case fieldName == "gender":
		return v.validateGender(value)
	}

	// Sanitize text content
	sanitized := v.sanitizeText(value, fieldName)
	if sanitized != value {
		// Note: In a real implementation, you'd need to modify the protobuf message
		// This is a simplified version for demonstration
	}

	return nil
}

// validateIntField validates integer fields
func (v *ValidationInterceptor) validateIntField(value int64, fieldName string) error {
	switch fieldName {
	case "school_id", "user_id", "teacher_id", "student_id":
		if value <= 0 {
			return ValidationError{Field: fieldName, Message: "must be a positive integer"}
		}
	}

	return nil
}

// validateBoolField validates boolean fields
func (v *ValidationInterceptor) validateBoolField(_ bool, _ string) error {
	// For now, no specific validation for booleans
	return nil
}

// validateFloatField validates float fields
func (v *ValidationInterceptor) validateFloatField(value float64, fieldName string) error {
	// Basic validation: check for invalid float values
	if value != value { // NaN check
		return ValidationError{Field: fieldName, Message: "cannot be NaN"}
	}
	return nil
}

// validateLength checks string length limits
func (v *ValidationInterceptor) validateLength(value, fieldName string) error {
	length := utf8.RuneCountInString(value)

	limits := map[string]int{
		"email":     255,
		"password":  128,
		"full_name": 100,
		"phone":     20,
		"address":   500,
		"name":      100,
		"code":      50,
		"description": 1000,
	}

	if limit, exists := limits[fieldName]; exists && length > limit {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("exceeds maximum length of %d characters", limit),
		}
	}

	// Minimum lengths
	minLimits := map[string]int{
		"full_name": 3,
		"password":  8,
		"name":      1,
		"code":      1,
	}

	if minLimit, exists := minLimits[fieldName]; exists && length < minLimit {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at least %d characters long", minLimit),
		}
	}

	return nil
}

// validateEmail validates email format
func (v *ValidationInterceptor) validateEmail(email string) error {
	if email == "" {
		return nil // Allow empty for optional fields
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}

	if len(email) > 255 {
		return ValidationError{Field: "email", Message: "email too long"}
	}

	return nil
}

// validatePassword validates password strength
func (v *ValidationInterceptor) validatePassword(password string) error {
	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "must be at least 8 characters long"}
	}

	// Check for required character types
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ValidationError{
			Field:   "password",
			Message: "must contain at least one uppercase letter, lowercase letter, number, and special character",
		}
	}

	return nil
}

// validatePhone validates phone number format
func (v *ValidationInterceptor) validatePhone(phone string) error {
	// Allow international format with + prefix
	phoneRegex := regexp.MustCompile(`^\+?[0-9\s\-\(\)]{7,20}$`)
	if !phoneRegex.MatchString(phone) {
		return ValidationError{Field: "phone", Message: "invalid phone number format"}
	}

	// Remove formatting and check digits
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")
	if len(cleaned) < 7 || len(cleaned) > 15 {
		return ValidationError{Field: "phone", Message: "phone number length invalid"}
	}

	return nil
}

// validateDate validates date format (YYYY-MM-DD)
func (v *ValidationInterceptor) validateDate(date string) error {
	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(date) {
		return ValidationError{Field: "date", Message: "must be in YYYY-MM-DD format"}
	}

	// Additional date validation could be added here
	return nil
}

// validateGender validates gender field
func (v *ValidationInterceptor) validateGender(gender string) error {
	validGenders := []string{"MALE", "FEMALE", "male", "female", "M", "F", "m", "f"}
	gender = strings.ToUpper(gender)

	for _, valid := range validGenders {
		if gender == valid {
			return nil
		}
	}

	return ValidationError{Field: "gender", Message: "must be MALE or FEMALE"}
}

// sanitizeText sanitizes text based on field type
func (v *ValidationInterceptor) sanitizeText(text, fieldName string) string {
	// Use strict policy for sensitive fields
	if v.isSensitiveField(fieldName) {
		return v.strictPolicy.Sanitize(text)
	}

	// Use UGC policy for user-generated content
	return v.sanitizer.Sanitize(text)
}

// isSensitiveField determines if a field contains sensitive data
func (v *ValidationInterceptor) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"password", "email", "phone", "nis", "nisn",
		"code", "token", "secret", "key",
	}

	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldName, sensitive) {
			return true
		}
	}

	return false
}