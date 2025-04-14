-- Script to initialize the PostgreSQL database for the SAST tool

-- Create the database if it doesn't exist
CREATE DATABASE sast;

-- Connect to the database
\c sast

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Initialize schema
-- (The actual schema creation will be handled by migrations)

-- Create role if needed
-- Uncomment and modify if needed
-- DO
-- $do$
-- BEGIN
--    IF NOT EXISTS (
--       SELECT FROM pg_catalog.pg_roles
--       WHERE  rolname = 'sast_user') THEN
--       
--       CREATE ROLE sast_user LOGIN PASSWORD 'sast_password';
--    END IF;
-- END
-- $do$; 