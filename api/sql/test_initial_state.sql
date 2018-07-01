TRUNCATE TABLE "users" CASCADE;
TRUNCATE TABLE "roles" CASCADE;
TRUNCATE TABLE "classes" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;
TRUNCATE TABLE "responsible_of" CASCADE;
TRUNCATE TABLE "children" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;
TRUNCATE TABLE "daycares" CASCADE;
TRUNCATE TABLE "teacher_classes" CASCADE;
TRUNCATE TABLE "schedules" CASCADE;

INSERT INTO daycares ("daycare_id", "name", "address_1", "address_2", "city", "state", "zip") VALUES ('peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon', 'peyredragon');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id1','elaria.sand@got.com','Elaria','Sand','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id2','john.snow@got.com','John','Snow','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id3','tyrion.lannister@got.com','Tyrion','Lannister','M','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id4','caitlyn.stark@got.com','Caitlyn','Stark','F','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id5','sansa.stark@got.com','Sansa','Stark','F','+3365651','address','floor','Peyredragon','WESTEROS','31400','http://image.com','peyredragon','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "roles" ("user_id","role") VALUES ('id1', 'admin');
INSERT INTO "roles" ("user_id","role") VALUES ('id2', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id3', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id4', 'teacher');
INSERT INTO "roles" ("user_id","role") VALUES ('id5', 'adult');

INSERT INTO daycares ("daycare_id", "name", "address_1", "address_2", "city", "state", "zip") VALUES ('namek', 'namek', 'namek', 'namek', 'namek', 'namek', 'namek');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id6','sangoku@dbz.com','San','Goku','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id7','vegeta@dbz.com','Vegeta','Vegeta','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id8','sangohan@dbz.com','San','Gohan','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id9','pan@dbz.com','Pan','Pan','F','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri","daycare_id","work_address_1","work_address_2","work_city","work_state","work_zip","work_phone") VALUES ('id10','krilin@dbz.com','Krilin','Krilin','M','+1234567','foo','bar','baz','NAMEK','31400','http://image.com','namek','work_address_1','work_address_2','work_city','work_state','work_zip','work_phone');
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
INSERT INTO "teacher_classes" ("teacher_id","class_id") VALUES ('id4', 'classid-2');
INSERT INTO "teacher_classes" ("teacher_id","class_id") VALUES ('id9', 'classid-1');

INSERT INTO "children" ("address_same_as","child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id","class_id") VALUES ('id6','childid-1','Goten','Goten','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','namek','classid-1');
INSERT INTO "children" ("address_same_as","child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id","class_id") VALUES ('id7','childid-2','Trunk','Trunk','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','namek','classid-1');
INSERT INTO "allergies" ("allergy_id","child_id","allergy","instruction") VALUES ('allergyid-1','childid-1','tomato','call the doctor');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id6','childid-1','father');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id7','childid-2','father');
INSERT INTO "special_instructions" ("special_instruction_id","child_id","instruction") VALUES ('specialinstruction-1','childid-1','this boy always sleeps please keep him awaken');

INSERT INTO "children" ("address_same_as", "child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id","class_id") VALUES ('id4','childid-3','Arya','Stark','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','peyredragon','classid-2');
INSERT INTO "children" ("address_same_as", "child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes","daycare_id","class_id") VALUES ('id3','childid-4','Joffrey','Baratheon','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes','peyredragon','classid-2');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id4','childid-3','mother');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id3','childid-4','father');

INSERT INTO "schedules" ("schedule_id","walk_in","monday_start","monday_end","tuesday_start","tuesday_end","wednesday_start","wednesday_end","thursday_start","thursday_end","friday_start","friday_end") VALUES ('scheduleid-1','false', '8:30 AM', '6:00 PM', '8:30 AM', '6:00 PM', '8:30 AM', '6:00 PM', '8:30 AM', '6:00 PM', '8:30 AM', '6:00 PM');
UPDATE users SET schedule_id = 'scheduleid-1' WHERE user_id = 'id9';
UPDATE children SET schedule_id = 'scheduleid-1' WHERE child_id = 'childid-1';

INSERT INTO "child_photos" ("photo_id","child_id","published_by","approved_by","image_uri","approved","publication_date") VALUES ('photoid-1','id6', 'id9',NULL,'foo/bar.jpg','false','1992-10-13T15:13:00Z');
