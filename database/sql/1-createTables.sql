BEGIN;

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	name TEXT,
	email TEXT NOT NULL UNIQUE,
	passwordHash TEXT NOT NULL,
	sessionToken TEXT NOT NULL,
	csrfToken TEXT NOT NULL,
	isAdmin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE uploaded_documents(
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	file_name TEXT NOT NULL,
	object_name TEXT NOT NULL,
	bucket_name TEXT NOT NULL,
	uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chunks (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	object_name TEXT NOT NULL,
	page_number INTEGER NOT NULL,
	bucket_name TEXT NOT NULL,
	chunk TEXT NOT NULL,
	embedding vector(768) NOT NULL
);

CREATE INDEX ON chunks USING hnsw (embedding vector_cosine_ops);

COMMIT;
