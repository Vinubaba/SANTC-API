TRUNCATE TABLE "users" CASCADE;
TRUNCATE TABLE "roles" CASCADE;
TRUNCATE TABLE "classes" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;
TRUNCATE TABLE "responsible_of" CASCADE;
TRUNCATE TABLE "children" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;

INSERT INTO daycares ("daycare_id", "name", "address_1", "address_2", "city", "state", "zip") VALUES ('peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id1','elaria.sand@got.com','Elaria','Sand','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id2','john.snow@got.com','John','Snow','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id3','tyrion.lannister@got.com','Tyrion','Lannister','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id4','caitlyn.stark@got.com','Caitlyn','Stark','F','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id5','sansa.stark@got.com','Sansa','Stark','F','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon');
INSERT INTO "roles" ("user_id","role") VALUES ('id1', 'admin');
INSERT INTO "roles" ("user_id","role") VALUES ('id2', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id3', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id4', 'teacher');
INSERT INTO "roles" ("user_id","role") VALUES ('id5', 'adult');

INSERT INTO daycares ("daycare_id", "name", "address_1", "address_2", "city", "state", "zip") VALUES ('namek', 'namek', 'namek', 'namek', 'namek', 'namek', 'namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id6','sangoku@dbz.com','San','Goku','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id7','vegeta@dbz.com','Vegeta','Vegeta','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id8','sangohan@dbz.com','San','Gohan','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id9','pan@dbz.com','Pan','Pan','F','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id") VALUES ('id10','krilin@dbz.com','Krilin','Krilin','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek');
INSERT INTO "roles" ("user_id","role") VALUES ('id6', 'admin');
INSERT INTO "roles" ("user_id","role") VALUES ('id7', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id8', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id9', 'teacher');
INSERT INTO "roles" ("user_id","role") VALUES ('id10', 'adult');

INSERT INTO "age_ranges" ("age_range_id","daycare_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-1','namek','infant','3','M','12','M');
INSERT INTO "age_ranges" ("age_range_id","daycare_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-2','peyredragon','toddlers','12','M','18','M');
INSERT INTO "age_ranges" ("age_range_id","daycare_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-3','namek','infant 2','4','M','13','M');
INSERT INTO "classes" ("class_id","daycare_id","age_range_id","name","description","image_uri") VALUES ('classid-1', 'namek', 'agerangeid-1', 'infant class','infant description','gs://foo/bar.jpg');
INSERT INTO "classes" ("class_id","daycare_id","age_range_id","name","description","image_uri") VALUES ('classid-2', 'peyredragon', 'agerangeid-2', 'toddlers class','toddlers description','gs://foo/bar.jpg');

INSERT INTO "children" ("child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id") VALUES ('childid-1','Goten','Goten','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','namek');
INSERT INTO "children" ("child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id") VALUES ('childid-2','Trunk','Trunk','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','namek');
INSERT INTO "allergies" ("allergy_id","child_id","allergy") VALUES ('allergyid-1','childid-1','tomato');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id6','childid-1','father');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id7','childid-2','father');
INSERT INTO "special_instructions" ("child_id","instruction") VALUES ('childid-1','this boy always sleeps please keep him awaken');

INSERT INTO "children" ("child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id") VALUES ('childid-3','Arya','Stark','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','peyredragon');
INSERT INTO "children" ("child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id") VALUES ('childid-4','Joffrey','Baratheon','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','peyredragon');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id4','childid-3','mother');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id3','childid-4','father');