-- TODO: string id for public ressources, int id for internal ressource
-- TODO: set default timezone to UTC

CREATE TABLE IF NOT EXISTS users (
  user_id varchar(64) NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS daycare (
  daycare_id varchar(64) NOT NULL PRIMARY KEY,
  owner_id varchar(64) REFERENCES users (user_id) UNIQUE,
  zipcode int NOT NULL,
  phone varchar(20)
);

CREATE TABLE IF NOT EXISTS daycare_class (
  class_id varchar(64) NOT NULL PRIMARY KEY,
  daycare_id varchar(64) REFERENCES daycare (daycare_id)
);

CREATE TABLE IF NOT EXISTS children (
  child_id varchar(64) REFERENCES users (user_id) UNIQUE,
  first_name varchar(64),
  last_name varchar(64),
  gender varchar(1) NOT NULL,
  birth_date date
);

CREATE TABLE IF NOT EXISTS adult_responsibles (
  responsible_id varchar(64) REFERENCES users (user_id) UNIQUE,
  first_name varchar(64),
  last_name varchar(64),
  gender varchar(1) NOT NULL
);

CREATE TABLE IF NOT EXISTS responsible_of (
  responsible_id varchar(64) references adult_responsibles (responsible_id),
  child_id varchar(64) REFERENCES children (child_id),
  relationship varchar (20) NOT NULL --mother, father, grandmother, grandfather, guardian
);

CREATE TABLE IF NOT EXISTS teachers (
  teacher_id varchar(64) REFERENCES users (user_id) UNIQUE,
  first_name varchar(64),
  last_name varchar(64)
);

CREATE TABLE IF NOT EXISTS teacher_teach (
  teacher_id varchar(64) REFERENCES teachers (teacher_id) UNIQUE,
  class_id varchar(64) REFERENCES daycare_class (class_id)
)

--CREATE TABLE roles (
--   role_id     int NOT NULL PRIMARY KEY,
--    role       varchar(40) NOT NULL
--);

--CREATE TABLE permissions
--(
--  permission_id int NOT NULL PRIMARY KEY,
--  user_id varchar(64) REFERENCES users (userid),
--  role_id int REFERENCES roles (roleid)
--);

-- E_CHECK_IN, E_CHECK_OUT, E_BREAKFAST, E_SNACK, E_LUNCH, E_DINNER, E_MOOD_HAPPY, E_MOOD_SAD, E_HEALTH, E_CHANGE_CLOTHES, E_NAP

--CREATE TABLE IF NOT EXISTS event_types (
--	event_id varchar(64) NOT NULL PRIMARY KEY,
--	name varchar(64) NOT NULL
--);

--CREATE TABLE IF NOT EXISTS reporting (
--  reporting_id varchar(64) NOT NULL PRIMARY KEY,
--  child_id varchar(64) REFERENCES children (child_id),
--  reported_by varchar(64) REFERENCES users (user_id),
--  picture_uri varchar(256),
--  event_type int REFERENCES event_types (event_id),
--  log_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
--  note varchar
--);