# User API

## Create User

Creates a new user with a unique corporate key.

### SMTP Address Generation

The system automatically generates a unique primary SMTP address based on:
- **firstName** and **lastName** (used for generation only, not stored)
- **countryCode** and **departmentCode** (determine the email domain)

Pattern: `{firstName}.{lastName}@{domain}`

**Uniqueness Guarantee:**
- User Service ensures the generated SMTP address is unique within its allocation at creation time
- If a conflict occurs at generation time, a numeric suffix is automatically appended (`.1`, `.2`, etc.)
- Allocated SMTP addresses are not reassigned by User Service to a different user, but the assigned address remains subject to asynchronous validation by SMTP Service and may be updated later if a conflict or policy violation is detected

### Request

**URL**: **POST** /api/users

**Body**:
```json
{
    "corpKey": "string",
    "firstName": "string",
    "lastName": "string",
    "fullName": "string",
    "countryCode": "string",
    "departmentCode": "string"
}
```

### Field Constraints
- `corpKey`: required, non-empty, alphanumeric
- `fullName`: required, non-empty
- `firstName`: required, non-empty
- `lastName`: required, non-empty
- `countryCode`: required, ISO 3166-1 alpha-2
- `departmentCode`: required, non-empty, alphanumeric

### Response

**Codes**: 
- 201 - Created
- 400 - Validation Errors
- 409 - User with Corporate Key already exists

**Body**:
```json
{
    "corpKey": "string",
    "fullName": "string",
    "countryCode": "string",
    "departmentCode": "string",
    "primarySMTP": "string",
    "secondarySMTPs": ["string"]
}
```

## Get User by Corp Key

Returns a user by corporate key.

### Request

**URL**: **GET** /api/users/:corpKey

**Path Parameters**:
- **corpKey** - User's unique identifier

### Response

**Codes**: 
- 200 - OK
- 404 - User not found

**Body**:
```json
{
    "corpKey": "string",
    "fullName": "string",
    "countryCode": "string",
    "departmentCode": "string",
    "primarySMTP": "string",
    "secondarySMTPs": ["string"]
}
```

## Get Users

Returns a page of users ordered by Full Name, then by Corp Key.

### Request

**URL**: **GET** /api/users?limit=&offset=

**Query Parameters**:
- **limit** - Number of users per page (default: 100, max: 1000)
- **offset** - Offset from the start of all users ordered by Full Name, then by Corp Key (default: 0)

### Response

**Codes**: 
- 200 - OK

**Body**:
```json
{
    "limit": 100,
    "offset": 0,
    "totalCount": 1000,
    "items": [
        {
            "corpKey": "string",
            "fullName": "string",
            "countryCode": "string",
            "departmentCode": "string",
            "primarySMTP": "string",
            "secondarySMTPs": ["string"]
        }
    ]
}
```

## Update User

Updates a user's data.

### SMTP Address Behavior on Update

**When Primary SMTP Changes:**

If any of the firstName, lastName, countryCode, or departmentCode fields are updated such that a different SMTP address would be generated, the system:
1. Generates a new unique primary SMTP address
2. Moves the old primary address to the `secondarySMTPs` array
3. Returns the updated user with the new primary address

**Secondary SMTP Addresses:**
- The `secondarySMTPs` array preserves all previous primary SMTP addresses for this user
- Addresses are never removed from this array (historical record)
- May be empty for newly created users or users whose primary address never changed

### Request

**URL**: **PUT** /api/users/:corpKey

**Path Parameters**:
- **corpKey** - User's unique identifier

**Body**:
```json
{
    "firstName": "string",
    "lastName": "string",
    "fullName": "string",
    "countryCode": "string",
    "departmentCode": "string"
}
```

### Field Constraints
- `fullName`: required, non-empty
- `firstName`: required, non-empty
- `lastName`: required, non-empty
- `countryCode`: required, ISO 3166-1 alpha-2
- `departmentCode`: required, non-empty, alphanumeric

### Response

**Codes**: 
- 200 - OK
- 400 - Validation Errors
- 404 - User not found

**Body**:
```json
{
    "corpKey": "string",
    "fullName": "string",
    "countryCode": "string",
    "departmentCode": "string",
    "primarySMTP": "string",
    "secondarySMTPs": ["string"]
}
```

## Delete User

Deletes a user from the system.

**Behavior:**
- Soft deletion—user data is retained but marked as deleted
- User's SMTP addresses become unavailable for reassignment
- Idempotent operation (calling twice returns 204 both times)
- Deleted users are excluded from GET /users queries

### Request

**URL**: **DELETE** /api/users/:corpKey

**Path Parameters**:
- **corpKey** - User's unique identifier

### Response

**Codes**: 
- 204 - No Content

## Error Response Format

### Structure
```json
{
    "type": "string",
    "status": number,
    "error": "string",
    "details": [
        {
            "field": "string",
            "message": "string"
        }
    ]
}
```

### Examples

#### Validation Error (400)
```json
{
    "type": "https://tools.ietf.org/html/rfc7231#section-6.5.1",
    "status": 400,
    "error": "One or more validation errors occurred",
    "details": [
        {"field": "corpKey", "message": "must be alphanumeric"},
        {"field": "firstName", "message": "is required"}
    ]
}
```

#### Conflict (409)
```json
{
    "type": "https://tools.ietf.org/html/rfc7231#section-6.5.8",
    "status": 409,
    "error": "User with Corp Key 'ABC123' already exists",
    "details": []
}
```