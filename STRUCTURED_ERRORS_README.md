# ACH v2 File Creation Endpoint with Structured Errors

## Summary

This implementation adds a new v2 endpoint for file creation that returns structured error responses instead of plain text errors, making it easier for clients to parse and handle validation failures programmatically.

## New Endpoint

**POST** `/files/v2/create`

### Request
- Same request format as the original endpoint
- Supports both JSON and plain text ACH files
- All the same validation query parameters

### Response

#### Success (200 OK)
```json
{
  "id": "file123",
  "file": { ... }, // Full file object
  "errors": []     // Empty errors array
}
```

#### Validation Errors (400 Bad Request)
```json
{
  "id": "file123",
  "file": null,
  "errors": [
    {
      "lineNumber": 2,
      "recordType": "EntryDetail",
      "errorType": "FieldError",
      "fieldName": "PaymentTypeCode",
      "message": "PaymentTypeCode is a required field"
    }
  ]
}
```

## Error Structure

Each error in the `errors` array contains:

- **`errorType`** (required): Type of validation error (e.g., "FieldError", "BatchError", "FileError")
- **`message`** (required): Human-readable error message
- **`fieldName`** (optional): Name of the field that caused the error
- **`recordType`** (optional): Type of record where the error occurred
- **`lineNumber`** (optional): Line number where the error occurred (future enhancement)

## Benefits

1. **Structured Data**: Clients can programmatically parse error information
2. **Error Mapping**: Each error can be mapped to specific R-codes (R26, R27, etc.)
3. **Better UX**: Frontend applications can highlight specific fields and provide targeted error messages
4. **Backward Compatibility**: Original v1 endpoint remains unchanged

## Implementation Details

- New endpoint: `/files/v2/create`
- New response type: `createFileV2Response`
- Error conversion: `convertErrorToStructured()` function handles various ACH error types
- HTTP status codes: Properly returns 400 for validation errors, 500 for server errors
- OpenAPI specification updated with new schemas

## Testing

Comprehensive test suite includes:
- Success scenarios with valid files
- Error scenarios with various validation failures
- JSON and plain text file format support
- Comparison between v1 and v2 endpoint behaviors