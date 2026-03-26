DROP TABLE IF EXISTS processed_events;
DROP TABLE IF EXISTS outbox_events;
DROP TABLE IF EXISTS kyc_documents;
DROP TABLE IF EXISTS rating_summaries;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS user_follows;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS seller_profiles;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS review_role;
DROP TYPE IF EXISTS business_type;
DROP TYPE IF EXISTS kyc_doc_status;
DROP TYPE IF EXISTS kyc_doc_type;
DROP TYPE IF EXISTS kyc_status;
DROP TYPE IF EXISTS gender_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
