# Modami User Service — Integration Guide

**Service:** `be-modami-user-service`
**Base URL:** `https://modami-user.techinsightsworld.com/v1/user-services`
**gRPC Host:** (internal) `user-service:<grpc_port>`

---

## Changelog

| Version | Date       | Changes |
|---------|------------|---------|
| v1.2.0  | 2026-04-02 | Unified API view (no admin/public split); add `username` and `date_of_birth` to profile response; migrate to `pkg-gokit` standard response envelope; embed DB migrations into binary (auto-run on startup); fix Kafka `UNKNOWN_TOPIC_ID` consumer recovery |
| v1.1.0  | 2026-03-28 | Add Swagger docs; `pkg-gokit` response/apperror integration; refactor handler error handling |
| v1.0.0  | 2026-03-27 | Initial release |

---

## Standard Response Envelope

All REST API responses are wrapped in this envelope:

### Success
```json
{
  "success": true,
  "data": { },
  "meta": { "timestamp": 1711584000 }
}
```

### Error
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "không tìm thấy",
    "detail": ""
  },
  "meta": { "timestamp": 1711584000 }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid input or parameters |
| `VALIDATION_ERROR` | 400 | Field-level validation failed |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Access denied (e.g. not admin) |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists or state conflict |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

---

## Authentication

| Icon | Meaning |
|------|---------|
| 🔒 | Requires `Authorization: Bearer <token>` |
| 🔑 | Requires admin role in addition to token |

```
Authorization: Bearer <token>
```

Token is issued by Keycloak and validated via JWKS.

---

## API Endpoints

---

### Users

#### `GET /users/search` — Search Users
Search users by name or email. No auth required.

**Query Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `q` | string | Yes | Search keyword |
| `limit` | int | No | Max results (1–100, default 20) |
| `cursor` | string | No | Pagination cursor from previous response |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "username": "nguyenvana",
        "full_name": "Nguyen Van A",
        "phone": "0901234567",
        "avatar_url": "https://...",
        "cover_url": "https://...",
        "bio": "Hello world",
        "gender": "male",
        "role": "buyer",
        "status": "active",
        "trust_score": 4.5,
        "follower_count": 100,
        "following_count": 50,
        "email_verified": true,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "cursor": "eyJpZCI6Ii4uLiJ9"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | `q` param missing |

---

#### `GET /users/{id}` — Get Public Profile
Returns a user's public profile. No auth required.

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "nguyenvana",
    "full_name": "Nguyen Van A",
    "phone": "0901234567",
    "avatar_url": "https://...",
    "cover_url": "https://...",
    "bio": "Hello world",
    "gender": "male",
    "date_of_birth": "1995-05-20",
    "role": "buyer",
    "status": "active",
    "trust_score": 4.5,
    "follower_count": 100,
    "following_count": 50,
    "email_verified": true,
    "created_at": "2026-01-01T00:00:00Z"
  },
  "meta": { "timestamp": 1711584000 }
}
```

> `date_of_birth` is omitted when not set.

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `NOT_FOUND` 404 | User not found |

---

#### `GET /users/me` — Get My Profile 🔒
Returns the authenticated user's own profile. Same response structure as `GET /users/{id}`.

**Errors**

| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | User not found |

---

#### `PUT /users/me` — Update My Profile 🔒

**Request Body** — all fields optional:
```json
{
  "full_name": "Nguyen Van B",
  "phone": "0901234567",
  "bio": "Updated bio",
  "gender": "male",
  "date_of_birth": "1995-05-20"
}
```

| Field | Type | Validation |
|-------|------|------------|
| `full_name` | string | min 1, max 255 |
| `phone` | string | — |
| `bio` | string | max 500 |
| `gender` | string | `male` \| `female` \| `other` \| `undisclosed` |
| `date_of_birth` | string | format `YYYY-MM-DD` |

**Success Response** `200` — updated `UserProfileResponse` (same as `GET /users/{id}`)

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `PUT /users/me/avatar` — Update Avatar 🔒

**Request Body**
```json
{
  "avatar_url": "https://cdn.example.com/avatar.jpg"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `avatar_url` | string | Yes | valid URL |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "avatar updated" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid body |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `PUT /users/me/cover` — Update Cover 🔒

**Request Body**
```json
{
  "cover_url": "https://cdn.example.com/cover.jpg"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `cover_url` | string | Yes | valid URL |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "cover updated" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid body |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `DELETE /users/me` — Deactivate Account 🔒

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "account deactivated" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `GET /admin/users` — List / Search All Users 🔒🔑
Admin-only. List or search all users.

**Query Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `q` | string | No | Search keyword |
| `limit` | int | No | Max results (1–100, default 20) |
| `cursor` | string | No | Pagination cursor |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "users": [ { "...": "UserProfileResponse" } ],
    "cursor": "eyJpZCI6Ii4uLiJ9"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `FORBIDDEN` 403 | Not admin |

---

#### `PUT /admin/users/{id}/status` — Update User Status 🔒🔑
Admin-only. Change a user's account status.

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Request Body**
```json
{
  "status": "suspended",
  "reason": "Violation of terms"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `status` | string | Yes | `active` \| `inactive` \| `suspended` \| `banned` |
| `reason` | string | No | max 500 |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "status updated" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID or validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `FORBIDDEN` 403 | Not admin |
| `NOT_FOUND` 404 | User not found |

---

### Follows

#### `POST /users/{id}/follow` — Follow User 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | Target user to follow |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "followed successfully" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID or trying to follow yourself |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `CONFLICT` 409 | Already following |

---

#### `DELETE /users/{id}/follow` — Unfollow User 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | Target user to unfollow |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "unfollowed successfully" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `GET /users/{id}/follow/status` — Check Follow Status 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | Target user to check |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "is_following": true },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `GET /users/{id}/followers` — Get Followers

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Query Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `limit` | int | No | Max results (1–100, default 20) |
| `cursor` | string | No | Pagination cursor |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "full_name": "Nguyen Van A",
        "avatar_url": "https://...",
        "followed_at": "2026-01-01T00:00:00Z"
      }
    ],
    "cursor": "eyJpZCI6Ii4uLiJ9"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |

---

#### `GET /users/{id}/following` — Get Following

Same structure as `GET /users/{id}/followers`.

---

### Reviews

#### `POST /users/{id}/reviews` — Create Review 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | Reviewee user ID |

**Request Body**
```json
{
  "order_id": "550e8400-e29b-41d4-a716-446655440000",
  "rating": 5,
  "comment": "Great seller!",
  "is_anonymous": false
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `order_id` | UUID | Yes | valid UUID |
| `rating` | int | Yes | 1–5 |
| `comment` | string | No | max 1000 |
| `is_anonymous` | bool | No | default false |

**Success Response** `201`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "reviewer_id": "uuid",
    "reviewee_id": "uuid",
    "order_id": "uuid",
    "rating": 5,
    "comment": "Great seller!",
    "role": "buyer",
    "is_anonymous": false,
    "created_at": "2026-01-01T00:00:00Z"
  },
  "meta": { "timestamp": 1711584000 }
}
```

> `reviewer_id` is omitted from the response when `is_anonymous` is `true`.

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed or invalid UUID |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `CONFLICT` 409 | Review already exists for this order |

---

#### `GET /users/{id}/reviews` — List Reviews

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | Reviewee user ID |

**Query Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `limit` | int | No | Max results (1–100, default 20) |
| `cursor` | string | No | Pagination cursor |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "reviews": [
      {
        "id": "uuid",
        "reviewer_id": "uuid",
        "reviewee_id": "uuid",
        "order_id": "uuid",
        "rating": 5,
        "comment": "Great!",
        "role": "buyer",
        "is_anonymous": false,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ],
    "cursor": "eyJpZCI6Ii4uLiJ9"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |

---

#### `GET /users/{id}/reviews/summary` — Rating Summary

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "user_id": "uuid",
    "avg_rating": 4.75,
    "total_reviews": 120,
    "count_1": 2,
    "count_2": 3,
    "count_3": 5,
    "count_4": 30,
    "count_5": 80
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |

---

### Addresses

#### `POST /users/me/addresses` — Add Address 🔒

**Request Body**
```json
{
  "label": "Home",
  "recipient_name": "Nguyen Van A",
  "phone": "0901234567",
  "address_line_1": "123 Nguyen Hue",
  "address_line_2": "Floor 5",
  "ward": "Ben Nghe",
  "district": "District 1",
  "province": "Ho Chi Minh",
  "postal_code": "700000",
  "country": "Vietnam",
  "is_default": true
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `label` | string | Yes | max 50 |
| `recipient_name` | string | Yes | max 255 |
| `phone` | string | Yes | — |
| `address_line_1` | string | Yes | max 512 |
| `address_line_2` | string | No | max 512 |
| `ward` | string | No | max 128 |
| `district` | string | No | max 128 |
| `province` | string | No | max 128 |
| `postal_code` | string | No | max 20 |
| `country` | string | No | max 64, default `Vietnam` |
| `is_default` | bool | No | default false (first address is always default) |

**Success Response** `201`
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "label": "Home",
    "recipient_name": "Nguyen Van A",
    "phone": "0901234567",
    "address_line_1": "123 Nguyen Hue",
    "address_line_2": "Floor 5",
    "ward": "Ben Nghe",
    "district": "District 1",
    "province": "Ho Chi Minh",
    "postal_code": "700000",
    "country": "Vietnam",
    "is_default": true,
    "created_at": "2026-01-01T00:00:00Z"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed or address limit reached (max 10) |
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `GET /users/me/addresses` — List My Addresses 🔒

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "addresses": [
      {
        "id": "uuid",
        "label": "Home",
        "recipient_name": "Nguyen Van A",
        "phone": "0901234567",
        "address_line_1": "123 Nguyen Hue",
        "address_line_2": "",
        "ward": "Ben Nghe",
        "district": "District 1",
        "province": "Ho Chi Minh",
        "postal_code": "700000",
        "country": "Vietnam",
        "is_default": true,
        "created_at": "2026-01-01T00:00:00Z"
      }
    ]
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `PUT /users/me/addresses/{addr_id}` — Update Address 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `addr_id` | UUID | Address ID |

**Request Body** — same fields as `POST /users/me/addresses`, all optional.

**Success Response** `200` — updated address (same structure as POST)

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID or validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | Address not found |

---

#### `DELETE /users/me/addresses/{addr_id}` — Delete Address 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `addr_id` | UUID | Address ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "address deleted" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | Address not found |

---

#### `PUT /users/me/addresses/{addr_id}/default` — Set Default Address 🔒

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `addr_id` | UUID | Address ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "default address updated" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | Address not found |

---

### Seller

#### `GET /users/{id}/shop` — Get Shop Profile
No auth required.

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User (seller) ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "user_id": "uuid",
    "shop_name": "My Shop",
    "shop_slug": "my-shop",
    "shop_description": "Best products",
    "shop_logo_url": "https://...",
    "shop_banner_url": "https://...",
    "business_type": "individual",
    "kyc_status": "approved",
    "avg_rating": 4.8,
    "total_reviews": 250,
    "created_at": "2026-01-01T00:00:00Z"
  },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID |
| `NOT_FOUND` 404 | Seller not found |

---

#### `POST /users/me/seller/register` — Register as Seller 🔒

**Request Body**
```json
{
  "shop_name": "My Shop",
  "shop_slug": "my-shop",
  "shop_description": "Best products",
  "business_type": "individual",
  "tax_id": "123456789",
  "bank_account": "0123456789",
  "bank_name": "Vietcombank"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `shop_name` | string | Yes | min 3, max 255 |
| `shop_slug` | string | Yes | min 3, max 255, alphanumeric + hyphen |
| `shop_description` | string | No | max 1000 |
| `business_type` | string | Yes | `individual` \| `business` |
| `tax_id` | string | No | — |
| `bank_account` | string | No | — |
| `bank_name` | string | No | max 128 |

**Success Response** `201` — `SellerProfileResponse` (same structure as `GET /users/{id}/shop`)

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `CONFLICT` 409 | User is already a seller |

---

#### `PUT /users/me/seller/profile` — Update Seller Profile 🔒

**Request Body** — all fields optional:
```json
{
  "shop_name": "New Shop Name",
  "shop_description": "Updated description",
  "shop_logo_url": "https://...",
  "shop_banner_url": "https://...",
  "tax_id": "987654321",
  "bank_account": "9876543210",
  "bank_name": "Techcombank"
}
```

**Success Response** `200` — updated `SellerProfileResponse`

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | Seller profile not found |

---

### KYC

#### `POST /users/me/seller/kyc` — Submit KYC Documents 🔒

**Request Body**
```json
{
  "documents": [
    {
      "doc_type": "id_card_front",
      "doc_url": "https://cdn.example.com/doc1.jpg"
    },
    {
      "doc_type": "id_card_back",
      "doc_url": "https://cdn.example.com/doc2.jpg"
    }
  ]
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `documents` | array | Yes | min 1 item |
| `documents[].doc_type` | string | Yes | `id_card_front` \| `id_card_back` \| `business_license` \| `selfie_with_id` |
| `documents[].doc_url` | string | Yes | valid URL |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "KYC documents submitted" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Validation failed or invalid KYC state (already approved) |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `NOT_FOUND` 404 | Seller profile not found |

---

#### `GET /users/me/seller/kyc/status` — Get My KYC Status 🔒

**Success Response** `200`
```json
{
  "success": true,
  "data": {
    "status": "pending"
  },
  "meta": { "timestamp": 1711584000 }
}
```

KYC status values: `none` | `pending` | `approved` | `rejected`

**Errors**

| Code | Condition |
|------|-----------|
| `UNAUTHORIZED` 401 | Invalid or missing token |

---

#### `PUT /admin/users/{id}/kyc/approve` — Approve KYC 🔒🔑
Admin-only. Approves KYC and upgrades user role to `seller`.

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "KYC approved" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID or KYC not in pending state |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `FORBIDDEN` 403 | Not admin |
| `NOT_FOUND` 404 | Seller profile not found |

---

#### `PUT /admin/users/{id}/kyc/reject` — Reject KYC 🔒🔑
Admin-only.

**Path Parameters**

| Param | Type | Description |
|-------|------|-------------|
| `id` | UUID | User ID |

**Request Body**
```json
{
  "reason": "Document quality too low"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `reason` | string | Yes | min 1, max 500 |

**Success Response** `200`
```json
{
  "success": true,
  "data": { "message": "KYC rejected" },
  "meta": { "timestamp": 1711584000 }
}
```

**Errors**

| Code | Condition |
|------|-----------|
| `BAD_REQUEST` 400 | Invalid UUID or validation failed |
| `UNAUTHORIZED` 401 | Invalid or missing token |
| `FORBIDDEN` 403 | Not admin |
| `NOT_FOUND` 404 | Seller profile not found |

---

## gRPC API

**Proto package:** `modami.user`
**Service:** `UserInternalService`

> For internal service-to-service calls only. Not accessible from FE/mobile.

---

### `GetUserBasic`
Get lightweight user info by ID.

**Request**
```protobuf
message GetUserBasicRequest {
  string user_id = 1; // UUID
}
```

**Response**
```protobuf
message UserBasicResponse {
  string id          = 1;
  string full_name   = 2;
  string avatar_url  = 3;
  string role        = 4;
  string status      = 5;
  double trust_score = 6;
}
```

**Errors:** `INVALID_ARGUMENT` (invalid UUID), `NOT_FOUND`, `INTERNAL`

---

### `GetUsersByIDs`
Batch fetch users by a list of IDs.

**Request**
```protobuf
message GetUsersByIDsRequest {
  repeated string user_ids = 1; // UUIDs
}
```

**Response**
```protobuf
message UsersResponse {
  repeated UserBasicResponse users = 1;
}
```

> Invalid or not-found IDs are silently skipped.

---

### `CheckUserStatus`
Check if a user is active.

**Request**
```protobuf
message CheckUserStatusRequest {
  string user_id = 1; // UUID
}
```

**Response**
```protobuf
message UserStatusResponse {
  string status    = 1; // active | inactive | suspended | banned
  bool   is_active = 2;
}
```

**Errors:** `INVALID_ARGUMENT` (invalid UUID), `NOT_FOUND`, `INTERNAL`

---

### `GetSellerInfo`
Get seller profile info by user ID.

**Request**
```protobuf
message GetSellerInfoRequest {
  string user_id = 1; // UUID
}
```

**Response**
```protobuf
message SellerInfoResponse {
  string user_id       = 1;
  string shop_name     = 2;
  string shop_slug     = 3;
  string shop_logo_url = 4;
  string kyc_status    = 5;
  double avg_rating    = 6;
  int32  total_reviews = 7;
}
```

**Errors:** `INVALID_ARGUMENT` (invalid UUID), `NOT_FOUND`, `INTERNAL`

---

## Pagination (Cursor-based)

All list endpoints use cursor-based pagination.

- Pass `cursor` from the previous response to get the next page.
- Empty `cursor` in the response means no more pages.
- Default `limit` is `20`, max is `100`.

---

## Enum Reference

| Enum | Values |
|------|--------|
| `gender` | `male` \| `female` \| `other` \| `undisclosed` |
| `role` | `buyer` \| `seller` |
| `user_status` | `active` \| `inactive` \| `suspended` \| `banned` |
| `business_type` | `individual` \| `business` |
| `kyc_status` | `none` \| `pending` \| `approved` \| `rejected` |
| `kyc_doc_type` | `id_card_front` \| `id_card_back` \| `business_license` \| `selfie_with_id` |
| `review_role` | `buyer` \| `seller` |
