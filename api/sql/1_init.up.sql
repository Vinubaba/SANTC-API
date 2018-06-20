CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS daycares (
  daycare_id varchar PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  name varchar NOT NULL,
  address_1 varchar NOT NULL,
  address_2 varchar,
  city varchar NOT NULL,
  state varchar NOT NULL,
  zip varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS age_ranges (
  age_range_id varchar PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  daycare_id varchar REFERENCES daycares (daycare_id) NOT NULL,
  stage varchar NOT NULL,
  min integer,
  min_unit varchar,
  max integer,
  max_unit varchar
);

CREATE TABLE IF NOT EXISTS classes (
  class_id varchar PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  daycare_id varchar REFERENCES daycares (daycare_id) ON DELETE SET NULL,
  age_range_id varchar REFERENCES age_ranges (age_range_id) ON DELETE SET NULL,
  name varchar UNIQUE NOT NULL,
  description varchar,
  image_uri varchar
);

CREATE TABLE IF NOT EXISTS users (
  user_id varchar PRIMARY KEY UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
  email varchar UNIQUE NOT NULL,
  first_name varchar,
  last_name varchar,
  phone varchar,
  address_1 varchar,
  address_2 varchar,
  city varchar,
  state varchar,
  zip varchar,
  gender varchar,
  image_uri varchar,
  daycare_id varchar REFERENCES daycares (daycare_id) ON DELETE SET NULL,
  work_address_1 varchar,
  work_address_2 varchar,
  work_city varchar,
  work_state varchar,
  work_zip varchar,
  work_phone varchar
);

CREATE TABLE IF NOT EXISTS children (
  child_id varchar UNIQUE NOT NULL PRIMARY KEY,
  daycare_id varchar REFERENCES daycares (daycare_id) NOT NULL,
  class_id varchar REFERENCES classes (class_id) ON DELETE SET NULL,
  first_name varchar NOT NULL,
  last_name varchar NOT NULL,
  gender varchar NOT NULL,
  birth_date date NOT NULL,
  image_uri varchar,
  start_date date NOT NULL,
  notes varchar
);

CREATE TABLE IF NOT EXISTS child_photos (
  photo_id varchar UNIQUE NOT NULL PRIMARY KEY,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  published_by varchar REFERENCES users (user_id) ON DELETE SET NULL,
  approved_by varchar REFERENCES users (user_id) ON DELETE SET NULL,
  image_uri varchar NOT NULL,
  approved boolean NOT NULL default false,
  publication_date date NOT NULL
);

CREATE TABLE IF NOT EXISTS allergies (
  allergy_id varchar UNIQUE NOT NULL PRIMARY KEY,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  allergy varchar NOT NULL,
  instruction varchar
);

CREATE TABLE IF NOT EXISTS special_instructions (
  special_instruction_id varchar UNIQUE NOT NULL PRIMARY KEY,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  instruction varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS responsible_of (
  responsible_id varchar REFERENCES users (user_id) ON DELETE CASCADE,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  relationship varchar  NOT NULL --mother, father, grandmother, grandfather, guardian
);

CREATE TABLE IF NOT EXISTS roles (
  user_id varchar REFERENCES users (user_id) ON DELETE CASCADE,
  role varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS teacher_classes (
  teacher_id varchar REFERENCES users (user_id) ON DELETE CASCADE,
  class_id varchar REFERENCES classes (class_id) ON DELETE CASCADE
);

INSERT INTO daycares ("daycare_id", "name", "address_1", "address_2", "city", "state", "zip") VALUES ('PUBLIC', 'PUBLIC', 'PUBLIC', 'PUBLIC', 'PUBLIC', 'PUBLIC', 'PUBLIC');

INSERT INTO users VALUES ('dd3a81f0-6432-4ddf-842a-b82a3911dadb', 'arthur.gustin@gmail.com', 'arthur', 'gustin', NULL, NULL, NULL, NULL, NULL, NULL, 'm', NULL, 'PUBLIC');
INSERT INTO users VALUES ('7ab5b1ca-8dfa-4695-a05f-3fc860900449', 'vinu.singh@gmail.com', 'vinu', 'singh', NULL, NULL, NULL, NULL, NULL, NULL, 'm', NULL, 'PUBLIC');
INSERT INTO users VALUES ('0f7892f6-dbd6-4c83-9bde-a35d0fd4b260', 'muneerkk66@gmail.com', 'muneer', 'kk', NULL, NULL, NULL, NULL, NULL, NULL, 'm', NULL, 'PUBLIC');
INSERT INTO users VALUES ('0f7892f6-dbd6-4c83-9bde-a35d0fd4b260', 'muneerkk66@gmail.com', 'muneer', 'kk', NULL, NULL, NULL, NULL, NULL, NULL, 'm', NULL, 'PUBLIC');
INSERT INTO users VALUES ('0f7892f6-dbd6-4c83-9bde-a35d0fd4b261', 'muneerkk7066@gmail.com', 'muneer', 'kk', NULL, NULL, NULL, NULL, NULL, NULL, 'm', NULL, 'PUBLIC');

INSERT INTO roles VALUES ('dd3a81f0-6432-4ddf-842a-b82a3911dadb', 'admin');
INSERT INTO roles VALUES ('7ab5b1ca-8dfa-4695-a05f-3fc860900449', 'officemanager');
INSERT INTO roles VALUES ('0f7892f6-dbd6-4c83-9bde-a35d0fd4b260', 'officemanager');
INSERT INTO roles VALUES ('0f7892f6-dbd6-4c83-9bde-a35d0fd4b261', 'officemanager');


-- E_CHECK_IN, E_CHECK_OUT, E_BREAKFAST, E_SNACK, E_LUNCH, E_DINNER, E_MOOD_HAPPY, E_MOOD_SAD, E_HEALTH, E_CHANGE_CLOTHES, E_NAP

--CREATE TABLE IF NOT EXISTS event_types (
--	event_id varchar NOT NULL PRIMARY KEY,
--	name varchar NOT NULL
--);

--CREATE TABLE IF NOT EXISTS reporting (
--  reporting_id varchar NOT NULL PRIMARY KEY,
--  child_id varchar REFERENCES children (child_id),
--  reported_by varchar REFERENCES users (user_id),
--  picture_uri varchar(256),
--  event_type int REFERENCES event_types (event_id),
--  log_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
--  note varchar
--);