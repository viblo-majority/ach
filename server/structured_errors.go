// Licensed to The Moov Authors under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. The Moov Authors licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package server

import (
	"fmt"
	"reflect"

	"github.com/moov-io/ach"
	"github.com/moov-io/base"
)

// StructuredError represents a single validation error in a structured format
type StructuredError struct {
	LineNumber int    `json:"lineNumber,omitempty"`
	RecordType string `json:"recordType,omitempty"`
	ErrorType  string `json:"errorType"`
	FieldName  string `json:"fieldName,omitempty"`
	Message    string `json:"message"`
}

// StructuredErrorResponse wraps the file creation response with structured errors
type StructuredErrorResponse struct {
	ID     string             `json:"id"`
	File   *ach.File          `json:"file,omitempty"`
	Errors []StructuredError  `json:"errors,omitempty"`
}

// convertErrorToStructured converts various ACH error types to structured errors
func convertErrorToStructured(err error) []StructuredError {
	if err == nil {
		return nil
	}

	var structuredErrors []StructuredError

	// Handle base.ErrorList (multiple errors)
	if errorList, ok := err.(base.ErrorList); ok {
		for _, e := range errorList {
			structuredErrors = append(structuredErrors, convertSingleErrorToStructured(e)...)
		}
		return structuredErrors
	}

	// Handle single error
	return convertSingleErrorToStructured(err)
}

// convertSingleErrorToStructured converts a single error to structured format
func convertSingleErrorToStructured(err error) []StructuredError {
	if err == nil {
		return nil
	}

	// Handle FieldError
	if fieldErr, ok := err.(*ach.FieldError); ok {
		return []StructuredError{{
			ErrorType: "FieldError", 
			FieldName: fieldErr.FieldName,
			Message:   fieldErr.Err.Error(),
		}}
	}

	// Handle BatchError
	if batchErr, ok := err.(*ach.BatchError); ok {
		recordType := "Batch"
		if batchErr.BatchType != "" {
			recordType = fmt.Sprintf("Batch (%s)", batchErr.BatchType)
		}
		return []StructuredError{{
			ErrorType:  "BatchError",
			RecordType: recordType,
			FieldName:  batchErr.FieldName,
			Message:    batchErr.Err.Error(),
		}}
	}

	// Handle FileError
	if fileErr, ok := err.(*ach.FileError); ok {
		return []StructuredError{{
			ErrorType: "FileError",
			RecordType: "File",
			FieldName: fileErr.FieldName,
			Message:   fileErr.Msg,
		}}
	}

	// Handle specific ACH errors using reflection and type checking
	errorValue := reflect.ValueOf(err)
	errorType := errorValue.Type()

	// Check for pointer types
	if errorType.Kind() == reflect.Ptr {
		errorType = errorType.Elem()
		if errorValue.IsValid() && !errorValue.IsNil() {
			errorValue = errorValue.Elem()
		}
	}

	// Try to extract structured information from known error types
	if errorType.Name() != "" {
		structErr := StructuredError{
			ErrorType: errorType.Name(),
			Message:   err.Error(),
		}

		// Try to extract field information if available
		if errorValue.IsValid() && errorValue.Kind() == reflect.Struct {
			// Look for common field names
			if fieldName := getFieldValue(errorValue, "FieldName"); fieldName != "" {
				structErr.FieldName = fieldName
			}
			if recordType := getFieldValue(errorValue, "RecordType"); recordType != "" {
				structErr.RecordType = recordType
			}
		}

		return []StructuredError{structErr}
	}

	// Fallback for unknown error types
	return []StructuredError{{
		ErrorType: "UnknownError",
		Message:   err.Error(),
	}}
}

// getFieldValue attempts to get a string value from a struct field using reflection
func getFieldValue(v reflect.Value, fieldName string) string {
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return ""
	}

	// Convert to string if possible
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", field.Int())
	default:
		if field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}

	return ""
}