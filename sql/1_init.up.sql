CREATE TABLE IF NOT EXISTS users (
  user_id varchar PRIMARY KEY UNIQUE NOT NULL,
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
  image_uri varchar
);

CREATE TABLE IF NOT EXISTS children (
  child_id varchar UNIQUE NOT NULL PRIMARY KEY,
  first_name varchar NOT NULL,
  last_name varchar NOT NULL,
  gender varchar NOT NULL,
  birth_date date NOT NULL,
  image_uri varchar
);

CREATE TABLE IF NOT EXISTS allergies (
  allergy_id varchar UNIQUE NOT NULL PRIMARY KEY,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  allergy varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS responsible_of (
  responsible_id varchar references users (user_id) ON DELETE CASCADE,
  child_id varchar REFERENCES children (child_id) ON DELETE CASCADE,
  relationship varchar  NOT NULL --mother, father, grandmother, grandfather, guardian
);

CREATE TABLE IF NOT EXISTS roles (
  user_id varchar REFERENCES users (user_id) ON DELETE CASCADE,
  role varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS pending_connexion_roles (
  email varchar NOT NULL,
  role varchar NOT NULL
);

INSERT INTO pending_connexion_roles ("email", "role") VALUES ('arthur.gustin@gmail.com', 'admin');
INSERT INTO pending_connexion_roles ("email", "role") VALUES ('vinu.singh@gmail.com', 'officemanager');
INSERT INTO pending_connexion_roles ("email", "role") VALUES ('johngallegodev@gmail.com', 'officemanager');

--CREATE TABLE IF NOT EXISTS teachers (
--  teacher_id varchar REFERENCES users (user_id) UNIQUE,
--  first_name varchar,
--  last_name varchar
--);

--CREATE TABLE IF NOT EXISTS teacher_teach (
--teacher_id varchar REFERENCES teachers (teacher_id) UNIQUE,
--  class_id varchar REFERENCES daycare_class (class_id)
--)

--CREATE TABLE roles (
--   role_id     int NOT NULL PRIMARY KEY,
--    role       varchar(40) NOT NULL
--);

--CREATE TABLE permissions
--(
--  permission_id int NOT NULL PRIMARY KEY,
--  user_id varchar REFERENCES users (userid),
--  role_id int REFERENCES roles (roleid)
--);

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