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
	"errors"
	"testing"

	"github.com/moov-io/ach"
	"github.com/moov-io/base"
	"github.com/stretchr/testify/require"
)

func TestConvertErrorToStructured(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		result := convertErrorToStructured(nil)
		require.Nil(t, result)
	})

	t.Run("FieldError", func(t *testing.T) {
		fieldErr := &ach.FieldError{
			FieldName: "PaymentTypeCode",
			Err:       errors.New("is a required field"),
			Value:     "",
		}

		result := convertErrorToStructured(fieldErr)
		require.Len(t, result, 1)
		require.Equal(t, "FieldError", result[0].ErrorType)
		require.Equal(t, "PaymentTypeCode", result[0].FieldName)
		require.Equal(t, "is a required field", result[0].Message)
	})

	t.Run("BatchError", func(t *testing.T) {
		batchErr := &ach.BatchError{
			BatchNumber: 1,
			BatchType:   "PPD",
			FieldName:   "ServiceClassCode",
			Err:         errors.New("is invalid"),
		}

		result := convertErrorToStructured(batchErr)
		require.Len(t, result, 1)
		require.Equal(t, "BatchError", result[0].ErrorType)
		require.Equal(t, "Batch (PPD)", result[0].RecordType)
		require.Equal(t, "ServiceClassCode", result[0].FieldName)
		require.Equal(t, "is invalid", result[0].Message)
	})

	t.Run("FileError", func(t *testing.T) {
		fileErr := &ach.FileError{
			FieldName: "ImmediateOrigin",
			Msg:       "is invalid",
		}

		result := convertErrorToStructured(fileErr)
		require.Len(t, result, 1)
		require.Equal(t, "FileError", result[0].ErrorType)
		require.Equal(t, "File", result[0].RecordType)
		require.Equal(t, "ImmediateOrigin", result[0].FieldName)
		require.Equal(t, "is invalid", result[0].Message)
	})

	t.Run("ErrorList", func(t *testing.T) {
		var errorList base.ErrorList
		errorList.Add(&ach.FieldError{
			FieldName: "Field1",
			Err:       errors.New("error1"),
		})
		errorList.Add(&ach.FieldError{
			FieldName: "Field2",
			Err:       errors.New("error2"),
		})

		result := convertErrorToStructured(errorList)
		require.Len(t, result, 2)
		require.Equal(t, "FieldError", result[0].ErrorType)
		require.Equal(t, "Field1", result[0].FieldName)
		require.Equal(t, "error1", result[0].Message)
		require.Equal(t, "FieldError", result[1].ErrorType)
		require.Equal(t, "Field2", result[1].FieldName)
		require.Equal(t, "error2", result[1].Message)
	})

	t.Run("Unknown error", func(t *testing.T) {
		unknownErr := errors.New("some unknown error")

		result := convertErrorToStructured(unknownErr)
		require.Len(t, result, 1)
		require.Equal(t, "errorString", result[0].ErrorType)
		require.Equal(t, "some unknown error", result[0].Message)
	})
}

func TestStructuredErrorResponse_Error(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		resp := createFileV2Response{
			ID:     "test123",
			Errors: []StructuredError{},
		}
		require.Nil(t, resp.error())
	})

	t.Run("with errors", func(t *testing.T) {
		resp := createFileV2Response{
			ID: "test123",
			Errors: []StructuredError{
				{
					ErrorType: "FieldError",
					Message:   "test error",
				},
			},
		}
		err := resp.error()
		require.NotNil(t, err)
		require.Equal(t, "validation error: test error", err.Error())
	})
}