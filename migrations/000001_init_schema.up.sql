-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enums
CREATE TYPE user_role AS ENUM ('buyer', 'seller');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'banned');
CREATE TYPE gender_type AS ENUM ('male', 'female', 'other', 'undisclosed');
CREATE TYPE kyc_status AS ENUM ('none', 'pending', 'approved', 'rejected');
CREATE TYPE kyc_doc_type AS ENUM ('id_card_front', 'id_card_back', 'business_license', 'selfie_with_id');
CREATE TYPE kyc_doc_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE business_type AS ENUM ('individual', 'business');
CREATE TYPE review_role AS ENUM ('buyer', 'seller');

-- Users table
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    keycloak_id     VARCHAR(255) UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    full_name       VARCHAR(255) NOT NULL DEFAULT '',
    phone           VARCHAR(50) DEFAULT '',
    avatar_url      TEXT DEFAULT '',
    cover_url       TEXT DEFAULT '',
    bio             TEXT DEFAULT '',
    gender          gender_type NOT NULL DEFAULT 'undisclosed',
    date_of_birth   DATE,
    role            user_role NOT NULL DEFAULT 'buyer',
    status          user_status NOT NULL DEFAULT 'active',
    email_verified  BOOLEAN NOT NULL DEFAULT false,
    trust_score     DECIMAL(5,2) NOT NULL DEFAULT 0,
    follower_count  INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_users_keycloak_id ON users(keycloak_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_full_name ON users USING gin(full_name gin_trgm_ops);

-- Seller profiles
CREATE TABLE seller_profiles (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_name        VARCHAR(255) NOT NULL,
    shop_slug        VARCHAR(255) UNIQUE NOT NULL,
    shop_description TEXT DEFAULT '',
    shop_logo_url    TEXT DEFAULT '',
    shop_banner_url  TEXT DEFAULT '',
    business_type    business_type NOT NULL DEFAULT 'individual',
    tax_id           VARCHAR(100) DEFAULT '',
    bank_account     VARCHAR(100) DEFAULT '',
    bank_name        VARCHAR(128) DEFAULT '',
    kyc_status       kyc_status NOT NULL DEFAULT 'none',
    kyc_verified_at  TIMESTAMPTZ,
    avg_rating       DECIMAL(3,2) NOT NULL DEFAULT 0,
    total_reviews    INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_seller_profiles_user_id ON seller_profiles(user_id);
CREATE INDEX idx_seller_profiles_shop_slug ON seller_profiles(shop_slug);
CREATE INDEX idx_seller_profiles_kyc_status ON seller_profiles(kyc_status);

-- Addresses
CREATE TABLE addresses (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label          VARCHAR(50) NOT NULL DEFAULT '',
    recipient_name VARCHAR(255) NOT NULL DEFAULT '',
    phone          VARCHAR(50) NOT NULL DEFAULT '',
    address_line_1 VARCHAR(512) NOT NULL DEFAULT '',
    address_line_2 VARCHAR(512) DEFAULT '',
    ward           VARCHAR(128) DEFAULT '',
    district       VARCHAR(128) DEFAULT '',
    province       VARCHAR(128) DEFAULT '',
    postal_code    VARCHAR(20) DEFAULT '',
    country        VARCHAR(64) DEFAULT 'Vietnam',
    is_default     BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE INDEX idx_addresses_is_default ON addresses(user_id, is_default);

-- User follows
CREATE TABLE user_follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);

CREATE INDEX idx_user_follows_follower_id ON user_follows(follower_id, created_at DESC);
CREATE INDEX idx_user_follows_following_id ON user_follows(following_id, created_at DESC);

-- Reviews
CREATE TABLE reviews (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reviewer_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reviewee_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id     UUID NOT NULL,
    rating       SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment      TEXT DEFAULT '',
    role         review_role NOT NULL DEFAULT 'buyer',
    is_anonymous BOOLEAN NOT NULL DEFAULT false,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id)
);

CREATE INDEX idx_reviews_reviewee_id ON reviews(reviewee_id, created_at DESC);
CREATE INDEX idx_reviews_reviewer_id ON reviews(reviewer_id);
CREATE INDEX idx_reviews_order_id ON reviews(order_id);

-- Rating summaries
CREATE TABLE rating_summaries (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    avg_rating    DECIMAL(3,2) NOT NULL DEFAULT 0,
    total_reviews INTEGER NOT NULL DEFAULT 0,
    count_1       INTEGER NOT NULL DEFAULT 0,
    count_2       INTEGER NOT NULL DEFAULT 0,
    count_3       INTEGER NOT NULL DEFAULT 0,
    count_4       INTEGER NOT NULL DEFAULT 0,
    count_5       INTEGER NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- KYC documents
CREATE TABLE kyc_documents (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doc_type    kyc_doc_type NOT NULL,
    doc_url     TEXT NOT NULL,
    status      kyc_doc_status NOT NULL DEFAULT 'pending',
    reason      TEXT DEFAULT '',
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_documents_user_id ON kyc_documents(user_id);
CREATE INDEX idx_kyc_documents_status ON kyc_documents(status);

-- Outbox events for transactional outbox pattern
CREATE TABLE outbox_events (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    topic      VARCHAR(255) NOT NULL,
    key        VARCHAR(255) NOT NULL,
    payload    JSONB NOT NULL,
    status     VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at    TIMESTAMPTZ
);

CREATE INDEX idx_outbox_events_status ON outbox_events(status, created_at);

-- Processed events for idempotency
CREATE TABLE processed_events (
    event_id     VARCHAR(255) PRIMARY KEY,
    topic        VARCHAR(255) NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processed_events_processed_at ON processed_events(processed_at);

-- Enable trigram extension for fuzzy search (needed for gin_trgm_ops)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
